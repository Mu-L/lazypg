# Data Panel Enhancements Design

Date: 2025-11-25

## Overview

This document describes three enhancements to the data panel: column sorting, horizontal scrolling, and data search.

---

## Feature 1: Column Sorting

### Interaction

| Key | Action |
|-----|--------|
| `s` | Sort by current column (toggle ASC/DESC) |
| `S` | Toggle NULL position (NULLS FIRST/LAST) |

Workflow:
1. Use `h/l` to select a column
2. Press `s` to sort ascending
3. Press `s` again to toggle to descending
4. Press `S` to toggle NULL position

### Visual Indicators

- Column header shows sort direction: `name ↑` (ASC) or `name ↓` (DESC)
- When NULLS FIRST is active: `name ↑ⁿ` or `name ↓ⁿ`

### Technical Implementation

**New fields in TableView:**
```go
type TableView struct {
    // ... existing fields
    SortColumn    int    // -1 means no sort
    SortDirection string // "ASC" or "DESC"
    NullsFirst    bool   // true = NULLS FIRST, false = NULLS LAST (default)
}
```

**SQL modification in metadata/data.go:**
```go
func QueryTableData(..., sortCol string, sortDir string, nullsFirst bool) {
    query := fmt.Sprintf("SELECT * FROM %s.%s", schema, table)
    if sortCol != "" {
        nullsClause := "NULLS LAST"
        if nullsFirst {
            nullsClause = "NULLS FIRST"
        }
        query += fmt.Sprintf(" ORDER BY %s %s %s", sortCol, sortDir, nullsClause)
    }
    query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
}
```

**Behavior:**
- Sort is server-side (ORDER BY in SQL)
- Changing sort reloads data from page 1
- Sort state persists until table changes

---

## Feature 2: Horizontal Scrolling

### Interaction

| Key | Action |
|-----|--------|
| `h/l` | Select column + auto-scroll to keep it visible |
| `H/L` | Jump scroll half screen horizontally |
| `0` | Jump to first column |
| `$` | Jump to last column |

### Visual Indicators

**Edge indicators:**
- `◀` on left edge when more columns exist to the left
- `▶` on right edge when more columns exist to the right

**Status bar:**
```
Cols 1-5 of 12 │ 1-100 of 5000 rows
```

### Technical Implementation

**New fields in TableView:**
```go
type TableView struct {
    // ... existing fields
    LeftColOffset int // First visible column index
    VisibleCols   int // Number of columns that fit in current width
}
```

**Rendering logic:**
```go
func (tv *TableView) View() string {
    // Calculate visible columns based on width
    tv.calculateVisibleCols()

    // Only render columns from LeftColOffset to LeftColOffset+VisibleCols
    visibleColumns := tv.Columns[tv.LeftColOffset : tv.LeftColOffset+tv.VisibleCols]

    // Add edge indicators
    leftIndicator := ""
    rightIndicator := ""
    if tv.LeftColOffset > 0 {
        leftIndicator = "◀ "
    }
    if tv.LeftColOffset+tv.VisibleCols < len(tv.Columns) {
        rightIndicator = " ▶"
    }
}
```

**Auto-scroll on selection:**
```go
func (tv *TableView) MoveSelectionHorizontal(delta int) {
    tv.SelectedCol += delta
    // Clamp to bounds
    tv.SelectedCol = clamp(tv.SelectedCol, 0, len(tv.Columns)-1)

    // Auto-scroll to keep selected column visible
    if tv.SelectedCol < tv.LeftColOffset {
        tv.LeftColOffset = tv.SelectedCol
    }
    if tv.SelectedCol >= tv.LeftColOffset+tv.VisibleCols {
        tv.LeftColOffset = tv.SelectedCol - tv.VisibleCols + 1
    }
}
```

---

## Feature 3: Data Search

### Interaction

| Key | Action |
|-----|--------|
| `/` or `Ctrl+F` | Open search box |
| `Tab` | Toggle Local/Table search mode |
| `Enter` | Execute search |
| `n` | Jump to next match |
| `N` | Jump to previous match |
| `Esc` | Close search / clear search |

### Two Search Modes

**Local Search (default):**
- Searches only loaded data in memory
- Instant results
- String contains matching (case-insensitive)

**Table Search:**
- Sends SQL query to database
- Searches entire table
- Uses `col::text ILIKE '%keyword%'` for all columns

### Visual Indicators

**Search box:**
```
┌─ Search [Local] ─────────────────┐
│ 🔍 keyword                       │
└──────────────────────────────────┘
```

**Match highlighting:**
- Current match: bright background color
- Other matches: subtle background color

**Status bar:**
```
Match 3 of 12 │ Cols 1-5 of 12 │ 1-100 of 5000 rows
```

### Technical Implementation

**New fields in TableView:**
```go
type TableView struct {
    // ... existing fields
    SearchMode    string      // "local" or "table"
    SearchQuery   string
    Matches       []MatchPos  // List of match positions
    CurrentMatch  int         // Index in Matches
    SearchActive  bool
}

type MatchPos struct {
    Row int
    Col int
}
```

**Local search:**
```go
func (tv *TableView) searchLocal(query string) {
    tv.Matches = nil
    query = strings.ToLower(query)
    for rowIdx, row := range tv.Rows {
        for colIdx, cell := range row {
            if strings.Contains(strings.ToLower(cell), query) {
                tv.Matches = append(tv.Matches, MatchPos{Row: rowIdx, Col: colIdx})
            }
        }
    }
    if len(tv.Matches) > 0 {
        tv.jumpToMatch(0)
    }
}
```

**Table search SQL:**
```go
func buildSearchQuery(schema, table string, columns []string, keyword string) string {
    var conditions []string
    for _, col := range columns {
        conditions = append(conditions, fmt.Sprintf("%s::text ILIKE '%%%s%%'", col, keyword))
    }
    return fmt.Sprintf("SELECT * FROM %s.%s WHERE %s",
        schema, table, strings.Join(conditions, " OR "))
}
```

---

## Implementation Order

1. **Column Sorting** - Most essential feature for data browsing
2. **Horizontal Scrolling** - Improves column navigation experience
3. **Data Search** - Adds powerful data discovery capability

## Files to Modify

| File | Changes |
|------|---------|
| `internal/ui/components/table_view.go` | Add sort state, scroll offset, search state, rendering logic |
| `internal/db/metadata/data.go` | Add ORDER BY support to QueryTableData |
| `internal/app/app.go` | Handle new key bindings, search messages |
| `internal/ui/help/help.go` | Document new keybindings |

## Key Bindings Summary

| Key | Action |
|-----|--------|
| `s` | Sort current column |
| `S` | Toggle NULLS FIRST/LAST |
| `h/l` | Select column + auto-scroll |
| `H/L` | Jump scroll half screen |
| `0` | First column |
| `$` | Last column |
| `/`, `Ctrl+F` | Open search |
| `n/N` | Next/previous match |
