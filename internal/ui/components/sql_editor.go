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

// MoveCursorLeft moves cursor left
func (e *SQLEditor) MoveCursorLeft() {
	if e.cursorCol > 0 {
		e.cursorCol--
	} else if e.cursorRow > 0 {
		// Move to end of previous line
		e.cursorRow--
		e.cursorCol = len(e.lines[e.cursorRow])
	}
}

// MoveCursorRight moves cursor right
func (e *SQLEditor) MoveCursorRight() {
	if e.cursorCol < len(e.lines[e.cursorRow]) {
		e.cursorCol++
	} else if e.cursorRow < len(e.lines)-1 {
		// Move to start of next line
		e.cursorRow++
		e.cursorCol = 0
	}
}

// MoveCursorUp moves cursor up
func (e *SQLEditor) MoveCursorUp() {
	if e.cursorRow > 0 {
		e.cursorRow--
		// Clamp column to line length
		if e.cursorCol > len(e.lines[e.cursorRow]) {
			e.cursorCol = len(e.lines[e.cursorRow])
		}
	}
}

// MoveCursorDown moves cursor down
func (e *SQLEditor) MoveCursorDown() {
	if e.cursorRow < len(e.lines)-1 {
		e.cursorRow++
		// Clamp column to line length
		if e.cursorCol > len(e.lines[e.cursorRow]) {
			e.cursorCol = len(e.lines[e.cursorRow])
		}
	}
}

// MoveCursorToLineStart moves cursor to start of line
func (e *SQLEditor) MoveCursorToLineStart() {
	e.cursorCol = 0
}

// MoveCursorToLineEnd moves cursor to end of line
func (e *SQLEditor) MoveCursorToLineEnd() {
	e.cursorCol = len(e.lines[e.cursorRow])
}

// MoveCursorToDocStart moves cursor to start of document
func (e *SQLEditor) MoveCursorToDocStart() {
	e.cursorRow = 0
	e.cursorCol = 0
}

// MoveCursorToDocEnd moves cursor to end of document
func (e *SQLEditor) MoveCursorToDocEnd() {
	e.cursorRow = len(e.lines) - 1
	e.cursorCol = len(e.lines[e.cursorRow])
}

// InsertChar inserts a character at cursor position
func (e *SQLEditor) InsertChar(ch rune) {
	line := e.lines[e.cursorRow]
	// Insert character at cursor position
	newLine := line[:e.cursorCol] + string(ch) + line[e.cursorCol:]
	e.lines[e.cursorRow] = newLine
	e.cursorCol++
}

// InsertNewline inserts a new line at cursor position
func (e *SQLEditor) InsertNewline() {
	line := e.lines[e.cursorRow]
	// Split line at cursor
	before := line[:e.cursorCol]
	after := line[e.cursorCol:]

	e.lines[e.cursorRow] = before
	// Insert new line after current
	newLines := make([]string, len(e.lines)+1)
	copy(newLines[:e.cursorRow+1], e.lines[:e.cursorRow+1])
	newLines[e.cursorRow+1] = after
	copy(newLines[e.cursorRow+2:], e.lines[e.cursorRow+1:])
	e.lines = newLines

	e.cursorRow++
	e.cursorCol = 0
}

// DeleteCharBefore deletes character before cursor (backspace)
func (e *SQLEditor) DeleteCharBefore() {
	if e.cursorCol > 0 {
		// Delete character before cursor
		line := e.lines[e.cursorRow]
		e.lines[e.cursorRow] = line[:e.cursorCol-1] + line[e.cursorCol:]
		e.cursorCol--
	} else if e.cursorRow > 0 {
		// Merge with previous line
		prevLine := e.lines[e.cursorRow-1]
		currLine := e.lines[e.cursorRow]
		e.cursorCol = len(prevLine)
		e.lines[e.cursorRow-1] = prevLine + currLine
		// Remove current line
		e.lines = append(e.lines[:e.cursorRow], e.lines[e.cursorRow+1:]...)
		e.cursorRow--
	}
}

// DeleteCharAfter deletes character after cursor (delete key)
func (e *SQLEditor) DeleteCharAfter() {
	line := e.lines[e.cursorRow]
	if e.cursorCol < len(line) {
		// Delete character at cursor
		e.lines[e.cursorRow] = line[:e.cursorCol] + line[e.cursorCol+1:]
	} else if e.cursorRow < len(e.lines)-1 {
		// Merge with next line
		nextLine := e.lines[e.cursorRow+1]
		e.lines[e.cursorRow] = line + nextLine
		// Remove next line
		e.lines = append(e.lines[:e.cursorRow+1], e.lines[e.cursorRow+2:]...)
	}
}
