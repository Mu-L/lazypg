# Phase 6: JSONB Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build comprehensive JSONB support with formatting, three-mode viewer (Formatted/Tree/Query), path extraction, and filtering integration.

**Architecture:** Create a JSONB viewer component with multiple display modes. Implement path extraction algorithm to generate JSON paths from nested structures. Integrate JSONB-specific filtering operators (@>, <@, ?) with the existing filter builder. Use Go's encoding/json for parsing and formatting.

**Tech Stack:** Go 1.21+, Bubble Tea, Lipgloss, encoding/json, pgx v5, existing filter system

---

## Task 1: JSONB Formatting and Detection

**Files:**
- Create: `internal/jsonb/formatter.go`
- Create: `internal/jsonb/detector.go`
- Modify: `internal/ui/components/table_view.go`

**Step 1: Create JSONB formatter**

Create `internal/jsonb/formatter.go`:

```go
package jsonb

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Format formats a JSONB value as a pretty-printed JSON string
func Format(value interface{}) (string, error) {
	if value == nil {
		return "null", nil
	}

	// Convert to JSON bytes
	var jsonBytes []byte
	switch v := value.(type) {
	case string:
		// Already a JSON string, parse it first
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return "", fmt.Errorf("invalid JSON: %w", err)
		}
		jsonBytes, _ = json.MarshalIndent(parsed, "", "  ")
	case []byte:
		// JSON bytes, parse and re-format
		var parsed interface{}
		if err := json.Unmarshal(v, &parsed); err != nil {
			return "", fmt.Errorf("invalid JSON: %w", err)
		}
		jsonBytes, _ = json.MarshalIndent(parsed, "", "  ")
	default:
		// Other types, marshal directly
		var err error
		jsonBytes, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to format: %w", err)
		}
	}

	return string(jsonBytes), nil
}

// Compact formats JSONB as compact (single-line) JSON
func Compact(value interface{}) (string, error) {
	if value == nil {
		return "null", nil
	}

	var jsonBytes []byte
	switch v := value.(type) {
	case string:
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return "", fmt.Errorf("invalid JSON: %w", err)
		}
		jsonBytes, _ = json.Marshal(parsed)
	case []byte:
		var parsed interface{}
		if err := json.Unmarshal(v, &parsed); err != nil {
			return "", fmt.Errorf("invalid JSON: %w", err)
		}
		jsonBytes, _ = json.Marshal(parsed)
	default:
		var err error
		jsonBytes, err = json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to compact: %w", err)
		}
	}

	return string(jsonBytes), nil
}

// Truncate truncates a JSON string for table display
func Truncate(jsonStr string, maxLen int) string {
	if len(jsonStr) <= maxLen {
		return jsonStr
	}

	// Try to truncate at a reasonable boundary
	truncated := jsonStr[:maxLen-3]

	// Find last space, comma, or bracket
	lastGood := strings.LastIndexAny(truncated, " ,{}[]")
	if lastGood > maxLen/2 {
		truncated = truncated[:lastGood]
	}

	return truncated + "..."
}
```

**Step 2: Create JSONB detector**

Create `internal/jsonb/detector.go`:

```go
package jsonb

import (
	"encoding/json"
	"strings"
)

// IsJSONB checks if a string value looks like JSONB
func IsJSONB(value string) bool {
	if value == "" {
		return false
	}

	// Quick check for JSON-like start
	value = strings.TrimSpace(value)
	if len(value) == 0 {
		return false
	}

	first := value[0]
	if first != '{' && first != '[' && first != '"' {
		// Could be null, true, false, or number
		if value == "null" || value == "true" || value == "false" {
			return true
		}
		// Try parsing as number
		var f float64
		err := json.Unmarshal([]byte(value), &f)
		return err == nil
	}

	// Try to parse as JSON
	var parsed interface{}
	err := json.Unmarshal([]byte(value), &parsed)
	return err == nil
}

// Type returns the type of a JSONB value (object, array, string, number, boolean, null)
func Type(value interface{}) string {
	if value == nil {
		return "null"
	}

	var parsed interface{}
	switch v := value.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return "unknown"
		}
	case []byte:
		if err := json.Unmarshal(v, &parsed); err != nil {
			return "unknown"
		}
	default:
		parsed = v
	}

	switch parsed.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}
```

**Step 3: Update table view to format JSONB cells**

In `internal/ui/components/table_view.go`, modify the `View()` method to detect and format JSONB:

Find the cell rendering code (around where cells are displayed) and add:

