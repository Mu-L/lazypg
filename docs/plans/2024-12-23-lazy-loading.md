# Tree Lazy Loading Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Reduce initial tree loading from 600+ queries to 1 aggregation query by implementing lazy loading.

**Architecture:** Initial load fetches only schema names and object counts via a single aggregation query. Tree displays skeleton with counts. When user expands a group node (e.g., "Tables (30)"), we fetch the actual objects for that group.

**Tech Stack:** Go, PostgreSQL system catalogs (pg_class, pg_namespace, pg_proc), Bubble Tea messages

---

### Task 1: Add SchemaObjectCounts type and GetSchemaObjectCounts function

**Files:**
- Modify: `internal/db/metadata/objects.go`

**Step 1: Add the SchemaObjectCounts struct**

Add at the top of objects.go after the existing type definitions:

```go
// SchemaObjectCounts holds the count of each object type in a schema
type SchemaObjectCounts struct {
	SchemaName        string
	Tables            int
	Views             int
	MaterializedViews int
	Sequences         int
	Functions         int
	Procedures        int
	TriggerFunctions  int
	CompositeTypes    int
	EnumTypes         int
	DomainTypes       int
	RangeTypes        int
}

// TotalObjects returns the total count of all objects in the schema
func (s *SchemaObjectCounts) TotalObjects() int {
	return s.Tables + s.Views + s.MaterializedViews + s.Sequences +
		s.Functions + s.Procedures + s.TriggerFunctions +
		s.CompositeTypes + s.EnumTypes + s.DomainTypes + s.RangeTypes
}
```

**Step 2: Add GetSchemaObjectCounts function**

Add after the struct definition:

```go
// GetSchemaObjectCounts returns object counts for all schemas in one query
func GetSchemaObjectCounts(ctx context.Context, pool *connection.Pool) ([]SchemaObjectCounts, error) {
	query := `
		WITH class_counts AS (
			SELECT
				n.nspname AS schema_name,
				SUM(CASE WHEN c.relkind = 'r' THEN 1 ELSE 0 END) AS tables,
				SUM(CASE WHEN c.relkind = 'v' THEN 1 ELSE 0 END) AS views,
				SUM(CASE WHEN c.relkind = 'm' THEN 1 ELSE 0 END) AS mat_views,
				SUM(CASE WHEN c.relkind = 'S' THEN 1 ELSE 0 END) AS sequences
			FROM pg_namespace n
			LEFT JOIN pg_class c ON c.relnamespace = n.oid
			WHERE n.nspname NOT LIKE 'pg_%'
			  AND n.nspname != 'information_schema'
			GROUP BY n.nspname
		),
		proc_counts AS (
			SELECT
				n.nspname AS schema_name,
				SUM(CASE WHEN p.prokind = 'f' AND p.prorettype != 'trigger'::regtype THEN 1 ELSE 0 END) AS functions,
				SUM(CASE WHEN p.prokind = 'p' THEN 1 ELSE 0 END) AS procedures,
				SUM(CASE WHEN p.prorettype = 'trigger'::regtype THEN 1 ELSE 0 END) AS trigger_funcs
			FROM pg_namespace n
			LEFT JOIN pg_proc p ON p.pronamespace = n.oid
			WHERE n.nspname NOT LIKE 'pg_%'
			  AND n.nspname != 'information_schema'
			GROUP BY n.nspname
		),
		type_counts AS (
			SELECT
				n.nspname AS schema_name,
				SUM(CASE WHEN t.typtype = 'c' THEN 1 ELSE 0 END) AS composite_types,
				SUM(CASE WHEN t.typtype = 'e' THEN 1 ELSE 0 END) AS enum_types,
				SUM(CASE WHEN t.typtype = 'd' THEN 1 ELSE 0 END) AS domain_types,
				SUM(CASE WHEN t.typtype = 'r' THEN 1 ELSE 0 END) AS range_types
			FROM pg_namespace n
			LEFT JOIN pg_type t ON t.typnamespace = n.oid
			WHERE n.nspname NOT LIKE 'pg_%'
			  AND n.nspname != 'information_schema'
			GROUP BY n.nspname
		)
		SELECT
			c.schema_name,
			COALESCE(c.tables, 0) AS tables,
			COALESCE(c.views, 0) AS views,
			COALESCE(c.mat_views, 0) AS mat_views,
			COALESCE(c.sequences, 0) AS sequences,
			COALESCE(p.functions, 0) AS functions,
			COALESCE(p.procedures, 0) AS procedures,
			COALESCE(p.trigger_funcs, 0) AS trigger_funcs,
			COALESCE(t.composite_types, 0) AS composite_types,
			COALESCE(t.enum_types, 0) AS enum_types,
			COALESCE(t.domain_types, 0) AS domain_types,
			COALESCE(t.range_types, 0) AS range_types
		FROM class_counts c
		LEFT JOIN proc_counts p ON c.schema_name = p.schema_name
		LEFT JOIN type_counts t ON c.schema_name = t.schema_name
		ORDER BY c.schema_name;
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema object counts: %w", err)
	}

	counts := make([]SchemaObjectCounts, 0, len(rows))
	for _, row := range rows {
		counts = append(counts, SchemaObjectCounts{
			SchemaName:        toString(row["schema_name"]),
			Tables:            int(toInt64(row["tables"])),
			Views:             int(toInt64(row["views"])),
			MaterializedViews: int(toInt64(row["mat_views"])),
			Sequences:         int(toInt64(row["sequences"])),
			Functions:         int(toInt64(row["functions"])),
			Procedures:        int(toInt64(row["procedures"])),
			TriggerFunctions:  int(toInt64(row["trigger_funcs"])),
			CompositeTypes:    int(toInt64(row["composite_types"])),
			EnumTypes:         int(toInt64(row["enum_types"])),
			DomainTypes:       int(toInt64(row["domain_types"])),
			RangeTypes:        int(toInt64(row["range_types"])),
		})
	}

	return counts, nil
}
```

**Step 3: Build and verify**

Run: `go build ./...`
Expected: Build succeeds

**Step 4: Commit**

```bash
git add internal/db/metadata/objects.go
git commit -m "feat(metadata): add GetSchemaObjectCounts aggregation query"
```

---

### Task 2: Add LoadNodeChildrenMsg and NodeChildrenLoadedMsg types

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add new message types**

Find the existing message type definitions (around line 179-200) and add:

```go
// LoadNodeChildrenMsg requests loading children for a tree node
type LoadNodeChildrenMsg struct {
	NodeID string
}

