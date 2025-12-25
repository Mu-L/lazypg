# Search Index Preload Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Pre-populate tree nodes with all object names at connection time so tree search works without manual expansion.

**Architecture:** Replace `GetSchemaObjectCounts()` with `GetAllSchemaObjects()` that returns object names. Modify `loadTree()` to create actual object nodes instead of empty group nodes.

**Tech Stack:** Go, pgx, Bubble Tea

---

## Task 1: Add GetAllSchemaObjects Query

**Files:**
- Modify: `internal/db/metadata/objects.go`

**Step 1: Add SchemaObject struct after SchemaObjectCounts (line ~112)**

```go
// SchemaObject represents a single database object for search indexing
type SchemaObject struct {
	SchemaName string
	ObjectType string // "table", "view", "matview", "function", "procedure", "trigger_function", "sequence", "composite_type", "enum_type", "domain_type", "range_type"
	ObjectName string
}
```

**Step 2: Add GetAllSchemaObjects function after GetSchemaObjectCounts (line ~853)**

```go
// GetAllSchemaObjects returns all object names grouped by schema and type
func GetAllSchemaObjects(ctx context.Context, pool *connection.Pool) ([]SchemaObject, error) {
	query := `
		-- Tables
		SELECT n.nspname AS schema_name, 'table' AS object_type, c.relname AS object_name
		FROM pg_class c
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE c.relkind = 'r'
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		UNION ALL

		-- Views
		SELECT n.nspname, 'view', c.relname
		FROM pg_class c
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE c.relkind = 'v'
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		UNION ALL

		-- Materialized Views
		SELECT n.nspname, 'matview', c.relname
		FROM pg_class c
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE c.relkind = 'm'
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		UNION ALL

		-- Sequences
		SELECT n.nspname, 'sequence', c.relname
		FROM pg_class c
		JOIN pg_namespace n ON c.relnamespace = n.oid
		WHERE c.relkind = 'S'
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		UNION ALL

		-- Functions (excluding trigger functions)
		SELECT n.nspname, 'function', p.proname
		FROM pg_proc p
		JOIN pg_namespace n ON p.pronamespace = n.oid
		WHERE p.prokind = 'f'
		  AND p.prorettype != 'trigger'::regtype
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		UNION ALL

		-- Procedures
		SELECT n.nspname, 'procedure', p.proname
		FROM pg_proc p
		JOIN pg_namespace n ON p.pronamespace = n.oid
		WHERE p.prokind = 'p'
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		UNION ALL

		-- Trigger Functions
		SELECT n.nspname, 'trigger_function', p.proname
		FROM pg_proc p
		JOIN pg_namespace n ON p.pronamespace = n.oid
		WHERE p.prorettype = 'trigger'::regtype
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		UNION ALL

		-- Composite Types
		SELECT n.nspname, 'composite_type', t.typname
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		LEFT JOIN pg_class c ON t.typrelid = c.oid
		WHERE t.typtype = 'c'
		  AND (c.relkind IS NULL OR c.relkind = 'c')
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		UNION ALL

		-- Enum Types
		SELECT n.nspname, 'enum_type', t.typname
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		WHERE t.typtype = 'e'
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		UNION ALL

		-- Domain Types
		SELECT n.nspname, 'domain_type', t.typname
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		WHERE t.typtype = 'd'
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		UNION ALL

		-- Range Types
		SELECT n.nspname, 'range_type', t.typname
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		WHERE t.typtype = 'r'
		  AND n.nspname NOT LIKE 'pg_%'
		  AND n.nspname != 'information_schema'

		ORDER BY schema_name, object_type, object_name;
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema objects: %w", err)
	}

	objects := make([]SchemaObject, 0, len(rows))
	for _, row := range rows {
		objects = append(objects, SchemaObject{
			SchemaName: toString(row["schema_name"]),
			ObjectType: toString(row["object_type"]),
			ObjectName: toString(row["object_name"]),
		})
	}

	return objects, nil
}
```

**Step 3: Build and verify no errors**

Run: `go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add internal/db/metadata/objects.go
git commit -m "feat(metadata): add GetAllSchemaObjects query for search preload"
```

---

## Task 2: Refactor loadTree to Pre-populate Nodes

**Files:**
- Modify: `internal/app/app.go:3568-3758` (loadTree function)

**Step 1: Replace GetSchemaObjectCounts with GetAllSchemaObjects**

Change line ~3585 from:
```go
schemaCounts, err := metadata.GetSchemaObjectCounts(ctx, conn.Pool)
```
to:
```go
schemaObjects, err := metadata.GetAllSchemaObjects(ctx, conn.Pool)
```

**Step 2: Add helper struct to organize objects by schema and type**

Add before the loop (after getting schemaObjects):
```go
// Organize objects by schema -> type -> names
type schemaData struct {
	tables           []string
	views            []string
	matViews         []string
	sequences        []string
	functions        []string
	procedures       []string
	triggerFunctions []string
	compositeTypes   []string
	enumTypes        []string
	domainTypes      []string
	rangeTypes       []string
}
schemaMap := make(map[string]*schemaData)

