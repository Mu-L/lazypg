package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/rebeliceyang/lazypg/internal/jsonb"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// TableView displays table data with virtual scrolling
type TableView struct {
	Columns      []string
	Rows         [][]string
	Width        int
	Height       int
	Style        lipgloss.Style
	Theme        theme.Theme // Color theme

	// Virtual scrolling state
	TopRow       int
	VisibleRows  int
	SelectedRow  int
	SelectedCol  int // Currently selected column
	TotalRows    int

	// Column widths (calculated)
	ColumnWidths []int
}

// NewTableView creates a new table view with theme
func NewTableView(th theme.Theme) *TableView {
	return &TableView{
		Columns:      []string{},
		Rows:         [][]string{},
		ColumnWidths: []int{},
		Theme:        th,
	}
}

// SetData sets the table data
func (tv *TableView) SetData(columns []string, rows [][]string, totalRows int) {
	tv.Columns = columns
	tv.Rows = rows
	tv.TotalRows = totalRows
	tv.calculateColumnWidths()
}

// calculateColumnWidths calculates optimal column widths
func (tv *TableView) calculateColumnWidths() {
	if len(tv.Columns) == 0 {
		return
	}

	numColumns := len(tv.Columns)
	tv.ColumnWidths = make([]int, numColumns)

	// Calculate separator space: 3 chars (" │ ") * (numColumns - 1)
	separatorWidth := 0
	if numColumns > 1 {
		separatorWidth = 3 * (numColumns - 1)
	}

	// Available width for actual column content
	availableWidth := tv.Width - separatorWidth
	if availableWidth < numColumns*10 {
		// Minimum 10 chars per column
		availableWidth = numColumns * 10
	}

	// Step 1: Calculate desired widths based on content
	desiredWidths := make([]int, numColumns)

	// Start with column header lengths
	for i, col := range tv.Columns {
		desiredWidths[i] = runewidth.StringWidth(col)
	}

	// Check row data
	for _, row := range tv.Rows {
		for i, cell := range row {
			if i < numColumns {
				cellLen := runewidth.StringWidth(cell)
				if cellLen > desiredWidths[i] {
					desiredWidths[i] = cellLen
				}
			}
		}
	}

	// Step 2: Apply constraints and distribute available width
	maxWidth := 50
	minWidth := 10

	// Calculate total desired width
	totalDesired := 0
	for _, w := range desiredWidths {
		if w > maxWidth {
			w = maxWidth
		}
		if w < minWidth {
			w = minWidth
		}
		totalDesired += w
	}

	// Step 3: Distribute width proportionally if we exceed available width
	if totalDesired > availableWidth {
		// Scale down proportionally
		scale := float64(availableWidth) / float64(totalDesired)
		for i := range desiredWidths {
			w := desiredWidths[i]
			if w > maxWidth {
				w = maxWidth
			}
			if w < minWidth {
				w = minWidth
			}
			tv.ColumnWidths[i] = int(float64(w) * scale)
			if tv.ColumnWidths[i] < minWidth {
				tv.ColumnWidths[i] = minWidth
			}
		}
	} else {
		// Use desired widths with constraints
		for i, w := range desiredWidths {
			if w > maxWidth {
				w = maxWidth
			}
			if w < minWidth {
				w = minWidth
			}
			tv.ColumnWidths[i] = w
		}
	}
}

// View renders the table
func (tv *TableView) View() string {
	if len(tv.Columns) == 0 {
		return tv.Style.Render("No data")
	}

	var b strings.Builder

	// Render header
	b.WriteString(tv.renderHeader())
	b.WriteString("\n")
	b.WriteString(tv.renderSeparator())
	b.WriteString("\n")

	// Calculate how many rows we can show
	// Height is already the content area height
	// Subtract 3 for header + separator + status line
	tv.VisibleRows = tv.Height - 3
	if tv.VisibleRows < 1 {
		tv.VisibleRows = 1
	}

	// Render visible rows
	endRow := tv.TopRow + tv.VisibleRows
	if endRow > len(tv.Rows) {
		endRow = len(tv.Rows)
	}

	for i := tv.TopRow; i < endRow; i++ {
		isSelected := i == tv.SelectedRow
		b.WriteString(tv.renderRow(tv.Rows[i], isSelected))
		if i < endRow-1 {
			b.WriteString("\n")
		}
	}

	// Render status
	b.WriteString("\n")
	b.WriteString(tv.renderStatus())

	return tv.Style.Width(tv.Width).Height(tv.Height).Render(b.String())
}

func (tv *TableView) renderHeader() string {
	s := make([]string, 0, len(tv.Columns)*2-1) // Account for separators

	// Create separator style
	separatorStyle := lipgloss.NewStyle().
		Foreground(tv.Theme.Border).
		Background(tv.Theme.Selection)
	separator := separatorStyle.Render(" │ ")

	for i, col := range tv.Columns {
		width := tv.ColumnWidths[i]
		if width <= 0 {
			continue
		}

		// Use runewidth.Truncate for proper truncation
		truncated := runewidth.Truncate(col, width, "…")

		// Create cell width style
		widthStyle := lipgloss.NewStyle().
			Width(width).
			MaxWidth(width).
			Inline(true)

		// Create header cell style
		headerCellStyle := lipgloss.NewStyle().
			Background(tv.Theme.Selection)

		// Render cell with header background
		renderedCell := headerCellStyle.Render(widthStyle.Render(truncated))
		s = append(s, renderedCell)

		// Add separator between columns (but not after the last column)
		if i < len(tv.Columns)-1 {
			s = append(s, separator)
		}
	}

	// Join header cells horizontally with separators
	headerRow := lipgloss.JoinHorizontal(lipgloss.Top, s...)

	// Apply bold and color to the entire row
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(tv.Theme.TableHeader)

	return headerStyle.Render(headerRow)
}

