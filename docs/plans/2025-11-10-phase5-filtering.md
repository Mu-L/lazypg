# Phase 5: Interactive Filtering Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build an interactive filter builder that generates SQL WHERE clauses with type-aware operators, quick filtering from cells, and live preview.

**Architecture:** Modal-based filter UI with operator selection, value input, and SQL generation engine. Filters are composable with AND/OR logic and can be saved/loaded. The system analyzes column types to provide appropriate operators (=, >, LIKE, @>, etc.).

**Tech Stack:** Bubble Tea, Lipgloss, pgx v5 for type introspection, SQL builder pattern

---

## Task 1: Filter Model and SQL Generator

**Files:**
- Create: `internal/models/filter.go`
- Create: `internal/filter/builder.go`
- Test: Manual testing (TDD not practical for SQL generation logic)

**Step 1: Create filter models**

Create `internal/models/filter.go`:

```go
package models

import "time"

// FilterOperator represents a filter comparison operator
type FilterOperator string

const (
	OpEqual          FilterOperator = "="
	OpNotEqual       FilterOperator = "!="
	OpGreaterThan    FilterOperator = ">"
	OpGreaterOrEqual FilterOperator = ">="
	OpLessThan       FilterOperator = "<"
	OpLessOrEqual    FilterOperator = "<="
	OpLike           FilterOperator = "LIKE"
	OpILike          FilterOperator = "ILIKE"
	OpIn             FilterOperator = "IN"
	OpNotIn          FilterOperator = "NOT IN"
	OpIsNull         FilterOperator = "IS NULL"
	OpIsNotNull      FilterOperator = "IS NOT NULL"
	OpContains       FilterOperator = "@>"  // JSONB contains
	OpContainedBy    FilterOperator = "<@"  // JSONB contained by
	OpHasKey         FilterOperator = "?"   // JSONB has key
	OpArrayOverlap   FilterOperator = "&&"  // Array overlap
)

// FilterCondition represents a single filter condition
type FilterCondition struct {
	Column   string
	Operator FilterOperator
	Value    interface{}
	Type     string // PostgreSQL type (text, integer, jsonb, etc.)
}

// FilterGroup represents a group of conditions with AND/OR logic
type FilterGroup struct {
	Conditions []FilterCondition
	Logic      string // "AND" or "OR"
	Groups     []FilterGroup
}

// Filter represents the complete filter state
type Filter struct {
	RootGroup FilterGroup
	TableName string
	Schema    string
}

// ColumnInfo represents column metadata for type-aware filtering
type ColumnInfo struct {
	Name     string
	DataType string
	IsArray  bool
	IsJsonb  bool
}
```

**Step 2: Create SQL builder**

Create `internal/filter/builder.go`:

```go
package filter

import (
	"fmt"
	"strings"

	"github.com/rebeliceyang/lazypg/internal/models"
)

// Builder generates SQL WHERE clauses from Filter models
type Builder struct{}

// NewBuilder creates a new filter builder
func NewBuilder() *Builder {
	return &Builder{}
}

// BuildWhere generates a WHERE clause from a Filter
func (b *Builder) BuildWhere(filter models.Filter) (string, []interface{}, error) {
	if len(filter.RootGroup.Conditions) == 0 && len(filter.RootGroup.Groups) == 0 {
		return "", nil, nil
	}

	clause, args, err := b.buildGroup(filter.RootGroup, 1)
	if err != nil {
		return "", nil, err
	}

	return "WHERE " + clause, args, nil
}

// buildGroup recursively builds a filter group
func (b *Builder) buildGroup(group models.FilterGroup, paramIndex int) (string, []interface{}, error) {
	var clauses []string
	var args []interface{}
	currentParam := paramIndex

	// Build conditions
	for _, cond := range group.Conditions {
		clause, condArgs, err := b.buildCondition(cond, currentParam)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, clause)
		args = append(args, condArgs...)
		currentParam += len(condArgs)
	}

	// Build sub-groups
	for _, subGroup := range group.Groups {
		clause, groupArgs, err := b.buildGroup(subGroup, currentParam)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, "("+clause+")")
		args = append(args, groupArgs...)
		currentParam += len(groupArgs)
	}

	logic := group.Logic
	if logic == "" {
		logic = "AND"
	}

	return strings.Join(clauses, " "+logic+" "), args, nil
}

// buildCondition builds a single filter condition
func (b *Builder) buildCondition(cond models.FilterCondition, paramIndex int) (string, []interface{}, error) {
	column := cond.Column

	switch cond.Operator {
	case models.OpIsNull:
		return fmt.Sprintf("%s IS NULL", column), nil, nil
	case models.OpIsNotNull:
		return fmt.Sprintf("%s IS NOT NULL", column), nil, nil
	case models.OpEqual, models.OpNotEqual, models.OpGreaterThan, models.OpGreaterOrEqual,
		models.OpLessThan, models.OpLessOrEqual:
		return fmt.Sprintf("%s %s $%d", column, cond.Operator, paramIndex), []interface{}{cond.Value}, nil
	case models.OpLike, models.OpILike:
		return fmt.Sprintf("%s %s $%d", column, cond.Operator, paramIndex), []interface{}{cond.Value}, nil
	case models.OpIn, models.OpNotIn:
		// For IN/NOT IN, value should be a slice
		return fmt.Sprintf("%s %s ($%d)", column, cond.Operator, paramIndex), []interface{}{cond.Value}, nil
	case models.OpContains, models.OpContainedBy, models.OpHasKey:
		// JSONB operators
		return fmt.Sprintf("%s %s $%d", column, cond.Operator, paramIndex), []interface{}{cond.Value}, nil
	case models.OpArrayOverlap:
		return fmt.Sprintf("%s && $%d", column, paramIndex), []interface{}{cond.Value}, nil
	default:
		return "", nil, fmt.Errorf("unsupported operator: %s", cond.Operator)
	}
}

// GetOperatorsForType returns available operators for a given PostgreSQL type
func GetOperatorsForType(dataType string) []models.FilterOperator {
	switch {
	case strings.Contains(dataType, "int") || strings.Contains(dataType, "numeric") ||
		strings.Contains(dataType, "real") || strings.Contains(dataType, "double"):
		return []models.FilterOperator{
			models.OpEqual, models.OpNotEqual,
			models.OpGreaterThan, models.OpGreaterOrEqual,
			models.OpLessThan, models.OpLessOrEqual,
			models.OpIsNull, models.OpIsNotNull,
		}
	case strings.Contains(dataType, "char") || strings.Contains(dataType, "text"):
		return []models.FilterOperator{
			models.OpEqual, models.OpNotEqual,
			models.OpLike, models.OpILike,
			models.OpIsNull, models.OpIsNotNull,
		}
	case strings.Contains(dataType, "jsonb"):
		return []models.FilterOperator{
			models.OpEqual, models.OpNotEqual,
			models.OpContains, models.OpContainedBy, models.OpHasKey,
			models.OpIsNull, models.OpIsNotNull,
		}
	case strings.Contains(dataType, "ARRAY"):
		return []models.FilterOperator{
			models.OpEqual, models.OpNotEqual,
			models.OpArrayOverlap, models.OpContains, models.OpContainedBy,
			models.OpIsNull, models.OpIsNotNull,
		}
	case strings.Contains(dataType, "bool"):
		return []models.FilterOperator{
			models.OpEqual, models.OpNotEqual,
			models.OpIsNull, models.OpIsNotNull,
		}
	case strings.Contains(dataType, "date") || strings.Contains(dataType, "time"):
		return []models.FilterOperator{
			models.OpEqual, models.OpNotEqual,
			models.OpGreaterThan, models.OpGreaterOrEqual,
			models.OpLessThan, models.OpLessOrEqual,
			models.OpIsNull, models.OpIsNotNull,
		}
	default:
		return []models.FilterOperator{
			models.OpEqual, models.OpNotEqual,
			models.OpIsNull, models.OpIsNotNull,
		}
	}
}
```

**Step 3: Build and verify**

Run: `go build ./...`
Expected: Clean build with no errors

**Step 4: Commit**

```bash
git add internal/models/filter.go internal/filter/builder.go
git commit -m "feat(filter): add filter models and SQL builder"
```

---

## Task 2: Filter UI Component

**Files:**
- Create: `internal/ui/components/filter_builder.go`

**Step 1: Create filter builder component**

Create `internal/ui/components/filter_builder.go`:

```go
package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rebeliceyang/lazypg/internal/filter"
	"github.com/rebeliceyang/lazypg/internal/models"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// ApplyFilterMsg is sent when a filter should be applied
type ApplyFilterMsg struct {
	Filter models.Filter
}

// CloseFilterBuilderMsg is sent when the filter builder should close
type CloseFilterBuilderMsg struct{}

// FilterBuilder provides an interactive UI for building SQL filters
type FilterBuilder struct {
	Width   int
	Height  int
	Theme   theme.Theme
	builder *filter.Builder

	// State
	columns       []models.ColumnInfo
	filter        models.Filter
	currentIndex  int // Index in conditions list
	editMode      string // "", "column", "operator", "value"
	columnInput   string
	operatorIndex int
	valueInput    string

	// UI elements
	selectedColumn   models.ColumnInfo
	availableOps     []models.FilterOperator
	previewSQL       string
}

// NewFilterBuilder creates a new filter builder
func NewFilterBuilder(th theme.Theme) *FilterBuilder {
	return &FilterBuilder{
		Width:   80,
		Height:  30,
		Theme:   th,
		builder: filter.NewBuilder(),
		filter: models.Filter{
			RootGroup: models.FilterGroup{
				Conditions: []models.FilterCondition{},
				Logic:      "AND",
				Groups:     []models.FilterGroup{},
			},
		},
		editMode: "",
	}
}

// SetColumns updates the available columns for filtering
func (fb *FilterBuilder) SetColumns(columns []models.ColumnInfo) {
	fb.columns = columns
}

// SetTable sets the table being filtered
func (fb *FilterBuilder) SetTable(schema, table string) {
	fb.filter.Schema = schema
	fb.filter.TableName = table
}

// Update handles keyboard input
func (fb *FilterBuilder) Update(msg tea.KeyMsg) (*FilterBuilder, tea.Cmd) {
	switch fb.editMode {
	case "":
		return fb.handleNavigationMode(msg)
	case "column":
		return fb.handleColumnMode(msg)
	case "operator":
		return fb.handleOperatorMode(msg)
	case "value":
		return fb.handleValueMode(msg)
	}
	return fb, nil
}

// handleNavigationMode handles keys in navigation mode
func (fb *FilterBuilder) handleNavigationMode(msg tea.KeyMsg) (*FilterBuilder, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if fb.currentIndex > 0 {
			fb.currentIndex--
		}
	case "down", "j":
		if fb.currentIndex < len(fb.filter.RootGroup.Conditions) {
			fb.currentIndex++
		}
	case "a", "n":
		// Add new condition
		fb.editMode = "column"
		fb.columnInput = ""
	case "d", "x":
		// Delete current condition
		if fb.currentIndex < len(fb.filter.RootGroup.Conditions) {
			fb.filter.RootGroup.Conditions = append(
				fb.filter.RootGroup.Conditions[:fb.currentIndex],
				fb.filter.RootGroup.Conditions[fb.currentIndex+1:]...,
			)
			if fb.currentIndex > 0 && fb.currentIndex >= len(fb.filter.RootGroup.Conditions) {
				fb.currentIndex--
			}
			fb.updatePreview()
		}
	case "enter":
		// Apply filter
		return fb, func() tea.Msg {
			return ApplyFilterMsg{Filter: fb.filter}
		}
	case "esc":
		return fb, func() tea.Msg {
			return CloseFilterBuilderMsg{}
		}
	}
	return fb, nil
}

// handleColumnMode handles column selection
func (fb *FilterBuilder) handleColumnMode(msg tea.KeyMsg) (*FilterBuilder, tea.Cmd) {
	switch msg.String() {
	case "esc":
		fb.editMode = ""
		fb.columnInput = ""
	case "enter":
		// Find matching column
		for _, col := range fb.columns {
			if strings.EqualFold(col.Name, fb.columnInput) {
				fb.selectedColumn = col
				fb.availableOps = filter.GetOperatorsForType(col.DataType)
				fb.editMode = "operator"
				fb.operatorIndex = 0
				return fb, nil
			}
		}
		// No match, stay in column mode
	case "backspace":
		if len(fb.columnInput) > 0 {
			fb.columnInput = fb.columnInput[:len(fb.columnInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			fb.columnInput += msg.String()
		}
	}
	return fb, nil
}

// handleOperatorMode handles operator selection
func (fb *FilterBuilder) handleOperatorMode(msg tea.KeyMsg) (*FilterBuilder, tea.Cmd) {
	switch msg.String() {
	case "esc":
		fb.editMode = "column"
	case "up", "k":
		if fb.operatorIndex > 0 {
			fb.operatorIndex--
		}
	case "down", "j":
		if fb.operatorIndex < len(fb.availableOps)-1 {
			fb.operatorIndex++
		}
	case "enter":
		// Check if operator needs a value
		selectedOp := fb.availableOps[fb.operatorIndex]
		if selectedOp == models.OpIsNull || selectedOp == models.OpIsNotNull {
			// No value needed, add condition immediately
			fb.filter.RootGroup.Conditions = append(fb.filter.RootGroup.Conditions, models.FilterCondition{
				Column:   fb.selectedColumn.Name,
				Operator: selectedOp,
				Value:    nil,
				Type:     fb.selectedColumn.DataType,
			})
			fb.editMode = ""
			fb.updatePreview()
		} else {
			fb.editMode = "value"
			fb.valueInput = ""
		}
	}
	return fb, nil
}

// handleValueMode handles value input
func (fb *FilterBuilder) handleValueMode(msg tea.KeyMsg) (*FilterBuilder, tea.Cmd) {
	switch msg.String() {
	case "esc":
		fb.editMode = "operator"
		fb.valueInput = ""
	case "enter":
		// Add condition
		fb.filter.RootGroup.Conditions = append(fb.filter.RootGroup.Conditions, models.FilterCondition{
			Column:   fb.selectedColumn.Name,
			Operator: fb.availableOps[fb.operatorIndex],
			Value:    fb.valueInput,
			Type:     fb.selectedColumn.DataType,
		})
		fb.editMode = ""
		fb.valueInput = ""
		fb.updatePreview()
	case "backspace":
		if len(fb.valueInput) > 0 {
			fb.valueInput = fb.valueInput[:len(fb.valueInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			fb.valueInput += msg.String()
		}
	}
	return fb, nil
}

// updatePreview updates the SQL preview
func (fb *FilterBuilder) updatePreview() {
	whereClause, _, err := fb.builder.BuildWhere(fb.filter)
	if err != nil {
		fb.previewSQL = fmt.Sprintf("Error: %s", err.Error())
	} else {
		if whereClause == "" {
			fb.previewSQL = "SELECT * FROM " + fb.filter.TableName
		} else {
			fb.previewSQL = fmt.Sprintf("SELECT * FROM %s %s", fb.filter.TableName, whereClause)
		}
	}
}

// View renders the filter builder
func (fb *FilterBuilder) View() string {
	var sections []string

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(fb.Theme.Text).
		Background(fb.Theme.Primary).
		Padding(0, 1).
		Bold(true)
	sections = append(sections, titleStyle.Render("Filter Builder"))

	// Instructions based on mode
	instructionStyle := lipgloss.NewStyle().
		Foreground(fb.Theme.SubtleText).
		Padding(0, 1)

	var instructions string
	switch fb.editMode {
	case "column":
		instructions = "Type column name, Enter to confirm, Esc to cancel"
	case "operator":
		instructions = "↑↓ Select operator, Enter to confirm, Esc to go back"
	case "value":
		instructions = "Type value, Enter to confirm, Esc to go back"
	default:
		instructions = "a=Add n=New d=Delete Enter=Apply Esc=Cancel"
	}
	sections = append(sections, instructionStyle.Render(instructions))

	// Conditions list
	if len(fb.filter.RootGroup.Conditions) > 0 {
		sections = append(sections, "\nConditions:")
		for i, cond := range fb.filter.RootGroup.Conditions {
			condStr := fmt.Sprintf("%s %s %v", cond.Column, cond.Operator, cond.Value)
			if cond.Operator == models.OpIsNull || cond.Operator == models.OpIsNotNull {
				condStr = fmt.Sprintf("%s %s", cond.Column, cond.Operator)
			}

			style := lipgloss.NewStyle().Padding(0, 1)
			if i == fb.currentIndex && fb.editMode == "" {
				style = style.Background(fb.Theme.SelectedBackground).Foreground(fb.Theme.SelectedText)
			}
			sections = append(sections, style.Render(fmt.Sprintf(" %d. %s", i+1, condStr)))
		}
	}

	// Edit area
	if fb.editMode != "" {
		sections = append(sections, "\n")
		switch fb.editMode {
		case "column":
			sections = append(sections, fmt.Sprintf("Column: %s_", fb.columnInput))
		case "operator":
			sections = append(sections, fmt.Sprintf("Column: %s", fb.selectedColumn.Name))
			sections = append(sections, "Select operator:")
			for i, op := range fb.availableOps {
				style := lipgloss.NewStyle().Padding(0, 1)
				if i == fb.operatorIndex {
					style = style.Background(fb.Theme.SelectedBackground).Foreground(fb.Theme.SelectedText)
				}
				sections = append(sections, style.Render(fmt.Sprintf("  %s", op)))
			}
		case "value":
			sections = append(sections, fmt.Sprintf("Column: %s %s", fb.selectedColumn.Name, fb.availableOps[fb.operatorIndex]))
			sections = append(sections, fmt.Sprintf("Value: %s_", fb.valueInput))
		}
	}

	// SQL Preview
	if fb.previewSQL != "" {
		sections = append(sections, "\nSQL Preview:")
		previewStyle := lipgloss.NewStyle().
			Foreground(fb.Theme.SubtleText).
			Background(fb.Theme.Background).
			Padding(0, 1).
			Italic(true)
		sections = append(sections, previewStyle.Render(fb.previewSQL))
	}

	content := strings.Join(sections, "\n")

	// Container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(fb.Theme.Border).
		Background(fb.Theme.Background).
		Foreground(fb.Theme.Text).
		Width(fb.Width).
		Height(fb.Height).
		Padding(1)

	return containerStyle.Render(content)
}
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: Clean build with no errors

**Step 3: Commit**

```bash
git add internal/ui/components/filter_builder.go
git commit -m "feat(filter): add interactive filter builder UI component"
```

---

## Task 3: Integrate Filter Builder with App

**Files:**
- Modify: `internal/app/app.go:27-65` (Add filter builder fields)
- Modify: `internal/app/app.go:110-180` (Initialize filter builder)
- Modify: `internal/app/app.go:190-497` (Handle filter messages)

**Step 1: Add filter builder to App struct**

In `internal/app/app.go`, add to the App struct (around line 60):

```go
// Filter builder
showFilterBuilder bool
filterBuilder     *components.FilterBuilder
activeFilter      *models.Filter
```

**Step 2: Initialize filter builder in New()**

In the `New()` function (around line 160), add:

```go
// Initialize filter builder
filterBuilder := components.NewFilterBuilder(th)
```

And add to the App struct initialization:

```go
showFilterBuilder: false,
filterBuilder:     filterBuilder,
activeFilter:      nil,
```

**Step 3: Add filter keyboard shortcut**

In the `Update()` method, add a new case for opening the filter builder (around line 340):

```go
case "f":
	// Open filter builder if on table view
	if a.state.FocusedPanel == models.RightPanel && a.state.CurrentNode != nil {
		if a.state.CurrentNode.Type == models.NodeTypeTable {
			// Get column info from current table
			columns := a.getTableColumns()
			a.filterBuilder.SetColumns(columns)
			a.filterBuilder.SetTable(a.state.CurrentNode.Schema, a.state.CurrentNode.Name)
			a.showFilterBuilder = true
		}
	}
	return a, nil