```go
import (
	"github.com/rebeliceyang/lazypg/internal/jsonb"
)

// In the cell rendering loop, add JSONB detection:
cellValue := tv.Rows[rowIdx][colIdx]

// Check if this looks like JSONB and truncate for display
if jsonb.IsJSONB(cellValue) {
	cellValue = jsonb.Truncate(cellValue, 50)
	// Add indicator that it's JSONB
	cellValue = "📦 " + cellValue
}
```

**Step 4: Build and verify**

Run: `go build ./...`
Expected: Clean build

**Step 5: Commit**

```bash
git add internal/jsonb/formatter.go internal/jsonb/detector.go internal/ui/components/table_view.go
git commit -m "feat(jsonb): add JSONB formatting and detection"
```

---

## Task 2: JSONB Path Extraction

**Files:**
- Create: `internal/jsonb/path.go`

**Step 1: Create path extraction algorithm**

Create `internal/jsonb/path.go`:

```go
package jsonb

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Path represents a JSON path (e.g., $.user.address.city)
type Path struct {
	Parts []string
}

// String returns the PostgreSQL JSONB path notation
func (p Path) String() string {
	if len(p.Parts) == 0 {
		return "$"
	}

	result := "$"
	for _, part := range p.Parts {
		// Check if part is numeric (array index)
		if _, err := strconv.Atoi(part); err == nil {
			result += "[" + part + "]"
		} else {
			result += "." + part
		}
	}
	return result
}

// PostgreSQLPath returns the PostgreSQL #> operator notation
func (p Path) PostgreSQLPath() string {
	if len(p.Parts) == 0 {
		return "{}"
	}
	return "{" + joinWithQuotes(p.Parts) + "}"
}

func joinWithQuotes(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += ","
		}
		result += part
	}
	return result
}

// ExtractPaths extracts all possible paths from a JSONB value
func ExtractPaths(value interface{}) []Path {
	var paths []Path

	var parsed interface{}
	switch v := value.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return paths
		}
	case []byte:
		if err := json.Unmarshal(v, &parsed); err != nil {
			return paths
		}
	default:
		parsed = v
	}

	extractPathsRecursive(parsed, []string{}, &paths)
	return paths
}

func extractPathsRecursive(value interface{}, currentPath []string, paths *[]Path) {
	if value == nil {
		*paths = append(*paths, Path{Parts: currentPath})
		return
	}

	switch v := value.(type) {
	case map[string]interface{}:
		// Add path for the object itself
		*paths = append(*paths, Path{Parts: currentPath})

		// Recurse into each key
		for key, val := range v {
			newPath := append([]string{}, currentPath...)
			newPath = append(newPath, key)
			extractPathsRecursive(val, newPath, paths)
		}

	case []interface{}:
		// Add path for the array itself
		*paths = append(*paths, Path{Parts: currentPath})

		// Recurse into array elements (limit to first 5 for performance)
		limit := len(v)
		if limit > 5 {
			limit = 5
		}
		for i := 0; i < limit; i++ {
			newPath := append([]string{}, currentPath...)
			newPath = append(newPath, strconv.Itoa(i))
			extractPathsRecursive(v[i], newPath, paths)
		}

	default:
		// Leaf value (string, number, boolean)
		*paths = append(*paths, Path{Parts: currentPath})
	}
}

// GetValueAtPath retrieves a value at a specific path
func GetValueAtPath(value interface{}, path Path) (interface{}, error) {
	var parsed interface{}
	switch v := value.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	case []byte:
		if err := json.Unmarshal(v, &parsed); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
	default:
		parsed = v
	}

	current := parsed
	for _, part := range path.Parts {
		switch curr := current.(type) {
		case map[string]interface{}:
			val, ok := curr[part]
			if !ok {
				return nil, fmt.Errorf("key '%s' not found", part)
			}
			current = val
		case []interface{}:
			idx, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", part)
			}
			if idx < 0 || idx >= len(curr) {
				return nil, fmt.Errorf("array index out of bounds: %d", idx)
			}
			current = curr[idx]
		default:
			return nil, fmt.Errorf("cannot traverse into %T", curr)
		}
	}

	return current, nil
}
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: Clean build

**Step 3: Commit**

```bash
git add internal/jsonb/path.go
git commit -m "feat(jsonb): add path extraction algorithm"
```

---

## Task 3: JSONB Tree Viewer Component

**Files:**
- Create: `internal/ui/components/jsonb_viewer.go`

**Step 1: Create JSONB tree viewer**

Create `internal/ui/components/jsonb_viewer.go`:

```go
package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rebeliceyang/lazypg/internal/jsonb"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// JSONBViewMode represents the display mode
type JSONBViewMode int

