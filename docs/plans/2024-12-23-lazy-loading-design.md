# Tree 懒加载优化设计

## 问题

当前 `loadTree()` 在初始化时会立即加载所有数据：
- 每个 Schema 查询 11 种对象类型
- 每个 Table 还查询 Indexes 和 Triggers
- 对于大型线上数据库（10 schema, 500 表），产生 600+ 次查询，耗时 5-10 秒

## 方案

采用 **Schema + 分组计数 + 懒加载** 方案：
- 初始只用 1 次聚合查询获取所有 schema 的对象计数
- 显示树骨架 + 数量
- 用户展开分组节点时才加载具体对象

## 初始显示效果

```
▼ postgres
  ▼ public
    ▸ Tables (30)
    ▸ Views (5)
    ▸ Functions (12)
  ▸ auth (8 objects)
  ▸ storage (3 objects)
```

## 聚合查询设计

用 1 次 SQL 获取所有 schema 的各类对象数量：

```sql
SELECT
    n.nspname AS schema_name,
    SUM(CASE WHEN c.relkind = 'r' THEN 1 ELSE 0 END) AS tables,
    SUM(CASE WHEN c.relkind = 'v' THEN 1 ELSE 0 END) AS views,
    SUM(CASE WHEN c.relkind = 'm' THEN 1 ELSE 0 END) AS materialized_views,
    SUM(CASE WHEN c.relkind = 'S' THEN 1 ELSE 0 END) AS sequences,
    -- functions, procedures 等用子查询
FROM pg_namespace n
LEFT JOIN pg_class c ON c.relnamespace = n.oid
WHERE n.nspname NOT LIKE 'pg_%' AND n.nspname != 'information_schema'
GROUP BY n.oid, n.nspname
```

## 懒加载机制

### 触发点

当用户展开一个分组节点（如 "Tables (30)"）时：
1. 检查 `node.Loaded == false`
2. 发送 `LoadNodeChildrenMsg{NodeID: "tables:postgres.public"}`
3. 显示内联 loading spinner
4. 查询具体对象列表
5. 填充子节点，设置 `node.Loaded = true`

### 节点类型与加载映射

| 节点类型 | 展开时加载 |
|---------|-----------|
| TableGroup | ListTables() |
| ViewGroup | ListViews() |
| MaterializedViewGroup | ListMaterializedViews() |
| FunctionGroup | ListFunctions() |
| ProcedureGroup | ListProcedures() |
| TriggerFunctionGroup | ListTriggerFunctions() |
| SequenceGroup | ListSequences() |
| CompositeTypeGroup | ListCompositeTypes() |
| EnumTypeGroup | ListEnumTypes() |
| DomainTypeGroup | ListDomainTypes() |
| RangeTypeGroup | ListRangeTypes() |
| Table (展开时) | ListTableIndexes() + ListTableTriggers() |

## 代码改动

### 需要修改/新增的文件

| 文件 | 改动 |
|------|------|
| `internal/db/metadata/schema.go` | 新增 `GetSchemaObjectCounts()` 聚合查询 |
| `internal/app/app.go` | 重构 `loadTree()` 为轻量级初始加载 |
| `internal/app/app.go` | 新增 `LoadNodeChildrenMsg` 和 handler |
| `internal/ui/components/tree_view.go` | 展开节点时触发懒加载 |

### loadTree() 改造

```go
// 改造前：加载所有数据
func (a *App) loadTree() tea.Msg {
    schemas := metadata.ListSchemas(...)
    for _, schema := range schemas {
        tables := metadata.ListTables(...)      // N 次
        views := metadata.ListViews(...)        // N 次
        // ... 11 种对象类型
    }
}

// 改造后：只加载骨架
func (a *App) loadTree() tea.Msg {
    counts := metadata.GetSchemaObjectCounts(...)  // 1 次聚合
    // 构建骨架树，所有分组节点 Loaded=false
}
```

### 新增消息类型

```go
// LoadNodeChildrenMsg 请求加载节点的子节点
type LoadNodeChildrenMsg struct {
    NodeID string
}

// NodeChildrenLoadedMsg 节点子节点加载完成
type NodeChildrenLoadedMsg struct {
    NodeID   string
    Children []*models.TreeNode
    Err      error
}
```

## 预期效果

| 场景 | 改造前 | 改造后 |
|------|--------|--------|
| 初始连接 (10 schema, 500 表) | ~600 次查询, 5-10秒 | 1 次查询, <1秒 |
| 展开 "Tables (30)" | 已加载 | 1 次查询, ~200ms |

## 边界情况

| 情况 | 处理方式 |
|------|---------|
| 空 schema (0 objects) | 不显示 |
| 分组为空 (如 Views=0) | 不显示该分组节点 |
| 懒加载失败 | 显示错误，保留节点可重试 |
| 快速连续展开多个节点 | 队列处理，避免并发冲突 |
