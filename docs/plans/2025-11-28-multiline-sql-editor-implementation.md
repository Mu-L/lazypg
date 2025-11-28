# Multiline SQL Editor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace single-line QuickQuery with a collapsible multiline SQL editor supporting syntax highlighting, line numbers, and multiple result tabs.

**Architecture:** New SQL Editor component manages text editing with syntax highlighting. Result Tabs component manages multiple query results. App layout restructured to split right panel into Data Panel (top) + SQL Editor (bottom).

**Tech Stack:** Go, Bubble Tea (TUI framework), Lipgloss (styling), existing theme system

---

## Task 1: Create SQL Editor Component - Basic Structure

**Files:**
- Create: `internal/ui/components/sql_editor.go`

**Step 1: Create basic SQL editor struct and constructor**

```go
package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
```

**Step 2: Run go build to verify syntax**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds with no errors

**Step 3: Commit**

```bash
git add internal/ui/components/sql_editor.go
git commit -m "feat(sql-editor): add basic SQL editor structure"
```

---

## Task 2: SQL Editor - Text Editing Operations

**Files:**
- Modify: `internal/ui/components/sql_editor.go`

**Step 1: Add cursor movement methods**

```go
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
```

**Step 2: Add text insertion and deletion methods**

```go
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
```

