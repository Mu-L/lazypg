package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/rebeliceyang/lazypg/internal/models"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// ColumnsView displays column information
type ColumnsView struct {
	Width   int
	Height  int
	Theme   theme.Theme
	Columns []models.ColumnDetail

	selectedRow int
	topRow      int
	visibleRows int
}

// NewColumnsView creates a new columns view
func NewColumnsView(th theme.Theme) *ColumnsView {
	return &ColumnsView{
		Theme:       th,
		Columns:     []models.ColumnDetail{},
		selectedRow: 0,
		topRow:      0,
	}
}

// SetColumns updates the columns data
func (cv *ColumnsView) SetColumns(columns []models.ColumnDetail) {
	cv.Columns = columns
	cv.selectedRow = 0
	cv.topRow = 0
}

// View renders the columns view
func (cv *ColumnsView) View() string {
	if len(cv.Columns) == 0 {
		return lipgloss.NewStyle().
			Foreground(cv.Theme.Metadata).
			Render("No columns to display")
	}

	var b strings.Builder

	// Render header
	b.WriteString(cv.renderHeader())
	b.WriteString("\n")
	b.WriteString(cv.renderSeparator())
	b.WriteString("\n")

	// Calculate visible rows
	cv.visibleRows = cv.Height - 3 // Header + separator + status
	if cv.visibleRows < 1 {
		cv.visibleRows = 1
	}

	// Render visible rows
	endRow := cv.topRow + cv.visibleRows
	if endRow > len(cv.Columns) {
		endRow = len(cv.Columns)
	}

	for i := cv.topRow; i < endRow; i++ {
		isSelected := i == cv.selectedRow
		b.WriteString(cv.renderRow(cv.Columns[i], isSelected))
		if i < endRow-1 {
			b.WriteString("\n")
		}
	}

	// Status line
	b.WriteString("\n")
	b.WriteString(cv.renderStatus())

	return lipgloss.NewStyle().
		Width(cv.Width).
		Height(cv.Height).
		Render(b.String())
}

func (cv *ColumnsView) renderHeader() string {
	headers := []string{"Name", "Type", "Nullable", "Default", "Constraints", "Comment"}
	widths := []int{20, 20, 10, 20, 15, 30}

	parts := make([]string, len(headers))
	for i, header := range headers {
		truncated := runewidth.Truncate(header, widths[i], "…")
		parts[i] = lipgloss.NewStyle().
			Width(widths[i]).
			Bold(true).
			Foreground(cv.Theme.TableHeader).
			Background(cv.Theme.Selection).
			Render(truncated)
	}

	separatorStyle := lipgloss.NewStyle().
		Foreground(cv.Theme.Border).
		Background(cv.Theme.Selection)

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		parts[0], separatorStyle.Render(" │ "),
		parts[1], separatorStyle.Render(" │ "),
		parts[2], separatorStyle.Render(" │ "),
		parts[3], separatorStyle.Render(" │ "),
		parts[4], separatorStyle.Render(" │ "),
		parts[5],
	)

	return row
}

func (cv *ColumnsView) renderSeparator() string {
	// Total width calculation: 20+20+10+20+15+30 + 5*3 (separators) = 130
	totalWidth := 130
	return lipgloss.NewStyle().
		Foreground(cv.Theme.Border).
		Render(strings.Repeat("─", totalWidth))
}

func (cv *ColumnsView) renderRow(col models.ColumnDetail, selected bool) string {
	widths := []int{20, 20, 10, 20, 15, 30}

	// Format constraint markers
	constraints := cv.formatConstraints(col)

	// Prepare cell values
	cells := []string{
		col.Name,
		col.DataType,
		cv.formatNullable(col.IsNullable),
		col.DefaultValue,
		constraints,
		col.Comment,
	}

	parts := make([]string, len(cells))
	for i, cell := range cells {
		truncated := runewidth.Truncate(cell, widths[i], "…")

		var cellStyle lipgloss.Style
		if selected {
			cellStyle = lipgloss.NewStyle().
				Background(cv.Theme.Selection).
				Foreground(cv.Theme.Foreground)
		} else {
			cellStyle = lipgloss.NewStyle()
		}

		parts[i] = cellStyle.Render(
			lipgloss.NewStyle().Width(widths[i]).Render(truncated),
		)
	}

	separatorStyle := lipgloss.NewStyle().Foreground(cv.Theme.Border)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		parts[0], separatorStyle.Render(" │ "),
		parts[1], separatorStyle.Render(" │ "),
		parts[2], separatorStyle.Render(" │ "),
		parts[3], separatorStyle.Render(" │ "),
		parts[4], separatorStyle.Render(" │ "),
		parts[5],
	)
}

func (cv *ColumnsView) formatConstraints(col models.ColumnDetail) string {
	markers := []string{}
	if col.IsPrimaryKey {
		markers = append(markers, "🔑 PK")
	}
	if col.IsForeignKey {
		markers = append(markers, "🔗 FK")
	}
	if col.IsUnique {
		markers = append(markers, "✓ UQ")
	}
	if col.HasCheck {
		markers = append(markers, "⚠️ CK")
	}
	if len(markers) == 0 {
		return "-"
	}
	return strings.Join(markers, ", ")
}

func (cv *ColumnsView) formatNullable(nullable bool) string {
	if nullable {
		return "YES"
	}
	return "NO"
}

func (cv *ColumnsView) renderStatus() string {
	showing := fmt.Sprintf(" 󰠵 %d columns", len(cv.Columns))
	return lipgloss.NewStyle().
		Foreground(cv.Theme.Metadata).
		Italic(true).
		Render(showing)
}

// MoveSelection moves the selected row up/down
func (cv *ColumnsView) MoveSelection(delta int) {
	cv.selectedRow += delta

	if cv.selectedRow < 0 {
		cv.selectedRow = 0
	}
	if cv.selectedRow >= len(cv.Columns) {
		cv.selectedRow = len(cv.Columns) - 1
	}

	// Adjust scroll
	if cv.selectedRow < cv.topRow {
		cv.topRow = cv.selectedRow
	}
	if cv.selectedRow >= cv.topRow+cv.visibleRows {
		cv.topRow = cv.selectedRow - cv.visibleRows + 1
	}
}

// GetSelectedColumn returns the currently selected column
func (cv *ColumnsView) GetSelectedColumn() *models.ColumnDetail {
	if cv.selectedRow < 0 || cv.selectedRow >= len(cv.Columns) {
		return nil
	}
	return &cv.Columns[cv.selectedRow]
}
