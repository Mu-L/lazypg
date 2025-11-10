# Phase 4: Command Palette & Query Implementation Plan

**Date:** 2025-01-10
**Status:** Ready for Implementation
**Estimated Time:** 2-3 weeks

## Overview

Implement the command palette (Ctrl+K/Cmd+K) and query execution system (quick query + full editor) as described in the design document.

## Prerequisites

- Phase 1-3 functionality working
- Bubble Tea foundation in place
- Database connection established

## Implementation Tasks

### Task 1: Command Palette Core Structure

**Goal:** Create the basic command palette modal overlay

**Files to Create/Modify:**
- `internal/ui/components/command_palette.go` (new)
- `internal/models/command.go` (new)
- `internal/app/app.go` (modify)

**Steps:**

1. Create command model structure:
```go
// internal/models/command.go
package models

type CommandType int

const (
    CommandTypeAction CommandType = iota
    CommandTypeObject
    CommandTypeHistory
    CommandTypeFavorite
)

type Command struct {
    ID          string
    Type        CommandType
    Label       string
    Description string
    Icon        string
    Tags        []string
    Score       int // For ranking
    Action      func() tea.Msg
}
```

2. Create command palette component:
```go
// internal/ui/components/command_palette.go
package components

type CommandPalette struct {
    Input        string
    Commands     []models.Command
    Filtered     []models.Command
    Selected     int
    Width        int
    Height       int
    Theme        theme.Theme
    Mode         string // "", ">", "?", "@", "#"
}

func NewCommandPalette(theme theme.Theme) *CommandPalette
func (cp *CommandPalette) Update(msg tea.KeyMsg) (*CommandPalette, tea.Cmd)
func (cp *CommandPalette) View() string
func (cp *CommandPalette) Filter()
```

3. Add state to App:
```go
// internal/app/app.go - add fields
showCommandPalette bool
commandPalette     *components.CommandPalette
```

4. Add keyboard handler in app.go Update():
```go
case "ctrl+k":
    a.showCommandPalette = true
    return a, nil
```

5. Add view rendering in app.go View():
```go
if a.showCommandPalette {
    return lipgloss.Place(
        a.state.Width, a.state.Height,
        lipgloss.Center, lipgloss.Center,
        a.commandPalette.View(),
    )
}
```

**Verification:**
```bash
go build -o bin/lazypg ./cmd/lazypg
# Run app, press Ctrl+K, should see command palette overlay
# Press Esc, should close
```

---

### Task 2: Command Registry and Built-in Commands

**Goal:** Create a registry of available commands

**Files to Create/Modify:**
- `internal/commands/registry.go` (new)
- `internal/commands/builtin.go` (new)

**Steps:**

1. Create command registry:
```go
// internal/commands/registry.go
package commands

import "github.com/rebeliceyang/lazypg/internal/models"

type Registry struct {
    commands map[string]models.Command
}

func NewRegistry() *Registry
func (r *Registry) Register(cmd models.Command)
func (r *Registry) Get(id string) (models.Command, bool)
func (r *Registry) GetAll() []models.Command
func (r *Registry) Search(query string) []models.Command
```