const (
	JSONBViewFormatted JSONBViewMode = iota
	JSONBViewTree
	JSONBViewQuery
)

// CloseJSONBViewerMsg is sent when viewer should close
type CloseJSONBViewerMsg struct{}

// JSONBViewer displays JSONB data in multiple modes
type JSONBViewer struct {
	Width  int
	Height int
	Theme  theme.Theme

	// Data
	value       interface{}
	formatted   string
	paths       []jsonb.Path
	currentMode JSONBViewMode

	// Tree view state
	selected     int
	offset       int
	expandedKeys map[string]bool
}

// NewJSONBViewer creates a new JSONB viewer
func NewJSONBViewer(th theme.Theme) *JSONBViewer {
	return &JSONBViewer{
		Width:        80,
		Height:       30,
		Theme:        th,
		currentMode:  JSONBViewFormatted,
		expandedKeys: make(map[string]bool),
	}
}

// SetValue sets the JSONB value to display
func (jv *JSONBViewer) SetValue(value interface{}) error {
	jv.value = value

	// Format the value
	formatted, err := jsonb.Format(value)
	if err != nil {
		return err
	}
	jv.formatted = formatted

	// Extract paths
	jv.paths = jsonb.ExtractPaths(value)
	jv.selected = 0
	jv.offset = 0

	return nil
}

// Update handles keyboard input
func (jv *JSONBViewer) Update(msg tea.KeyMsg) (*JSONBViewer, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		return jv, func() tea.Msg {
			return CloseJSONBViewerMsg{}
		}
	case "1":
		jv.currentMode = JSONBViewFormatted
	case "2":
		jv.currentMode = JSONBViewTree
	case "3":
		jv.currentMode = JSONBViewQuery
	case "up", "k":
		if jv.currentMode == JSONBViewTree && jv.selected > 0 {
			jv.selected--
			if jv.selected < jv.offset {
				jv.offset = jv.selected
			}
		}
	case "down", "j":
		if jv.currentMode == JSONBViewTree && jv.selected < len(jv.paths)-1 {
			jv.selected++
			visibleHeight := jv.Height - 8
			if jv.selected >= jv.offset+visibleHeight {
				jv.offset = jv.selected - visibleHeight + 1
			}
		}
	}

	return jv, nil
}

// View renders the JSONB viewer
func (jv *JSONBViewer) View() string {
	var sections []string

	// Title bar with mode indicators
	titleStyle := lipgloss.NewStyle().
		Foreground(jv.Theme.Foreground).
		Background(jv.Theme.Info).
		Padding(0, 1).
		Bold(true)

	modes := []string{"1:Formatted", "2:Tree", "3:Query"}
	for i, mode := range modes {
		if JSONBViewMode(i) == jv.currentMode {
			modes[i] = "[" + mode + "]"
		}
	}
	title := "JSONB Viewer    " + strings.Join(modes, "  ")
	sections = append(sections, titleStyle.Render(title))

	// Instructions
	instrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a6adc8")).
		Padding(0, 1)
	sections = append(sections, instrStyle.Render("1-3: Switch mode  ↑↓: Navigate  Esc: Close"))

	// Content based on mode
	contentHeight := jv.Height - 6
	var content string

	switch jv.currentMode {
	case JSONBViewFormatted:
		content = jv.renderFormatted(contentHeight)
	case JSONBViewTree:
		content = jv.renderTree(contentHeight)
	case JSONBViewQuery:
		content = jv.renderQuery(contentHeight)
	}

	sections = append(sections, content)

	// Container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(jv.Theme.Border).
		Width(jv.Width).
		Height(jv.Height).
		Padding(1)

	return containerStyle.Render(strings.Join(sections, "\n"))
}

func (jv *JSONBViewer) renderFormatted(height int) string {
	lines := strings.Split(jv.formatted, "\n")

	// Limit to visible height
	if len(lines) > height {
		lines = lines[:height]
		lines = append(lines, "...")
	}

	style := lipgloss.NewStyle().
		Foreground(jv.Theme.Foreground).
		Padding(0, 1)

	return style.Render(strings.Join(lines, "\n"))
}

