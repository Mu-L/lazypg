package components

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// NodeType represents the type of a JSON node
type NodeType int

const (
	NodeObject NodeType = iota
	NodeArray
	NodeString
	NodeNumber
	NodeBoolean
	NodeNull
)

// TreeNode represents a node in the JSON tree
type TreeNode struct {
	Key        string      // Key name (for object properties)
	Value      interface{} // Raw value
	Type       NodeType    // Type of this node
	IsExpanded bool        // Whether this node is expanded (for objects/arrays)
	Children   []*TreeNode // Child nodes
	Parent     *TreeNode   // Parent node
	Path       []string    // Full path from root to this node
	Level      int         // Indentation level (depth in tree)
}

// CloseJSONBViewerMsg is sent when viewer should close
type CloseJSONBViewerMsg struct{}

// JSONBViewer displays JSONB data as an interactive collapsible tree
type JSONBViewer struct {
	Width  int
	Height int
	Theme  theme.Theme

	// Tree structure
	root *TreeNode

	// Flattened list of visible nodes (for rendering and navigation)
	visibleNodes []*TreeNode

	// Navigation state
	selectedIndex int // Index in visibleNodes
	scrollOffset  int // Scroll offset for viewport

	// Search state
	searchMode   bool
	searchQuery  string
	searchResult []*TreeNode // Nodes matching search
}

// NewJSONBViewer creates a new tree-based JSONB viewer
func NewJSONBViewer(th theme.Theme) *JSONBViewer {
	return &JSONBViewer{
		Width:         80,
		Height:        30,
		Theme:         th,
		selectedIndex: 0,
		scrollOffset:  0,
	}
}

// SetValue parses JSON and builds the tree structure
func (jv *JSONBViewer) SetValue(value interface{}) error {
	// Parse JSON if it's a string
	var parsed interface{}
	switch v := value.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
	case []byte:
		if err := json.Unmarshal(v, &parsed); err != nil {
			return fmt.Errorf("invalid JSON: %w", err)
		}
	default:
		parsed = v
	}

	// Build tree
	jv.root = jv.buildTree("root", parsed, nil, []string{}, 0)
	jv.root.IsExpanded = true // Root is always expanded

	// Flatten tree to get visible nodes
	jv.rebuildVisibleNodes()

	// Reset navigation
	jv.selectedIndex = 0
	jv.scrollOffset = 0

	return nil
}

// buildTree recursively builds the tree structure from JSON
func (jv *JSONBViewer) buildTree(key string, value interface{}, parent *TreeNode, path []string, level int) *TreeNode {
	node := &TreeNode{
		Key:        key,
		Value:      value,
		Parent:     parent,
		Path:       path,
		Level:      level,
		IsExpanded: false, // Collapsed by default
	}

	// Determine type and build children
	if value == nil {
		node.Type = NodeNull
		return node
	}

	switch v := value.(type) {
	case map[string]interface{}:
		node.Type = NodeObject
		node.Children = make([]*TreeNode, 0, len(v))
		for childKey, childValue := range v {
			childPath := append([]string{}, path...)
			childPath = append(childPath, childKey)
			childNode := jv.buildTree(childKey, childValue, node, childPath, level+1)
			node.Children = append(node.Children, childNode)
		}

	case []interface{}:
		node.Type = NodeArray
		node.Children = make([]*TreeNode, 0, len(v))
		for i, childValue := range v {
			childKey := fmt.Sprintf("[%d]", i)
			childPath := append([]string{}, path...)
			childPath = append(childPath, fmt.Sprintf("%d", i))
			childNode := jv.buildTree(childKey, childValue, node, childPath, level+1)
			node.Children = append(node.Children, childNode)
		}

	case string:
		node.Type = NodeString

	case float64:
		node.Type = NodeNumber

	case bool:
		node.Type = NodeBoolean

	default:
		node.Type = NodeNull
	}

	return node
}

// rebuildVisibleNodes flattens the tree into a list of visible nodes (respecting collapse state)
func (jv *JSONBViewer) rebuildVisibleNodes() {
	jv.visibleNodes = []*TreeNode{}
	if jv.root != nil {
		jv.flattenTree(jv.root)
	}
}

// flattenTree recursively flattens the tree into visibleNodes
func (jv *JSONBViewer) flattenTree(node *TreeNode) {
	jv.visibleNodes = append(jv.visibleNodes, node)

	// Only recurse into children if node is expanded
	if node.IsExpanded && len(node.Children) > 0 {
		for _, child := range node.Children {
			jv.flattenTree(child)
		}
	}
}

