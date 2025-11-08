package models

import (
	"testing"
)

func TestNewTreeNode(t *testing.T) {
	node := NewTreeNode("db:postgres", TreeNodeTypeDatabase, "postgres")

	if node.ID != "db:postgres" {
		t.Errorf("Expected ID 'db:postgres', got '%s'", node.ID)
	}

	if node.Type != TreeNodeTypeDatabase {
		t.Errorf("Expected type Database, got %s", node.Type)
	}

	if node.Label != "postgres" {
		t.Errorf("Expected label 'postgres', got '%s'", node.Label)
	}

	if !node.Selectable {
		t.Error("Database node should be selectable")
	}

	if node.Expanded {
		t.Error("New node should not be expanded")
	}

	if node.Loaded {
		t.Error("New node should not be loaded")
	}
}

func TestAddChild(t *testing.T) {
	parent := NewTreeNode("db:postgres", TreeNodeTypeDatabase, "postgres")
	child := NewTreeNode("schema:postgres.public", TreeNodeTypeSchema, "public")

	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(parent.Children))
	}

	if child.Parent != parent {
		t.Error("Child's parent should be set")
	}
}

func TestToggle(t *testing.T) {
	node := NewTreeNode("db:postgres", TreeNodeTypeDatabase, "postgres")

	// Toggle on unloaded node should expand (for lazy loading)
	node.Toggle()
	if !node.Expanded {
		t.Error("Unloaded node should expand (for lazy loading)")
	}

	// Toggle again should collapse
	node.Toggle()
	if node.Expanded {
		t.Error("Node should collapse")
	}

	// Mark as loaded with no children
	node.Loaded = true
	node.Toggle()
	if node.Expanded {
		t.Error("Loaded node with no children should not expand")
	}

	// Add a child
	child := NewTreeNode("schema:postgres.public", TreeNodeTypeSchema, "public")
	node.AddChild(child)

	// Now it should toggle
	node.Toggle()
	if !node.Expanded {
		t.Error("Node with children should expand")
	}

	node.Toggle()
	if node.Expanded {
		t.Error("Node should collapse")
	}

	// Test that column nodes cannot be toggled
	colNode := NewTreeNode("column:postgres.public.users.id", TreeNodeTypeColumn, "id (integer)")
	colNode.Toggle()
	if colNode.Expanded {
		t.Error("Column nodes should never expand")
	}
}

func TestFlatten(t *testing.T) {
	// Build a simple tree:
	// root
	//   ├─ db1
	//   │   └─ schema1
	//   └─ db2
	root := NewTreeNode("root", TreeNodeTypeRoot, "Databases")
	root.Expanded = true

	db1 := NewTreeNode("db:db1", TreeNodeTypeDatabase, "db1")
	db2 := NewTreeNode("db:db2", TreeNodeTypeDatabase, "db2")
	root.AddChild(db1)
	root.AddChild(db2)

	schema1 := NewTreeNode("schema:db1.schema1", TreeNodeTypeSchema, "schema1")
	db1.AddChild(schema1)

	// Without expanding db1, should only see db1 and db2
	flat := root.Flatten()
	if len(flat) != 2 {
		t.Errorf("Expected 2 visible nodes, got %d", len(flat))
	}

	// Expand db1
	db1.Expanded = true
	flat = root.Flatten()
	if len(flat) != 3 {
		t.Errorf("Expected 3 visible nodes after expansion, got %d", len(flat))
	}

	// Verify order: db1, schema1, db2
	if flat[0] != db1 || flat[1] != schema1 || flat[2] != db2 {
		t.Error("Nodes in wrong order")
	}
}

func TestFindByID(t *testing.T) {
	root := BuildDatabaseTree([]string{"postgres", "mydb"}, "postgres")

	// Find postgres database
	node := root.FindByID("db:postgres")
	if node == nil {
		t.Fatal("Should find postgres node")
	}

	if node.Label != "postgres" {
		t.Errorf("Expected label 'postgres', got '%s'", node.Label)
	}

	// Find non-existent node
	node = root.FindByID("db:nonexistent")
	if node != nil {
		t.Error("Should not find non-existent node")
	}
}