2. Create built-in commands:
```go
// internal/commands/builtin.go
package commands

func GetBuiltinCommands() []models.Command {
    return []models.Command{
        {
            ID:          "connect",
            Type:        models.CommandTypeAction,
            Label:       "Connect to Database",
            Description: "Open connection dialog",
            Icon:        "🔌",
            Tags:        []string{"connection", "database"},
        },
        {
            ID:          "disconnect",
            Type:        models.CommandTypeAction,
            Label:       "Disconnect",
            Description: "Close current connection",
            Icon:        "🔴",
            Tags:        []string{"connection"},
        },
        {
            ID:          "refresh",
            Type:        models.CommandTypeAction,
            Label:       "Refresh",
            Description: "Refresh current view",
            Icon:        "🔄",
            Tags:        []string{"view"},
        },
        {
            ID:          "query",
            Type:        models.CommandTypeAction,
            Label:       "Quick Query",
            Description: "Execute a quick SQL query",
            Icon:        "⚡",
            Tags:        []string{"query", "sql"},
        },
        {
            ID:          "editor",
            Type:        models.CommandTypeAction,
            Label:       "Query Editor",
            Description: "Open full query editor",
            Icon:        "📝",
            Tags:        []string{"query", "sql", "editor"},
        },
        {
            ID:          "history",
            Type:        models.CommandTypeAction,
            Label:       "Query History",
            Description: "View query history",
            Icon:        "📜",
            Tags:        []string{"query", "history"},
        },
        {
            ID:          "help",
            Type:        models.CommandTypeAction,
            Label:       "Help",
            Description: "Show keyboard shortcuts",
            Icon:        "❓",
            Tags:        []string{"help"},
        },
        {
            ID:          "settings",
            Type:        models.CommandTypeAction,
            Label:       "Settings",
            Description: "Open settings",
            Icon:        "⚙️",
            Tags:        []string{"config", "settings"},
        },
    }
}
```

3. Initialize registry in app.go:
```go
// internal/app/app.go - add field
commandRegistry *commands.Registry

// In New():
app.commandRegistry = commands.NewRegistry()
for _, cmd := range commands.GetBuiltinCommands() {
    app.commandRegistry.Register(cmd)
}
app.commandPalette = components.NewCommandPalette(th, app.commandRegistry)
```

**Verification:**
```bash
go build -o bin/lazypg ./cmd/lazypg
# Press Ctrl+K
# Type "conn" - should filter to "Connect to Database"
# Type "query" - should show query-related commands
```

---

### Task 3: Fuzzy Search and Ranking

**Goal:** Implement smart fuzzy search with ranking

**Files to Create/Modify:**
- `internal/search/fuzzy.go` (new)
- `internal/ui/components/command_palette.go` (modify)

**Steps:**

1. Create fuzzy search utility:
```go
// internal/search/fuzzy.go
package search

type Match struct {
    Score    int
    Indices  []int
    Matched  bool
}

func FuzzyMatch(query, target string) Match
func RankMatches(query string, targets []string) []int
```

2. Implement simple fuzzy matching algorithm:
```go
func FuzzyMatch(query, target string) Match {
    // Convert to lowercase
    query = strings.ToLower(query)
    target = strings.ToLower(target)

    // Simple substring matching with scoring
    if idx := strings.Index(target, query); idx >= 0 {
        score := 100 - idx // Earlier matches score higher
        return Match{Score: score, Matched: true}
    }

    // Character-by-character fuzzy match
    queryIdx := 0
    score := 0
    indices := []int{}

    for i, ch := range target {
        if queryIdx < len(query) && rune(query[queryIdx]) == ch {
            indices = append(indices, i)
            score += 10
            queryIdx++
        }
    }

    if queryIdx == len(query) {
        return Match{Score: score, Indices: indices, Matched: true}
    }

    return Match{Matched: false}
}
```

3. Update command palette filtering:
```go
// internal/ui/components/command_palette.go
func (cp *CommandPalette) Filter() {
    if cp.Input == "" {
        cp.Filtered = cp.Commands
        return
    }

    filtered := []models.Command{}
    for _, cmd := range cp.Commands {
        match := search.FuzzyMatch(cp.Input, cmd.Label)
        if !match.Matched {
            match = search.FuzzyMatch(cp.Input, cmd.Description)
        }
        if !match.Matched {
            for _, tag := range cmd.Tags {
                match = search.FuzzyMatch(cp.Input, tag)
                if match.Matched {
                    break
                }
            }
        }

        if match.Matched {
            cmd.Score = match.Score
            filtered = append(filtered, cmd)
        }
    }

    // Sort by score
    sort.Slice(filtered, func(i, j int) bool {
        return filtered[i].Score > filtered[j].Score
    })

    cp.Filtered = filtered
    cp.Selected = 0
}
```

