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

// ConstraintsView displays constraint information
type ConstraintsView struct {
	Width       int
	Height      int
	Theme       theme.Theme
	Constraints []models.Constraint

	selectedRow int
	topRow      int
	visibleRows int
}

// NewConstraintsView creates a new constraints view
func NewConstraintsView(th theme.Theme) *ConstraintsView {
	return &ConstraintsView{
		Theme:       th,
		Constraints: []models.Constraint{},
		selectedRow: 0,
		topRow:      0,
	}
}

// SetConstraints updates the constraints data
func (cv *ConstraintsView) SetConstraints(constraints []models.Constraint) {
	cv.Constraints = constraints
	cv.selectedRow = 0
	cv.topRow = 0
}

// View renders the constraints view
func (cv *ConstraintsView) View() string {
	if len(cv.Constraints) == 0 {
		return lipgloss.NewStyle().
			Foreground(cv.Theme.Metadata).
			Render("No constraints to display")
	}

	var b strings.Builder

	b.WriteString(cv.renderHeader())
	b.WriteString("\n")
	b.WriteString(cv.renderSeparator())
	b.WriteString("\n")

	cv.visibleRows = cv.Height - 3
	if cv.visibleRows < 1 {
		cv.visibleRows = 1
	}

	endRow := cv.topRow + cv.visibleRows
	if endRow > len(cv.Constraints) {
		endRow = len(cv.Constraints)
	}

	for i := cv.topRow; i < endRow; i++ {
		isSelected := i == cv.selectedRow
		b.WriteString(cv.renderRow(cv.Constraints[i], isSelected))
		if i < endRow-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(cv.renderStatus())

	return lipgloss.NewStyle().
		Width(cv.Width).
		Height(cv.Height).
		Render(b.String())
}

func (cv *ConstraintsView) renderHeader() string {
	headers := []string{"Type", "Name", "Columns", "Definition", "Description"}
	widths := []int{6, 25, 20, 45, 30}

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

	return lipgloss.JoinHorizontal(lipgloss.Top,
		parts[0], separatorStyle.Render(" │ "),
		parts[1], separatorStyle.Render(" │ "),
		parts[2], separatorStyle.Render(" │ "),
		parts[3], separatorStyle.Render(" │ "),
		parts[4],
	)
}

func (cv *ConstraintsView) renderSeparator() string {
	totalWidth := 6 + 25 + 20 + 45 + 30 + 4*3 // widths + separators
	return lipgloss.NewStyle().
		Foreground(cv.Theme.Border).
		Render(strings.Repeat("─", totalWidth))
}

func (cv *ConstraintsView) renderRow(con models.Constraint, selected bool) string {
	widths := []int{6, 25, 20, 45, 30}

	// Format type with color
	typeLabel := metadata.FormatConstraintType(con.Type)
	typeColor := cv.getTypeColor(con.Type)
	typeCell := lipgloss.NewStyle().
		Foreground(typeColor).
		Bold(true).
		Render(typeLabel)

	// Format columns
	columnsStr := strings.Join(con.Columns, ", ")

	// Format definition
	definition := cv.formatDefinition(con)

	// Format description
	description := cv.formatDescription(con)

	cells := []string{
		typeCell,
		con.Name,
		columnsStr,
		definition,
		description,
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
		parts[4],
	)
}

func (cv *ConstraintsView) getTypeColor(conType string) lipgloss.Color {
	switch conType {
	case "p":
		return cv.Theme.Info // Blue
	case "f":
		return cv.Theme.Warning // Orange
	case "u":
		return cv.Theme.Success // Green
	case "c":
		return cv.Theme.Metadata // Gray
	default:
		return cv.Theme.Foreground
	}
}

func (cv *ConstraintsView) formatDefinition(con models.Constraint) string {
	if con.Type == "f" && con.ForeignTable != "" {
		// Format as: → table(columns)
		fkCols := strings.Join(con.ForeignCols, ", ")
		return fmt.Sprintf("→ %s(%s)", con.ForeignTable, fkCols)
	}
	return con.Definition
}

func (cv *ConstraintsView) formatDescription(con models.Constraint) string {
	switch con.Type {
	case "p":
		return "Primary key constraint"
	case "f":
		return fmt.Sprintf("References %s", con.ForeignTable)
	case "u":
		return "Unique constraint"
	case "c":
		return "Check constraint"
	default:
		return "-"
	}
}

func (cv *ConstraintsView) renderStatus() string {
	showing := fmt.Sprintf(" 󰌆 %d constraints", len(cv.Constraints))
	return lipgloss.NewStyle().
		Foreground(cv.Theme.Metadata).
		Italic(true).
		Render(showing)
}

// MoveSelection moves the selected row up/down
func (cv *ConstraintsView) MoveSelection(delta int) {
	cv.selectedRow += delta

	if cv.selectedRow < 0 {
		cv.selectedRow = 0
	}
	if cv.selectedRow >= len(cv.Constraints) {
		cv.selectedRow = len(cv.Constraints) - 1
	}

	if cv.selectedRow < cv.topRow {
		cv.topRow = cv.selectedRow
	}
	if cv.selectedRow >= cv.topRow+cv.visibleRows {
		cv.topRow = cv.selectedRow - cv.visibleRows + 1
	}
}

// GetSelectedConstraint returns the currently selected constraint
func (cv *ConstraintsView) GetSelectedConstraint() *models.Constraint {
	if cv.selectedRow < 0 || cv.selectedRow >= len(cv.Constraints) {
		return nil
	}
	return &cv.Constraints[cv.selectedRow]
}