// Update handles keyboard input
func (jv *JSONBViewer) Update(msg tea.KeyMsg) (*JSONBViewer, tea.Cmd) {
	// Handle search mode
	if jv.searchMode {
		switch msg.String() {
		case "esc":
			jv.searchMode = false
			jv.searchQuery = ""
			jv.searchResult = nil
			return jv, nil
		case "enter":
			jv.searchMode = false
			return jv, nil
		case "backspace":
			if len(jv.searchQuery) > 0 {
				jv.searchQuery = jv.searchQuery[:len(jv.searchQuery)-1]
				jv.performSearch()
			}
			return jv, nil
		default:
			// Append character to search query
			if len(msg.String()) == 1 {
				jv.searchQuery += msg.String()
				jv.performSearch()
			}
			return jv, nil
		}
	}

	// Normal navigation mode
	switch msg.String() {
	case "esc", "q":
		return jv, func() tea.Msg {
			return CloseJSONBViewerMsg{}
		}

	case "up", "k":
		if jv.selectedIndex > 0 {
			jv.selectedIndex--
			jv.adjustScroll()
		}

	case "down", "j":
		if jv.selectedIndex < len(jv.visibleNodes)-1 {
			jv.selectedIndex++
			jv.adjustScroll()
		}

	case " ", "enter":
		// Toggle expand/collapse
		if jv.selectedIndex < len(jv.visibleNodes) {
			node := jv.visibleNodes[jv.selectedIndex]
			if node.Type == NodeObject || node.Type == NodeArray {
				node.IsExpanded = !node.IsExpanded
				jv.rebuildVisibleNodes()
			}
		}

	case "E":
		// Expand all
		jv.expandAll(jv.root)
		jv.rebuildVisibleNodes()

	case "C":
		// Collapse all
		jv.collapseAll(jv.root)
		jv.rebuildVisibleNodes()

	case "/":
		// Enter search mode
		jv.searchMode = true
		jv.searchQuery = ""
		jv.searchResult = nil
	}

	return jv, nil
}

// adjustScroll adjusts scroll offset to keep selected node visible
func (jv *JSONBViewer) adjustScroll() {
	contentHeight := jv.Height - 5 // Account for header and footer
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Scroll up if selected is above viewport
	if jv.selectedIndex < jv.scrollOffset {
		jv.scrollOffset = jv.selectedIndex
	}

	// Scroll down if selected is below viewport
	if jv.selectedIndex >= jv.scrollOffset+contentHeight {
		jv.scrollOffset = jv.selectedIndex - contentHeight + 1
	}
}

// expandAll recursively expands all nodes
func (jv *JSONBViewer) expandAll(node *TreeNode) {
	if node.Type == NodeObject || node.Type == NodeArray {
		node.IsExpanded = true
		for _, child := range node.Children {
			jv.expandAll(child)
		}
	}
}

// collapseAll recursively collapses all nodes
func (jv *JSONBViewer) collapseAll(node *TreeNode) {
	if node.Type == NodeObject || node.Type == NodeArray {
		node.IsExpanded = false
		for _, child := range node.Children {
			jv.collapseAll(child)
		}
	}
}

// performSearch searches for nodes matching the query
func (jv *JSONBViewer) performSearch() {
	jv.searchResult = []*TreeNode{}
	if jv.searchQuery == "" {
		return
	}

	query := strings.ToLower(jv.searchQuery)
	for _, node := range jv.visibleNodes {
		// Search in key name
		if strings.Contains(strings.ToLower(node.Key), query) {
			jv.searchResult = append(jv.searchResult, node)
			continue
		}

		// Search in value (for primitives)
		if node.Type == NodeString || node.Type == NodeNumber || node.Type == NodeBoolean {
			valueStr := fmt.Sprintf("%v", node.Value)
			if strings.Contains(strings.ToLower(valueStr), query) {
				jv.searchResult = append(jv.searchResult, node)
			}
		}
	}
}

// View renders the JSONB viewer
func (jv *JSONBViewer) View() string {
	var sections []string

	// Title bar
	titleStyle := lipgloss.NewStyle().
		Foreground(jv.Theme.Background).
		Background(jv.Theme.Info).
		Padding(0, 1).
		Bold(true)

	title := " JSONB Tree Viewer"
	sections = append(sections, titleStyle.Render(title))

	// Instructions or search bar
	instrStyle := lipgloss.NewStyle().
		Foreground(jv.Theme.Metadata).
		Padding(0, 1)

	if jv.searchMode {
		searchBar := fmt.Sprintf("Search: %s_", jv.searchQuery)
		if len(jv.searchResult) > 0 {
			searchBar += fmt.Sprintf("  (%d matches)", len(jv.searchResult))
		}
		sections = append(sections, instrStyle.Render(searchBar))
	} else {
		instr := "↑↓: Navigate  Space/Enter: Expand/Collapse  E: Expand All  C: Collapse All  /: Search  Esc: Close"
		sections = append(sections, instrStyle.Render(instr))
	}

	// Content (tree view)
	contentHeight := jv.Height - 5
	if contentHeight < 1 {
		contentHeight = 1
	}

	content := jv.renderTree(contentHeight)
	sections = append(sections, content)

	// Status bar
	statusBar := jv.renderStatus()
	sections = append(sections, statusBar)

	// Container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(jv.Theme.Border).
		Width(jv.Width).
		Padding(1)

	return lipgloss.Place(
		jv.Width,
		jv.Height,
		lipgloss.Center,
		lipgloss.Center,
		containerStyle.Render(strings.Join(sections, "\n")),
	)
}