**Verification:**
```bash
go build -o bin/lazypg ./cmd/lazypg
# Press Ctrl+K
# Type "qry" - should match "Quick Query" and "Query Editor"
# Type "hlp" - should match "Help"
# Results should be ranked by relevance
```

---

### Task 4: Command Palette UI Polish

**Goal:** Create beautiful command palette UI with proper styling

**Files to Modify:**
- `internal/ui/components/command_palette.go`

**Steps:**

1. Implement View() method with modern styling:
```go
func (cp *CommandPalette) View() string {
    // Input box
    inputStyle := lipgloss.NewStyle().
        Foreground(cp.Theme.Foreground).
        Background(cp.Theme.Selection).
        Padding(0, 1).
        Width(cp.Width - 4)

    prefix := lipgloss.NewStyle().
        Foreground(cp.Theme.BorderFocused).
        Bold(true).
        Render("> ")

    input := inputStyle.Render(prefix + cp.Input + "█")

    // Results list
    maxResults := 8
    results := []string{}

    for i, cmd := range cp.Filtered {
        if i >= maxResults {
            break
        }

        style := lipgloss.NewStyle().
            Foreground(cp.Theme.Foreground).
            Padding(0, 1).
            Width(cp.Width - 4)

        if i == cp.Selected {
            style = style.
                Background(cp.Theme.BorderFocused).
                Foreground(cp.Theme.Background).
                Bold(true)
        }

        icon := lipgloss.NewStyle().
            Foreground(cp.Theme.Info).
            Render(cmd.Icon + " ")

        label := lipgloss.NewStyle().
            Bold(i == cp.Selected).
            Render(cmd.Label)

        desc := lipgloss.NewStyle().
            Foreground(cp.Theme.Metadata).
            Render(" - " + cmd.Description)

        line := style.Render(icon + label + desc)
        results = append(results, line)
    }

    // Empty state
    if len(cp.Filtered) == 0 {
        emptyStyle := lipgloss.NewStyle().
            Foreground(cp.Theme.Comment).
            Italic(true).
            Padding(2, 0).
            Width(cp.Width - 4).
            Align(lipgloss.Center)
        results = append(results, emptyStyle.Render("No commands found"))
    }

    // Combine
    content := lipgloss.JoinVertical(
        lipgloss.Left,
        input,
        lipgloss.NewStyle().
            Foreground(cp.Theme.Border).
            Render(strings.Repeat("─", cp.Width-4)),
        lipgloss.JoinVertical(lipgloss.Left, results...),
    )

    // Box
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(cp.Theme.BorderFocused).
        Padding(1, 2).
        Width(cp.Width).
        MaxHeight(cp.Height)

    return boxStyle.Render(content)
}
```

2. Add navigation in Update():
```go
func (cp *CommandPalette) Update(msg tea.KeyMsg) (*CommandPalette, tea.Cmd) {
    switch msg.String() {
    case "up", "ctrl+p":
        if cp.Selected > 0 {
            cp.Selected--
        }
    case "down", "ctrl+n":
        if cp.Selected < len(cp.Filtered)-1 {
            cp.Selected++
        }
    case "enter":
        if cp.Selected < len(cp.Filtered) {
            cmd := cp.Filtered[cp.Selected]
            if cmd.Action != nil {
                return cp, cmd.Action
            }
        }
    case "esc", "ctrl+c":
        return cp, func() tea.Msg {
            return CloseCommandPaletteMsg{}
        }
    case "backspace":
        if len(cp.Input) > 0 {
            cp.Input = cp.Input[:len(cp.Input)-1]
            cp.Filter()
        }
    default:
        key := msg.String()
        if len(key) == 1 {
            cp.Input += key
            cp.Filter()
        }
    }
    return cp, nil
}
```