for _, obj := range schemaObjects {
	sd, ok := schemaMap[obj.SchemaName]
	if !ok {
		sd = &schemaData{}
		schemaMap[obj.SchemaName] = sd
	}
	switch obj.ObjectType {
	case "table":
		sd.tables = append(sd.tables, obj.ObjectName)
	case "view":
		sd.views = append(sd.views, obj.ObjectName)
	case "matview":
		sd.matViews = append(sd.matViews, obj.ObjectName)
	case "sequence":
		sd.sequences = append(sd.sequences, obj.ObjectName)
	case "function":
		sd.functions = append(sd.functions, obj.ObjectName)
	case "procedure":
		sd.procedures = append(sd.procedures, obj.ObjectName)
	case "trigger_function":
		sd.triggerFunctions = append(sd.triggerFunctions, obj.ObjectName)
	case "composite_type":
		sd.compositeTypes = append(sd.compositeTypes, obj.ObjectName)
	case "enum_type":
		sd.enumTypes = append(sd.enumTypes, obj.ObjectName)
	case "domain_type":
		sd.domainTypes = append(sd.domainTypes, obj.ObjectName)
	case "range_type":
		sd.rangeTypes = append(sd.rangeTypes, obj.ObjectName)
	}
}
```

**Step 3: Rewrite the schema loop to create actual object nodes**

Replace the entire `for _, sc := range schemaCounts` loop with:

```go
// Build tree with actual object nodes
for schemaName, sd := range schemaMap {
	schemaNode := models.NewTreeNode(
		fmt.Sprintf("schema:%s.%s", currentDB, schemaName),
		models.TreeNodeTypeSchema,
		schemaName,
	)
	schemaNode.Selectable = true

	// Tables group with actual table nodes
	if len(sd.tables) > 0 {
		tablesGroup := models.NewTreeNode(
			fmt.Sprintf("tables:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeTableGroup,
			fmt.Sprintf("Tables (%d)", len(sd.tables)),
		)
		tablesGroup.Selectable = false
		for _, tableName := range sd.tables {
			tableNode := models.NewTreeNode(
				fmt.Sprintf("table:%s.%s.%s", currentDB, schemaName, tableName),
				models.TreeNodeTypeTable,
				tableName,
			)
			tableNode.Selectable = true
			tableNode.Loaded = false // Columns/indexes still lazy load
			tablesGroup.AddChild(tableNode)
		}
		tablesGroup.Loaded = true
		schemaNode.AddChild(tablesGroup)
	}

	// Views group
	if len(sd.views) > 0 {
		viewsGroup := models.NewTreeNode(
			fmt.Sprintf("views:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeViewGroup,
			fmt.Sprintf("Views (%d)", len(sd.views)),
		)
		viewsGroup.Selectable = false
		for _, viewName := range sd.views {
			viewNode := models.NewTreeNode(
				fmt.Sprintf("view:%s.%s.%s", currentDB, schemaName, viewName),
				models.TreeNodeTypeView,
				viewName,
			)
			viewNode.Selectable = true
			viewNode.Loaded = true
			viewsGroup.AddChild(viewNode)
		}
		viewsGroup.Loaded = true
		schemaNode.AddChild(viewsGroup)
	}

	// Materialized Views group
	if len(sd.matViews) > 0 {
		matViewsGroup := models.NewTreeNode(
			fmt.Sprintf("matviews:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeMaterializedViewGroup,
			fmt.Sprintf("Materialized Views (%d)", len(sd.matViews)),
		)
		matViewsGroup.Selectable = false
		for _, mvName := range sd.matViews {
			mvNode := models.NewTreeNode(
				fmt.Sprintf("matview:%s.%s.%s", currentDB, schemaName, mvName),
				models.TreeNodeTypeMaterializedView,
				mvName,
			)
			mvNode.Selectable = true
			mvNode.Loaded = true
			matViewsGroup.AddChild(mvNode)
		}
		matViewsGroup.Loaded = true
		schemaNode.AddChild(matViewsGroup)
	}

	// Functions group
	if len(sd.functions) > 0 {
		funcsGroup := models.NewTreeNode(
			fmt.Sprintf("functions:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeFunctionGroup,
			fmt.Sprintf("Functions (%d)", len(sd.functions)),
		)
		funcsGroup.Selectable = false
		for _, funcName := range sd.functions {
			funcNode := models.NewTreeNode(
				fmt.Sprintf("function:%s.%s.%s", currentDB, schemaName, funcName),
				models.TreeNodeTypeFunction,
				funcName,
			)
			funcNode.Selectable = true
			funcNode.Loaded = true
			funcsGroup.AddChild(funcNode)
		}
		funcsGroup.Loaded = true
		schemaNode.AddChild(funcsGroup)
	}

	// Procedures group
	if len(sd.procedures) > 0 {
		procsGroup := models.NewTreeNode(
			fmt.Sprintf("procedures:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeProcedureGroup,
			fmt.Sprintf("Procedures (%d)", len(sd.procedures)),
		)
		procsGroup.Selectable = false
		for _, procName := range sd.procedures {
			procNode := models.NewTreeNode(
				fmt.Sprintf("procedure:%s.%s.%s", currentDB, schemaName, procName),
				models.TreeNodeTypeProcedure,
				procName,
			)
			procNode.Selectable = true
			procNode.Loaded = true
			procsGroup.AddChild(procNode)
		}
		procsGroup.Loaded = true
		schemaNode.AddChild(procsGroup)
	}

	// Trigger Functions group
	if len(sd.triggerFunctions) > 0 {
		trigFuncsGroup := models.NewTreeNode(
			fmt.Sprintf("triggerfuncs:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeTriggerFunctionGroup,
			fmt.Sprintf("Trigger Functions (%d)", len(sd.triggerFunctions)),
		)
		trigFuncsGroup.Selectable = false
		for _, tfName := range sd.triggerFunctions {
			tfNode := models.NewTreeNode(
				fmt.Sprintf("triggerfunction:%s.%s.%s", currentDB, schemaName, tfName),
				models.TreeNodeTypeTriggerFunction,
				tfName,
			)
			tfNode.Selectable = true
			tfNode.Loaded = true
			trigFuncsGroup.AddChild(tfNode)
		}
		trigFuncsGroup.Loaded = true
		schemaNode.AddChild(trigFuncsGroup)
	}

	// Sequences group
	if len(sd.sequences) > 0 {
		seqsGroup := models.NewTreeNode(
			fmt.Sprintf("sequences:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeSequenceGroup,
			fmt.Sprintf("Sequences (%d)", len(sd.sequences)),
		)
		seqsGroup.Selectable = false
		for _, seqName := range sd.sequences {
			seqNode := models.NewTreeNode(
				fmt.Sprintf("sequence:%s.%s.%s", currentDB, schemaName, seqName),
				models.TreeNodeTypeSequence,
				seqName,
			)
			seqNode.Selectable = true
			seqNode.Loaded = true
			seqsGroup.AddChild(seqNode)
		}
		seqsGroup.Loaded = true
		schemaNode.AddChild(seqsGroup)
	}

	// Composite Types group
	if len(sd.compositeTypes) > 0 {
		compTypesGroup := models.NewTreeNode(
			fmt.Sprintf("compositetypes:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeCompositeTypeGroup,
			fmt.Sprintf("Composite Types (%d)", len(sd.compositeTypes)),
		)
		compTypesGroup.Selectable = false
		for _, ctName := range sd.compositeTypes {
			ctNode := models.NewTreeNode(
				fmt.Sprintf("compositetype:%s.%s.%s", currentDB, schemaName, ctName),
				models.TreeNodeTypeCompositeType,
				ctName,
			)
			ctNode.Selectable = true
			ctNode.Loaded = true
			compTypesGroup.AddChild(ctNode)
		}
		compTypesGroup.Loaded = true
		schemaNode.AddChild(compTypesGroup)
	}

	// Enum Types group
	if len(sd.enumTypes) > 0 {
		enumTypesGroup := models.NewTreeNode(
			fmt.Sprintf("enumtypes:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeEnumTypeGroup,
			fmt.Sprintf("Enum Types (%d)", len(sd.enumTypes)),
		)
		enumTypesGroup.Selectable = false
		for _, etName := range sd.enumTypes {
			etNode := models.NewTreeNode(
				fmt.Sprintf("enumtype:%s.%s.%s", currentDB, schemaName, etName),
				models.TreeNodeTypeEnumType,
				etName,
			)
			etNode.Selectable = true
			etNode.Loaded = true
			enumTypesGroup.AddChild(etNode)
		}
		enumTypesGroup.Loaded = true
		schemaNode.AddChild(enumTypesGroup)
	}

	// Domain Types group
	if len(sd.domainTypes) > 0 {
		domainTypesGroup := models.NewTreeNode(
			fmt.Sprintf("domaintypes:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeDomainTypeGroup,
			fmt.Sprintf("Domain Types (%d)", len(sd.domainTypes)),
		)
		domainTypesGroup.Selectable = false
		for _, dtName := range sd.domainTypes {
			dtNode := models.NewTreeNode(
				fmt.Sprintf("domaintype:%s.%s.%s", currentDB, schemaName, dtName),
				models.TreeNodeTypeDomainType,
				dtName,
			)
			dtNode.Selectable = true
			dtNode.Loaded = true
			domainTypesGroup.AddChild(dtNode)
		}
		domainTypesGroup.Loaded = true
		schemaNode.AddChild(domainTypesGroup)
	}

	// Range Types group
	if len(sd.rangeTypes) > 0 {
		rangeTypesGroup := models.NewTreeNode(
			fmt.Sprintf("rangetypes:%s.%s", currentDB, schemaName),
			models.TreeNodeTypeRangeTypeGroup,
			fmt.Sprintf("Range Types (%d)", len(sd.rangeTypes)),
		)
		rangeTypesGroup.Selectable = false
		for _, rtName := range sd.rangeTypes {
			rtNode := models.NewTreeNode(
				fmt.Sprintf("rangetype:%s.%s.%s", currentDB, schemaName, rtName),
				models.TreeNodeTypeRangeType,
				rtName,
			)
			rtNode.Selectable = true
			rtNode.Loaded = true
			rangeTypesGroup.AddChild(rtNode)
		}
		rangeTypesGroup.Loaded = true
		schemaNode.AddChild(rangeTypesGroup)
	}

	schemaNode.Loaded = true
	dbNode.AddChild(schemaNode)
}
```

**Step 4: Build and run tests**

Run: `go build ./... && go test ./...`
Expected: All pass

**Step 5: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(tree): pre-populate all object nodes for search support"
```

---

## Task 3: Remove Obsolete Lazy Loading Code

**Files:**
- Modify: `internal/app/app.go`

**Step 1: Check if loadNodeChildren is still needed for group nodes**

The `loadNodeChildren` function handles lazy loading for group nodes (tables, views, etc.). Since we now pre-populate these, we can simplify this function.

However, table nodes still need lazy loading for their children (columns, indexes, triggers). Keep the table expansion logic, remove the group loading logic.

**Step 2: Update loadNodeChildren to only handle table/view expansion**

Find the `loadNodeChildren` function and remove the cases for:
- `tables:*`
- `views:*`
- `matviews:*`
- `functions:*`
- `procedures:*`
- `triggerfuncs:*`
- `sequences:*`
- `compositetypes:*`
- `enumtypes:*`
- `domaintypes:*`
- `rangetypes:*`

Keep only the `table:*` case for loading columns/indexes/triggers.

**Step 3: Build and test**

Run: `go build ./... && go test ./...`
Expected: All pass

**Step 4: Commit**

```bash
git add internal/app/app.go
git commit -m "refactor(tree): remove obsolete group lazy loading"
```

---

## Task 4: Manual Testing

**Step 1: Start the app and connect to a database**

Run: `go run ./cmd/lazypg`

**Step 2: Verify tree shows all objects**

- Expand a schema
- Verify Tables, Views, Functions etc. show actual object names (not just counts)

**Step 3: Test search**

- Press `/` to start search
- Type a table name
- Verify it appears in search results without needing to expand the tree first

**Step 4: Commit all changes**

```bash
git add -A
git commit -m "feat: enable tree search with preloaded object nodes

- Add GetAllSchemaObjects query to fetch all object names
- Pre-populate tree nodes at connection time
- Search now works without manual tree expansion
- Table columns/indexes still lazy load on demand"
```
