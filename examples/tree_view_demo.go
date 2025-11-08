//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rebeliceyang/lazypg/internal/models"
	"github.com/rebeliceyang/lazypg/internal/ui/components"
	"github.com/rebeliceyang/lazypg/internal/ui/theme"
)

// model represents the application state for the demo
type model struct {
	treeView *components.TreeView
	root     *models.TreeNode
	width    int
	height   int
	message  string
}

func initialModel() model {
	// Build a sample tree
	databases := []string{"postgres", "myapp_db", "template1"}
	root := models.BuildDatabaseTree(databases, "postgres")

	// Add schemas to postgres
	postgres := root.FindByID("db:postgres")
	if postgres != nil {
		schemas := models.BuildSchemaNodes("postgres", []string{"public", "information_schema"})
		models.RefreshTreeChildren(postgres, schemas)
	}

	// Add tables to public schema
	public := root.FindByID("schema:postgres.public")
	if public != nil {
		tables := models.BuildTableNodes("postgres", "public", []string{"users", "posts", "comments"})
		models.RefreshTreeChildren(public, tables)
	}

	// Add columns to users table
	users := root.FindByID("table:postgres.public.users")
	if users != nil {
		columns := []models.ColumnInfo{
			{Name: "id", DataType: "integer", PrimaryKey: true, Nullable: false},
			{Name: "email", DataType: "varchar(255)", PrimaryKey: false, Nullable: false},
			{Name: "name", DataType: "varchar(100)", PrimaryKey: false, Nullable: true},
			{Name: "created_at", DataType: "timestamp", PrimaryKey: false, Nullable: false},
		}
		columnNodes := models.BuildColumnNodes("postgres", "public", "users", columns)
		models.RefreshTreeChildren(users, columnNodes)
	}

	// Add tables to myapp_db
	myappDB := root.FindByID("db:myapp_db")
	if myappDB != nil {
		schemas := models.BuildSchemaNodes("myapp_db", []string{"public"})
		models.RefreshTreeChildren(myappDB, schemas)
	}

	myappPublic := root.FindByID("schema:myapp_db.public")
	if myappPublic != nil {
		tables := models.BuildTableNodes("myapp_db", "public", []string{"products", "orders"})
		models.RefreshTreeChildren(myappPublic, tables)

		// Add row count metadata to products table
		products := root.FindByID("table:myapp_db.public.products")
		if products != nil {
			products.Metadata = map[string]interface{}{
				"row_count": int64(1250),
			}
		}

		orders := root.FindByID("table:myapp_db.public.orders")
		if orders != nil {
			orders.Metadata = map[string]interface{}{
				"row_count": int64(15432),
			}
		}
	}

	testTheme := theme.DefaultTheme()
	treeView := components.NewTreeView(root, testTheme)

	return model{
		treeView: treeView,
		root:     root,
		width:    80,
		height:   24,
		message:  "Press ? for help",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.treeView.Width = msg.Width / 2
		m.treeView.Height = msg.Height - 4
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "?":
			m.message = "Keys: ↑↓/jk=move, →←/hl=expand/collapse, space=toggle, enter=select, g/G=top/bottom, q=quit"
			return m, nil

		default:
			// Handle tree navigation
			var cmd tea.Cmd
			m.treeView, cmd = m.treeView.Update(msg)

			// Check for messages
			if cmd != nil {
				return m, tea.Batch(cmd, func() tea.Msg {
					return cmd()
				})
			}

			return m, nil
		}

	case components.TreeNodeSelectedMsg:
		currentNode := msg.Node
		if currentNode != nil {
			path := currentNode.GetPath()
			m.message = fmt.Sprintf("Selected: %s (Type: %s)", currentNode.Label, currentNode.Type)
			_ = path // We have the path if needed
		}
		return m, nil

	case components.TreeNodeExpandedMsg:
		if msg.Expanded {
			m.message = fmt.Sprintf("Expanded: %s", msg.Node.Label)
		} else {
			m.message = fmt.Sprintf("Collapsed: %s", msg.Node.Label)
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	// Update tree view dimensions
	m.treeView.Width = m.width/2 - 4
	m.treeView.Height = m.height - 6

	// Create the panel border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Width(m.width/2 - 2).
		Height(m.height - 4)

	// Create title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("75"))

	title := titleStyle.Render("Database Navigator Demo")

	// Render tree view
	treeContent := m.treeView.View()

	// Create panel with tree
	panel := borderStyle.Render(treeContent)

	// Create help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	help := helpStyle.Render(m.message)

	// Current node info
	currentNode := m.treeView.GetCurrentNode()
	nodeInfo := ""
	if currentNode != nil {
		path := currentNode.GetPath()
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

		nodeInfo = infoStyle.Render(fmt.Sprintf("Current: %s", pathToString(path)))
	}

	// Combine everything
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		panel,
		nodeInfo,
		help,
	)

	return content
}

func pathToString(path []string) string {
	if len(path) == 0 {
		return ""
	}
	result := ""
	for i, p := range path {
		if i > 0 {
			result += " > "
		}
		result += p
	}
	return result
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
