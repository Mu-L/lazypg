package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// ErrorOverlay represents an error message overlay
type ErrorOverlay struct {
	Title   string
	Message string
	Width   int
	Height  int
	Theme   theme.Theme
}

// NewErrorOverlay creates a new error overlay
func NewErrorOverlay(th theme.Theme) *ErrorOverlay {
	return &ErrorOverlay{
		Theme:  th,
		Width:  60,
		Height: 15,
	}
}

// SetError sets the error title and message
func (e *ErrorOverlay) SetError(title, message string) {
	e.Title = title
	e.Message = message
}

// View renders the error overlay
func (e *ErrorOverlay) View() string {
	if e.Width <= 0 || e.Height <= 0 {
		return ""
	}

	// Title style with error color
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(e.Theme.Error).
		Padding(0, 1)

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(e.Theme.Foreground).
		Padding(1, 2).
		Width(e.Width - 4)

	// Footer style (dimmed)
	footerStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(e.Theme.Foreground).
		Align(lipgloss.Center).
		Width(e.Width - 4)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("Error: " + e.Title))
	content.WriteString("\n\n")

	// Message - wrap text to fit width
	wrappedMessage := wrapText(e.Message, e.Width-8)
	content.WriteString(messageStyle.Render(wrappedMessage))
	content.WriteString("\n")

	// Footer
	content.WriteString(footerStyle.Render("Press Enter or Esc to dismiss"))

	// Box style with error border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(e.Theme.Error).
		Padding(1, 2).
		Width(e.Width).
		MaxWidth(e.Width).
		Background(e.Theme.Background)

	return boxStyle.Render(content.String())
}

// wrapText wraps text to fit within the specified width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	var wrapped []string

	for _, line := range lines {
		if len(line) <= width {
			wrapped = append(wrapped, line)
			continue
		}

		// Wrap long lines
		words := strings.Fields(line)
		if len(words) == 0 {
			wrapped = append(wrapped, line)
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				wrapped = append(wrapped, currentLine)
				currentLine = word
			}
		}
		wrapped = append(wrapped, currentLine)
	}

	return strings.Join(wrapped, "\n")
}
