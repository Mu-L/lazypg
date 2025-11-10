# Phase 8: Favorites & Polish Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build query favorites system with YAML storage, add export functionality, enhance help system, refine error handling, and complete final documentation.

**Architecture:** Create favorites manager with YAML persistence stored in ~/.config/lazypg/favorites.yaml. Integrate with command palette and quick query. Add export functionality for query results (CSV, JSON). Enhance help system with context-aware tips. Improve error handling with user-friendly messages.

**Tech Stack:** Go 1.21+, Bubble Tea, Lipgloss, gopkg.in/yaml.v3, encoding/csv, existing command system

---

## Task 1: Favorites Storage and Model

**Files:**
- Create: `internal/favorites/manager.go`
- Create: `internal/favorites/schema.yaml` (example)
- Create: `internal/models/favorite.go`

**Step 1: Create favorite model**

Create `internal/models/favorite.go`:

```go
package models

import "time"

// Favorite represents a saved query
type Favorite struct {
	ID          string    `yaml:"id"`
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Query       string    `yaml:"query"`
	Tags        []string  `yaml:"tags"`
	Connection  string    `yaml:"connection"`  // Connection name
	Database    string    `yaml:"database"`    // Database name
	CreatedAt   time.Time `yaml:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at"`
	UsageCount  int       `yaml:"usage_count"`
	LastUsed    time.Time `yaml:"last_used"`
}
```

**Step 2: Create favorites manager**

Create `internal/favorites/manager.go`:

```go
package favorites

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rebeliceyang/lazypg/internal/models"
	"gopkg.in/yaml.v3"
)

// Manager manages query favorites
type Manager struct {
	path      string
	favorites []models.Favorite
}

// NewManager creates a new favorites manager
func NewManager(configDir string) (*Manager, error) {
	path := filepath.Join(configDir, "favorites.yaml")

	m := &Manager{
		path:      path,
		favorites: []models.Favorite{},
	}

	// Load existing favorites if file exists
	if _, err := os.Stat(path); err == nil {
		if err := m.Load(); err != nil {
			return nil, fmt.Errorf("failed to load favorites: %w", err)
		}
	}

	return m, nil
}

// Load loads favorites from YAML file
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return fmt.Errorf("failed to read favorites file: %w", err)
	}

	if err := yaml.Unmarshal(data, &m.favorites); err != nil {
		return fmt.Errorf("failed to parse favorites: %w", err)
	}

	return nil
}

// Save saves favorites to YAML file
func (m *Manager) Save() error {
	data, err := yaml.Marshal(m.favorites)
	if err != nil {
		return fmt.Errorf("failed to marshal favorites: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(m.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write favorites file: %w", err)
	}

	return nil
}

// Add adds a new favorite
func (m *Manager) Add(name, description, query, connection, database string, tags []string) (*models.Favorite, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	favorite := models.Favorite{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Query:       query,
		Tags:        tags,
		Connection:  connection,
		Database:    database,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UsageCount:  0,
		LastUsed:    time.Time{},
	}

	m.favorites = append(m.favorites, favorite)

	if err := m.Save(); err != nil {
		return nil, err
	}

	return &favorite, nil
}

// Update updates an existing favorite
func (m *Manager) Update(id string, name, description, query string, tags []string) error {
	for i, fav := range m.favorites {
		if fav.ID == id {
			m.favorites[i].Name = name
			m.favorites[i].Description = description
			m.favorites[i].Query = query
			m.favorites[i].Tags = tags
			m.favorites[i].UpdatedAt = time.Now()
			return m.Save()
		}
	}
	return fmt.Errorf("favorite not found: %s", id)
}

// Delete deletes a favorite by ID
func (m *Manager) Delete(id string) error {
	for i, fav := range m.favorites {
		if fav.ID == id {
			m.favorites = append(m.favorites[:i], m.favorites[i+1:]...)
			return m.Save()
		}
	}
	return fmt.Errorf("favorite not found: %s", id)
}

// Get returns a favorite by ID
func (m *Manager) Get(id string) (*models.Favorite, error) {
	for _, fav := range m.favorites {
		if fav.ID == id {
			return &fav, nil
		}
	}
	return nil, fmt.Errorf("favorite not found: %s", id)
}

