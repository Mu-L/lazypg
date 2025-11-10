package models

import tea "github.com/charmbracelet/bubbletea"

// CommandType represents the type of command
type CommandType int

const (
	CommandTypeAction CommandType = iota
	CommandTypeObject
	CommandTypeHistory
	CommandTypeFavorite
)

// Command represents an executable command in the command palette
type Command struct {
	ID          string
	Type        CommandType
	Label       string
	Description string
	Icon        string
	Tags        []string
	Score       int // For ranking in search results
	Action      func() tea.Msg
}