func TestGetPath(t *testing.T) {
	// Build a deeper tree
	root := NewTreeNode("root", TreeNodeTypeRoot, "Databases")
	db := NewTreeNode("db:postgres", TreeNodeTypeDatabase, "postgres")
	schema := NewTreeNode("schema:postgres.public", TreeNodeTypeSchema, "public")
	table := NewTreeNode("table:postgres.public.users", TreeNodeTypeTable, "users")

	root.AddChild(db)
	db.AddChild(schema)
	schema.AddChild(table)

	path := table.GetPath()
	expected := []string{"postgres", "public", "users"}

	if len(path) != len(expected) {
		t.Fatalf("Expected path length %d, got %d", len(expected), len(path))
	}

	for i, p := range path {
		if p != expected[i] {
			t.Errorf("Expected path[%d] = '%s', got '%s'", i, expected[i], p)
		}
	}
}

func TestGetDepth(t *testing.T) {
	root := NewTreeNode("root", TreeNodeTypeRoot, "Databases")
	db := NewTreeNode("db:postgres", TreeNodeTypeDatabase, "postgres")
	schema := NewTreeNode("schema:postgres.public", TreeNodeTypeSchema, "public")

	root.AddChild(db)
	db.AddChild(schema)

	if root.GetDepth() != 0 {
		t.Errorf("Root depth should be 0, got %d", root.GetDepth())
	}

	if db.GetDepth() != 1 {
		t.Errorf("Database depth should be 1, got %d", db.GetDepth())
	}

	if schema.GetDepth() != 2 {
		t.Errorf("Schema depth should be 2, got %d", schema.GetDepth())
	}
}

func TestBuildDatabaseTree(t *testing.T) {
	databases := []string{"postgres", "mydb", "testdb"}
	root := BuildDatabaseTree(databases, "postgres")

	if root.Type != TreeNodeTypeRoot {
		t.Error("Root should be TreeNodeTypeRoot")
	}

	if !root.Expanded {
		t.Error("Root should be expanded by default")
	}

	if len(root.Children) != 3 {
		t.Errorf("Expected 3 database nodes, got %d", len(root.Children))
	}

	// Check postgres is marked as active
	pgNode := root.FindByID("db:postgres")
	if pgNode == nil {
		t.Fatal("Should find postgres node")
	}

	meta, ok := pgNode.Metadata.(map[string]interface{})
	if !ok {
		t.Fatal("Metadata should be a map")
	}

	if active, ok := meta["active"].(bool); !ok || !active {
		t.Error("Postgres should be marked as active")
	}
}

func TestRefreshTreeChildren(t *testing.T) {
	db := NewTreeNode("db:postgres", TreeNodeTypeDatabase, "postgres")

	// Add some initial children
	schema1 := NewTreeNode("schema:postgres.public", TreeNodeTypeSchema, "public")
	db.AddChild(schema1)

	if len(db.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(db.Children))
	}

	// Refresh with new children
	newSchemas := BuildSchemaNodes("postgres", []string{"public", "private", "test"})
	RefreshTreeChildren(db, newSchemas)

	if len(db.Children) != 3 {
		t.Errorf("Expected 3 children after refresh, got %d", len(db.Children))
	}

	if !db.Loaded {
		t.Error("Node should be marked as loaded")
	}
}

func TestBuildSchemaNodes(t *testing.T) {
	schemas := []string{"public", "private"}
	nodes := BuildSchemaNodes("postgres", schemas)

	if len(nodes) != 2 {
		t.Errorf("Expected 2 schema nodes, got %d", len(nodes))
	}

	if nodes[0].ID != "schema:postgres.public" {
		t.Errorf("Expected ID 'schema:postgres.public', got '%s'", nodes[0].ID)
	}

	if nodes[0].Type != TreeNodeTypeSchema {
		t.Error("Node should be schema type")
	}
}

func TestBuildTableNodes(t *testing.T) {
	tables := []string{"users", "posts"}
	nodes := BuildTableNodes("postgres", "public", tables)

	if len(nodes) != 2 {
		t.Errorf("Expected 2 table nodes, got %d", len(nodes))
	}

	if nodes[0].ID != "table:postgres.public.users" {
		t.Errorf("Expected ID 'table:postgres.public.users', got '%s'", nodes[0].ID)
	}
}