// GetAll returns all favorites
func (m *Manager) GetAll() []models.Favorite {
	return m.favorites
}

// Search searches favorites by name, description, or tags
func (m *Manager) Search(query string) []models.Favorite {
	if query == "" {
		return m.favorites
	}

	query = strings.ToLower(query)
	var results []models.Favorite

	for _, fav := range m.favorites {
		// Search in name
		if strings.Contains(strings.ToLower(fav.Name), query) {
			results = append(results, fav)
			continue
		}

		// Search in description
		if strings.Contains(strings.ToLower(fav.Description), query) {
			results = append(results, fav)
			continue
		}

		// Search in tags
		for _, tag := range fav.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, fav)
				break
			}
		}
	}

	return results
}

// RecordUsage updates usage statistics for a favorite
func (m *Manager) RecordUsage(id string) error {
	for i, fav := range m.favorites {
		if fav.ID == id {
			m.favorites[i].UsageCount++
			m.favorites[i].LastUsed = time.Now()
			return m.Save()
		}
	}
	return fmt.Errorf("favorite not found: %s", id)
}

// GetMostUsed returns the most frequently used favorites
func (m *Manager) GetMostUsed(limit int) []models.Favorite {
	sorted := make([]models.Favorite, len(m.favorites))
	copy(sorted, m.favorites)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].UsageCount > sorted[j].UsageCount
	})

	if limit > 0 && limit < len(sorted) {
		sorted = sorted[:limit]
	}

	return sorted
}

// GetRecent returns the most recently used favorites
func (m *Manager) GetRecent(limit int) []models.Favorite {
	sorted := make([]models.Favorite, len(m.favorites))
	copy(sorted, m.favorites)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LastUsed.After(sorted[j].LastUsed)
	})

	if limit > 0 && limit < len(sorted) {
		sorted = sorted[:limit]
	}

	return sorted
}
```

**Step 3: Create example favorites schema**

Create `internal/favorites/schema.yaml` (documentation/example file):

```yaml
# Example favorites.yaml structure
- id: "550e8400-e29b-41d4-a716-446655440000"
  name: "Active Users"
  description: "Get all active users from the last 30 days"
  query: "SELECT * FROM users WHERE last_login > NOW() - INTERVAL '30 days'"
  tags:
    - users
    - active
  connection: "production"
  database: "myapp"
  created_at: 2025-11-10T10:00:00Z
  updated_at: 2025-11-10T10:00:00Z
  usage_count: 5
  last_used: 2025-11-10T14:30:00Z

- id: "660e8400-e29b-41d4-a716-446655440001"
  name: "Orders This Month"
  description: "All orders placed in the current month"
  query: "SELECT * FROM orders WHERE created_at >= date_trunc('month', CURRENT_DATE)"
  tags:
    - orders
    - reports
  connection: "production"
  database: "sales"
  created_at: 2025-11-10T11:00:00Z
  updated_at: 2025-11-10T11:00:00Z
  usage_count: 12
  last_used: 2025-11-10T15:00:00Z
```

**Step 4: Build and verify**

Run: `go get gopkg.in/yaml.v3`
Run: `go get github.com/google/uuid`
Run: `go build ./...`
Expected: Clean build

**Step 5: Commit**

```bash
git add internal/models/favorite.go internal/favorites/manager.go internal/favorites/schema.yaml
git commit -m "feat(favorites): add favorites storage and YAML manager"
```

---

## Task 2: Favorites UI Component

**Files:**
- Create: `internal/ui/components/favorites_dialog.go`

**Step 1: Create favorites dialog component**

Create `internal/ui/components/favorites_dialog.go`:

```go
package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rebeliceyang/lazypg/internal/models"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// FavoritesMode represents the dialog mode
type FavoritesMode int

const (
	FavoritesModeList FavoritesMode = iota
	FavoritesModeAdd
	FavoritesModeEdit
)

// ExecuteFavoriteMsg is sent when a favorite should be executed
type ExecuteFavoriteMsg struct {
	Favorite models.Favorite
}

// CloseFavoritesDialogMsg is sent when dialog should close
type CloseFavoritesDialogMsg struct{}