**Verification:**
```bash
go build -o bin/lazypg ./cmd/lazypg
# Press Ctrl+K - beautiful styled command palette appears
# Type text - see live filtering
# Use ↑↓ - navigate results with highlight
# Press Enter - command should execute
# Press Esc - closes palette
```

---

### Task 5: Quick Query Mode

**Goal:** Implement quick query input (Ctrl+P) at bottom of screen

**Files to Create/Modify:**
- `internal/ui/components/quick_query.go` (new)
- `internal/app/app.go` (modify)

**Steps:**

1. Create quick query component:
```go
// internal/ui/components/quick_query.go
package components

type QuickQuery struct {
    Input    string
    Width    int
    Theme    theme.Theme
    History  []string
    HistIdx  int
}

func NewQuickQuery(theme theme.Theme) *QuickQuery
func (qq *QuickQuery) Update(msg tea.KeyMsg) (*QuickQuery, tea.Cmd)
func (qq *QuickQuery) View() string
func (qq *QuickQuery) Execute() tea.Cmd
```

2. Implement View():
```go
func (qq *QuickQuery) View() string {
    prefix := lipgloss.NewStyle().
        Foreground(qq.Theme.Info).
        Bold(true).
        Render("SQL> ")

    inputStyle := lipgloss.NewStyle().
        Foreground(qq.Theme.Foreground).
        Background(qq.Theme.Selection).
        Padding(0, 1).
        Width(qq.Width - 10)

    cursor := "█"
    input := inputStyle.Render(qq.Input + cursor)

    hint := lipgloss.NewStyle().
        Foreground(qq.Theme.Comment).
        Italic(true).
        Render(" [Enter: Execute | Ctrl+E: Full Editor | Esc: Cancel]")

    return prefix + input + hint
}
```

3. Add to app.go:
```go
// Add fields
showQuickQuery bool
quickQuery     *components.QuickQuery

// Initialize
app.quickQuery = components.NewQuickQuery(th)

// Handler
case "ctrl+p":
    a.showQuickQuery = true
    return a, nil

// In View() at bottom
if a.showQuickQuery {
    return lipgloss.JoinVertical(
        lipgloss.Left,
        mainView,
        a.quickQuery.View(),
    )
}
```

**Verification:**
```bash
go build -o bin/lazypg ./cmd/lazypg
# Press Ctrl+P - quick query bar appears at bottom
# Type "SELECT 1" - see text
# Press Esc - closes
```

---

### Task 6: Query Execution Engine

**Goal:** Execute SQL queries and display results

**Files to Create/Modify:**
- `internal/db/query/executor.go` (new)
- `internal/models/query_result.go` (new)

**Steps:**

1. Create query result model:
```go
// internal/models/query_result.go
package models

import "time"

type QueryResult struct {
    Columns      []string
    Rows         [][]string
    RowsAffected int64
    Duration     time.Duration
    Error        error
}
```

2. Create query executor:
```go
// internal/db/query/executor.go
package query

import (
    "context"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rebeliceyang/lazypg/internal/models"
)

func Execute(ctx context.Context, pool *pgxpool.Pool, sql string) models.QueryResult {
    start := time.Now()

    rows, err := pool.Query(ctx, sql)
    if err != nil {
        return models.QueryResult{Error: err, Duration: time.Since(start)}
    }
    defer rows.Close()

    // Get columns
    fieldDescs := rows.FieldDescriptions()
    columns := make([]string, len(fieldDescs))
    for i, fd := range fieldDescs {
        columns[i] = string(fd.Name)
    }

    // Get rows
    var result [][]string
    for rows.Next() {
        values, err := rows.Values()
        if err != nil {
            return models.QueryResult{Error: err, Duration: time.Since(start)}
        }

        row := make([]string, len(values))
        for i, v := range values {
            if v == nil {
                row[i] = "NULL"
            } else {
                row[i] = fmt.Sprintf("%v", v)
            }
        }
        result = append(result, row)
    }

    return models.QueryResult{
        Columns:      columns,
        Rows:         result,
        RowsAffected: int64(len(result)),
        Duration:     time.Since(start),
    }
}
```