func TestBuildColumnNodes(t *testing.T) {
	columns := []ColumnInfo{
		{Name: "id", DataType: "integer", PrimaryKey: true, Nullable: false},
		{Name: "email", DataType: "varchar", PrimaryKey: false, Nullable: false},
	}

	nodes := BuildColumnNodes("postgres", "public", "users", columns)

	if len(nodes) != 2 {
		t.Errorf("Expected 2 column nodes, got %d", len(nodes))
	}

	if nodes[0].Type != TreeNodeTypeColumn {
		t.Error("Node should be column type")
	}

	if nodes[0].Selectable {
		t.Error("Column nodes should not be selectable")
	}

	// Check metadata
	meta, ok := nodes[0].Metadata.(ColumnInfo)
	if !ok {
		t.Fatal("Metadata should be ColumnInfo")
	}

	if !meta.PrimaryKey {
		t.Error("First column should be primary key")
	}
}

func TestParseNodeID(t *testing.T) {
	tests := []struct {
		id         string
		expectType string
		expectComp []string
	}{
		{"db:postgres", "db", []string{"postgres"}},
		{"schema:postgres.public", "schema", []string{"postgres", "public"}},
		{"table:postgres.public.users", "table", []string{"postgres", "public", "users"}},
		{"column:postgres.public.users.id", "column", []string{"postgres", "public", "users", "id"}},
	}

	for _, tt := range tests {
		nodeType, components := ParseNodeID(tt.id)

		if nodeType != tt.expectType {
			t.Errorf("For ID '%s': expected type '%s', got '%s'", tt.id, tt.expectType, nodeType)
		}

		if len(components) != len(tt.expectComp) {
			t.Errorf("For ID '%s': expected %d components, got %d", tt.id, len(tt.expectComp), len(components))
			continue
		}

		for i, comp := range components {
			if comp != tt.expectComp[i] {
				t.Errorf("For ID '%s': expected component[%d] = '%s', got '%s'", tt.id, i, tt.expectComp[i], comp)
			}
		}
	}
}

func TestGetDatabaseFromNode(t *testing.T) {
	root := NewTreeNode("root", TreeNodeTypeRoot, "Databases")
	db := NewTreeNode("db:postgres", TreeNodeTypeDatabase, "postgres")
	schema := NewTreeNode("schema:postgres.public", TreeNodeTypeSchema, "public")
	table := NewTreeNode("table:postgres.public.users", TreeNodeTypeTable, "users")

	root.AddChild(db)
	db.AddChild(schema)
	schema.AddChild(table)

	// Test from different levels
	if GetDatabaseFromNode(db) != "postgres" {
		t.Error("Should get database name from database node")
	}

	if GetDatabaseFromNode(schema) != "postgres" {
		t.Error("Should get database name from schema node")
	}

	if GetDatabaseFromNode(table) != "postgres" {
		t.Error("Should get database name from table node")
	}

	if GetDatabaseFromNode(root) != "" {
		t.Error("Should return empty string from root node")
	}
}

func TestGetSchemaFromNode(t *testing.T) {
	db := NewTreeNode("db:postgres", TreeNodeTypeDatabase, "postgres")
	schema := NewTreeNode("schema:postgres.public", TreeNodeTypeSchema, "public")
	table := NewTreeNode("table:postgres.public.users", TreeNodeTypeTable, "users")

	db.AddChild(schema)
	schema.AddChild(table)

	// Test from different levels
	if GetSchemaFromNode(schema) != "public" {
		t.Error("Should get schema name from schema node")
	}

	if GetSchemaFromNode(table) != "public" {
		t.Error("Should get schema name from table node")
	}

	if GetSchemaFromNode(db) != "" {
		t.Error("Should return empty string from database node")
	}
}

func TestIsAncestorOf(t *testing.T) {
	root := NewTreeNode("root", TreeNodeTypeRoot, "Databases")
	db := NewTreeNode("db:postgres", TreeNodeTypeDatabase, "postgres")
	schema := NewTreeNode("schema:postgres.public", TreeNodeTypeSchema, "public")
	table := NewTreeNode("table:postgres.public.users", TreeNodeTypeTable, "users")

	root.AddChild(db)
	db.AddChild(schema)
	schema.AddChild(table)

	if !root.IsAncestorOf(table) {
		t.Error("Root should be ancestor of table")
	}

	if !db.IsAncestorOf(table) {
		t.Error("Database should be ancestor of table")
	}

	if !schema.IsAncestorOf(table) {
		t.Error("Schema should be ancestor of table")
	}

	if table.IsAncestorOf(schema) {
		t.Error("Table should not be ancestor of schema")
	}

	if db.IsAncestorOf(db) {
		t.Error("Node should not be ancestor of itself")
	}
}