**Step 3: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add internal/ui/components/sql_editor.go
git commit -m "feat(sql-editor): add cursor movement and text editing"
```

---

## Task 3: SQL Editor - Syntax Highlighting

**Files:**
- Modify: `internal/ui/components/sql_editor.go`

**Step 1: Add SQL keyword lists and tokenizer**

```go
import (
	"regexp"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// SQL keywords for syntax highlighting
var sqlKeywords = map[string]bool{
	"SELECT": true, "FROM": true, "WHERE": true, "AND": true, "OR": true,
	"INSERT": true, "INTO": true, "VALUES": true, "UPDATE": true, "SET": true,
	"DELETE": true, "CREATE": true, "TABLE": true, "DROP": true, "ALTER": true,
	"INDEX": true, "VIEW": true, "JOIN": true, "LEFT": true, "RIGHT": true,
	"INNER": true, "OUTER": true, "FULL": true, "ON": true, "AS": true,
	"ORDER": true, "BY": true, "GROUP": true, "HAVING": true, "LIMIT": true,
	"OFFSET": true, "UNION": true, "ALL": true, "DISTINCT": true, "CASE": true,
	"WHEN": true, "THEN": true, "ELSE": true, "END": true, "NULL": true,
	"NOT": true, "IN": true, "EXISTS": true, "BETWEEN": true, "LIKE": true,
	"IS": true, "TRUE": true, "FALSE": true, "ASC": true, "DESC": true,
	"PRIMARY": true, "KEY": true, "FOREIGN": true, "REFERENCES": true,
	"CONSTRAINT": true, "UNIQUE": true, "CHECK": true, "DEFAULT": true,
	"CASCADE": true, "NULLS": true, "FIRST": true, "LAST": true,
	"BEGIN": true, "COMMIT": true, "ROLLBACK": true, "TRANSACTION": true,
	"WITH": true, "RECURSIVE": true, "RETURNING": true, "COALESCE": true,
	"CAST": true, "COUNT": true, "SUM": true, "AVG": true, "MIN": true, "MAX": true,
}

// TokenType represents the type of a syntax token
type TokenType int

const (
	TokenText TokenType = iota
	TokenKeyword
	TokenString
	TokenNumber
	TokenComment
	TokenOperator
)

// Token represents a syntax-highlighted token
type Token struct {
	Type  TokenType
	Value string
}

// tokenizeLine tokenizes a single line for syntax highlighting
func (e *SQLEditor) tokenizeLine(line string) []Token {
	var tokens []Token
	i := 0

	for i < len(line) {
		// Skip whitespace
		if unicode.IsSpace(rune(line[i])) {
			start := i
			for i < len(line) && unicode.IsSpace(rune(line[i])) {
				i++
			}
			tokens = append(tokens, Token{Type: TokenText, Value: line[start:i]})
			continue
		}

		// Comment (-- to end of line)
		if i+1 < len(line) && line[i:i+2] == "--" {
			tokens = append(tokens, Token{Type: TokenComment, Value: line[i:]})
			break
		}

		// String literal (single quotes)
		if line[i] == '\'' {
			start := i
			i++
			for i < len(line) {
				if line[i] == '\'' {
					if i+1 < len(line) && line[i+1] == '\'' {
						// Escaped quote
						i += 2
					} else {
						i++
						break
					}
				} else {
					i++
				}
			}
			tokens = append(tokens, Token{Type: TokenString, Value: line[start:i]})
			continue
		}

		// Number
		if unicode.IsDigit(rune(line[i])) || (line[i] == '.' && i+1 < len(line) && unicode.IsDigit(rune(line[i+1]))) {
			start := i
			for i < len(line) && (unicode.IsDigit(rune(line[i])) || line[i] == '.') {
				i++
			}
			tokens = append(tokens, Token{Type: TokenNumber, Value: line[start:i]})
			continue
		}

		// Identifier or keyword
		if unicode.IsLetter(rune(line[i])) || line[i] == '_' {
			start := i
			for i < len(line) && (unicode.IsLetter(rune(line[i])) || unicode.IsDigit(rune(line[i])) || line[i] == '_') {
				i++
			}
			word := line[start:i]
			if sqlKeywords[strings.ToUpper(word)] {
				tokens = append(tokens, Token{Type: TokenKeyword, Value: word})
			} else {
				tokens = append(tokens, Token{Type: TokenText, Value: word})
			}
			continue
		}

		// Operators
		if strings.ContainsRune("=<>!+-*/%&|^~", rune(line[i])) {
			start := i
			for i < len(line) && strings.ContainsRune("=<>!+-*/%&|^~", rune(line[i])) {
				i++
			}
			tokens = append(tokens, Token{Type: TokenOperator, Value: line[start:i]})
			continue
		}

		// Other single characters (parens, commas, etc.)
		tokens = append(tokens, Token{Type: TokenText, Value: string(line[i])})
		i++
	}

	return tokens
}

// renderTokens renders tokens with syntax highlighting
func (e *SQLEditor) renderTokens(tokens []Token) string {
	var result strings.Builder

	for _, token := range tokens {
		var style lipgloss.Style
		switch token.Type {
		case TokenKeyword:
			style = lipgloss.NewStyle().Foreground(e.Theme.Keyword).Bold(true)
		case TokenString:
			style = lipgloss.NewStyle().Foreground(e.Theme.String)
		case TokenNumber:
			style = lipgloss.NewStyle().Foreground(e.Theme.Number)
		case TokenComment:
			style = lipgloss.NewStyle().Foreground(e.Theme.Comment).Italic(true)
		case TokenOperator:
			style = lipgloss.NewStyle().Foreground(e.Theme.Operator)
		default:
			style = lipgloss.NewStyle().Foreground(e.Theme.Foreground)
		}
		result.WriteString(style.Render(token.Value))
	}

	return result.String()
}
```

**Step 2: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/ui/components/sql_editor.go
git commit -m "feat(sql-editor): add syntax highlighting tokenizer"
```

---

## Task 4: SQL Editor - View Rendering

**Files:**
- Modify: `internal/ui/components/sql_editor.go`

**Step 1: Add View method with line numbers and cursor**

```go
import (
	"fmt"
	// ... existing imports
)

// View renders the SQL editor
func (e *SQLEditor) View() string {
	// Calculate visible lines based on height
	contentHeight := e.Height - 2 // Account for borders
	if contentHeight < 1 {
		contentHeight = 1
	}

	// Determine which lines to show
	var visibleLines []string
	var startLine int

	if e.expanded {
		// Show all lines that fit, scroll if needed
		if e.cursorRow >= contentHeight {
			startLine = e.cursorRow - contentHeight + 1
		}
		endLine := startLine + contentHeight
		if endLine > len(e.lines) {
			endLine = len(e.lines)
		}

		for i := startLine; i < endLine; i++ {
			visibleLines = append(visibleLines, e.renderLine(i, i == e.cursorRow))
		}

		// Pad with empty lines if needed
		for len(visibleLines) < contentHeight {
			visibleLines = append(visibleLines, e.renderEmptyLine(len(e.lines)+len(visibleLines)-len(e.lines)+startLine))
		}
	} else {
		// Collapsed: show first 2 lines
		for i := 0; i < 2 && i < len(e.lines); i++ {
			visibleLines = append(visibleLines, e.renderLine(i, false))
		}
		// Pad if less than 2 lines
		for len(visibleLines) < 2 {
			visibleLines = append(visibleLines, e.renderEmptyLine(len(visibleLines)))
		}
	}

	content := strings.Join(visibleLines, "\n")

	// Container style
	borderColor := e.Theme.Border
	if e.expanded {
		borderColor = e.Theme.BorderFocused
	}

	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(e.Width - 2). // Account for border
		Height(contentHeight)

	return containerStyle.Render(content)
}

// renderLine renders a single line with line number and syntax highlighting
func (e *SQLEditor) renderLine(lineNum int, hasCursor bool) string {
	// Line number
	lineNumWidth := e.getLineNumberWidth()
	lineNumStr := fmt.Sprintf("%*d", lineNumWidth-3, lineNum+1)

	lineNumStyle := lipgloss.NewStyle().Foreground(e.Theme.Metadata)
	sepStyle := lipgloss.NewStyle().Foreground(e.Theme.Border)

	lineNumPart := lineNumStyle.Render(lineNumStr) + sepStyle.Render(" │ ")

	// Line content with syntax highlighting
	line := e.lines[lineNum]
	tokens := e.tokenizeLine(line)
	contentPart := e.renderTokens(tokens)

	// Insert cursor if this line has it
	if hasCursor && e.expanded {
		contentPart = e.insertCursor(line, tokens)
	}

	return lineNumPart + contentPart
}

// renderEmptyLine renders an empty line placeholder
func (e *SQLEditor) renderEmptyLine(lineNum int) string {
	lineNumWidth := e.getLineNumberWidth()
	lineNumStr := fmt.Sprintf("%*s", lineNumWidth-3, "~")

	lineNumStyle := lipgloss.NewStyle().Foreground(e.Theme.Metadata)
	sepStyle := lipgloss.NewStyle().Foreground(e.Theme.Border)

	return lineNumStyle.Render(lineNumStr) + sepStyle.Render(" │ ")
}

// getLineNumberWidth returns the width needed for line numbers
func (e *SQLEditor) getLineNumberWidth() int {
	maxLine := len(e.lines)
	if maxLine < 10 {
		maxLine = 10
	}
	digits := len(fmt.Sprintf("%d", maxLine))
	if digits < 2 {
		digits = 2
	}
	return digits + 3 // digits + space + separator
}

// insertCursor inserts the cursor character into the rendered line
func (e *SQLEditor) insertCursor(line string, tokens []Token) string {
	// Rebuild line with cursor
	var result strings.Builder
	charIdx := 0

	cursorStyle := lipgloss.NewStyle().
		Foreground(e.Theme.Background).
		Background(e.Theme.Cursor)

	for _, token := range tokens {
		var style lipgloss.Style
		switch token.Type {
		case TokenKeyword:
			style = lipgloss.NewStyle().Foreground(e.Theme.Keyword).Bold(true)
		case TokenString:
			style = lipgloss.NewStyle().Foreground(e.Theme.String)
		case TokenNumber:
			style = lipgloss.NewStyle().Foreground(e.Theme.Number)
		case TokenComment:
			style = lipgloss.NewStyle().Foreground(e.Theme.Comment).Italic(true)
		case TokenOperator:
			style = lipgloss.NewStyle().Foreground(e.Theme.Operator)
		default:
			style = lipgloss.NewStyle().Foreground(e.Theme.Foreground)
		}

		for _, ch := range token.Value {
			if charIdx == e.cursorCol {
				result.WriteString(cursorStyle.Render(string(ch)))
			} else {
				result.WriteString(style.Render(string(ch)))
			}
			charIdx++
		}
	}

	// Cursor at end of line
	if e.cursorCol >= charIdx {
		result.WriteString(cursorStyle.Render(" "))
	}

	return result.String()
}
```

**Step 2: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/ui/components/sql_editor.go
git commit -m "feat(sql-editor): add view rendering with line numbers"
```

---

## Task 5: SQL Editor - Update Handler and History

**Files:**
- Modify: `internal/ui/components/sql_editor.go`

**Step 1: Add Update method for keyboard handling**

```go
// Update handles keyboard input
func (e *SQLEditor) Update(msg tea.KeyMsg) (*SQLEditor, tea.Cmd) {
	switch msg.String() {
	// Cursor movement
	case "left":
		e.MoveCursorLeft()
	case "right":
		e.MoveCursorRight()
	case "up":
		e.MoveCursorUp()
	case "down":
		e.MoveCursorDown()
	case "home":
		e.MoveCursorToLineStart()
	case "end":
		e.MoveCursorToLineEnd()
	case "ctrl+home":
		e.MoveCursorToDocStart()
	case "ctrl+end":
		e.MoveCursorToDocEnd()

	// Text editing
	case "backspace":
		e.DeleteCharBefore()
	case "delete":
		e.DeleteCharAfter()
	case "enter":
		e.InsertNewline()
	case "ctrl+u":
		e.Clear()

	// History navigation
	case "ctrl+up":
		e.HistoryPrev()
	case "ctrl+down":
		e.HistoryNext()

	// Execute
	case "ctrl+enter":
		sql := e.GetCurrentStatement()
		if sql != "" {
			e.AddToHistory(e.GetContent())
			return e, func() tea.Msg {
				return ExecuteQueryMsg{SQL: sql}
			}
		}

	default:
		// Handle printable characters
		if len(msg.String()) == 1 {
			ch := rune(msg.String()[0])
			if ch >= 32 && ch <= 126 {
				e.InsertChar(ch)
			}
		} else if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				e.InsertChar(r)
			}
		}
	}

	return e, nil
}

