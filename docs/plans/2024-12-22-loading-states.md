# Loading States Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add loading indicators to Tree and Table components for better user experience during data loading.

**Architecture:** Add loading state fields to TreeView and TableView components. Share the existing spinner from app.go via pointer. Handle loading state transitions in app.go when loading commands are issued and completed.

**Tech Stack:** Bubble Tea, Lip Gloss, bubbles/spinner

---

## Task 1: Add loading state fields to TreeView

**Files:**
- Modify: `internal/ui/components/tree_view.go:55-68`

**Step 1: Add loading fields to TreeView struct**

Add after line 67 (after `MatchPositions` field):

```go
	// Loading state
	IsLoading      bool           // True when initial tree is loading
	LoadingNodeID  string         // ID of node currently loading children (for inline spinner)
	LoadingStart   time.Time      // When loading started (for elapsed time)
```

**Step 2: Add spinner import**

Add to imports section:

```go
	"time"

	"github.com/charmbracelet/bubbles/spinner"
```

**Step 3: Add Spinner field to TreeView struct**

Add after LoadingStart:

```go
	Spinner        *spinner.Model // Shared spinner instance
```

**Step 4: Run tests to verify no compilation errors**

Run: `go build ./...`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add internal/ui/components/tree_view.go
git commit -m "feat(tree): add loading state fields to TreeView"
```

---

## Task 2: Add loading view for Tree initial connection

**Files:**
- Modify: `internal/ui/components/tree_view.go:172-211`

**Step 1: Add loadingState method to TreeView**

Add after `emptyState()` method (around line 755):

```go
// loadingState returns the loading state view for initial tree loading
func (tv *TreeView) loadingState() string {
	spinnerView := ""
	if tv.Spinner != nil {
		spinnerView = tv.Spinner.View() + " "
	}

	elapsed := time.Since(tv.LoadingStart)
	elapsedStr := fmt.Sprintf("(%.1fs)", elapsed.Seconds())

	loadingStyle := lipgloss.NewStyle().
		Foreground(tv.Theme.Foreground)

	elapsedStyle := lipgloss.NewStyle().
		Foreground(tv.Theme.Metadata)

	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		spinnerView+loadingStyle.Render("Loading databases...")+elapsedStyle.Render(" "+elapsedStr),
		"",
	)

	return lipgloss.Place(tv.Width, tv.Height, lipgloss.Center, lipgloss.Center, content)
}
```

**Step 2: Modify View() to check loading state first**

Modify `View()` method. Replace lines 172-175:

```go
func (tv *TreeView) View() string {
	// Show loading state for initial tree load
	if tv.IsLoading && tv.Root == nil {
		return tv.loadingState()
	}

	if tv.Root == nil {
		return tv.emptyState()
	}
```

**Step 3: Run tests**

Run: `go test ./internal/ui/components/... -v`
Expected: Tests pass

**Step 4: Commit**

```bash
git add internal/ui/components/tree_view.go
git commit -m "feat(tree): add loading view for initial tree connection"
```

---

## Task 3: Add inline loading for Tree node expansion

**Files:**
- Modify: `internal/ui/components/tree_view.go:397-448`

**Step 1: Add inlineLoadingNode method**

Add after `loadingState()` method:

```go
// inlineLoadingNode returns a loading indicator for a specific node
func (tv *TreeView) inlineLoadingNode(node *models.TreeNode) string {
	// Calculate indentation
	depth := node.GetDepth() - 1
	if depth < 0 {
		depth = 0
	}
	indent := strings.Repeat("  ", depth+1) // Extra indent for child position

	spinnerView := ""
	if tv.Spinner != nil {
		spinnerView = tv.Spinner.View()
	}

	loadingStyle := lipgloss.NewStyle().
		Foreground(tv.Theme.Comment).
		Italic(true)

	content := fmt.Sprintf("%s%s %s", indent, spinnerView, loadingStyle.Render("Loading..."))

	maxWidth := tv.Width - 2
	style := lipgloss.NewStyle().Width(maxWidth)
	return style.Render(content)
}
```

**Step 2: Modify View() to insert loading line after expanding node**

In the View() method, after rendering nodes (around line 250), add check for loading node:

Find this block:
```go
	for i := startIdx; i < endIdx; i++ {
		node := visibleNodes[i]
		line := tv.renderNode(node, i == tv.CursorIndex)
		// Wrap each row with zone mark for click detection
		// Use visible row index (i - startIdx) for zone ID
		zoneID := fmt.Sprintf("%s%d", ZoneTreeRowPrefix, i-startIdx)
		lines = append(lines, zone.Mark(zoneID, line))
	}
```

Replace with:
```go
	for i := startIdx; i < endIdx; i++ {
		node := visibleNodes[i]
		line := tv.renderNode(node, i == tv.CursorIndex)
		// Wrap each row with zone mark for click detection
		zoneID := fmt.Sprintf("%s%d", ZoneTreeRowPrefix, i-startIdx)
		lines = append(lines, zone.Mark(zoneID, line))

		// If this node is loading, show inline loading indicator after it
		if tv.LoadingNodeID != "" && node.ID == tv.LoadingNodeID {
			lines = append(lines, tv.inlineLoadingNode(node))
		}
	}
```

**Step 3: Run tests**

Run: `go test ./internal/ui/components/... -v`
Expected: Tests pass

**Step 4: Commit**

```bash
git add internal/ui/components/tree_view.go
git commit -m "feat(tree): add inline loading indicator for node expansion"
```

---

## Task 4: Add loading state fields to TableView

**Files:**
- Modify: `internal/ui/components/table_view.go:22-72`

**Step 1: Add loading fields to TableView struct**

Add after `PendingG` field (around line 68):

```go
	// Loading state
	IsLoading     bool           // True when first loading table data
	IsPaginating  bool           // True when loading more rows (pagination)
	LoadingStart  time.Time      // When loading started
	Spinner       *spinner.Model // Shared spinner instance
```

**Step 2: Add time import**

Verify `time` is already imported (line 7). If not, add it.

**Step 3: Add spinner import**

Add to imports:

```go
	"github.com/charmbracelet/bubbles/spinner"
```

**Step 4: Run build**

Run: `go build ./...`
Expected: Build succeeds

**Step 5: Commit**

```bash
git add internal/ui/components/table_view.go
git commit -m "feat(table): add loading state fields to TableView"
```

---

## Task 5: Add loading view for Table first load

**Files:**
- Modify: `internal/ui/components/table_view.go:331-429`

**Step 1: Add tableLoadingState method**

Add before `View()` method:

```go
// tableLoadingState returns the loading state view for initial table data load
func (tv *TableView) tableLoadingState(width, height int) string {
	spinnerView := ""
	if tv.Spinner != nil {
		spinnerView = tv.Spinner.View() + " "
	}

	elapsed := time.Since(tv.LoadingStart)
	elapsedStr := fmt.Sprintf("(%.1fs)", elapsed.Seconds())

	loadingStyle := lipgloss.NewStyle().
		Foreground(tv.Theme.Foreground)

	elapsedStyle := lipgloss.NewStyle().
		Foreground(tv.Theme.Metadata)

	cancelHint := lipgloss.NewStyle().
		Foreground(tv.Theme.Border).
		Render("Press Esc to cancel")

	content := lipgloss.JoinVertical(lipgloss.Center,
		"",
		spinnerView+loadingStyle.Render("Loading table data...")+elapsedStyle.Render(" "+elapsedStr),
		"",
		cancelHint,
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
```

**Step 2: Modify View() to check loading state**

In `View()` method, after calculating contentWidth/contentHeight (around line 341), add loading check:

Find:
```go
	contentWidth := tv.Width - containerStyle.GetHorizontalFrameSize()
	contentHeight := tv.Height - containerStyle.GetVerticalFrameSize()

	if len(tv.Columns) == 0 {
		return containerStyle.Width(contentWidth).Height(contentHeight).Render("No data")
	}
```

Replace with:
```go
	contentWidth := tv.Width - containerStyle.GetHorizontalFrameSize()
	contentHeight := tv.Height - containerStyle.GetVerticalFrameSize()

	// Show loading state for initial table load
	if tv.IsLoading {
		loadingContent := tv.tableLoadingState(contentWidth, contentHeight)
		return containerStyle.Width(contentWidth).Height(contentHeight).Render(loadingContent)
	}

	if len(tv.Columns) == 0 {
		return containerStyle.Width(contentWidth).Height(contentHeight).Render("No data")
	}
```

**Step 3: Run tests**

Run: `go test ./internal/ui/components/... -v`
Expected: Tests pass

**Step 4: Commit**

```bash
git add internal/ui/components/table_view.go
git commit -m "feat(table): add loading view for initial table data load"
```

---

## Task 6: Add pagination loading indicator to status bar

**Files:**
- Modify: `internal/ui/components/table_view.go:594-618`

**Step 1: Modify renderStatus() to show pagination spinner**

Replace `renderStatus()` method:

```go
func (tv *TableView) renderStatus() string {
	// Show pagination loading indicator
	if tv.IsPaginating {
		spinnerView := ""
		if tv.Spinner != nil {
			spinnerView = tv.Spinner.View() + " "
		}
		paginatingText := spinnerView + "Loading..."
		return tv.cachedStyles.status.Render(paginatingText)
	}

	endRow := tv.TopRow + len(tv.Rows)
	if endRow > tv.TotalRows {
		endRow = tv.TotalRows
	}

	// Search match info
	matchInfo := ""
	if tv.SearchActive && len(tv.Matches) > 0 {
		matchInfo = fmt.Sprintf("Match %d of %d │ ", tv.CurrentMatch+1, len(tv.Matches))
	}

	// Column info for horizontal scrolling
	colInfo := ""
	if len(tv.Columns) > tv.VisibleCols {
		endCol := tv.LeftColOffset + tv.VisibleCols
		if endCol > len(tv.Columns) {
			endCol = len(tv.Columns)
		}
		colInfo = fmt.Sprintf("Cols %d-%d of %d │ ", tv.LeftColOffset+1, endCol, len(tv.Columns))
	}

	showing := fmt.Sprintf(" 󰈙 %s%s%d-%d of %d rows", matchInfo, colInfo, tv.TopRow+1, endRow, tv.TotalRows)
	return tv.cachedStyles.status.Render(showing)
}
```

**Step 2: Run tests**

Run: `go test ./internal/ui/components/... -v`
Expected: Tests pass

**Step 3: Commit**

```bash
git add internal/ui/components/table_view.go
git commit -m "feat(table): add pagination loading indicator to status bar"
```

---

## Task 7: Wire up Tree loading states in App

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Pass spinner to TreeView on creation**

Find in `NewApp()` (around line 319):
```go
		treeView:          components.NewTreeView(emptyRoot, th),
```

After the app creation, add spinner assignment. Find (around line 341):
```go
	app := &App{
		...
		executeSpinner:    s,
```

After `app := &App{...}` block, add:
```go
	// Share spinner with TreeView
	app.treeView.Spinner = &app.executeSpinner
```

**Step 2: Set loading state when LoadTreeMsg is received**

Find case `LoadTreeMsg:` (around line 1426):
```go
	case LoadTreeMsg:
		return a, a.loadTree
```

Replace with:
```go
	case LoadTreeMsg:
		a.treeView.IsLoading = true
		a.treeView.LoadingStart = time.Now()
		return a, tea.Batch(a.loadTree, a.executeSpinner.Tick)
```

**Step 3: Clear loading state when TreeLoadedMsg is received**

Find case `TreeLoadedMsg:` (around line 1429):
```go
	case TreeLoadedMsg:
		if msg.Err != nil {
```

Add loading state reset at the start:
```go
	case TreeLoadedMsg:
		a.treeView.IsLoading = false
		a.treeView.LoadingNodeID = ""
		if msg.Err != nil {
```

**Step 4: Update spinner tick to also check tree loading**

Find case `spinner.TickMsg:` (around line 389):
```go
	case spinner.TickMsg:
		// Update spinner when there's a pending query
		if a.resultTabs.HasPendingQuery() {
			var cmd tea.Cmd
			a.executeSpinner, cmd = a.executeSpinner.Update(msg)
			return a, cmd
		}
		return a, nil
```

Replace with:
```go
	case spinner.TickMsg:
		// Update spinner when there's a pending query or tree is loading
		if a.resultTabs.HasPendingQuery() || a.treeView.IsLoading || a.treeView.LoadingNodeID != "" {
			var cmd tea.Cmd
			a.executeSpinner, cmd = a.executeSpinner.Update(msg)
			return a, cmd
		}
		return a, nil
```

**Step 5: Run build**

Run: `go build ./...`
Expected: Build succeeds

**Step 6: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): wire up tree loading states with spinner"
```

---

## Task 8: Wire up Table loading states in App

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Pass spinner to TableView when creating tabs**

Find where TableView is created for tabs (around line 1493):
```go
				tableView := components.NewTableView(a.theme)
```

Add spinner assignment after creation:
```go
				tableView := components.NewTableView(a.theme)
				tableView.Spinner = &a.executeSpinner
```

**Step 2: Set loading state when creating new table tab**

Find (around line 1497):
```go
				// Add as a new tab
				a.resultTabs.AddTableData(objectID, msg.Node.Label, structureView)
```

Add before this line:
```go
				// Set loading state
				tableView.IsLoading = true
				tableView.LoadingStart = time.Now()
```

**Step 3: Clear loading state when TabTableDataLoadedMsg is received**

Find the handler for `TabTableDataLoadedMsg`. Search for it:

```go
	case TabTableDataLoadedMsg:
```

At the start of this case, add:
```go
	case TabTableDataLoadedMsg:
		// Clear loading state for the tab's table view
		if tab := a.resultTabs.GetTabByObjectID(msg.ObjectID); tab != nil {
			if sv := tab.StructureView; sv != nil {
				if tv := sv.GetTableView(); tv != nil {
					tv.IsLoading = false
				}
			}
		}
```

**Step 4: Set pagination loading state in checkLazyLoad**

Find `checkLazyLoad()` (around line 1716). Before returning the command, set pagination state:

```go
func (a *App) checkLazyLoad() tea.Cmd {
	// Check if we need to load more data (lazy loading)
	if a.tableView.SelectedRow >= len(a.tableView.Rows)-10 &&
		len(a.tableView.Rows) < a.tableView.TotalRows &&
		a.currentTable != "" {
		// Set pagination loading state
		a.tableView.IsPaginating = true
		// Parse schema and table from currentTable
```

**Step 5: Clear pagination state when TableDataLoadedMsg is received**

Find the handler for `TableDataLoadedMsg`. At the end of successful handling, add:

```go
		a.tableView.IsPaginating = false
```

**Step 6: Update spinner tick to check table loading**

Update case `spinner.TickMsg:` (already modified in Task 7):
```go
	case spinner.TickMsg:
		// Update spinner when there's a pending query, tree or table is loading
		needsSpinner := a.resultTabs.HasPendingQuery() ||
			a.treeView.IsLoading ||
			a.treeView.LoadingNodeID != "" ||
			a.tableView.IsPaginating

		// Also check active tab's table view
		if activeTab := a.resultTabs.GetActiveTab(); activeTab != nil {
			if sv := activeTab.StructureView; sv != nil {
				if tv := sv.GetTableView(); tv != nil && tv.IsLoading {
					needsSpinner = true
				}
			}
		}

		if needsSpinner {
			var cmd tea.Cmd
			a.executeSpinner, cmd = a.executeSpinner.Update(msg)
			return a, cmd
		}
		return a, nil
```

**Step 7: Run build and tests**

Run: `go build ./... && go test ./...`
Expected: Build and tests succeed

**Step 8: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): wire up table loading states with spinner"
```

---

## Task 9: Add helper methods for StructureView

**Files:**
- Modify: `internal/ui/components/structure_view.go`

**Step 1: Check if GetTableView method exists**

Search for `GetTableView` method in structure_view.go. If it doesn't exist, add:

```go
// GetTableView returns the underlying TableView for the Data tab
func (sv *StructureView) GetTableView() *TableView {
	return sv.tableView
}
```

**Step 2: Run build**

Run: `go build ./...`
Expected: Build succeeds

**Step 3: Commit (if changes were made)**

```bash
git add internal/ui/components/structure_view.go
git commit -m "feat(structure): add GetTableView helper method"
```

---

## Task 10: Add helper method to ResultTabs

**Files:**
- Modify: `internal/ui/components/result_tabs.go`

**Step 1: Check if GetTabByObjectID method exists**

Search for `GetTabByObjectID` in result_tabs.go. If it doesn't exist, add:

```go
// GetTabByObjectID returns a tab by its object ID
func (rt *ResultTabs) GetTabByObjectID(objectID string) *ResultTab {
	for _, tab := range rt.tabs {
		if tab.ObjectID == objectID {
			return tab
		}
	}
	return nil
}
```

**Step 2: Run build**

Run: `go build ./...`
Expected: Build succeeds

**Step 3: Commit (if changes were made)**

```bash
git add internal/ui/components/result_tabs.go
git commit -m "feat(tabs): add GetTabByObjectID helper method"
```

---

## Task 11: Manual testing

**Step 1: Build and run the application**

Run: `go build -o lazypg ./cmd/lazypg && ./lazypg`

**Step 2: Test Tree initial loading**

1. Connect to a database
2. Verify spinner shows "Loading databases... (X.Xs)" while loading
3. Verify tree displays correctly after loading

**Step 3: Test Tree node expansion loading**

1. Collapse a schema node
2. Expand it
3. Verify inline "Loading..." appears (if loading takes time)

**Step 4: Test Table first load**

1. Select a table from the tree
2. Verify spinner shows "Loading table data... (X.Xs)" with cancel hint
3. Press Esc to test cancel functionality

**Step 5: Test Table pagination**

1. Load a table with many rows
2. Scroll down to trigger lazy loading
3. Verify status bar shows "Loading..." during pagination

**Step 6: Commit final changes if needed**

```bash
git add -A
git commit -m "fix: address any issues found during testing"
```

---

## Summary

This plan adds loading indicators to:
1. Tree initial connection (centered spinner with elapsed time)
2. Tree node expansion (inline spinner)
3. Table first load (centered spinner with elapsed time and cancel hint)
4. Table pagination (status bar spinner)

All loading states share the existing spinner from app.go for consistency.