// FavoritesDialog manages favorite queries
type FavoritesDialog struct {
	Width  int
	Height int
	Theme  theme.Theme

	// State
	mode      FavoritesMode
	favorites []models.Favorite
	selected  int
	offset    int

	// Add/Edit state
	nameInput        string
	descriptionInput string
	queryInput       string
	tagsInput        string
	currentField     int // 0=name, 1=description, 2=query, 3=tags

	// Search
	searchQuery string
}

// NewFavoritesDialog creates a new favorites dialog
func NewFavoritesDialog(th theme.Theme) *FavoritesDialog {
	return &FavoritesDialog{
		Width:     80,
		Height:    30,
		Theme:     th,
		mode:      FavoritesModeList,
		favorites: []models.Favorite{},
		selected:  0,
		offset:    0,
	}
}

// SetFavorites updates the favorites list
func (fd *FavoritesDialog) SetFavorites(favorites []models.Favorite) {
	fd.favorites = favorites
	fd.selected = 0
	fd.offset = 0
}

// Update handles keyboard input
func (fd *FavoritesDialog) Update(msg tea.KeyMsg) (*FavoritesDialog, tea.Cmd) {
	switch fd.mode {
	case FavoritesModeList:
		return fd.handleListMode(msg)
	case FavoritesModeAdd, FavoritesModeEdit:
		return fd.handleEditMode(msg)
	}
	return fd, nil
}

func (fd *FavoritesDialog) handleListMode(msg tea.KeyMsg) (*FavoritesDialog, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		return fd, func() tea.Msg {
			return CloseFavoritesDialogMsg{}
		}
	case "up", "k":
		if fd.selected > 0 {
			fd.selected--
			if fd.selected < fd.offset {
				fd.offset = fd.selected
			}
		}
	case "down", "j":
		if fd.selected < len(fd.favorites)-1 {
			fd.selected++
			visibleHeight := fd.Height - 10
			if fd.selected >= fd.offset+visibleHeight {
				fd.offset = fd.selected - visibleHeight + 1
			}
		}
	case "enter":
		// Execute selected favorite
		if fd.selected < len(fd.favorites) {
			fav := fd.favorites[fd.selected]
			return fd, func() tea.Msg {
				return ExecuteFavoriteMsg{Favorite: fav}
			}
		}
	case "a", "n":
		// Add new favorite
		fd.mode = FavoritesModeAdd
		fd.nameInput = ""
		fd.descriptionInput = ""
		fd.queryInput = ""
		fd.tagsInput = ""
		fd.currentField = 0
	case "e":
		// Edit selected favorite
		if fd.selected < len(fd.favorites) {
			fav := fd.favorites[fd.selected]
			fd.mode = FavoritesModeEdit
			fd.nameInput = fav.Name
			fd.descriptionInput = fav.Description
			fd.queryInput = fav.Query
			fd.tagsInput = strings.Join(fav.Tags, ", ")
			fd.currentField = 0
		}
	case "d", "x":
		// Delete - handled by parent
	}
	return fd, nil
}

func (fd *FavoritesDialog) handleEditMode(msg tea.KeyMsg) (*FavoritesDialog, tea.Cmd) {
	switch msg.String() {
	case "esc":
		fd.mode = FavoritesModeList
	case "tab":
		fd.currentField = (fd.currentField + 1) % 4
	case "shift+tab":
		fd.currentField = (fd.currentField - 1 + 4) % 4
	case "backspace":
		fd.deleteChar()
	case "enter":
		if fd.currentField == 3 {
			// Save and close
			fd.mode = FavoritesModeList
			// Parent will handle actual save
		} else {
			fd.currentField++
		}
	default:
		if len(msg.String()) == 1 {
			fd.addChar(msg.String())
		}
	}
	return fd, nil
}

func (fd *FavoritesDialog) addChar(ch string) {
	switch fd.currentField {
	case 0:
		fd.nameInput += ch
	case 1:
		fd.descriptionInput += ch
	case 2:
		fd.queryInput += ch
	case 3:
		fd.tagsInput += ch
	}
}