// AddToHistory adds content to history
func (e *SQLEditor) AddToHistory(content string) {
	if content == "" {
		return
	}
	// Avoid duplicates
	if len(e.history) > 0 && e.history[len(e.history)-1] == content {
		return
	}
	e.history = append(e.history, content)
	e.historyIdx = len(e.history)
}

// HistoryPrev navigates to previous history entry
func (e *SQLEditor) HistoryPrev() {
	if len(e.history) == 0 {
		return
	}
	if e.historyIdx > 0 {
		e.historyIdx--
		e.SetContent(e.history[e.historyIdx])
	}
}

// HistoryNext navigates to next history entry
func (e *SQLEditor) HistoryNext() {
	if len(e.history) == 0 {
		return
	}
	if e.historyIdx < len(e.history)-1 {
		e.historyIdx++
		e.SetContent(e.history[e.historyIdx])
	} else {
		e.historyIdx = len(e.history)
		e.Clear()
	}
}

// GetCurrentStatement returns the SQL statement at cursor position
func (e *SQLEditor) GetCurrentStatement() string {
	content := e.GetContent()
	if content == "" {
		return ""
	}

	// Find statement boundaries using semicolons
	statements := splitStatements(content)
	if len(statements) == 0 {
		return strings.TrimSpace(content)
	}

	// Find which statement the cursor is in
	charPos := 0
	for row := 0; row < e.cursorRow; row++ {
		charPos += len(e.lines[row]) + 1 // +1 for newline
	}
	charPos += e.cursorCol

	// Find the statement containing this position
	currentPos := 0
	for _, stmt := range statements {
		stmtLen := len(stmt)
		if charPos >= currentPos && charPos <= currentPos+stmtLen {
			return strings.TrimSpace(stmt)
		}
		currentPos += stmtLen + 1 // +1 for semicolon
	}

	// Return last statement if cursor is at end
	return strings.TrimSpace(statements[len(statements)-1])
}