3. Add execute message and handler:
```go
// internal/app/app.go
type ExecuteQueryMsg struct {
    SQL string
}

type QueryResultMsg struct {
    Result models.QueryResult
}

// Handler
case ExecuteQueryMsg:
    return a, func() tea.Msg {
        conn, err := a.connectionManager.GetActive()
        if err != nil {
            return QueryResultMsg{
                Result: models.QueryResult{
                    Error: fmt.Errorf("no active connection: %w", err),
                },
            }
        }

        result := query.Execute(context.Background(), conn.Pool, msg.SQL)
        return QueryResultMsg{Result: result}
    }

case QueryResultMsg:
    if msg.Result.Error != nil {
        a.ShowError("Query Error", msg.Result.Error.Error())
        a.showQuickQuery = false
        return a, nil
    }

    // Display results in table view
    a.tableView.SetData(msg.Result.Columns, msg.Result.Rows, len(msg.Result.Rows))
    a.state.FocusedPanel = models.RightPanel
    a.updatePanelStyles()
    a.showQuickQuery = false

    return a, nil
```

**Verification:**
```bash
go build -o bin/lazypg ./cmd/lazypg
# Connect to database
# Press Ctrl+P
# Type "SELECT * FROM pg_tables LIMIT 5"
# Press Enter
# Should see results in table view
# Try "SELECT 1 + 1" - should work
# Try invalid SQL - should show error overlay
```

---

### Task 7: Query History Storage

**Goal:** Store query history in SQLite

**Files to Create/Modify:**
- `internal/history/store.go` (new)
- `internal/history/schema.sql` (new)

**Steps:**

1. Create schema:
```sql
-- internal/history/schema.sql
CREATE TABLE IF NOT EXISTS query_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    connection_name TEXT,
    database_name TEXT,
    query TEXT NOT NULL,
    executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    duration_ms INTEGER,
    rows_affected INTEGER,
    success BOOLEAN NOT NULL,
    error_message TEXT
);

CREATE INDEX idx_executed_at ON query_history(executed_at DESC);
CREATE INDEX idx_connection ON query_history(connection_name);
```

2. Create history store:
```go
// internal/history/store.go
package history

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "time"
)

type Store struct {
    db *sql.DB
}

func NewStore(path string) (*Store, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, err
    }

    // Create schema
    _, err = db.Exec(schemaSQL)
    if err != nil {
        return nil, err
    }

    return &Store{db: db}, nil
}

func (s *Store) Add(entry HistoryEntry) error {
    _, err := s.db.Exec(`
        INSERT INTO query_history
        (connection_name, database_name, query, duration_ms, rows_affected, success, error_message)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
        entry.ConnectionName, entry.DatabaseName, entry.Query,
        entry.Duration.Milliseconds(), entry.RowsAffected,
        entry.Success, entry.ErrorMessage,
    )
    return err
}

