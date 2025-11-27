package components

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/rebeliceyang/lazypg/internal/jsonb"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// PreviewPane displays full content for truncated values
type PreviewPane struct {
	Width     int
	MaxHeight int    // Maximum height (screen 1/3)
	Content   string // Raw content to display
	Title     string // Title (column name or JSON path)

	// Visibility state
	Visible       bool // Whether pane should be shown
	ForceHidden   bool // User manually hid the pane (overrides auto-show)
	IsTruncated   bool // Whether content was truncated in parent view

	// Scrolling
	scrollY       int
	contentLines  []string // Formatted content split into lines

	// Styling
	Theme theme.Theme
	style lipgloss.Style
}

// NewPreviewPane creates a new preview pane
func NewPreviewPane(th theme.Theme) *PreviewPane {
	return &PreviewPane{
		Width:     80,
		MaxHeight: 10,
		Theme:     th,
		style: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(th.Border).
			Padding(0, 1),
	}
}

// SetContent sets the content to display
// isTruncated indicates whether the content was truncated in the parent view
func (p *PreviewPane) SetContent(content, title string, isTruncated bool) {
	p.Content = content
	p.Title = title
	p.IsTruncated = isTruncated
	p.scrollY = 0

	// Format content
	p.formatContent()

	// Update visibility (only auto-show if not force hidden)
	if !p.ForceHidden {
		p.Visible = isTruncated && content != "" && content != "NULL"
	}
}

// formatContent formats the raw content for display
func (p *PreviewPane) formatContent() {
	if p.Content == "" {
		p.contentLines = []string{}
		return
	}

	// Calculate available width for content
	contentWidth := p.Width - p.style.GetHorizontalFrameSize()
	if contentWidth < 10 {
		contentWidth = 10
	}

	// Try to format as JSON if it looks like JSONB
	formatted := p.Content
	if jsonb.IsJSONB(p.Content) {
		var parsed interface{}
		if err := json.Unmarshal([]byte(p.Content), &parsed); err == nil {
			if pretty, err := json.MarshalIndent(parsed, "", "  "); err == nil {
				formatted = string(pretty)
			}
		}
	}

	// Wrap lines to fit width
	p.contentLines = p.wrapText(formatted, contentWidth)
}

// wrapText wraps text to fit within maxWidth
func (p *PreviewPane) wrapText(text string, maxWidth int) []string {
	var result []string
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if runewidth.StringWidth(line) <= maxWidth {
			result = append(result, line)
			continue
		}

		// Wrap long lines
		current := ""
		currentWidth := 0
		for _, r := range line {
			rWidth := runewidth.RuneWidth(r)
			if currentWidth+rWidth > maxWidth {
				result = append(result, current)
				current = string(r)
				currentWidth = rWidth
			} else {
				current += string(r)
				currentWidth += rWidth
			}
		}
		if current != "" {
			result = append(result, current)
		}
	}

	return result
}

// Toggle toggles the preview pane visibility
// When toggled off, sets ForceHidden to prevent auto-show
// When toggled on, clears ForceHidden to allow auto-show
func (p *PreviewPane) Toggle() {
	if p.Visible {
		p.Visible = false
		p.ForceHidden = true
	} else {
		p.ForceHidden = false
		// Only show if content is truncated
		p.Visible = p.IsTruncated && p.Content != "" && p.Content != "NULL"
	}
}

// Height returns the actual rendered height including borders
// Returns 0 if not visible
func (p *PreviewPane) Height() int {
	if !p.Visible {
		return 0
	}

	// Calculate content height
	contentHeight := len(p.contentLines)
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Apply max height constraint
	maxContentHeight := p.MaxHeight - p.style.GetVerticalFrameSize()
	if maxContentHeight < 1 {
		maxContentHeight = 1
	}
	if contentHeight > maxContentHeight {
		contentHeight = maxContentHeight
	}

	// Add frame size for total height
	return contentHeight + p.style.GetVerticalFrameSize()
}

// IsScrollable returns true if content exceeds visible area
func (p *PreviewPane) IsScrollable() bool {
	maxContentHeight := p.MaxHeight - p.style.GetVerticalFrameSize()
	return len(p.contentLines) > maxContentHeight
}

// ScrollUp scrolls content up
func (p *PreviewPane) ScrollUp() {
	if p.scrollY > 0 {
		p.scrollY--
	}
}

// ScrollDown scrolls content down
func (p *PreviewPane) ScrollDown() {
	maxContentHeight := p.MaxHeight - p.style.GetVerticalFrameSize()
	maxScroll := len(p.contentLines) - maxContentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if p.scrollY < maxScroll {
		p.scrollY++
	}
}