// splitStatements splits SQL content into individual statements
func splitStatements(content string) []string {
	var statements []string
	var current strings.Builder
	inString := false

	for i := 0; i < len(content); i++ {
		ch := content[i]

		if ch == '\'' && !inString {
			inString = true
			current.WriteByte(ch)
		} else if ch == '\'' && inString {
			// Check for escaped quote
			if i+1 < len(content) && content[i+1] == '\'' {
				current.WriteByte(ch)
				current.WriteByte(content[i+1])
				i++
			} else {
				inString = false
				current.WriteByte(ch)
			}
		} else if ch == ';' && !inString {
			stmt := current.String()
			if strings.TrimSpace(stmt) != "" {
				statements = append(statements, stmt)
			}
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}

	// Add remaining content
	stmt := current.String()
	if strings.TrimSpace(stmt) != "" {
		statements = append(statements, stmt)
	}

	return statements
}
```

**Step 2: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/ui/components/sql_editor.go
git commit -m "feat(sql-editor): add keyboard handling and history"
```

---

## Task 6: Create Result Tabs Component

**Files:**
- Create: `internal/ui/components/result_tabs.go`

**Step 1: Create ResultTab and ResultTabs structures**

```go
package components

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rebeliceyang/lazypg/internal/models"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

const MaxResultTabs = 10

// ResultTab represents a single query result tab
type ResultTab struct {
	ID         int
	Title      string
	SQL        string
	Result     models.QueryResult
	CreatedAt  time.Time
	TableView  *TableView
}

// ResultTabs manages multiple query result tabs
type ResultTabs struct {
	tabs       []*ResultTab
	activeIdx  int
	nextID     int
	Theme      theme.Theme
	Width      int
}

// NewResultTabs creates a new result tabs manager
func NewResultTabs(th theme.Theme) *ResultTabs {
	return &ResultTabs{
		tabs:      []*ResultTab{},
		activeIdx: 0,
		nextID:    1,
		Theme:     th,
	}
}

// AddResult adds a new query result as a tab
func (rt *ResultTabs) AddResult(sql string, result models.QueryResult) {
	// Create TableView for this result
	tableView := NewTableView(rt.Theme)
	tableView.SetData(result.Columns, result.Rows, len(result.Rows))

	tab := &ResultTab{
		ID:        rt.nextID,
		Title:     rt.generateTitle(sql, result),
		SQL:       sql,
		Result:    result,
		CreatedAt: time.Now(),
		TableView: tableView,
	}
	rt.nextID++

	// Add to tabs
	rt.tabs = append(rt.tabs, tab)

	// Remove oldest if exceeding max
	if len(rt.tabs) > MaxResultTabs {
		rt.tabs = rt.tabs[1:]
		// Adjust active index
		if rt.activeIdx > 0 {
			rt.activeIdx--
		}
	}

	// Set new tab as active
	rt.activeIdx = len(rt.tabs) - 1
}

// generateTitle generates a smart title for the tab
func (rt *ResultTabs) generateTitle(sql string, result models.QueryResult) string {
	// Check for custom comment title
	if title := rt.extractCommentTitle(sql); title != "" {
		return title
	}

	// Extract table name from SQL
	if tableName := rt.extractTableName(sql); tableName != "" {
		return tableName
	}

	// Fallback to truncated SQL
	cleaned := strings.TrimSpace(sql)
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	if len(cleaned) > 20 {
		cleaned = cleaned[:17] + "..."
	}
	return cleaned
}

// extractCommentTitle extracts title from SQL comment (-- title or /* title */)
func (rt *ResultTabs) extractCommentTitle(sql string) string {
	// Match -- comment at start
	dashComment := regexp.MustCompile(`^\s*--\s*(.+)$`)
	lines := strings.Split(sql, "\n")
	if len(lines) > 0 {
		if matches := dashComment.FindStringSubmatch(lines[0]); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// Match /* comment */ at start
	blockComment := regexp.MustCompile(`^\s*/\*\s*(.+?)\s*\*/`)
	if matches := blockComment.FindStringSubmatch(sql); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// extractTableName extracts the main table name from SQL
func (rt *ResultTabs) extractTableName(sql string) string {
	upperSQL := strings.ToUpper(sql)

	// SELECT ... FROM table
	fromRegex := regexp.MustCompile(`(?i)\bFROM\s+([a-zA-Z_][a-zA-Z0-9_.]*)`+
		`(?:\s+(?:AS\s+)?[a-zA-Z_][a-zA-Z0-9_]*)?`)
	if matches := fromRegex.FindStringSubmatch(sql); len(matches) > 1 {
		tableName := matches[1]
		// Check for JOIN
		if strings.Contains(upperSQL, "JOIN") {
			return tableName + "(+)"
		}
		return tableName
	}

	// UPDATE table
	updateRegex := regexp.MustCompile(`(?i)\bUPDATE\s+([a-zA-Z_][a-zA-Z0-9_.]*)`)
	if matches := updateRegex.FindStringSubmatch(sql); len(matches) > 1 {
		return "UPDATE " + matches[1]
	}

	// DELETE FROM table
	deleteRegex := regexp.MustCompile(`(?i)\bDELETE\s+FROM\s+([a-zA-Z_][a-zA-Z0-9_.]*)`)
	if matches := deleteRegex.FindStringSubmatch(sql); len(matches) > 1 {
		return "DELETE " + matches[1]
	}

	// INSERT INTO table
	insertRegex := regexp.MustCompile(`(?i)\bINSERT\s+INTO\s+([a-zA-Z_][a-zA-Z0-9_.]*)`)
	if matches := insertRegex.FindStringSubmatch(sql); len(matches) > 1 {
		return "INSERT " + matches[1]
	}

	return ""
}

// GetActiveTab returns the currently active tab
func (rt *ResultTabs) GetActiveTab() *ResultTab {
	if len(rt.tabs) == 0 || rt.activeIdx < 0 || rt.activeIdx >= len(rt.tabs) {
		return nil
	}
	return rt.tabs[rt.activeIdx]
}

// GetActiveTableView returns the TableView of the active tab
func (rt *ResultTabs) GetActiveTableView() *TableView {
	tab := rt.GetActiveTab()
	if tab == nil {
		return nil
	}
	return tab.TableView
}

// NextTab switches to the next tab
func (rt *ResultTabs) NextTab() {
	if len(rt.tabs) > 0 {
		rt.activeIdx = (rt.activeIdx + 1) % len(rt.tabs)
	}
}

// PrevTab switches to the previous tab
func (rt *ResultTabs) PrevTab() {
	if len(rt.tabs) > 0 {
		rt.activeIdx = (rt.activeIdx - 1 + len(rt.tabs)) % len(rt.tabs)
	}
}

// TabCount returns the number of tabs
func (rt *ResultTabs) TabCount() int {
	return len(rt.tabs)
}

// HasTabs returns whether there are any tabs
func (rt *ResultTabs) HasTabs() bool {
	return len(rt.tabs) > 0
}
```

**Step 2: Add View method for tab bar rendering**

```go
// RenderTabBar renders the tab bar
func (rt *ResultTabs) RenderTabBar(width int) string {
	if len(rt.tabs) == 0 {
		return ""
	}

	var tabViews []string

	for i, tab := range rt.tabs {
		// Format: [index] title (rows)
		rowCount := len(tab.Result.Rows)
		rowStr := fmt.Sprintf("%d rows", rowCount)
		if rowCount == 1 {
			rowStr = "1 row"
		}

		label := fmt.Sprintf("[%d] %s (%s)", i+1, tab.Title, rowStr)

		// Truncate if too long
		maxLabelLen := width / MaxResultTabs
		if maxLabelLen < 15 {
			maxLabelLen = 15
		}
		if len(label) > maxLabelLen {
			// Try without row count
			label = fmt.Sprintf("[%d] %s", i+1, tab.Title)
			if len(label) > maxLabelLen {
				label = label[:maxLabelLen-3] + "..."
			}
		}

		var style lipgloss.Style
		if i == rt.activeIdx {
			style = lipgloss.NewStyle().
				Foreground(rt.Theme.Background).
				Background(rt.Theme.Info).
				Bold(true).
				Padding(0, 1)
		} else {
			style = lipgloss.NewStyle().
				Foreground(rt.Theme.Foreground).
				Background(rt.Theme.Selection).
				Padding(0, 1)
		}

		tabViews = append(tabViews, style.Render(label))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabViews...)
}
```

**Step 3: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add internal/ui/components/result_tabs.go
git commit -m "feat(result-tabs): add result tabs component"
```

---

## Task 7: Integrate SQL Editor into App - State and Initialization

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add SQL editor and result tabs to App struct**

Add these fields to the `App` struct (around line 65-67, replacing quickQuery):

```go
// SQL Editor (replaces Quick Query)
sqlEditor   *components.SQLEditor
resultTabs  *components.ResultTabs
sqlEditorFocused bool // true when SQL editor has focus
```

**Step 2: Update New() function to initialize components**

Replace `quickQuery: components.NewQuickQuery(th),` with:

```go
sqlEditor:  components.NewSQLEditor(th),
resultTabs: components.NewResultTabs(th),
```

Remove the `showQuickQuery` field initialization.

**Step 3: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build may have errors for removed quickQuery references - that's expected, we'll fix them in next tasks

**Step 4: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): add sql editor and result tabs state"
```

---

## Task 8: Update App Layout - Right Panel Split

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Update renderRightPanel to include SQL editor**

Replace the `renderRightPanel` method (around line 1472-1532):

```go
// renderRightPanel renders the right panel content based on current state
func (a *App) renderRightPanel(width, height int) string {
	// Calculate SQL editor height
	editorHeight := a.sqlEditor.GetCollapsedHeight()
	if a.sqlEditor.IsExpanded() {
		editorHeight = int(float64(height) * a.sqlEditor.GetHeightRatio())
		if editorHeight < 5 {
			editorHeight = 5
		}
	}

	// Calculate data panel height
	dataPanelHeight := height - editorHeight - 1 // -1 for tab bar
	if dataPanelHeight < 5 {
		dataPanelHeight = 5
	}

	// Render tab bar
	tabBar := ""
	if a.resultTabs.HasTabs() {
		tabBar = a.resultTabs.RenderTabBar(width)
	}

	// Render data panel
	dataPanel := a.renderDataPanel(width, dataPanelHeight)

	// Render SQL editor
	a.sqlEditor.Width = width
	a.sqlEditor.Height = editorHeight
	sqlEditorView := a.sqlEditor.View()

	// Combine vertically: tab bar + data + editor
	if tabBar != "" {
		return lipgloss.JoinVertical(lipgloss.Left, tabBar, dataPanel, sqlEditorView)
	}
	return lipgloss.JoinVertical(lipgloss.Left, dataPanel, sqlEditorView)
}

// renderDataPanel renders the data panel (table view or structure view)
func (a *App) renderDataPanel(width, height int) string {
	// If we have result tabs, show the active tab's table view
	if a.resultTabs.HasTabs() {
		activeTable := a.resultTabs.GetActiveTableView()
		if activeTable != nil {
			activeTable.Width = width
			activeTable.Height = height
			return activeTable.View()
		}
	}

	// If table is selected in tree, show structure view
	if a.currentTable != "" {
		// ... existing structure view code ...
		// (copy from original renderRightPanel)
		a.structureView.Width = width
		a.structureView.Height = height

		conn, err := a.connectionManager.GetActive()
		if err == nil && conn != nil && conn.Pool != nil {
			parts := strings.Split(a.currentTable, ".")
			if len(parts) == 2 {
				if !a.structureView.HasTableLoaded(parts[0], parts[1]) {
					ctx := context.Background()
					err := a.structureView.SetTable(ctx, conn.Pool, parts[0], parts[1])
					if err != nil {
						log.Printf("Failed to load structure: %v", err)
					}
				}
			}
		}

		return a.structureView.View()
	}

	// No data - show placeholder
	placeholderStyle := lipgloss.NewStyle().
		Foreground(a.theme.Comment).
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center)

	return placeholderStyle.Render("No data to display\n\nPress Ctrl+E to open SQL editor")
}
```

**Step 2: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): update layout with sql editor and tabs"
```

---

## Task 9: Update App - Keyboard Handling for SQL Editor

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add SQL editor toggle and focus handling in Update method**

Find the keyboard handling section in Update() and add these cases:

```go
// In the tea.KeyMsg switch statement, add:

case "ctrl+e":
	// Toggle SQL editor expand/collapse
	a.sqlEditor.Toggle()
	if a.sqlEditor.IsExpanded() {
		a.sqlEditorFocused = true
	}
	return a, nil

case "ctrl+shift+up":
	if a.sqlEditorFocused && a.sqlEditor.IsExpanded() {
		a.sqlEditor.IncreaseHeight()
	}
	return a, nil

case "ctrl+shift+down":
	if a.sqlEditorFocused && a.sqlEditor.IsExpanded() {
		a.sqlEditor.DecreaseHeight()
	}
	return a, nil

case "[":
	// Previous result tab
	if a.resultTabs.HasTabs() && !a.sqlEditorFocused {
		a.resultTabs.PrevTab()
	}
	return a, nil

case "]":
	// Next result tab
	if a.resultTabs.HasTabs() && !a.sqlEditorFocused {
		a.resultTabs.NextTab()
	}
	return a, nil
```

**Step 2: Update Tab key for focus cycling to include SQL editor**

Find the existing Tab handling and update:

```go
case "tab":
	if a.sqlEditorFocused {
		// From SQL editor to sidebar
		a.sqlEditorFocused = false
		a.state.FocusedPanel = models.LeftPanel
	} else if a.state.FocusedPanel == models.LeftPanel {
		// From sidebar to data panel
		a.state.FocusedPanel = models.RightPanel
	} else {
		// From data panel to SQL editor (if expanded) or sidebar
		if a.sqlEditor.IsExpanded() {
			a.sqlEditorFocused = true
		} else {
			a.state.FocusedPanel = models.LeftPanel
		}
	}
	a.updatePanelStyles()
	return a, nil