func (jv *JSONBViewer) renderTree(height int) string {
	var lines []string

	visibleStart := jv.offset
	visibleEnd := jv.offset + height
	if visibleEnd > len(jv.paths) {
		visibleEnd = len(jv.paths)
	}

	for i := visibleStart; i < visibleEnd; i++ {
		path := jv.paths[i]
		indent := strings.Repeat("  ", len(path.Parts))

		// Get path label
		label := "$"
		if len(path.Parts) > 0 {
			label = path.Parts[len(path.Parts)-1]
		}

		// Get value at path
		value, err := jsonb.GetValueAtPath(jv.value, path)
		valueStr := ""
		if err == nil {
			switch v := value.(type) {
			case string:
				valueStr = fmt.Sprintf(": \"%s\"", truncate(v, 30))
			case float64:
				valueStr = fmt.Sprintf(": %v", v)
			case bool:
				valueStr = fmt.Sprintf(": %v", v)
			case nil:
				valueStr = ": null"
			case map[string]interface{}:
				valueStr = fmt.Sprintf(" {%d keys}", len(v))
			case []interface{}:
				valueStr = fmt.Sprintf(" [%d items]", len(v))
			}
		}

		line := indent + label + valueStr

		// Highlight selected
		style := lipgloss.NewStyle().Padding(0, 1)
		if i == jv.selected {
			style = style.Background(jv.Theme.Selection).Foreground(jv.Theme.Foreground)
		}

		lines = append(lines, style.Render(line))
	}

	return strings.Join(lines, "\n")
}