```

**Step 4: Handle filter builder messages**

Add message handling in `Update()` (around line 290):

```go
case components.ApplyFilterMsg:
	// Apply the filter and reload table data
	a.showFilterBuilder = false
	a.activeFilter = &msg.Filter

	// Reload table with filter
	if a.state.CurrentNode != nil && a.state.CurrentNode.Type == models.NodeTypeTable {
		return a, a.loadTableDataWithFilter(*a.activeFilter)
	}
	return a, nil

case components.CloseFilterBuilderMsg:
	a.showFilterBuilder = false
	return a, nil
```

**Step 5: Handle filter builder input when visible**

In the `Update()` method, add handling before normal key handling (around line 320):

```go
// Handle filter builder input
if a.showFilterBuilder {
	return a.handleFilterBuilder(msg)
}
```

**Step 6: Add handleFilterBuilder method**

Add this method after `handleQuickQuery()` (around line 940):

```go
// handleFilterBuilder handles key events when filter builder is visible
func (a *App) handleFilterBuilder(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.filterBuilder, cmd = a.filterBuilder.Update(msg)
	return a, cmd
}
```

**Step 7: Add getTableColumns helper method**

Add this method after `handleFilterBuilder()`:

```go
// getTableColumns returns column info for the current table
func (a *App) getTableColumns() []models.ColumnInfo {
	if a.state.CurrentNode == nil || a.state.CurrentNode.Type != models.NodeTypeTable {
		return nil
	}

	// Extract columns from table view
	columns := []models.ColumnInfo{}

	// Get column names from the current table view
	// This is a placeholder - actual implementation depends on how you store column metadata
	// For now, we'll return columns from the tableView
	if len(a.tableView.Headers) > 0 {
		for _, header := range a.tableView.Headers {
			columns = append(columns, models.ColumnInfo{
				Name:     header,
				DataType: "text", // Default type, should be enhanced with actual type info
				IsArray:  false,
				IsJsonb:  false,
			})
		}
	}

	return columns
}
```

**Step 8: Render filter builder in View()**

In the `View()` method, add filter builder rendering (around line 520, after error overlay):

```go
// Render filter builder if visible
if a.showFilterBuilder {
	view = lipgloss.Place(
		a.state.Width,
		a.state.Height,
		lipgloss.Center,
		lipgloss.Center,
		a.filterBuilder.View(),
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#555555")),
	)
}
```

**Step 9: Build and verify**

Run: `go build ./...`
Expected: Clean build with no errors

**Step 10: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(filter): integrate filter builder with main app"
```

