// Package delegates provides message handling delegates for the App.
// Each delegate handles a specific domain of messages, enabling
// separation of concerns and parallel development.
package delegates

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Delegate handles messages for a specific domain.
// Each delegate processes relevant messages and returns whether
// the message was handled and any resulting command.
type Delegate interface {
	// Name returns the delegate name (for debugging/logging)
	Name() string

	// Update processes a message and returns whether it was handled.
	// If handled is true, no other delegates will process this message.
	Update(msg tea.Msg, app AppAccess) (handled bool, cmd tea.Cmd)
}
