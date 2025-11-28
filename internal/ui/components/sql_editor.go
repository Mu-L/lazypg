package components

import (
	"strings"

	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// SQLEditorHeightPreset defines the height presets for the editor
type SQLEditorHeightPreset int

const (
	SQLEditorSmall  SQLEditorHeightPreset = iota // 20% of available height
	SQLEditorMedium                              // 35% of available height
	SQLEditorLarge                               // 50% of available height
)

// SQLEditor is a multiline SQL editor component
type SQLEditor struct {
	// Content
	lines      []string // Lines of text
	cursorRow  int      // Current cursor row (0-indexed)
	cursorCol  int      // Current cursor column (0-indexed)

	// Dimensions
	Width  int
	Height int

	// State
	expanded     bool
	heightPreset SQLEditorHeightPreset

	// Theme
	Theme theme.Theme

	// History
	history    []string
	historyIdx int
}

// NewSQLEditor creates a new SQL editor
func NewSQLEditor(th theme.Theme) *SQLEditor {
	return &SQLEditor{
		lines:        []string{""},
		cursorRow:    0,
		cursorCol:    0,
		expanded:     false,
		heightPreset: SQLEditorMedium,
		Theme:        th,
		history:      []string{},
		historyIdx:   -1,
	}
}

// IsExpanded returns whether the editor is expanded
func (e *SQLEditor) IsExpanded() bool {
	return e.expanded
}

// Toggle expands or collapses the editor
func (e *SQLEditor) Toggle() {
	e.expanded = !e.expanded
}

// Expand expands the editor
func (e *SQLEditor) Expand() {
	e.expanded = true
}

// Collapse collapses the editor
func (e *SQLEditor) Collapse() {
	e.expanded = false
}

// GetHeightPreset returns the current height preset
func (e *SQLEditor) GetHeightPreset() SQLEditorHeightPreset {
	return e.heightPreset
}

// IncreaseHeight increases the height preset
func (e *SQLEditor) IncreaseHeight() {
	if e.heightPreset < SQLEditorLarge {
		e.heightPreset++
	}
}

// DecreaseHeight decreases the height preset
func (e *SQLEditor) DecreaseHeight() {
	if e.heightPreset > SQLEditorSmall {
		e.heightPreset--
	}
}

// GetHeightRatio returns the height ratio for the current preset
func (e *SQLEditor) GetHeightRatio() float64 {
	switch e.heightPreset {
	case SQLEditorSmall:
		return 0.20
	case SQLEditorMedium:
		return 0.35
	case SQLEditorLarge:
		return 0.50
	default:
		return 0.35
	}
}

// GetContent returns the full content as a single string
func (e *SQLEditor) GetContent() string {
	return strings.Join(e.lines, "\n")
}

// SetContent sets the editor content
func (e *SQLEditor) SetContent(content string) {
	if content == "" {
		e.lines = []string{""}
	} else {
		e.lines = strings.Split(content, "\n")
	}
	e.cursorRow = len(e.lines) - 1
	e.cursorCol = len(e.lines[e.cursorRow])
}

// Clear clears the editor content
func (e *SQLEditor) Clear() {
	e.lines = []string{""}
	e.cursorRow = 0
	e.cursorCol = 0
}

// GetCollapsedHeight returns the height when collapsed (2 lines + border)
func (e *SQLEditor) GetCollapsedHeight() int {
	return 4 // 2 content lines + 2 border lines
}