```

**Step 3: Route keyboard input to SQL editor when focused**

Add this near the beginning of the Update method's tea.KeyMsg handling:

```go
case tea.KeyMsg:
	// If SQL editor is focused and expanded, route input there
	if a.sqlEditorFocused && a.sqlEditor.IsExpanded() {
		// Handle escape to unfocus
		if msg.String() == "esc" {
			a.sqlEditorFocused = false
			a.sqlEditor.Collapse()
			a.state.FocusedPanel = models.RightPanel
			return a, nil
		}

		// Handle ctrl+e to collapse
		if msg.String() == "ctrl+e" {
			a.sqlEditor.Collapse()
			a.sqlEditorFocused = false
			return a, nil
		}

		// Route to SQL editor
		_, cmd := a.sqlEditor.Update(msg)
		return a, cmd
	}

	// ... rest of existing keyboard handling
```

**Step 4: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): add keyboard handling for sql editor"
```

---

## Task 10: Update App - Query Result Handling

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Update QueryResultMsg handling to use result tabs**

Find the `QueryResultMsg` case in Update() and update it:

```go
case QueryResultMsg:
	// Record query to history
	if a.historyStore != nil {
		connName := ""
		dbName := ""
		if a.state.ActiveConnection != nil {
			connName = a.state.ActiveConnection.Config.Name
			dbName = a.state.ActiveConnection.Config.Database
		}

		entry := history.HistoryEntry{
			ConnectionName: connName,
			DatabaseName:   dbName,
			Query:          msg.SQL,
			Duration:       msg.Result.Duration,
			RowsAffected:   msg.Result.RowsAffected,
			Success:        msg.Result.Error == nil,
		}

		if msg.Result.Error != nil {
			entry.ErrorMessage = msg.Result.Error.Error()
		}

		_ = a.historyStore.Add(entry)
	}

	// Handle query result
	if msg.Result.Error != nil {
		a.ShowError("Query Error", msg.Result.Error.Error())
		return a, nil
	}

	// Add result to tabs
	a.resultTabs.AddResult(msg.SQL, msg.Result)

	// Collapse editor and focus data panel
	a.sqlEditor.Collapse()
	a.sqlEditorFocused = false
	a.state.FocusedPanel = models.RightPanel
	a.updatePanelStyles()

	return a, nil
```