// renderTree renders the visible portion of the tree
func (jv *JSONBViewer) renderTree(height int) string {
	if len(jv.visibleNodes) == 0 {
		return lipgloss.NewStyle().
			Foreground(jv.Theme.Metadata).
			Italic(true).
			Render("No data")
	}

	var lines []string

	endIndex := jv.scrollOffset + height
	if endIndex > len(jv.visibleNodes) {
		endIndex = len(jv.visibleNodes)
	}

	for i := jv.scrollOffset; i < endIndex; i++ {
		node := jv.visibleNodes[i]
		line := jv.renderNode(node, i == jv.selectedIndex)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderNode renders a single tree node with proper indentation and styling
func (jv *JSONBViewer) renderNode(node *TreeNode, isSelected bool) string {
	// Indentation
	indent := strings.Repeat("  ", node.Level)

	// Expand/collapse indicator
	var indicator string
	if node.Type == NodeObject || node.Type == NodeArray {
		if node.IsExpanded {
			indicator = "▼ "
		} else {
			indicator = "▶ "
		}
	} else {
		indicator = "  "
	}

	// Key with syntax highlighting
	keyStyle := lipgloss.NewStyle().Foreground(jv.Theme.Info) // Blue for keys
	keyPart := keyStyle.Render(node.Key)

	// Value with syntax highlighting
	var valuePart string
	switch node.Type {
	case NodeObject:
		count := len(node.Children)
		valuePart = lipgloss.NewStyle().
			Foreground(jv.Theme.Metadata).
			Render(fmt.Sprintf(" { %d properties }", count))

	case NodeArray:
		count := len(node.Children)
		valuePart = lipgloss.NewStyle().
			Foreground(jv.Theme.Metadata).
			Render(fmt.Sprintf(" [ %d items ]", count))

	case NodeString:
		str := fmt.Sprintf("%v", node.Value)
		if len(str) > 50 {
			str = str[:47] + "..."
		}
		valuePart = lipgloss.NewStyle().
			Foreground(jv.Theme.Success). // Green for strings
			Render(fmt.Sprintf(": \"%s\"", str))

	case NodeNumber:
		valuePart = lipgloss.NewStyle().
			Foreground(jv.Theme.Warning). // Yellow/orange for numbers
			Render(fmt.Sprintf(": %v", node.Value))

	case NodeBoolean:
		valuePart = lipgloss.NewStyle().
			Foreground(jv.Theme.Error). // Red for booleans
			Render(fmt.Sprintf(": %v", node.Value))

	case NodeNull:
		valuePart = lipgloss.NewStyle().
			Foreground(jv.Theme.Metadata).
			Italic(true).
			Render(": null")
	}

	line := indent + indicator + keyPart + valuePart

	// Highlight selected row
	if isSelected {
		return lipgloss.NewStyle().
			Background(jv.Theme.Selection).
			Foreground(jv.Theme.Background).
			Bold(true).
			Width(jv.Width - 6). // Account for container padding
			Render(line)
	}

	return line
}

// renderStatus renders the status bar at the bottom
func (jv *JSONBViewer) renderStatus() string {
	totalNodes := len(jv.visibleNodes)
	currentPos := jv.selectedIndex + 1

	var pathStr string
	if jv.selectedIndex < len(jv.visibleNodes) {
		node := jv.visibleNodes[jv.selectedIndex]
		if len(node.Path) > 0 {
			pathStr = "Path: $." + strings.Join(node.Path, ".")
		} else {
			pathStr = "Path: $"
		}

		// Truncate if too long
		maxPathLen := jv.Width - 30
		if len(pathStr) > maxPathLen {
			pathStr = pathStr[:maxPathLen-3] + "..."
		}
	}

	status := fmt.Sprintf(" %d/%d  %s", currentPos, totalNodes, pathStr)

	return lipgloss.NewStyle().
		Foreground(jv.Theme.Metadata).
		Italic(true).
		Render(status)
}
