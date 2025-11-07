package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rebeliceyang/lazypg/internal/models"
)

// App is the main application model
type App struct {
	state models.AppState
}

// New creates a new App instance
func New() *App {
	return &App{
		state: models.AppState{
			Width:  80,
			Height: 24,
		},
	}
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	case tea.WindowSizeMsg:
		a.state.Width = msg.Width
		a.state.Height = msg.Height
	}
	return a, nil
}

// View implements tea.Model
func (a *App) View() string {
	return "lazypg - Press 'q' to quit\n"
}