**Step 2: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): update query result handling for tabs"
```

---

## Task 11: Remove QuickQuery References

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Remove all quickQuery references**

Search for and remove/replace:
- Remove `showQuickQuery bool` field from App struct
- Remove `quickQuery *components.QuickQuery` field from App struct
- Remove quickQuery initialization in New()
- Remove quickQuery handling in Update()
- Remove quickQuery rendering in View()
- Update `commands.QuickQueryCommandMsg` to toggle SQL editor instead

**Step 2: Update QuickQueryCommandMsg handler**

```go
case commands.QuickQueryCommandMsg:
	// Open SQL editor
	a.sqlEditor.Expand()
	a.sqlEditorFocused = true
	return a, nil
```

**Step 3: Remove quickQuery rendering in View()**

Remove this section from renderNormalView():
```go
// If quick query is showing, replace bottom bar with it
if a.showQuickQuery {
	// ... remove this entire block
}
```

**Step 4: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds with no errors

**Step 5: Commit**

```bash
git add internal/app/app.go
git commit -m "refactor(app): remove quickQuery, use sql editor"
```

---

## Task 12: External Editor Support

**Files:**
- Modify: `internal/ui/components/sql_editor.go`
- Modify: `internal/app/app.go`

**Step 1: Add external editor message type**

In `sql_editor.go`:

```go
// OpenExternalEditorMsg requests opening an external editor
type OpenExternalEditorMsg struct {
	Content string
}

