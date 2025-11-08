//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"strings"

	"github.com/rebeliceyang/lazypg/internal/models"
)

// This example demonstrates the tree model functionality
func main() {
	fmt.Println("=== Navigation Tree Demo ===\n")

	// 1. Build a database tree
	databases := []string{"postgres", "myapp_db", "template1"}
	root := models.BuildDatabaseTree(databases, "postgres")

	fmt.Println("1. Initial tree with databases:")
	printTree(root, 0)
	fmt.Println()

	// 2. Get flattened view
	flat := root.Flatten()
	fmt.Printf("2. Flattened view (%d visible nodes):\n", len(flat))
	for i, node := range flat {
		fmt.Printf("   [%d] %s (ID: %s)\n", i, node.Label, node.ID)
	}
	fmt.Println()

	// 3. Expand postgres database and add schemas
	postgres := root.FindByID("db:postgres")
	if postgres != nil {
		fmt.Println("3. Expanding postgres database and adding schemas...")
		schemas := models.BuildSchemaNodes("postgres", []string{"public", "information_schema"})
		models.RefreshTreeChildren(postgres, schemas)
		postgres.Expanded = true

		printTree(root, 0)
		fmt.Println()
	}

	// 4. Expand public schema and add tables
	public := root.FindByID("schema:postgres.public")
	if public != nil {
		fmt.Println("4. Expanding public schema and adding tables...")
		tables := models.BuildTableNodes("postgres", "public", []string{"users", "posts", "comments"})
		models.RefreshTreeChildren(public, tables)
		public.Expanded = true

		printTree(root, 0)
		fmt.Println()
	}

	// 5. Expand users table and add columns
	users := root.FindByID("table:postgres.public.users")
	if users != nil {
		fmt.Println("5. Expanding users table and adding columns...")
		columns := []models.ColumnInfo{
			{Name: "id", DataType: "integer", PrimaryKey: true, Nullable: false},
			{Name: "email", DataType: "varchar(255)", PrimaryKey: false, Nullable: false},
			{Name: "created_at", DataType: "timestamp", PrimaryKey: false, Nullable: false},
		}
		columnNodes := models.BuildColumnNodes("postgres", "public", "users", columns)
		models.RefreshTreeChildren(users, columnNodes)
		users.Expanded = true

		printTree(root, 0)
		fmt.Println()
	}

	// 6. Demonstrate path and depth
	if users != nil {
		fmt.Println("6. Node information for 'users' table:")
		path := users.GetPath()
		fmt.Printf("   Path: %s\n", strings.Join(path, " > "))
		fmt.Printf("   Depth: %d\n", users.GetDepth())
		fmt.Printf("   Type: %s\n", users.Type)
		fmt.Printf("   Selectable: %t\n", users.Selectable)
		fmt.Println()
	}

	// 7. Demonstrate flattened view after full expansion
	flat = root.Flatten()
	fmt.Printf("7. Final flattened view (%d visible nodes):\n", len(flat))
	for i, node := range flat {
		indent := strings.Repeat("  ", node.GetDepth()-1)
		icon := "▸"
		if node.Expanded {
			icon = "▾"
		} else if node.Type == models.TreeNodeTypeColumn {
			icon = "•"
		}
		fmt.Printf("   [%d] %s%s %s\n", i, indent, icon, node.Label)
	}
	fmt.Println()

	// 8. Demonstrate collapsing
	fmt.Println("8. Collapsing public schema...")
	if public != nil {
		public.Toggle()
		printTree(root, 0)
		fmt.Println()

		flat = root.Flatten()
		fmt.Printf("   Flattened view now has %d visible nodes\n", len(flat))
	}
}

// printTree is a helper function to print the tree structure
func printTree(node *models.TreeNode, depth int) {
	if node.Type != models.TreeNodeTypeRoot {
		indent := strings.Repeat("  ", depth)
		icon := "▸"
		if node.Expanded {
			icon = "▾"
		} else if node.Type == models.TreeNodeTypeColumn {
			icon = "•"
		} else if !node.Loaded && len(node.Children) == 0 {
			icon = "▸"
		}

		active := ""
		if meta, ok := node.Metadata.(map[string]interface{}); ok {
			if isActive, ok := meta["active"].(bool); ok && isActive {
				active = " (active)"
			}
		}

		fmt.Printf("%s%s %s%s\n", indent, icon, node.Label, active)
	}

	if node.Expanded || node.Type == models.TreeNodeTypeRoot {
		for _, child := range node.Children {
			printTree(child, depth+1)
		}
	}
}
