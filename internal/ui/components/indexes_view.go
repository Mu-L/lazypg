package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/rebeliceyang/lazypg/internal/db/metadata"
	"github.com/rebeliceyang/lazypg/internal/models"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// IndexesView displays index information
type IndexesView struct {
	Width   int
	Height  int
	Theme   theme.Theme
	Indexes []models.IndexInfo

	selectedRow int
	topRow      int
	visibleRows int
}

// NewIndexesView creates a new indexes view
func NewIndexesView(th theme.Theme) *IndexesView {
	return &IndexesView{
		Theme:       th,
		Indexes:     []models.IndexInfo{},
		selectedRow: 0,
		topRow:      0,
	}
}

// SetIndexes updates the indexes data
func (iv *IndexesView) SetIndexes(indexes []models.IndexInfo) {
	iv.Indexes = indexes
	iv.selectedRow = 0
	iv.topRow = 0
}

// View renders the indexes view
func (iv *IndexesView) View() string {
	if len(iv.Indexes) == 0 {
		return lipgloss.NewStyle().
			Foreground(iv.Theme.Metadata).
			Render("No indexes to display")
	}

	var b strings.Builder

	b.WriteString(iv.renderHeader())
	b.WriteString("\n")
	b.WriteString(iv.renderSeparator())
	b.WriteString("\n")

	iv.visibleRows = iv.Height - 3
	if iv.visibleRows < 1 {
		iv.visibleRows = 1
	}

	endRow := iv.topRow + iv.visibleRows
	if endRow > len(iv.Indexes) {
		endRow = len(iv.Indexes)
	}

	for i := iv.topRow; i < endRow; i++ {
		isSelected := i == iv.selectedRow
		b.WriteString(iv.renderRow(iv.Indexes[i], isSelected))
		if i < endRow-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(iv.renderStatus())

	return lipgloss.NewStyle().
		Width(iv.Width).
		Height(iv.Height).
		Render(b.String())
}

func (iv *IndexesView) renderHeader() string {
	headers := []string{"Name", "Type", "Columns", "Properties", "Size", "Definition"}
	widths := []int{25, 10, 20, 20, 10, 40}

	parts := make([]string, len(headers))
	for i, header := range headers {
		truncated := runewidth.Truncate(header, widths[i], "…")
		parts[i] = lipgloss.NewStyle().
			Width(widths[i]).
			Bold(true).
			Foreground(iv.Theme.TableHeader).
			Background(iv.Theme.Selection).
			Render(truncated)
	}

	separatorStyle := lipgloss.NewStyle().
		Foreground(iv.Theme.Border).
		Background(iv.Theme.Selection)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		parts[0], separatorStyle.Render(" │ "),
		parts[1], separatorStyle.Render(" │ "),
		parts[2], separatorStyle.Render(" │ "),
		parts[3], separatorStyle.Render(" │ "),
		parts[4], separatorStyle.Render(" │ "),
		parts[5],
	)
}

func (iv *IndexesView) renderSeparator() string {
	totalWidth := 25 + 10 + 20 + 20 + 10 + 40 + 5*3 // widths + separators
	return lipgloss.NewStyle().
		Foreground(iv.Theme.Border).
		Render(strings.Repeat("─", totalWidth))
}

func (iv *IndexesView) renderRow(idx models.IndexInfo, selected bool) string {
	widths := []int{25, 10, 20, 20, 10, 40}

	// Format columns
	columnsStr := strings.Join(idx.Columns, ", ")

	// Format properties
	properties := iv.formatProperties(idx)

	// Format size
	sizeStr := metadata.FormatSize(idx.Size)

	// Format definition
	definition := idx.Definition

	cells := []string{
		idx.Name,
		idx.Type,
		columnsStr,
		properties,
		sizeStr,
		definition,
	}

	parts := make([]string, len(cells))
	for i, cell := range cells {
		truncated := runewidth.Truncate(cell, widths[i], "…")

		var cellStyle lipgloss.Style
		if selected {
			cellStyle = lipgloss.NewStyle().
				Background(iv.Theme.Selection).
				Foreground(iv.Theme.Foreground)
		} else {
			cellStyle = lipgloss.NewStyle()
		}

		parts[i] = cellStyle.Render(
			lipgloss.NewStyle().Width(widths[i]).Render(truncated),
		)
	}

	separatorStyle := lipgloss.NewStyle().Foreground(iv.Theme.Border)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		parts[0], separatorStyle.Render(" │ "),
		parts[1], separatorStyle.Render(" │ "),
		parts[2], separatorStyle.Render(" │ "),
		parts[3], separatorStyle.Render(" │ "),
		parts[4], separatorStyle.Render(" │ "),
		parts[5],
	)
}

func (iv *IndexesView) formatProperties(idx models.IndexInfo) string {
	props := []string{}
	if idx.IsPrimary {
		props = append(props, "🔑 PK")
	}
	if idx.IsUnique {
		props = append(props, "✓ UQ")
	}
	if idx.IsPartial {
		props = append(props, "📋 Partial")
	}
	if len(props) == 0 {
		return "-"
	}
	return strings.Join(props, ", ")
}

func (iv *IndexesView) renderStatus() string {
	showing := fmt.Sprintf(" 󰘚 %d indexes", len(iv.Indexes))
	return lipgloss.NewStyle().
		Foreground(iv.Theme.Metadata).
		Italic(true).
		Render(showing)
}

// MoveSelection moves the selected row up/down
func (iv *IndexesView) MoveSelection(delta int) {
	iv.selectedRow += delta

	if iv.selectedRow < 0 {
		iv.selectedRow = 0
	}
	if iv.selectedRow >= len(iv.Indexes) {
		iv.selectedRow = len(iv.Indexes) - 1
	}

	if iv.selectedRow < iv.topRow {
		iv.topRow = iv.selectedRow
	}
	if iv.selectedRow >= iv.topRow+iv.visibleRows {
		iv.topRow = iv.selectedRow - iv.visibleRows + 1
	}
}

// GetSelectedIndex returns the currently selected index
func (iv *IndexesView) GetSelectedIndex() *models.IndexInfo {
	if iv.selectedRow < 0 || iv.selectedRow >= len(iv.Indexes) {
		return nil
	}
	return &iv.Indexes[iv.selectedRow]
}