func (tv *TableView) renderSeparator() string {
	// Calculate total width of all columns
	totalWidth := 0
	for _, width := range tv.ColumnWidths {
		totalWidth += width
	}

	// Add width for separators: 3 chars (" │ ") * (number of separators)
	// Number of separators = number of columns - 1
	if len(tv.ColumnWidths) > 1 {
		totalWidth += 3 * (len(tv.ColumnWidths) - 1)
	}

	// Create a simple horizontal line
	separatorStyle := lipgloss.NewStyle().
		Foreground(tv.Theme.Border)

	return separatorStyle.Render(strings.Repeat("─", totalWidth))
}

func (tv *TableView) renderRow(row []string, selected bool) string {
	s := make([]string, 0, len(row)*2-1) // Account for separators

	// Create separator style (always uses border color, no background)
	separatorStyle := lipgloss.NewStyle().
		Foreground(tv.Theme.Border)
	separator := separatorStyle.Render(" │ ")

	for i, value := range row {
		if i >= len(tv.ColumnWidths) {
			break
		}
		width := tv.ColumnWidths[i]
		if width <= 0 {
			continue
		}

		// Check if this looks like JSONB and format for display
		cellValue := value
		if jsonb.IsJSONB(cellValue) {
			cellValue = jsonb.Truncate(cellValue, 50)
		}

		// Use runewidth.Truncate for proper truncation (handles multibyte chars)
		truncated := runewidth.Truncate(cellValue, width, "…")

		// Create cell width style
		widthStyle := lipgloss.NewStyle().
			Width(width).
			MaxWidth(width).
			Inline(true)

		// Determine cell background based on selection
		var cellStyle lipgloss.Style
		if selected && i == tv.SelectedCol {
			// Selected cell - bright highlight
			cellStyle = lipgloss.NewStyle().
				Background(tv.Theme.BorderFocused).
				Foreground(tv.Theme.Background).
				Bold(true)
		} else if selected {
			// Selected row but not selected column - dim highlight
			cellStyle = lipgloss.NewStyle().
				Background(tv.Theme.Selection).
				Foreground(tv.Theme.Foreground)
		} else {
			// Normal cell
			cellStyle = lipgloss.NewStyle()
		}

		// Render cell: first apply width, then apply cell style
		renderedCell := cellStyle.Render(widthStyle.Render(truncated))
		s = append(s, renderedCell)

		// Add separator between columns (but not after the last column)
		if i < len(row)-1 && i < len(tv.ColumnWidths)-1 {
			s = append(s, separator)
		}
	}

	// Join cells horizontally with separators
	rowStr := lipgloss.JoinHorizontal(lipgloss.Top, s...)

	return rowStr
}

func (tv *TableView) renderStatus() string {
	endRow := tv.TopRow + len(tv.Rows)
	if endRow > tv.TotalRows {
		endRow = tv.TotalRows
	}

	showing := fmt.Sprintf(" 󰈙 %d-%d of %d rows", tv.TopRow+1, endRow, tv.TotalRows)
	return lipgloss.NewStyle().
		Foreground(tv.Theme.Metadata).
		Italic(true).
		Render(showing)
}

func (tv *TableView) pad(s string, width int) string {
	if len(s) > width {
		return s[:width-3] + "..."
	}
	return s + strings.Repeat(" ", width-len(s))
}

// MoveSelection moves the selection up or down
func (tv *TableView) MoveSelection(delta int) {
	tv.SelectedRow += delta

	// Bounds checking
	if tv.SelectedRow < 0 {
		tv.SelectedRow = 0
	}
	if tv.SelectedRow >= len(tv.Rows) {
		tv.SelectedRow = len(tv.Rows) - 1
	}

	// Adjust visible window if needed
	if tv.SelectedRow < tv.TopRow {
		tv.TopRow = tv.SelectedRow
	}
	if tv.SelectedRow >= tv.TopRow+tv.VisibleRows {
		tv.TopRow = tv.SelectedRow - tv.VisibleRows + 1
	}
}

// PageUp/PageDown
func (tv *TableView) PageUp() {
	tv.SelectedRow -= tv.VisibleRows
	if tv.SelectedRow < 0 {
		tv.SelectedRow = 0
	}
	tv.TopRow = tv.SelectedRow
}

func (tv *TableView) PageDown() {
	tv.SelectedRow += tv.VisibleRows
	if tv.SelectedRow >= len(tv.Rows) {
		tv.SelectedRow = len(tv.Rows) - 1
	}
	tv.TopRow = tv.SelectedRow
	if tv.TopRow+tv.VisibleRows > len(tv.Rows) {
		tv.TopRow = len(tv.Rows) - tv.VisibleRows
		if tv.TopRow < 0 {
			tv.TopRow = 0
		}
	}
}

// GetSelectedCell returns the currently selected row and column indices
func (tv *TableView) GetSelectedCell() (row int, col int) {
	return tv.SelectedRow, tv.SelectedCol
}

// MoveSelectionHorizontal moves the selected column left or right
func (tv *TableView) MoveSelectionHorizontal(delta int) {
	tv.SelectedCol += delta

	// Bounds checking
	if tv.SelectedCol < 0 {
		tv.SelectedCol = 0
	}
	if tv.SelectedCol >= len(tv.Columns) {
		tv.SelectedCol = len(tv.Columns) - 1
	}
}