func (fd *FavoritesDialog) deleteChar() {
	switch fd.currentField {
	case 0:
		if len(fd.nameInput) > 0 {
			fd.nameInput = fd.nameInput[:len(fd.nameInput)-1]
		}
	case 1:
		if len(fd.descriptionInput) > 0 {
			fd.descriptionInput = fd.descriptionInput[:len(fd.descriptionInput)-1]
		}
	case 2:
		if len(fd.queryInput) > 0 {
			fd.queryInput = fd.queryInput[:len(fd.queryInput)-1]
		}
	case 3:
		if len(fd.tagsInput) > 0 {
			fd.tagsInput = fd.tagsInput[:len(fd.tagsInput)-1]
		}
	}
}

// View renders the dialog
func (fd *FavoritesDialog) View() string {
	switch fd.mode {
	case FavoritesModeList:
		return fd.renderList()
	case FavoritesModeAdd, FavoritesModeEdit:
		return fd.renderEdit()
	}
	return ""
}

func (fd *FavoritesDialog) renderList() string {
	var sections []string

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(fd.Theme.Foreground).
		Background(fd.Theme.Info).
		Padding(0, 1).
		Bold(true)
	sections = append(sections, titleStyle.Render("Favorite Queries"))

	// Instructions
	instrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a6adc8")).
		Padding(0, 1)
	sections = append(sections, instrStyle.Render("↑↓: Navigate  Enter: Execute  a: Add  e: Edit  d: Delete  Esc: Close"))

	// Favorites list
	if len(fd.favorites) == 0 {
		sections = append(sections, "\nNo favorites yet. Press 'a' to add one.")
	} else {
		sections = append(sections, "")
		visibleStart := fd.offset
		visibleEnd := fd.offset + fd.Height - 10
		if visibleEnd > len(fd.favorites) {
			visibleEnd = len(fd.favorites)
		}

		for i := visibleStart; i < visibleEnd; i++ {
			fav := fd.favorites[i]

			// Format favorite entry
			name := fav.Name
			if len(name) > 40 {
				name = name[:37] + "..."
			}

			desc := fav.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}

			line := fmt.Sprintf("%s\n  %s", name, desc)
			if len(fav.Tags) > 0 {
				line += fmt.Sprintf(" [%s]", strings.Join(fav.Tags, ", "))
			}

			style := lipgloss.NewStyle().Padding(0, 1)
			if i == fd.selected {
				style = style.Background(fd.Theme.Selection).Foreground(fd.Theme.Foreground)
			}
			sections = append(sections, style.Render(line))
		}
	}

	// Container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(fd.Theme.Border).
		Width(fd.Width).
		Height(fd.Height).
		Padding(1)

	return containerStyle.Render(strings.Join(sections, "\n"))
}

func (fd *FavoritesDialog) renderEdit() string {
	var sections []string

	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(fd.Theme.Foreground).
		Background(fd.Theme.Info).
		Padding(0, 1).
		Bold(true)

	title := "Add Favorite"
	if fd.mode == FavoritesModeEdit {
		title = "Edit Favorite"
	}
	sections = append(sections, titleStyle.Render(title))

	// Instructions
	instrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a6adc8")).
		Padding(0, 1)
	sections = append(sections, instrStyle.Render("Tab: Next field  Enter: Save  Esc: Cancel"))

	// Fields
	sections = append(sections, "")
	sections = append(sections, fd.renderField("Name:", fd.nameInput, fd.currentField == 0))
	sections = append(sections, fd.renderField("Description:", fd.descriptionInput, fd.currentField == 1))
	sections = append(sections, fd.renderField("Query:", fd.queryInput, fd.currentField == 2))
	sections = append(sections, fd.renderField("Tags (comma separated):", fd.tagsInput, fd.currentField == 3))

	// Container
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(fd.Theme.Border).
		Width(fd.Width).
		Height(fd.Height).
		Padding(1)

	return containerStyle.Render(strings.Join(sections, "\n"))
}

func (fd *FavoritesDialog) renderField(label, value string, active bool) string {
	style := lipgloss.NewStyle().Padding(0, 1)
	if active {
		style = style.Background(fd.Theme.Selection).Foreground(fd.Theme.Foreground)
		value = value + "_"
	}
	return style.Render(fmt.Sprintf("%s %s", label, value))
}