func (s *Store) GetRecent(limit int) ([]HistoryEntry, error) {
    rows, err := s.db.Query(`
        SELECT id, connection_name, database_name, query, executed_at,
               duration_ms, rows_affected, success, error_message
        FROM query_history
        ORDER BY executed_at DESC
        LIMIT ?`, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var entries []HistoryEntry
    for rows.Next() {
        var e HistoryEntry
        var durationMs int64
        err := rows.Scan(&e.ID, &e.ConnectionName, &e.DatabaseName, &e.Query,
            &e.ExecutedAt, &durationMs, &e.RowsAffected, &e.Success, &e.ErrorMessage)
        if err != nil {
            return nil, err
        }
        e.Duration = time.Duration(durationMs) * time.Millisecond
        entries = append(entries, e)
    }

    return entries, nil
}
```

3. Integrate with query execution:
```go
// internal/app/app.go
historyStore *history.Store

// Initialize in New()
historyPath := filepath.Join(os.Getenv("HOME"), ".config", "lazypg", "history.db")
store, err := history.NewStore(historyPath)
if err != nil {
    log.Printf("Warning: Could not open history: %v", err)
} else {
    app.historyStore = store
}

// After query execution
if a.historyStore != nil {
    entry := history.HistoryEntry{
        ConnectionName: "current", // TODO: get from connection
        DatabaseName:   conn.Config.Database,
        Query:          msg.SQL,
        Duration:       msg.Result.Duration,
        RowsAffected:   msg.Result.RowsAffected,
        Success:        msg.Result.Error == nil,
    }
    if msg.Result.Error != nil {
        entry.ErrorMessage = msg.Result.Error.Error()
    }
    a.historyStore.Add(entry)
}
```

**Verification:**
```bash
go build -o bin/lazypg ./cmd/lazypg
# Execute several queries via Ctrl+P
# Check ~/.config/lazypg/history.db exists
sqlite3 ~/.config/lazypg/history.db "SELECT * FROM query_history;"
# Should see all executed queries
```

---

### Task 8: Command Palette Integration with History

**Goal:** Show recent queries in command palette

**Files to Modify:**
- `internal/ui/components/command_palette.go`
- `internal/app/app.go`

**Steps:**

1. Add history commands dynamically:
```go
// internal/app/app.go
func (a *App) updateCommandPalette() {
    // Start with builtin commands
    commands := a.commandRegistry.GetAll()

    // Add recent queries from history
    if a.historyStore != nil {
        recent, err := a.historyStore.GetRecent(10)
        if err == nil {
            for _, entry := range recent {
                cmd := models.Command{
                    ID:          fmt.Sprintf("history:%d", entry.ID),
                    Type:        models.CommandTypeHistory,
                    Label:       truncate(entry.Query, 60),
                    Description: fmt.Sprintf("%s ago", timeAgo(entry.ExecutedAt)),
                    Icon:        "💾",
                    Tags:        []string{"history", "query"},
                    Action: func() tea.Msg {
                        return ExecuteQueryMsg{SQL: entry.Query}
                    },
                }
                commands = append(commands, cmd)
            }
        }
    }

    a.commandPalette.SetCommands(commands)
}

// Call before showing palette
case "ctrl+k":
    a.updateCommandPalette()
    a.showCommandPalette = true
```

**Verification:**
```bash
go build -o bin/lazypg ./cmd/lazypg
# Execute a few queries
# Press Ctrl+K
# Type "select" - should see recent SELECT queries
# Select one and press Enter - should re-execute
```

---

## Success Criteria

Phase 4 is complete when:

- ✅ Ctrl+K opens command palette with fuzzy search
- ✅ Built-in commands (connect, disconnect, refresh, query, help) work
- ✅ Command palette has beautiful UI with icons and descriptions
- ✅ Ctrl+P opens quick query input at bottom
- ✅ Queries can be executed and results displayed in table view
- ✅ Query errors show in error overlay
- ✅ Query history is persisted to SQLite
- ✅ Recent queries appear in command palette
- ✅ Can re-execute queries from history
- ✅ No crashes or data loss

## Testing Checklist

- [ ] Command palette opens/closes smoothly
- [ ] Fuzzy search finds commands by name, description, tags
- [ ] Navigation (↑↓) works in command palette
- [ ] Quick query accepts SQL and executes
- [ ] Query results display correctly
- [ ] Query errors are handled gracefully
- [ ] History is persisted across sessions
- [ ] Recent queries appear in command palette
- [ ] Can re-execute from history
- [ ] Performance is acceptable (< 100ms response)

## Notes

- Query editor (full multi-line editor) is deferred to next iteration
- Syntax highlighting will be added in polish phase
- Autocomplete will be added in polish phase
- For now, focus on core functionality