---

## Task 4: Load Table Data with Filters

**Files:**
- Create: `internal/db/metadata/columns.go`
- Modify: `internal/app/app.go:1037-1060` (Update loadTableData)

**Step 1: Create column metadata query**

Create `internal/db/metadata/columns.go`:

```go
package metadata

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rebeliceyang/lazypg/internal/models"
)

// GetTableColumns retrieves column metadata for a table
func GetTableColumns(ctx context.Context, pool *pgxpool.Pool, schema, table string) ([]models.ColumnInfo, error) {
	query := `
		SELECT
			column_name,
			data_type,
			udt_name,
			CASE WHEN data_type = 'ARRAY' THEN true ELSE false END as is_array
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`

	rows, err := pool.Query(ctx, query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	defer rows.Close()

	var columns []models.ColumnInfo
	for rows.Next() {
		var col models.ColumnInfo
		var udtName string
		err := rows.Scan(&col.Name, &col.DataType, &udtName, &col.IsArray)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		// Check if it's JSONB
		col.IsJsonb = udtName == "jsonb"

		columns = append(columns, col)
	}

	return columns, rows.Err()
}
```

**Step 2: Update getTableColumns to use metadata query**

In `internal/app/app.go`, update the `getTableColumns()` method:

```go
// getTableColumns returns column info for the current table
func (a *App) getTableColumns() []models.ColumnInfo {
	if a.state.CurrentNode == nil || a.state.CurrentNode.Type != models.NodeTypeTable {
		return nil
	}

	conn, err := a.connectionManager.GetActive()
	if err != nil {
		return nil
	}

	columns, err := metadata.GetTableColumns(
		context.Background(),
		conn.Pool.GetPool(),
		a.state.CurrentNode.Schema,
		a.state.CurrentNode.Name,
	)
	if err != nil {
		return nil
	}

	return columns
}
```

