# 搜索索引预加载设计

## 问题

懒加载优化后，树节点只在展开时才加载子节点。这导致搜索功能失效 —— `FilterTree` 遍历 `node.Children` 时，未展开的分组节点 `Children` 为空。

## 方案

**预填充树节点**：连接时加载所有对象名称，直接填充为树节点。搜索逻辑无需改动。

## 改造对比

| 阶段 | 改造前（懒加载） | 改造后 |
|------|-----------------|--------|
| 初始查询 | `GetSchemaObjectCounts()` 获取计数 | `GetAllObjectNames()` 获取所有对象名 |
| 分组节点 | `Children = []`, `Loaded = false` | `Children = [实际节点]`, `Loaded = true` |
| 搜索 | 只能搜已展开节点 | 可搜所有对象 |
| 详情加载 | 展开时加载 | 选中时加载（表的列/索引等） |

## 数据库查询

新增 `GetAllObjectNames()` 查询，一次获取所有可搜索对象：

```sql
-- Tables
SELECT n.nspname as schema_name, 'table' as object_type, c.relname as object_name
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

-- Functions (excluding triggers and procedures)
SELECT n.nspname, 'function', p.proname
FROM pg_proc p
JOIN pg_namespace n ON p.pronamespace = n.oid
WHERE p.prokind = 'f'
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
WHERE t.typtype = 'c'
  AND t.typrelid != 0
  AND NOT EXISTS (SELECT 1 FROM pg_class c WHERE c.oid = t.typrelid AND c.relkind != 'c')
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
```

## 返回数据结构

```go
type SchemaObject struct {
    SchemaName string
    ObjectType string // "table", "view", "function", etc.
    ObjectName string
}

func GetAllSchemaObjects(ctx context.Context, pool *pgxpool.Pool) ([]SchemaObject, error)
```

## 代码改动

### 文件改动

| 文件 | 改动 |
|------|------|
| `internal/db/metadata/schema.go` | 新增 `GetAllSchemaObjects()` |
| `internal/app/app.go` | 改造 `loadTree()` 用新查询填充节点 |

### loadTree() 改造

```go
// 改造前
func (a *App) loadTree() tea.Msg {
    counts, _ := metadata.GetSchemaObjectCounts(ctx, pool)
    // 创建空分组节点，Loaded=false
}

// 改造后
func (a *App) loadTree() tea.Msg {
    objects, _ := metadata.GetAllSchemaObjects(ctx, pool)
    // 按 schema 和 type 分组
    // 创建分组节点并填充实际对象节点，Loaded=true
}
```

## 保留的懒加载

表/视图的子节点（列、索引、触发器）仍然懒加载：
- 用户选中表时才加载列信息
- 用户展开表节点时才加载索引/触发器

## 预期效果

| 场景 | 改造前 | 改造后 |
|------|--------|--------|
| 初始加载 (10 schema, 500 对象) | 1 次计数查询, <1s | 1 次名称查询, ~1s |
| 搜索 "users" | 只能搜已展开节点 | 立即找到所有匹配 |
| 展开 Tables 组 | 需要加载表列表 | 已有，无需加载 |
| 选中表查看数据 | 加载表数据 | 加载表数据（不变） |
