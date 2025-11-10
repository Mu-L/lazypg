package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// ExecuteQueryMsg is sent when a query should be executed
type ExecuteQueryMsg struct {
	SQL string
}

// QuickQuery provides a single-line SQL input at the bottom of the screen
type QuickQuery struct {
	Input   string
	Width   int
	Theme   theme.Theme
	History []string
	HistIdx int
}

// NewQuickQuery creates a new quick query component
func NewQuickQuery(th theme.Theme) *QuickQuery {
	return &QuickQuery{
		Input:   "",
		Width:   80,
		Theme:   th,
		History: []string{},
		HistIdx: -1,
	}
}

// Update handles keyboard input for quick query
func (qq *QuickQuery) Update(msg tea.KeyMsg) (*QuickQuery, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if qq.Input != "" {
			sql := qq.Input
			// Add to history
			qq.History = append(qq.History, sql)
			qq.HistIdx = len(qq.History)
			// Clear input
			qq.Input = ""
			// Send execute message
			return qq, func() tea.Msg {
				return ExecuteQueryMsg{SQL: sql}
			}
		}
		return qq, nil

	case "esc", "ctrl+c":
		return qq, nil

	case "up", "ctrl+p":
		// Navigate history backward
		if len(qq.History) > 0 {
			if qq.HistIdx > 0 {
				qq.HistIdx--
				qq.Input = qq.History[qq.HistIdx]
			} else if qq.HistIdx == -1 {
				qq.HistIdx = len(qq.History) - 1
				qq.Input = qq.History[qq.HistIdx]
			}
		}
		return qq, nil

	case "down", "ctrl+n":
		// Navigate history forward
		if len(qq.History) > 0 && qq.HistIdx >= 0 {
			if qq.HistIdx < len(qq.History)-1 {
				qq.HistIdx++
				qq.Input = qq.History[qq.HistIdx]
			} else {
				qq.HistIdx = len(qq.History)
				qq.Input = ""
			}
		}
		return qq, nil

	case "backspace":
		if len(qq.Input) > 0 {
			qq.Input = qq.Input[:len(qq.Input)-1]
			// Reset history index when editing
			qq.HistIdx = len(qq.History)
		}
		return qq, nil

	case "ctrl+u":
		// Clear input
		qq.Input = ""
		qq.HistIdx = len(qq.History)
		return qq, nil

	default:
		// Handle printable characters
		key := msg.String()
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			qq.Input += key
			// Reset history index when editing
			qq.HistIdx = len(qq.History)
		}
		return qq, nil
	}
}

// View renders the quick query input
func (qq *QuickQuery) View() string {
	prefix := lipgloss.NewStyle().
		Foreground(qq.Theme.Info).
		Bold(true).
		Render("SQL> ")

	cursor := lipgloss.NewStyle().
		Foreground(qq.Theme.Cursor).
		Render("█")

	inputStyle := lipgloss.NewStyle().
		Foreground(qq.Theme.Foreground).
		Background(qq.Theme.Selection).
		Padding(0, 1)

	input := inputStyle.Render(qq.Input + cursor)

	hint := lipgloss.NewStyle().
		Foreground(qq.Theme.Comment).
		Italic(true).
		Render(" [Enter: Execute | ↑↓: History | Esc: Cancel]")

	// Full width bar
	barStyle := lipgloss.NewStyle().
		Width(qq.Width).
		Background(qq.Theme.Selection).
		Padding(0, 1)

	content := prefix + input + hint
	return barStyle.Render(content)
}