func (jv *JSONBViewer) renderQuery(height int) string {
	if jv.selected >= len(jv.paths) {
		return ""
	}

	selectedPath := jv.paths[jv.selected]

	var lines []string
	lines = append(lines, "Selected Path:")
	lines = append(lines, "  " + selectedPath.String())
	lines = append(lines, "")
	lines = append(lines, "PostgreSQL Queries:")
	lines = append(lines, "")

	// Assume column name is 'data'
	colName := "data"

	// #> operator (returns JSONB)
	lines = append(lines, fmt.Sprintf("Get JSONB value:"))
	lines = append(lines, fmt.Sprintf("  %s #> '%s'", colName, selectedPath.PostgreSQLPath()))
	lines = append(lines, "")

	// #>> operator (returns text)
	lines = append(lines, fmt.Sprintf("Get text value:"))
	lines = append(lines, fmt.Sprintf("  %s #>> '%s'", colName, selectedPath.PostgreSQLPath()))
	lines = append(lines, "")

	// @> operator (contains)
	lines = append(lines, fmt.Sprintf("Filter rows containing this path:"))
	lines = append(lines, fmt.Sprintf("  %s @> '{...}'", colName))

	style := lipgloss.NewStyle().
		Foreground(jv.Theme.Foreground).
		Padding(0, 1)

	return style.Render(strings.Join(lines, "\n"))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: Clean build

**Step 3: Commit**

```bash
git add internal/ui/components/jsonb_viewer.go
git commit -m "feat(jsonb): add three-mode JSONB viewer component"
```

---

## Task 4: Integrate JSONB Viewer with App

**Files:**
- Modify: `internal/app/app.go:60-70` (Add JSONB viewer fields)
- Modify: `internal/app/app.go:155-175` (Initialize JSONB viewer)
- Modify: `internal/app/app.go:370-390` (Add keyboard shortcut)
- Modify: `internal/app/app.go:500-530` (Render JSONB viewer)

**Step 1: Add JSONB viewer to App struct**

In `internal/app/app.go`, add to the App struct (around line 70):

```go
// JSONB viewer
showJSONBViewer bool
jsonbViewer     *components.JSONBViewer
```

**Step 2: Initialize JSONB viewer in New()**

In the `New()` function (around line 170), add:

```go
jsonbViewer := components.NewJSONBViewer(th)
```

And in the App struct initialization:

```go
showJSONBViewer: false,
jsonbViewer:     jsonbViewer,
```

**Step 3: Add keyboard shortcut to open JSONB viewer**

In the `Update()` method, add a case for opening JSONB viewer on selected cell (around line 380):

```go
case "j":
	// Open JSONB viewer if cell contains JSONB
	if a.state.FocusedPanel == models.RightPanel && a.tableView != nil {
		selectedRow, selectedCol := a.tableView.GetSelectedCell()
		if selectedRow >= 0 && selectedCol >= 0 {
			cellValue := a.tableView.Rows[selectedRow][selectedCol]
			if jsonb.IsJSONB(cellValue) {
				if err := a.jsonbViewer.SetValue(cellValue); err == nil {
					a.showJSONBViewer = true
				}
			}
		}
	}
	return a, nil
```

**Step 4: Handle JSONB viewer messages**

Add message handling in `Update()` (around line 310):

```go
case components.CloseJSONBViewerMsg:
	a.showJSONBViewer = false
	return a, nil
```

**Step 5: Handle JSONB viewer input when visible**

In the `Update()` method, add handling for JSONB viewer (around line 330):

```go
// Handle JSONB viewer input
if a.showJSONBViewer {
	return a.handleJSONBViewer(msg)
}
```

**Step 6: Add handleJSONBViewer method**

Add after `handleFilterBuilder()`:

```go
// handleJSONBViewer handles key events when JSONB viewer is visible
func (a *App) handleJSONBViewer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.jsonbViewer, cmd = a.jsonbViewer.Update(msg)
	return a, cmd
}
```

**Step 7: Render JSONB viewer in View()**

In the `View()` method, add JSONB viewer rendering (around line 700, after filter builder):

```go
// Render JSONB viewer if visible
if a.showJSONBViewer {
	mainView = lipgloss.Place(
		a.state.Width,
		a.state.Height,
		lipgloss.Center,
		lipgloss.Center,
		a.jsonbViewer.View(),
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#555555")),
	)
}
```

**Step 8: Update help documentation**

In `internal/ui/help/help.go`, add JSONB shortcut:

```go
{"j", "Open JSONB viewer (on JSONB cell)"},
```

**Step 9: Build and verify**

Run: `go build ./...`
Expected: Clean build

**Step 10: Commit**

```bash
git add internal/app/app.go internal/ui/help/help.go
git commit -m "feat(jsonb): integrate JSONB viewer with main app"
```

---

## Task 5: JSONB Filtering Integration

**Files:**
- Modify: `internal/ui/components/filter_builder.go:160-180` (Add JSONB path input mode)
- Modify: `internal/filter/builder.go:70-90` (Enhance JSONB operators)

**Step 1: Add JSONB path input to filter builder**

In `internal/ui/components/filter_builder.go`, add a new field to the FilterBuilder struct:

```go
jsonbPath string // For JSONB path input (e.g., $.user.name)
```

And add a new edit mode in `handleValueMode`:

```go
// After entering value, if operator is JSONB-specific and value looks like it needs a path
if jv.availableOps[jv.operatorIndex] == models.OpContains ||
   jv.availableOps[jv.operatorIndex] == models.OpContainedBy ||
   jv.availableOps[jv.operatorIndex] == models.OpHasKey {
	// Value should be JSON or a key name
	// Allow users to optionally specify a path
	if strings.HasPrefix(jv.valueInput, "$.") {
		// This is a path, extract it
		jv.jsonbPath = jv.valueInput
	}
}
```

**Step 2: Update JSONB operator handling in builder**

In `internal/filter/builder.go`, update the JSONB operator cases in `buildCondition`:

```go
case models.OpContains:
	// JSONB @> operator
	// If value is a string, treat it as JSON literal
	return fmt.Sprintf(`"%s" @> $%d::jsonb`, cond.Column, paramIndex), []interface{}{cond.Value}, nil

case models.OpContainedBy:
	// JSONB <@ operator
	return fmt.Sprintf(`"%s" <@ $%d::jsonb`, cond.Column, paramIndex), []interface{}{cond.Value}, nil

case models.OpHasKey:
	// JSONB ? operator (has key)
	return fmt.Sprintf(`"%s" ? $%d`, cond.Column, paramIndex), []interface{}{cond.Value}, nil
```

**Step 3: Build and verify**

Run: `go build ./...`
Expected: Clean build

**Step 4: Commit**

```bash
git add internal/ui/components/filter_builder.go internal/filter/builder.go
git commit -m "feat(jsonb): integrate JSONB operators with filter builder"
```

---

## Task 6: Documentation and Testing

**Files:**
- Create: `docs/features/jsonb.md`
- Modify: `README.md`

**Step 1: Create JSONB documentation**

Create `docs/features/jsonb.md`:

```markdown
# JSONB Support

lazypg provides comprehensive JSONB support with multiple viewing modes and filtering capabilities.

## Detecting JSONB Columns

JSONB columns are automatically detected and displayed with a 📦 icon in table views. Values are truncated for display but can be expanded using the JSONB viewer.

## Opening the JSONB Viewer

1. Navigate to a cell containing JSONB data
2. Press `j` to open the JSONB viewer

The viewer has three modes:

### Mode 1: Formatted View

Press `1` to see pretty-printed JSON with proper indentation.

```json
{
  "user": {
    "name": "Alice",
    "age": 30,
    "address": {
      "city": "San Francisco"
    }
  }
}
```

### Mode 2: Tree View

Press `2` to see an interactive tree representation:

```
$ {3 keys}
  user {3 keys}
    name: "Alice"
    age: 30
    address {1 keys}
      city: "San Francisco"
```

Use ↑↓ to navigate through paths.

### Mode 3: Query Mode

Press `3` to see PostgreSQL query examples for the selected path:

```sql
-- Get JSONB value
data #> '{user,address,city}'

-- Get text value
data #>> '{user,address,city}'

-- Filter rows containing this path
data @> '{"user": {"address": {"city": "San Francisco"}}}'
```

## JSONB Filtering

The filter builder supports JSONB-specific operators:

### Contains (@>)

Filter rows where JSONB contains a specific value:

1. Press `f` to open filter builder
2. Select a JSONB column
3. Choose `@>` operator
4. Enter JSON value: `{"user": {"name": "Alice"}}`

### Contained By (<@)

Filter rows where JSONB is contained by a value:

1. Select JSONB column
2. Choose `<@` operator
3. Enter containing JSON value

### Has Key (?)

Filter rows where JSONB has a specific key:

1. Select JSONB column
2. Choose `?` operator
3. Enter key name: `"user"`

## Path Extraction

The tree view automatically extracts all JSON paths from nested structures. This makes it easy to:

- See the structure of complex JSON documents
- Find specific paths to use in queries
- Navigate deeply nested data

## Keyboard Shortcuts

- `j` - Open JSONB viewer (on JSONB cell)
- `1` - Switch to Formatted view
- `2` - Switch to Tree view
- `3` - Switch to Query view
- `↑`/`↓` - Navigate tree paths
- `Esc` - Close viewer

## Examples

### View User Profile

```
1. Run: SELECT * FROM users WHERE id = 123
2. Navigate to profile_data column
3. Press 'j' to open viewer
4. Press '2' for tree view
5. Navigate to $.profile.email
6. Press '3' to see query for that path
```

### Filter by Nested Value

```
1. Press 'f' to open filter builder
2. Column: profile_data
3. Operator: @>
4. Value: {"profile": {"city": "NYC"}}
5. Press Enter to apply
```
```

**Step 2: Update README**

In `README.md`, update the status section:

```markdown
## Status

**Phase 6 (JSONB Support) Complete**
- JSONB detection and formatting
- Three-mode viewer (Formatted/Tree/Query)
- Path extraction algorithm
- JSONB filtering operators (@>, <@, ?)
```

**Step 3: Manual testing checklist**

Test the following:
- [ ] JSONB cells show 📦 icon in table view
- [ ] Press `j` on JSONB cell opens viewer
- [ ] Formatted mode (1) shows pretty JSON
- [ ] Tree mode (2) shows navigable paths
- [ ] Query mode (3) shows PostgreSQL examples
- [ ] Can navigate tree with ↑↓ keys
- [ ] Filter builder offers JSONB operators for JSONB columns
- [ ] @> operator filters correctly
- [ ] ? operator finds keys correctly
- [ ] Esc closes viewer

**Step 4: Build final binary**

Run: `go build -o bin/lazypg ./cmd/lazypg`
Expected: Clean build

**Step 5: Commit**

```bash
git add docs/features/jsonb.md README.md
git commit -m "docs: add JSONB documentation and complete Phase 6"
```

---

## Summary

Phase 6 Implementation adds:

1. **JSONB Formatting** - Pretty-print and compact JSON display
2. **JSONB Detection** - Automatic detection in table cells
3. **Path Extraction** - Extract all paths from nested JSON
4. **Three-Mode Viewer** - Formatted, Tree, and Query modes
5. **Tree Navigation** - Interactive exploration of JSON structure
6. **Query Generation** - PostgreSQL examples for selected paths
7. **JSONB Filtering** - Integration with filter builder (@>, <@, ? operators)
8. **Documentation** - Comprehensive user guide

**Keyboard Shortcuts:**
- `j` - Open JSONB viewer
- `1` - Formatted view
- `2` - Tree view
- `3` - Query view
- `↑`/`↓` - Navigate paths
- `Esc` - Close

**Total Files Created:** 5
**Total Files Modified:** 5
**Estimated Implementation Time:** 2 weeks
