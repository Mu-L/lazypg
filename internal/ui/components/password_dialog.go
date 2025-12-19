package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/rebelice/lazypg/internal/ui/theme"
)

// Zone IDs for password dialog
const (
	ZonePasswordSubmit = "password-submit"
	ZonePasswordCancel = "password-cancel"
)

// PasswordSubmitMsg is sent when password is submitted
type PasswordSubmitMsg struct {
	Password string
}

// PasswordCancelMsg is sent when password dialog is cancelled
type PasswordCancelMsg struct{}

// PasswordDialog represents a password input dialog
type PasswordDialog struct {
	Title       string
	Description string
	Width       int
	Height      int
	Theme       theme.Theme

	input    textinput.Model
	host     string
	port     int
	database string
	user     string
}

// NewPasswordDialog creates a new password dialog
func NewPasswordDialog(th theme.Theme) *PasswordDialog {
	input := textinput.New()
	input.Placeholder = "Enter password"
	input.EchoMode = textinput.EchoPassword
	input.EchoCharacter = '•'
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7"))
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
	input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8"))
	input.CharLimit = 256
	input.Width = 40
	input.Focus()

	return &PasswordDialog{
		Theme:  th,
		Width:  50,
		Height: 12,
		input:  input,
	}
}

// SetConnectionInfo sets the connection info to display
func (p *PasswordDialog) SetConnectionInfo(host string, port int, database, user string) {
	p.host = host
	p.port = port
	p.database = database
	p.user = user
	p.Title = "Password Required"
	p.Description = fmt.Sprintf("Enter password for %s@%s:%d/%s", user, host, port, database)
	p.input.SetValue("")
	p.input.Focus()
}

// Init initializes the password dialog
func (p *PasswordDialog) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (p *PasswordDialog) Update(msg tea.Msg) (*PasswordDialog, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return p, func() tea.Msg {
				return PasswordSubmitMsg{Password: p.input.Value()}
			}
		case "esc":
			return p, func() tea.Msg {
				return PasswordCancelMsg{}
			}
		}
	}

	p.input, cmd = p.input.Update(msg)
	return p, cmd
}

// View renders the password dialog
func (p *PasswordDialog) View() string {
	if p.Width <= 0 || p.Height <= 0 {
		return ""
	}

	// Title style
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Theme.Info).
		Padding(0, 1)

	// Description style
	descStyle := lipgloss.NewStyle().
		Foreground(p.Theme.Foreground).
		Faint(true).
		Padding(0, 1)

	// Label style
	labelStyle := lipgloss.NewStyle().
		Foreground(p.Theme.Info).
		Padding(0, 1)

	// Footer style
	footerStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(p.Theme.Foreground).
		Padding(0, 1)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render(p.Title))
	content.WriteString("\n\n")

	// Description
	content.WriteString(descStyle.Render(p.Description))
	content.WriteString("\n\n")

	// Password input
	content.WriteString(labelStyle.Render("Password:"))
	content.WriteString("\n")
	content.WriteString("  ")
	content.WriteString(p.input.View())
	content.WriteString("\n\n")

	// Footer with buttons
	submitBtn := zone.Mark(ZonePasswordSubmit, footerStyle.Render("[Enter] Submit"))
	cancelBtn := zone.Mark(ZonePasswordCancel, footerStyle.Render("[Esc] Cancel"))
	content.WriteString(submitBtn)
	content.WriteString("  ")
	content.WriteString(cancelBtn)

	// Box style
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Theme.Info).
		Padding(1, 2).
		MaxWidth(p.Width).
		Background(p.Theme.Background)

	return boxStyle.Render(content.String())
}

// HandleMouseClick handles mouse click events
func (p *PasswordDialog) HandleMouseClick(msg tea.MouseMsg) (handled bool, cmd tea.Cmd) {
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionPress {
		return false, nil
	}

	if zone.Get(ZonePasswordSubmit).InBounds(msg) {
		return true, func() tea.Msg {
			return PasswordSubmitMsg{Password: p.input.Value()}
		}
	}

	if zone.Get(ZonePasswordCancel).InBounds(msg) {
		return true, func() tea.Msg {
			return PasswordCancelMsg{}
		}
	}

	return false, nil
}

// GetPassword returns the entered password
func (p *PasswordDialog) GetPassword() string {
	return p.input.Value()
}

// GetConnectionInfo returns the connection info
func (p *PasswordDialog) GetConnectionInfo() (host string, port int, database, user string) {
	return p.host, p.port, p.database, p.user
}