**Step 3: Add loadTableDataWithFilter method**

In `internal/app/app.go`, add this method after `loadTableData()`:

```go
// loadTableDataWithFilter loads table data with an applied filter
func (a *App) loadTableDataWithFilter(filter models.Filter) tea.Cmd {
	return func() tea.Msg {
		conn, err := a.connectionManager.GetActive()
		if err != nil {
			return ErrorMsg{Title: "Connection Error", Message: err.Error()}
		}

		node := a.state.CurrentNode
		if node == nil || node.Type != models.NodeTypeTable {
			return ErrorMsg{Title: "Error", Message: "No table selected"}
		}

		// Build filtered query
		builder := filter2.NewBuilder()
		whereClause, args, err := builder.BuildWhere(filter)
		if err != nil {
			return ErrorMsg{Title: "Filter Error", Message: err.Error()}
		}

		// Construct query
		query := fmt.Sprintf(
			"SELECT * FROM %s.%s %s LIMIT 100",
			node.Schema,
			node.Name,
			whereClause,
		)

		// Execute query
		result, err := conn.Pool.QueryWithColumns(context.Background(), query, args...)
		if err != nil {
			return ErrorMsg{Title: "Query Error", Message: err.Error()}
		}

		// Convert to string rows for display
		var rows [][]string
		for _, row := range result.Rows {
			var strRow []string
			for _, col := range result.Columns {
				val := row[col]
				if val == nil {
					strRow = append(strRow, "NULL")
				} else {
					strRow = append(strRow, fmt.Sprintf("%v", val))
				}
			}
			rows = append(rows, strRow)
		}

		return TableDataLoadedMsg{
			Columns: result.Columns,
			Rows:    rows,
			Total:   len(rows),
			Offset:  0,
		}
	}
}
```

Note: You'll need to import the filter package with an alias to avoid conflict:
```go
import (
	filter2 "github.com/rebeliceyang/lazypg/internal/filter"
)
```

**Step 4: Build and verify**

Run: `go build ./...`
Expected: Clean build with no errors

**Step 5: Commit**

```bash
git add internal/db/metadata/columns.go internal/app/app.go
git commit -m "feat(filter): add filtered table data loading"
```

---

## Task 5: Quick Filter from Cell

**Files:**
- Modify: `internal/app/app.go:350-380` (Add quick filter shortcut)

**Step 1: Add quick filter keyboard shortcut**

In `internal/app/app.go`, in the key handling section (around line 370), add:

```go
case "ctrl+f":
	// Quick filter from current cell
	if a.state.FocusedPanel == models.RightPanel && a.tableView != nil {
		selectedRow, selectedCol := a.tableView.GetSelectedCell()
		if selectedRow >= 0 && selectedCol >= 0 && selectedCol < len(a.tableView.Headers) {
			// Get column name and value
			columnName := a.tableView.Headers[selectedCol]
			cellValue := a.tableView.Data[selectedRow][selectedCol]

			// Create quick filter
			columns := a.getTableColumns()
			var columnInfo models.ColumnInfo
			for _, col := range columns {
				if col.Name == columnName {
					columnInfo = col
					break
				}
			}

			// Create filter with single condition
			quickFilter := models.Filter{
				Schema:    a.state.CurrentNode.Schema,
				TableName: a.state.CurrentNode.Name,
				RootGroup: models.FilterGroup{
					Conditions: []models.FilterCondition{
						{
							Column:   columnName,
							Operator: models.OpEqual,
							Value:    cellValue,
							Type:     columnInfo.DataType,
						},
					},
					Logic: "AND",
				},
			}

			a.activeFilter = &quickFilter
			return a, a.loadTableDataWithFilter(quickFilter)
		}
	}
	return a, nil
```

**Step 2: Add GetSelectedCell method to TableView**

In `internal/ui/components/table_view.go`, add this method:

```go
// GetSelectedCell returns the currently selected row and column indices
func (tv *TableView) GetSelectedCell() (row int, col int) {
	return tv.Selected, tv.SelectedCol
}
```

If TableView doesn't have a SelectedCol field, add it to the struct:

```go
type TableView struct {
	// ... existing fields ...
	SelectedCol int // Currently selected column
}
```

And update the keyboard navigation to track column selection (left/right keys).

**Step 3: Build and verify**

Run: `go build ./...`
Expected: Clean build with no errors

**Step 4: Commit**

```bash
git add internal/app/app.go internal/ui/components/table_view.go
git commit -m "feat(filter): add quick filter from cell with Ctrl+F"
```

---

## Task 6: Clear Filter & Status Display

**Files:**
- Modify: `internal/app/app.go:390-400` (Add clear filter shortcut)
- Modify: `internal/app/app.go:550-650` (Show filter status)

**Step 1: Add clear filter shortcut**

In `internal/app/app.go`, add keyboard shortcut (around line 390):

```go
case "ctrl+r":
	// Clear filter and reload
	if a.activeFilter != nil && a.state.CurrentNode != nil {
		a.activeFilter = nil
		return a, a.loadTableData(LoadTableDataMsg{
			Schema: a.state.CurrentNode.Schema,
			Table:  a.state.CurrentNode.Name,
			Limit:  100,
			Offset: 0,
		})
	}
	return a, nil
```

**Step 2: Show filter status in bottom bar**

In `internal/app/app.go`, update the bottom status bar rendering (around line 590):

Find where the status bar is rendered and add filter information:

```go
// Build status line
statusParts := []string{}

// Add filter indicator if active
if a.activeFilter != nil && len(a.activeFilter.RootGroup.Conditions) > 0 {
	filterCount := len(a.activeFilter.RootGroup.Conditions)
	filterIndicator := fmt.Sprintf("🔍 %d filter%s active", filterCount, pluralize(filterCount))
	statusParts = append(statusParts, filterIndicator)
}

// ... existing status parts ...

// Helper function for pluralization
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
```

**Step 3: Update help text to show filter shortcuts**

In `internal/ui/help/help.go`, add filter-related shortcuts:

```go
{"f", "Open filter builder"},
{"Ctrl+F", "Quick filter from cell"},
{"Ctrl+R", "Clear filter"},
```

**Step 4: Build and verify**

Run: `go build ./...`
Expected: Clean build with no errors

**Step 5: Commit**

```bash
git add internal/app/app.go internal/ui/help/help.go
git commit -m "feat(filter): add clear filter and status display"
```

---

## Task 7: Filter Preview and Validation

**Files:**
- Modify: `internal/filter/builder.go:100-120` (Add validation)
- Modify: `internal/ui/components/filter_builder.go:200-220` (Show validation errors)

**Step 1: Add filter validation**

In `internal/filter/builder.go`, add validation method:

```go
// Validate checks if a filter is valid
func (b *Builder) Validate(filter models.Filter) error {
	if filter.TableName == "" {
		return fmt.Errorf("table name is required")
	}

	return b.validateGroup(filter.RootGroup)
}

// validateGroup validates a filter group
func (b *Builder) validateGroup(group models.FilterGroup) error {
	for _, cond := range group.Conditions {
		if err := b.validateCondition(cond); err != nil {
			return err
		}
	}

	for _, subGroup := range group.Groups {
		if err := b.validateGroup(subGroup); err != nil {
			return err
		}
	}

	return nil
}

// validateCondition validates a single condition
func (b *Builder) validateCondition(cond models.FilterCondition) error {
	if cond.Column == "" {
		return fmt.Errorf("column name is required")
	}

	// Check if value is required for operator
	requiresValue := cond.Operator != models.OpIsNull && cond.Operator != models.OpIsNotNull
	if requiresValue && cond.Value == nil {
		return fmt.Errorf("value is required for operator %s", cond.Operator)
	}

	return nil
}
```

**Step 2: Add validation error display to filter builder**

In `internal/ui/components/filter_builder.go`, add error field to struct:

```go
type FilterBuilder struct {
	// ... existing fields ...
	validationError string
}
```

Update the `handleNavigationMode` to validate before applying:

```go
case "enter":
	// Validate and apply filter
	if err := fb.builder.Validate(fb.filter); err != nil {
		fb.validationError = err.Error()
		return fb, nil
	}

	fb.validationError = ""
	return fb, func() tea.Msg {
		return ApplyFilterMsg{Filter: fb.filter}
	}
```

Add error display in `View()` method:

```go
// Show validation error if present
if fb.validationError != "" {
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555")).
		Padding(0, 1).
		Bold(true)
	sections = append(sections, errorStyle.Render("⚠ "+fb.validationError))
}
```

**Step 3: Build and verify**

Run: `go build ./...`
Expected: Clean build with no errors

**Step 4: Commit**

```bash
git add internal/filter/builder.go internal/ui/components/filter_builder.go
git commit -m "feat(filter): add validation and error display"
```

---