// GetEditData returns the current edit data
func (fd *FavoritesDialog) GetEditData() (name, description, query string, tags []string) {
	name = fd.nameInput
	description = fd.descriptionInput
	query = fd.queryInput

	// Parse tags
	if fd.tagsInput != "" {
		parts := strings.Split(fd.tagsInput, ",")
		for _, part := range parts {
			tag := strings.TrimSpace(part)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	return
}
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: Clean build

**Step 3: Commit**

```bash
git add internal/ui/components/favorites_dialog.go
git commit -m "feat(favorites): add favorites dialog UI component"
```

---

## Task 3: Integrate Favorites with App

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/commands/builtin.go`

**Step 1: Add favorites manager to App**

In `internal/app/app.go`, add to imports and App struct:

```go
import (
	"github.com/rebeliceyang/lazypg/internal/favorites"
)

// In App struct (around line 75):
// Favorites
favoritesManager *favorites.Manager
showFavorites    bool
favoritesDialog  *components.FavoritesDialog
```

**Step 2: Initialize favorites in New()**

In the `New()` function (around line 175):

```go
// Initialize favorites manager
favoritesManager, err := favorites.NewManager(configDir)
if err != nil {
	log.Printf("Warning: Could not load favorites: %v", err)
}

favoritesDialog := components.NewFavoritesDialog(th)
```

And in App initialization:

```go
favoritesManager: favoritesManager,
showFavorites:    false,
favoritesDialog:  favoritesDialog,
```

**Step 3: Add favorites keyboard shortcut**

In the `Update()` method (around line 400):

```go
case "ctrl+b":
	// Open favorites dialog
	if a.favoritesManager != nil {
		favorites := a.favoritesManager.GetAll()
		a.favoritesDialog.SetFavorites(favorites)
		a.showFavorites = true
	}
	return a, nil
```

**Step 4: Handle favorites messages**

Add message handling in `Update()` (around line 330):

```go
case components.ExecuteFavoriteMsg:
	// Execute favorite query
	a.showFavorites = false

	// Record usage
	if a.favoritesManager != nil {
		_ = a.favoritesManager.RecordUsage(msg.Favorite.ID)
	}

	// Execute the query
	return a, func() tea.Msg {
		return components.ExecuteQueryMsg{SQL: msg.Favorite.Query}
	}

case components.CloseFavoritesDialogMsg:
	a.showFavorites = false
	return a, nil
```

**Step 5: Handle favorites input**

In the `Update()` method (around line 380):

```go
// Handle favorites dialog input
if a.showFavorites {
	return a.handleFavoritesDialog(msg)
}
```

**Step 6: Add handleFavoritesDialog method**

After `handleJSONBViewer()`:

```go
// handleFavoritesDialog handles key events when favorites dialog is visible
func (a *App) handleFavoritesDialog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.favoritesDialog, cmd = a.favoritesDialog.Update(msg)
	return a, cmd
}
```

**Step 7: Render favorites dialog**

In the `View()` method (around line 860):

```go
// Render favorites dialog if visible
if a.showFavorites {
	mainView = lipgloss.Place(
		a.state.Width,
		a.state.Height,
		lipgloss.Center,
		lipgloss.Center,
		a.favoritesDialog.View(),
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#555555")),
	)
}
```

**Step 8: Add favorites command to registry**

In `internal/commands/builtin.go`, add:

```go
// FavoritesCommandMsg is sent when favorites command is triggered
type FavoritesCommandMsg struct{}

func RegisterFavorites(registry *Registry) {
	registry.Register(models.Command{
		Type:        models.CommandTypeAction,
		Label:       "Favorites",
		Description: "Manage favorite queries",
		Icon:        "⭐",
		Tags:        []string{"favorites", "saved", "queries"},
		Action: func() tea.Cmd {
			return func() tea.Msg {
				return FavoritesCommandMsg{}
			}
		},
	})
}
```

And call it in `RegisterBuiltinCommands()`:

```go
RegisterFavorites(registry)
```

**Step 9: Build and verify**

Run: `go build ./...`
Expected: Clean build

**Step 10: Commit**

```bash
git add internal/app/app.go internal/commands/builtin.go
git commit -m "feat(favorites): integrate favorites with main app"
```

---

## Task 4: Export Functionality

**Files:**
- Create: `internal/export/csv.go`
- Create: `internal/export/json.go`
- Modify: `internal/app/app.go`

**Step 1: Create CSV exporter**

Create `internal/export/csv.go`:

```go
package export

import (
	"encoding/csv"
	"fmt"
	"os"
)

// ToCSV exports data to CSV format
func ToCSV(filename string, columns []string, rows [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(columns); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}
```

**Step 2: Create JSON exporter**

Create `internal/export/json.go`:

```go
package export

import (
	"encoding/json"
	"fmt"
	"os"
)

// ToJSON exports data to JSON format
func ToJSON(filename string, columns []string, rows [][]string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Convert to array of objects
	var data []map[string]string
	for _, row := range rows {
		obj := make(map[string]string)
		for i, col := range columns {
			if i < len(row) {
				obj[col] = row[i]
			}
		}
		data = append(data, obj)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
```

**Step 3: Add export command**

In `internal/app/app.go`, add export shortcut (around line 560):

```go
case "ctrl+e":
	// Export current table data
	if a.state.FocusedPanel == models.RightPanel && a.tableView != nil {
		if len(a.tableView.Rows) > 0 {
			// Show export format selector (for now, default to CSV)
			timestamp := time.Now().Format("20060102_150405")
			tableName := "data"
			if a.state.TreeSelected != nil {
				tableName = a.state.TreeSelected.Label
			}
			filename := fmt.Sprintf("%s_%s.csv", tableName, timestamp)

			err := export.ToCSV(filename, a.tableView.Columns, a.tableView.Rows)
			if err != nil {
				a.ShowError("Export Error", err.Error())
			} else {
				a.ShowError("Export Success", fmt.Sprintf("Exported to %s", filename))
			}
		}
	}
	return a, nil
```

**Step 4: Build and verify**

Run: `go build ./...`
Expected: Clean build

**Step 5: Commit**

```bash
git add internal/export/csv.go internal/export/json.go internal/app/app.go
git commit -m "feat(export): add CSV and JSON export functionality"
```

---

## Task 5: Polish Error Handling and Help

**Files:**
- Modify: `internal/ui/help/help.go`
- Modify: `internal/app/app.go`

**Step 1: Enhance help system**

In `internal/ui/help/help.go`, add more comprehensive help sections:

```go
// Add new section for favorites
var favoritesKeys = []KeyBinding{
	{"Ctrl+B", "Open favorites"},
	{"a/n", "Add new favorite (in dialog)"},
	{"e", "Edit selected favorite"},
	{"d/x", "Delete favorite"},
	{"Enter", "Execute favorite"},
}

// Add export section
var exportKeys = []KeyBinding{
	{"Ctrl+E", "Export current table (CSV)"},
}

// Update GetHelp to include new sections
func GetHelp() string {
	sections := []HelpSection{
		{Title: "Global", Keys: globalKeys},
		{Title: "Navigation", Keys: navigationKeys},
		{Title: "Data View", Keys: dataViewKeys},
		{Title: "Favorites", Keys: favoritesKeys},
		{Title: "Export", Keys: exportKeys},
	}
	// ... rest of rendering
}
```

**Step 2: Improve error messages**

In `internal/app/app.go`, enhance `ShowError` to provide more context:

```go
// ShowError displays an error overlay with context-aware suggestions
func (a *App) ShowError(title, message string) {
	// Add helpful suggestions based on error type
	suggestions := ""

	if strings.Contains(message, "connection") {
		suggestions = "\n\nTip: Check your connection settings with 'c' key"
	} else if strings.Contains(message, "permission") {
		suggestions = "\n\nTip: Verify database user permissions"
	} else if strings.Contains(message, "syntax") {
		suggestions = "\n\nTip: Check SQL syntax in your query"
	}

	a.errorOverlay.SetError(title, message+suggestions)
	a.showError = true
}
```

**Step 3: Build and verify**

Run: `go build ./...`
Expected: Clean build

**Step 4: Commit**

```bash
git add internal/ui/help/help.go internal/app/app.go
git commit -m "feat(polish): enhance help system and error messages"
```

---

## Task 6: Final Documentation

**Files:**
- Create: `docs/features/favorites.md`
- Modify: `README.md`
- Update: All feature docs

**Step 1: Create favorites documentation**

Create `docs/features/favorites.md`:

```markdown
# Favorite Queries

lazypg allows you to save frequently used queries as favorites for quick access.

## Opening Favorites

Press `Ctrl+B` to open the favorites dialog.

## Managing Favorites

### Add a Favorite

1. Press `Ctrl+B` to open favorites
2. Press `a` or `n` to add new
3. Fill in:
   - Name: Short descriptive name
   - Description: Longer explanation
   - Query: The SQL query
   - Tags: Comma-separated tags (optional)
4. Press Tab to move between fields
5. Press Enter to save

### Execute a Favorite

1. Open favorites with `Ctrl+B`
2. Use ↑↓ to select a favorite
3. Press Enter to execute

### Edit a Favorite

1. Open favorites with `Ctrl+B`
2. Select the favorite to edit
3. Press `e`
4. Modify fields
5. Press Enter to save

### Delete a Favorite

1. Open favorites with `Ctrl+B`
2. Select the favorite to delete
3. Press `d` or `x`

## Usage Statistics

Favorites track:
- Usage count (how many times executed)
- Last used timestamp
- Creation and update times

## Storage

Favorites are stored in `~/.config/lazypg/favorites.yaml` in YAML format.

## Examples

**Example favorite for active users:**
```yaml
name: "Active Users Last Month"
description: "Users who logged in within the last 30 days"
query: "SELECT * FROM users WHERE last_login > NOW() - INTERVAL '30 days'"
tags: [users, active, reports]
```

**Example favorite for sales report:**
```yaml
name: "Monthly Sales"
description: "Total sales grouped by month"
query: "SELECT DATE_TRUNC('month', created_at) as month, SUM(amount) FROM orders GROUP BY month ORDER BY month DESC"
tags: [sales, reports, monthly]
```

## Keyboard Shortcuts

- `Ctrl+B` - Open favorites dialog
- `a`/`n` - Add new favorite
- `e` - Edit selected
- `d`/`x` - Delete selected
- `↑`/`↓` - Navigate list
- `Enter` - Execute favorite
- `Esc` - Close dialog
```

**Step 2: Update README**

In `README.md`, update status and features:

```markdown
## Status

**Phase 8 (Favorites & Polish) Complete**
- Query favorites with YAML storage
- Favorites management UI
- CSV/JSON export functionality
- Enhanced help system
- Improved error messages

## Features

### Favorites
- Save frequently used queries
- Tag and organize favorites
- Track usage statistics
- Quick execution with Ctrl+B
- YAML storage in ~/.config/lazypg/

### Export
- Export table data to CSV
- Export to JSON format
- Automatic filename generation
- One-key export with Ctrl+E
```

**Step 3: Create main documentation index**

Create `docs/README.md`:

```markdown
# lazypg Documentation

## Features

- [Filtering](features/filtering.md) - Interactive filter builder
- [JSONB Support](features/jsonb.md) - JSONB viewer and filtering
- [Favorites](features/favorites.md) - Save and manage favorite queries

## Configuration

See `~/.config/lazypg/config.yaml` for configuration options.

## Keyboard Shortcuts

See the help screen (press `?`) for all keyboard shortcuts.
```

**Step 4: Build final binary**

Run: `go build -o bin/lazypg ./cmd/lazypg`
Expected: Clean build

**Step 5: Commit**

```bash
git add docs/features/favorites.md docs/README.md README.md
git commit -m "docs: add favorites documentation and complete Phase 8"
```

---

## Summary

Phase 8 Implementation adds:

1. **Favorites Storage** - YAML-based persistence with full CRUD operations
2. **Favorites UI** - Dialog for managing saved queries
3. **Usage Tracking** - Count and timestamp tracking
4. **Export Functionality** - CSV and JSON export
5. **Enhanced Help** - Comprehensive keyboard shortcuts and tips
6. **Better Error Handling** - Context-aware error messages
7. **Complete Documentation** - User guides for all features

**Keyboard Shortcuts:**
- `Ctrl+B` - Open favorites
- `Ctrl+E` - Export data
- `a` - Add favorite
- `e` - Edit favorite
- `d` - Delete favorite

**Total Files Created:** 8
**Total Files Modified:** 5
**Estimated Implementation Time:** 1-2 weeks
