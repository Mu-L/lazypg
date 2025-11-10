package components

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rebeliceyang/lazypg/internal/models"
	"github.com/rebeliceyang/lazypg/internal/search"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// CloseCommandPaletteMsg is sent when the command palette should close
type CloseCommandPaletteMsg struct{}

// CommandPalette provides fuzzy search over commands
type CommandPalette struct {
	Input    string
	Commands []models.Command
	Filtered []models.Command
	Selected int
	Width    int
	Height   int
	Theme    theme.Theme
	Mode     string // "", ">", "?", "@", "#"
}

// NewCommandPalette creates a new command palette
func NewCommandPalette(th theme.Theme) *CommandPalette {
	return &CommandPalette{
		Input:    "",
		Commands: []models.Command{},
		Filtered: []models.Command{},
		Selected: 0,
		Width:    80,
		Height:   20,
		Theme:    th,
		Mode:     "",
	}
}

// SetCommands updates the available commands
func (cp *CommandPalette) SetCommands(commands []models.Command) {
	cp.Commands = commands
	cp.Filter()
}

// Update handles keyboard input for the command palette
func (cp *CommandPalette) Update(msg tea.KeyMsg) (*CommandPalette, tea.Cmd) {
	switch msg.String() {
	case "up", "ctrl+p":
		if cp.Selected > 0 {
			cp.Selected--
		}
		return cp, nil

	case "down", "ctrl+n":
		if cp.Selected < len(cp.Filtered)-1 {
			cp.Selected++
		}
		return cp, nil

	case "enter":
		if cp.Selected < len(cp.Filtered) && cp.Selected >= 0 {
			cmd := cp.Filtered[cp.Selected]
			if cmd.Action != nil {
				return cp, func() tea.Msg {
					result := cmd.Action()
					// Also close the palette
					return result
				}
			}
		}
		return cp, func() tea.Msg {
			return CloseCommandPaletteMsg{}
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
		return cp, nil

	default:
		key := msg.String()
		// Only accept single printable characters
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			cp.Input += key
			cp.Filter()
		}
		return cp, nil
	}
}

// Filter filters commands based on input and updates the filtered list
func (cp *CommandPalette) Filter() {
	if cp.Input == "" {
		cp.Filtered = cp.Commands
		cp.Selected = 0
		return
	}

	filtered := []models.Command{}

	for _, cmd := range cp.Commands {
		// Try fuzzy matching on label first
		matchLabel := search.FuzzyMatch(cp.Input, cmd.Label)
		matchDesc := search.FuzzyMatch(cp.Input, cmd.Description)

		// Check tags
		var matchTag search.Match
		for _, tag := range cmd.Tags {
			tagMatch := search.FuzzyMatch(cp.Input, tag)
			if tagMatch.Matched && tagMatch.Score > matchTag.Score {
				matchTag = tagMatch
			}
		}

		// Use best match
		bestScore := 0
		matched := false

		if matchLabel.Matched {
			bestScore = matchLabel.Score + 50 // Bonus for label match
			matched = true
		}
		if matchDesc.Matched && matchDesc.Score > bestScore {
			bestScore = matchDesc.Score + 25 // Bonus for description match
			matched = true
		}
		if matchTag.Matched && matchTag.Score > bestScore {
			bestScore = matchTag.Score + 10 // Bonus for tag match
			matched = true
		}

		if matched {
			cmd.Score = bestScore
			filtered = append(filtered, cmd)
		}
	}

	// Sort by score (descending)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	cp.Filtered = filtered
	cp.Selected = 0
}

// View renders the command palette
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

	cursor := lipgloss.NewStyle().
		Foreground(cp.Theme.Cursor).
		Render("█")

	input := inputStyle.Render(prefix + cp.Input + cursor)

	// Separator
	separator := lipgloss.NewStyle().
		Foreground(cp.Theme.Border).
		Render(strings.Repeat("─", cp.Width-4))

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
		separator,
		lipgloss.JoinVertical(lipgloss.Left, results...),
	)

	// Box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cp.Theme.BorderFocused).
		Padding(1, 2).
		Width(cp.Width)

	return boxStyle.Render(content)
}