## Task 8: Final Testing and Documentation

**Files:**
- Create: `docs/features/filtering.md`
- Modify: `README.md` (Add filtering section)

**Step 1: Create filtering documentation**

Create `docs/features/filtering.md`:

```markdown
# Interactive Filtering

lazypg provides a powerful interactive filter builder that generates SQL WHERE clauses.

## Opening the Filter Builder

- Press `f` while viewing a table to open the filter builder
- Press `Ctrl+F` on a cell to quickly filter by that cell's value

## Building Filters

1. Press `a` or `n` to add a new condition
2. Type the column name and press Enter
3. Select an operator using ↑↓ and press Enter
4. Type the value and press Enter
5. Repeat to add more conditions
6. Press Enter to apply the filter

## Operators by Type

### Numeric (int, numeric, real, double)
- `=`, `!=`, `>`, `>=`, `<`, `<=`
- `IS NULL`, `IS NOT NULL`

### Text (char, varchar, text)
- `=`, `!=`
- `LIKE`, `ILIKE` (case-insensitive)
- `IS NULL`, `IS NOT NULL`

### JSONB
- `=`, `!=`
- `@>` (contains)
- `<@` (contained by)
- `?` (has key)
- `IS NULL`, `IS NOT NULL`

### Arrays
- `=`, `!=`
- `&&` (overlap)
- `@>` (contains)
- `<@` (contained by)
- `IS NULL`, `IS NOT NULL`

### Boolean
- `=`, `!=`
- `IS NULL`, `IS NOT NULL`

### Date/Time
- `=`, `!=`, `>`, `>=`, `<`, `<=`
- `IS NULL`, `IS NOT NULL`

## Managing Filters

- Press `d` or `x` to delete a condition
- Press `Ctrl+R` to clear all filters
- Press `Esc` to close without applying

## SQL Preview

The filter builder shows a live SQL preview at the bottom, so you can see exactly what query will be executed.

## Multiple Conditions

All conditions are combined with AND logic. The filter will only show rows that match ALL conditions.

## Examples

### Find users named "Alice"
1. Press `f`
2. Press `a`
3. Type "name", Enter
4. Select `=`, Enter
5. Type "Alice", Enter
6. Press Enter to apply

### Find recent orders (last 7 days)
1. Press `f`
2. Press `a`
3. Type "created_at", Enter
4. Select `>`, Enter
5. Type "2024-01-01", Enter
6. Press Enter to apply

### Quick filter from cell
1. Navigate to a cell with Ctrl+F
2. Filter is automatically applied
```

**Step 2: Update README with filtering section**

In `README.md`, add a filtering section:

```markdown
## Filtering

Press `f` while viewing a table to open the interactive filter builder, or `Ctrl+F` to quickly filter by the selected cell.

See [docs/features/filtering.md](docs/features/filtering.md) for detailed documentation.
```

**Step 3: Manual testing checklist**

Test the following scenarios:

- [ ] Open filter builder with `f` key
- [ ] Add a condition with text column and `=` operator
- [ ] Add a condition with numeric column and `>` operator
- [ ] Delete a condition with `d` key
- [ ] Apply filter and verify data is filtered
- [ ] Clear filter with `Ctrl+R`
- [ ] Quick filter from cell with `Ctrl+F`
- [ ] Try invalid input (empty value) and verify error message
- [ ] Check SQL preview updates correctly
- [ ] Test with different column types (text, int, date)

**Step 4: Build final binary**

Run: `go build -o bin/lazypg ./cmd/lazypg`
Expected: Clean build with working binary

**Step 5: Final commit**

```bash
git add docs/features/filtering.md README.md
git commit -m "docs: add filtering documentation and complete Phase 5"
```

---

## Summary

Phase 5 Implementation adds:

1. **Filter Model & SQL Builder** - Type-safe filter representation and SQL generation
2. **Interactive Filter UI** - Modal-based filter builder with keyboard navigation
3. **Type-Aware Operators** - Appropriate operators based on PostgreSQL column types
4. **Quick Filter** - One-key filtering from any cell (Ctrl+F)
5. **Filter Status Display** - Visual indicator when filters are active
6. **SQL Preview** - Live preview of generated WHERE clause
7. **Validation** - Error checking before applying filters
8. **Clear Filter** - Easy removal of all filters (Ctrl+R)

**Keyboard Shortcuts:**
- `f` - Open filter builder
- `Ctrl+F` - Quick filter from selected cell
- `Ctrl+R` - Clear active filters
- `a`/`n` - Add new condition (in filter builder)
- `d`/`x` - Delete condition (in filter builder)
- `↑`/`↓` - Navigate operators/conditions
- `Enter` - Confirm/Apply
- `Esc` - Cancel/Close

**Total Files Created:** 4
**Total Files Modified:** 5
**Estimated Implementation Time:** 2 weeks