// ExternalEditorResultMsg contains the result from external editor
type ExternalEditorResultMsg struct {
	Content string
	Error   error
}
```

**Step 2: Add Ctrl+O handling in SQL editor Update**

In `sql_editor.go` Update method:

```go
case "ctrl+o":
	return e, func() tea.Msg {
		return OpenExternalEditorMsg{Content: e.GetContent()}
	}
```

**Step 3: Add external editor handling in app.go**

```go
import (
	"os"
	"os/exec"
	// ... other imports
)

// In Update():
case components.OpenExternalEditorMsg:
	// Open external editor
	return a, a.openExternalEditor(msg.Content)

case components.ExternalEditorResultMsg:
	if msg.Error != nil {
		a.ShowError("Editor Error", msg.Error.Error())
		return a, nil
	}
	a.sqlEditor.SetContent(msg.Content)
	return a, nil

// Add this method:
func (a *App) openExternalEditor(content string) tea.Cmd {
	return func() tea.Msg {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		// Create temp file
		tmpFile, err := os.CreateTemp("", "lazypg-*.sql")
		if err != nil {
			return components.ExternalEditorResultMsg{Error: err}
		}
		defer os.Remove(tmpFile.Name())

		// Write content
		if _, err := tmpFile.WriteString(content); err != nil {
			return components.ExternalEditorResultMsg{Error: err}
		}
		tmpFile.Close()

		// Open editor
		cmd := exec.Command(editor, tmpFile.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return components.ExternalEditorResultMsg{Error: err}
		}

		// Read result
		result, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			return components.ExternalEditorResultMsg{Error: err}
		}

		return components.ExternalEditorResultMsg{Content: string(result)}
	}
}
```

**Step 4: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add internal/ui/components/sql_editor.go internal/app/app.go
git commit -m "feat(sql-editor): add external editor support"
```

---

## Task 13: Update Bottom Status Bar

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Update bottom bar to show SQL editor shortcuts**

In `renderNormalView()`, update the `bottomBarLeft` section to include SQL editor hints:

```go
// Update bottom bar hints based on focus
var bottomBarLeft string
if a.sqlEditorFocused {
	// SQL editor mode
	bottomBarLeft = keyStyle.Render("Ctrl+Enter") + dimStyle.Render(" execute") +
		separatorStyle.Render(" │ ") +
		keyStyle.Render("Ctrl+O") + dimStyle.Render(" editor") +
		separatorStyle.Render(" │ ") +
		keyStyle.Render("Esc") + dimStyle.Render(" close")
} else if a.state.FocusedPanel == models.LeftPanel {
	// Tree navigation keys
	bottomBarLeft = keyStyle.Render("↑↓") + dimStyle.Render(" navigate") +
		separatorStyle.Render(" │ ") +
		keyStyle.Render("→←") + dimStyle.Render(" expand") +
		separatorStyle.Render(" │ ") +
		keyStyle.Render("Enter") + dimStyle.Render(" select")
} else {
	// Data panel
	bottomBarLeft = keyStyle.Render("↑↓") + dimStyle.Render(" navigate") +
		separatorStyle.Render(" │ ") +
		keyStyle.Render("[]") + dimStyle.Render(" tabs") +
		separatorStyle.Render(" │ ") +
		keyStyle.Render("Ctrl+E") + dimStyle.Render(" sql")
}
```

**Step 2: Run go build to verify**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): update status bar for sql editor"
```

---

## Task 14: Final Testing and Polish

**Files:**
- All modified files

**Step 1: Run full build**

Run: `cd /Users/rebeliceyang/Github/lazypg && go build ./...`
Expected: Build succeeds with no errors

**Step 2: Run application and test manually**

Run: `cd /Users/rebeliceyang/Github/lazypg && go run .`

Test checklist:
- [ ] Ctrl+E opens/closes SQL editor
- [ ] Typing in editor works with syntax highlighting
- [ ] Line numbers display correctly
- [ ] Ctrl+Enter executes query
- [ ] Result appears in new tab
- [ ] [ and ] switch between tabs
- [ ] Ctrl+Shift+Up/Down changes editor height
- [ ] Tab cycles focus between panels
- [ ] History navigation with Ctrl+Up/Down
- [ ] Ctrl+O opens external editor

**Step 3: Fix any issues found**

**Step 4: Final commit**

```bash
git add .
git commit -m "feat: complete multiline sql editor implementation"
```

---

## Summary

This implementation plan covers:

1. **Tasks 1-5**: SQL Editor component with text editing, syntax highlighting, and view rendering
2. **Task 6**: Result Tabs component for managing multiple query results
3. **Tasks 7-11**: App integration with layout changes, keyboard handling, and removal of QuickQuery
4. **Task 12**: External editor support
5. **Task 13-14**: Polish and testing

Each task is designed to be independently testable and commits at logical boundaries.