// NodeChildrenLoadedMsg is sent when node children are loaded
type NodeChildrenLoadedMsg struct {
	NodeID   string
	Children []*models.TreeNode
	Err      error
}
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): add LoadNodeChildrenMsg and NodeChildrenLoadedMsg types"
```

---

### Task 3: Refactor loadTree to build skeleton tree with counts

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Replace the loadTree function**

Find the `loadTree` function (around line 3493) and replace it entirely:

```go
// loadTree loads the database structure skeleton with object counts (fast)
func (a *App) loadTree() tea.Msg {
	ctx := context.Background()

	conn, err := a.connectionManager.GetActive()
	if err != nil {
		return TreeLoadedMsg{Err: fmt.Errorf("no active connection: %w", err)}
	}

	currentDB := conn.Config.Database

	// Build root with database node
	root := models.BuildDatabaseTree([]string{currentDB}, currentDB)

	// Load extensions (usually fast, small number)
	extensions, _ := metadata.ListExtensions(ctx, conn.Pool)

	// Get all schema object counts in ONE query
	schemaCounts, err := metadata.GetSchemaObjectCounts(ctx, conn.Pool)
	if err != nil {
		return TreeLoadedMsg{Err: fmt.Errorf("failed to load schema counts: %w", err)}
	}

	dbNode := root.FindByID(fmt.Sprintf("db:%s", currentDB))
	if dbNode == nil {
		return TreeLoadedMsg{Root: root}
	}

	// Add extensions group
	if len(extensions) > 0 {
		extGroup := models.NewTreeNode(
			fmt.Sprintf("extensions:%s", currentDB),
			models.TreeNodeTypeExtensionGroup,
			fmt.Sprintf("Extensions (%d)", len(extensions)),
		)
		extGroup.Selectable = false
		for _, ext := range extensions {
			extNode := models.NewTreeNode(
				fmt.Sprintf("extension:%s.%s", currentDB, ext.Name),
				models.TreeNodeTypeExtension,
				fmt.Sprintf("%s v%s", ext.Name, ext.Version),
			)
			extNode.Selectable = true
			extNode.Metadata = ext
			extNode.Loaded = true
			extGroup.AddChild(extNode)
		}
		extGroup.Loaded = true
		dbNode.AddChild(extGroup)
	}

	// Build skeleton tree with counts (no actual objects loaded)
	for _, sc := range schemaCounts {
		if sc.TotalObjects() == 0 {
			continue
		}

		schemaNode := models.NewTreeNode(
			fmt.Sprintf("schema:%s.%s", currentDB, sc.SchemaName),
			models.TreeNodeTypeSchema,
			sc.SchemaName,
		)
		schemaNode.Selectable = true

		// Add group nodes with counts - Loaded=false means lazy load on expand
		if sc.Tables > 0 {
			tablesGroup := models.NewTreeNode(
				fmt.Sprintf("tables:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeTableGroup,
				fmt.Sprintf("Tables (%d)", sc.Tables),
			)
			tablesGroup.Selectable = false
			tablesGroup.Loaded = false // Will lazy load
			schemaNode.AddChild(tablesGroup)
		}

		if sc.Views > 0 {
			viewsGroup := models.NewTreeNode(
				fmt.Sprintf("views:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeViewGroup,
				fmt.Sprintf("Views (%d)", sc.Views),
			)
			viewsGroup.Selectable = false
			viewsGroup.Loaded = false
			schemaNode.AddChild(viewsGroup)
		}

		if sc.MaterializedViews > 0 {
			matViewsGroup := models.NewTreeNode(
				fmt.Sprintf("matviews:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeMaterializedViewGroup,
				fmt.Sprintf("Materialized Views (%d)", sc.MaterializedViews),
			)
			matViewsGroup.Selectable = false
			matViewsGroup.Loaded = false
			schemaNode.AddChild(matViewsGroup)
		}

		if sc.Functions > 0 {
			funcsGroup := models.NewTreeNode(
				fmt.Sprintf("functions:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeFunctionGroup,
				fmt.Sprintf("Functions (%d)", sc.Functions),
			)
			funcsGroup.Selectable = false
			funcsGroup.Loaded = false
			schemaNode.AddChild(funcsGroup)
		}

		if sc.Procedures > 0 {
			procsGroup := models.NewTreeNode(
				fmt.Sprintf("procedures:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeProcedureGroup,
				fmt.Sprintf("Procedures (%d)", sc.Procedures),
			)
			procsGroup.Selectable = false
			procsGroup.Loaded = false
			schemaNode.AddChild(procsGroup)
		}

		if sc.TriggerFunctions > 0 {
			trigFuncsGroup := models.NewTreeNode(
				fmt.Sprintf("triggerfuncs:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeTriggerFunctionGroup,
				fmt.Sprintf("Trigger Functions (%d)", sc.TriggerFunctions),
			)
			trigFuncsGroup.Selectable = false
			trigFuncsGroup.Loaded = false
			schemaNode.AddChild(trigFuncsGroup)
		}

		if sc.Sequences > 0 {
			seqsGroup := models.NewTreeNode(
				fmt.Sprintf("sequences:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeSequenceGroup,
				fmt.Sprintf("Sequences (%d)", sc.Sequences),
			)
			seqsGroup.Selectable = false
			seqsGroup.Loaded = false
			schemaNode.AddChild(seqsGroup)
		}

		if sc.CompositeTypes > 0 {
			compTypesGroup := models.NewTreeNode(
				fmt.Sprintf("compositetypes:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeCompositeTypeGroup,
				fmt.Sprintf("Composite Types (%d)", sc.CompositeTypes),
			)
			compTypesGroup.Selectable = false
			compTypesGroup.Loaded = false
			schemaNode.AddChild(compTypesGroup)
		}

		if sc.EnumTypes > 0 {
			enumTypesGroup := models.NewTreeNode(
				fmt.Sprintf("enumtypes:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeEnumTypeGroup,
				fmt.Sprintf("Enum Types (%d)", sc.EnumTypes),
			)
			enumTypesGroup.Selectable = false
			enumTypesGroup.Loaded = false
			schemaNode.AddChild(enumTypesGroup)
		}

		if sc.DomainTypes > 0 {
			domainTypesGroup := models.NewTreeNode(
				fmt.Sprintf("domaintypes:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeDomainTypeGroup,
				fmt.Sprintf("Domain Types (%d)", sc.DomainTypes),
			)
			domainTypesGroup.Selectable = false
			domainTypesGroup.Loaded = false
			schemaNode.AddChild(domainTypesGroup)
		}

		if sc.RangeTypes > 0 {
			rangeTypesGroup := models.NewTreeNode(
				fmt.Sprintf("rangetypes:%s.%s", currentDB, sc.SchemaName),
				models.TreeNodeTypeRangeTypeGroup,
				fmt.Sprintf("Range Types (%d)", sc.RangeTypes),
			)
			rangeTypesGroup.Selectable = false
			rangeTypesGroup.Loaded = false
			schemaNode.AddChild(rangeTypesGroup)
		}

		schemaNode.Loaded = true
		dbNode.AddChild(schemaNode)
	}

	return TreeLoadedMsg{Root: root}
}
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "refactor(app): loadTree now builds skeleton with counts only"
```

---

### Task 4: Add loadNodeChildren function

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add loadNodeChildren function**

Add after the loadTree function:

```go
// loadNodeChildren loads children for a specific node (lazy loading)
func (a *App) loadNodeChildren(nodeID string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		conn, err := a.connectionManager.GetActive()
		if err != nil {
			return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
		}

		// Parse node ID to determine what to load
		// Format: "type:db.schema" or "type:db.schema.table"
		node := a.treeView.Root.FindByID(nodeID)
		if node == nil {
			return NodeChildrenLoadedMsg{NodeID: nodeID, Err: fmt.Errorf("node not found: %s", nodeID)}
		}

		var children []*models.TreeNode
		currentDB := conn.Config.Database

		switch node.Type {
		case models.TreeNodeTypeTableGroup:
			schema := extractSchemaFromNodeID(nodeID)
			tables, err := metadata.ListTables(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, table := range tables {
				tableNode := models.NewTreeNode(
					fmt.Sprintf("table:%s.%s.%s", currentDB, schema, table.Name),
					models.TreeNodeTypeTable,
					table.Name,
				)
				tableNode.Selectable = true
				tableNode.Loaded = false // Has indexes/triggers to lazy load
				children = append(children, tableNode)
			}

		case models.TreeNodeTypeViewGroup:
			schema := extractSchemaFromNodeID(nodeID)
			views, err := metadata.ListViews(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, view := range views {
				viewNode := models.NewTreeNode(
					fmt.Sprintf("view:%s.%s.%s", currentDB, schema, view.Name),
					models.TreeNodeTypeView,
					view.Name,
				)
				viewNode.Selectable = true
				viewNode.Loaded = true
				children = append(children, viewNode)
			}

		case models.TreeNodeTypeMaterializedViewGroup:
			schema := extractSchemaFromNodeID(nodeID)
			matViews, err := metadata.ListMaterializedViews(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, mv := range matViews {
				mvNode := models.NewTreeNode(
					fmt.Sprintf("matview:%s.%s.%s", currentDB, schema, mv.Name),
					models.TreeNodeTypeMaterializedView,
					mv.Name,
				)
				mvNode.Selectable = true
				mvNode.Loaded = true
				children = append(children, mvNode)
			}

		case models.TreeNodeTypeFunctionGroup:
			schema := extractSchemaFromNodeID(nodeID)
			funcs, err := metadata.ListFunctions(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, f := range funcs {
				label := f.Name
				if f.Arguments != "" {
					label = fmt.Sprintf("%s(%s)", f.Name, f.Arguments)
				}
				funcNode := models.NewTreeNode(
					fmt.Sprintf("function:%s.%s.%s(%s)", currentDB, schema, f.Name, f.Arguments),
					models.TreeNodeTypeFunction,
					label,
				)
				funcNode.Selectable = true
				funcNode.Metadata = f
				funcNode.Loaded = true
				children = append(children, funcNode)
			}

		case models.TreeNodeTypeProcedureGroup:
			schema := extractSchemaFromNodeID(nodeID)
			procs, err := metadata.ListProcedures(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, p := range procs {
				label := p.Name
				if p.Arguments != "" {
					label = fmt.Sprintf("%s(%s)", p.Name, p.Arguments)
				}
				procNode := models.NewTreeNode(
					fmt.Sprintf("procedure:%s.%s.%s(%s)", currentDB, schema, p.Name, p.Arguments),
					models.TreeNodeTypeProcedure,
					label,
				)
				procNode.Selectable = true
				procNode.Metadata = p
				procNode.Loaded = true
				children = append(children, procNode)
			}

		case models.TreeNodeTypeTriggerFunctionGroup:
			schema := extractSchemaFromNodeID(nodeID)
			trigFuncs, err := metadata.ListTriggerFunctions(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, tf := range trigFuncs {
				tfNode := models.NewTreeNode(
					fmt.Sprintf("triggerfunc:%s.%s.%s", currentDB, schema, tf.Name),
					models.TreeNodeTypeTriggerFunction,
					tf.Name,
				)
				tfNode.Selectable = true
				tfNode.Metadata = tf
				tfNode.Loaded = true
				children = append(children, tfNode)
			}

		case models.TreeNodeTypeSequenceGroup:
			schema := extractSchemaFromNodeID(nodeID)
			seqs, err := metadata.ListSequences(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, seq := range seqs {
				seqNode := models.NewTreeNode(
					fmt.Sprintf("sequence:%s.%s.%s", currentDB, schema, seq.Name),
					models.TreeNodeTypeSequence,
					seq.Name,
				)
				seqNode.Selectable = true
				seqNode.Metadata = seq
				seqNode.Loaded = true
				children = append(children, seqNode)
			}

		case models.TreeNodeTypeCompositeTypeGroup:
			schema := extractSchemaFromNodeID(nodeID)
			compTypes, err := metadata.ListCompositeTypes(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, ct := range compTypes {
				ctNode := models.NewTreeNode(
					fmt.Sprintf("compositetype:%s.%s.%s", currentDB, schema, ct.Name),
					models.TreeNodeTypeCompositeType,
					ct.Name,
				)
				ctNode.Selectable = true
				ctNode.Metadata = ct
				ctNode.Loaded = true
				children = append(children, ctNode)
			}

		case models.TreeNodeTypeEnumTypeGroup:
			schema := extractSchemaFromNodeID(nodeID)
			enumTypes, err := metadata.ListEnumTypes(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, et := range enumTypes {
				etNode := models.NewTreeNode(
					fmt.Sprintf("enumtype:%s.%s.%s", currentDB, schema, et.Name),
					models.TreeNodeTypeEnumType,
					et.Name,
				)
				etNode.Selectable = true
				etNode.Metadata = et
				etNode.Loaded = true
				children = append(children, etNode)
			}

		case models.TreeNodeTypeDomainTypeGroup:
			schema := extractSchemaFromNodeID(nodeID)
			domainTypes, err := metadata.ListDomainTypes(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, dt := range domainTypes {
				dtNode := models.NewTreeNode(
					fmt.Sprintf("domaintype:%s.%s.%s", currentDB, schema, dt.Name),
					models.TreeNodeTypeDomainType,
					dt.Name,
				)
				dtNode.Selectable = true
				dtNode.Metadata = dt
				dtNode.Loaded = true
				children = append(children, dtNode)
			}

		case models.TreeNodeTypeRangeTypeGroup:
			schema := extractSchemaFromNodeID(nodeID)
			rangeTypes, err := metadata.ListRangeTypes(ctx, conn.Pool, schema)
			if err != nil {
				return NodeChildrenLoadedMsg{NodeID: nodeID, Err: err}
			}
			for _, rt := range rangeTypes {
				rtNode := models.NewTreeNode(
					fmt.Sprintf("rangetype:%s.%s.%s", currentDB, schema, rt.Name),
					models.TreeNodeTypeRangeType,
					rt.Name,
				)
				rtNode.Selectable = true
				rtNode.Metadata = rt
				rtNode.Loaded = true
				children = append(children, rtNode)
			}

		case models.TreeNodeTypeTable:
			// Load indexes and triggers for a table
			schema, table := extractSchemaAndTableFromNodeID(nodeID)
			indexes, _ := metadata.ListTableIndexes(ctx, conn.Pool, schema, table)
			triggers, _ := metadata.ListTableTriggers(ctx, conn.Pool, schema, table)

			if len(indexes) > 0 {
				indexGroup := models.NewTreeNode(
					fmt.Sprintf("indexes:%s.%s.%s", currentDB, schema, table),
					models.TreeNodeTypeIndexGroup,
					fmt.Sprintf("Indexes (%d)", len(indexes)),
				)
				indexGroup.Selectable = false
				for _, idx := range indexes {
					idxNode := models.NewTreeNode(
						fmt.Sprintf("index:%s.%s.%s.%s", currentDB, schema, table, idx.Name),
						models.TreeNodeTypeIndex,
						idx.Name,
					)
					idxNode.Selectable = true
					idxNode.Metadata = idx
					idxNode.Loaded = true
					indexGroup.AddChild(idxNode)
				}
				indexGroup.Loaded = true
				children = append(children, indexGroup)
			}

			if len(triggers) > 0 {
				triggerGroup := models.NewTreeNode(
					fmt.Sprintf("triggers:%s.%s.%s", currentDB, schema, table),
					models.TreeNodeTypeTriggerGroup,
					fmt.Sprintf("Triggers (%d)", len(triggers)),
				)
				triggerGroup.Selectable = false
				for _, trg := range triggers {
					trgNode := models.NewTreeNode(
						fmt.Sprintf("trigger:%s.%s.%s.%s", currentDB, schema, table, trg.Name),
						models.TreeNodeTypeTrigger,
						trg.Name,
					)
					trgNode.Selectable = true
					trgNode.Metadata = trg
					trgNode.Loaded = true
					triggerGroup.AddChild(trgNode)
				}
				triggerGroup.Loaded = true
				children = append(children, triggerGroup)
			}
		}

		return NodeChildrenLoadedMsg{NodeID: nodeID, Children: children}
	}
}

// extractSchemaFromNodeID extracts schema name from node ID like "tables:db.schema"
func extractSchemaFromNodeID(nodeID string) string {
	parts := strings.Split(nodeID, ":")
	if len(parts) < 2 {
		return ""
	}
	dbSchema := strings.Split(parts[1], ".")
	if len(dbSchema) < 2 {
		return ""
	}
	return dbSchema[1]
}

// extractSchemaAndTableFromNodeID extracts schema and table from node ID like "table:db.schema.table"
func extractSchemaAndTableFromNodeID(nodeID string) (string, string) {
	parts := strings.Split(nodeID, ":")
	if len(parts) < 2 {
		return "", ""
	}
	dbSchemaTable := strings.Split(parts[1], ".")
	if len(dbSchemaTable) < 3 {
		return "", ""
	}
	return dbSchemaTable[1], dbSchemaTable[2]
}
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): add loadNodeChildren for lazy loading"
```

---

### Task 5: Add message handlers for lazy loading

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Add LoadNodeChildrenMsg handler**

Find the Update function's switch statement and add after the TreeLoadedMsg case:

```go
	case LoadNodeChildrenMsg:
		a.treeView.LoadingNodeID = msg.NodeID
		return a, tea.Batch(a.loadNodeChildren(msg.NodeID), a.executeSpinner.Tick)

	case NodeChildrenLoadedMsg:
		a.treeView.LoadingNodeID = ""
		if msg.Err != nil {
			a.ShowError("Load Error", fmt.Sprintf("Failed to load children:\n\n%v", msg.Err))
			return a, nil
		}
		// Find the node and add children
		node := a.treeView.Root.FindByID(msg.NodeID)
		if node != nil {
			for _, child := range msg.Children {
				child.Parent = node
				node.AddChild(child)
			}
			node.Loaded = true
			node.Expanded = true
		}
		return a, nil
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): add handlers for LoadNodeChildrenMsg and NodeChildrenLoadedMsg"
```

---

### Task 6: Modify TreeNodeExpandedMsg handler to trigger lazy loading

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Find and update TreeNodeExpandedMsg handler**

Find the `case components.TreeNodeExpandedMsg:` handler and update it to check for lazy loading:

```go
	case components.TreeNodeExpandedMsg:
		// Check if this node needs lazy loading
		if msg.Expanded && msg.Node != nil && !msg.Node.Loaded && len(msg.Node.Children) == 0 {
			// Trigger lazy load
			return a, func() tea.Msg {
				return LoadNodeChildrenMsg{NodeID: msg.Node.ID}
			}
		}
		return a, nil
```

**Step 2: Build and verify**

Run: `go build ./...`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(app): trigger lazy loading on node expand"
```

---

### Task 7: Manual testing

**Step 1: Build the application**

```bash
make build
```

**Step 2: Test with a database**

1. Run `./bin/lazypg`
2. Connect to a database
3. Verify initial load is fast (< 1 second)
4. Verify tree shows schema + group counts
5. Expand a "Tables (N)" group
6. Verify loading spinner appears
7. Verify tables load correctly

**Step 3: Commit any fixes if needed**

```bash
git add -A
git commit -m "fix: address issues found in manual testing"
```
