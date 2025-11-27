# Preview Pane Design

统一预览面板，用于查看 Data View 表格和 JSONB Tree Viewer 中被截断的内容。

## 核心行为

### 显示逻辑

- 当选中内容被截断时，自动在底部显示预览面板
- 当选中内容完整显示（未截断）时，自动隐藏预览面板
- 按 `p` 键可手动切换显示/隐藏（覆盖自动行为）

### 高度自适应

- 根据内容行数自动调整高度
- 最小高度：3 行（标题 + 1 行内容 + 边框）
- 最大高度：屏幕高度的 1/3
- 超出最大高度时可滚动

### 格式化

- JSON/JSONB 内容自动美化缩进显示
- 普通文本保持原样
- JSON/JSONB 内容显示提示：可按 `J` 打开 JSONB Tree Viewer

### 视觉示例

```
┌─ Data ─────────────────────────────────────────────────────┐
│ id   │ metadata                        │ config           │
├──────┼─────────────────────────────────┼──────────────────┤
│ 1165 │ {"characterSet":"UTF8","col...  │ {"timeout":30... │
│ 1166 │ {"name":"test"}                 │ {"debug":true}   │
└──────┴─────────────────────────────────┴──────────────────┘
┌─ Preview ──────────────────────────────────────────────────┐
│ {                                                          │
│   "characterSet": "UTF8",                                  │
│   "collation": "en_US.utf8",                               │
│   "encoding": "UTF8"                                       │
│ }                                                          │
└──────────────────────────────────── p: Toggle │ J: Tree ───┘
```

## 适用场景

### Data View 表格

- 选中单元格时，检测该单元格内容是否被截断
- 被截断时显示完整单元格内容
- 支持左右移动切换列时自动更新预览

### JSONB Tree Viewer

- 选中节点时，检测该节点的值是否被截断（字符串超过 50 字符）
- 被截断时显示完整节点值
- 如果节点是对象/数组，显示其 JSON 格式化内容

### 截断检测逻辑

```
表格单元格：实际内容长度 > 列宽显示长度
树节点：字符串长度 > 50 或 显示时带有 "..."
```

### 预览面板标题

- 表格场景：`Preview: [列名]`
- 树场景：`Preview: [JSON 路径]`（如 `$.extensions.4.description`）

## 组件架构

### 新增组件

```
internal/ui/components/preview_pane.go
```

### 组件接口

```go
type PreviewPane struct {
    Width      int
    MaxHeight  int             // 最大高度（屏幕 1/3）
    Content    string          // 原始内容
    Title      string          // 标题（列名或 JSON 路径）
    Visible    bool            // 是否可见
    Forced     bool            // 用户手动切换（覆盖自动行为）
    scrollY    int             // 滚动偏移
    style      lipgloss.Style  // 容器样式
}

// 核心方法
func (p *PreviewPane) SetContent(content, title string, isTruncated bool)
func (p *PreviewPane) Toggle()           // p 键触发
func (p *PreviewPane) View() string
func (p *PreviewPane) Update(msg tea.Msg) (PreviewPane, tea.Cmd)
func (p *PreviewPane) Height() int       // 返回实际渲染高度（含边框）
```

### 布局计算（使用 lipgloss 内置方法）

```go
// 定义样式
p.style = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#cba6f7")).
    Padding(0, 1)

// 计算内容可用宽度
contentWidth := p.Width - p.style.GetHorizontalFrameSize()

// 计算内容可用高度
maxContentHeight := p.MaxHeight - p.style.GetVerticalFrameSize()

// 实际渲染高度
actualHeight := min(contentLines, maxContentHeight) + p.style.GetVerticalFrameSize()
```

### 集成方式

- `TableView` 和 `JSONBViewer` 各自持有 `PreviewPane` 实例
- 选中项变化时调用 `SetContent()`
- 父组件根据 `PreviewPane.Height()` 调整主内容区域高度

## 交互与快捷键

### 预览面板内操作

| 按键 | 功能 |
|------|------|
| `p` | 切换预览面板显示/隐藏 |
| `↑/k` | 内容向上滚动（当内容超出高度时） |
| `↓/j` | 内容向下滚动 |
| `y` | 复制预览内容到剪贴板 |

### 焦点说明

- 预览面板不独立获取焦点
- 滚动按键在主视图按键之后处理（避免冲突）
- 即：在表格中 `j/k` 优先移动行，预览面板跟随更新

### 状态持久化

- `Forced` 状态在切换选中项时保持
- 用户手动隐藏后，即使遇到截断内容也保持隐藏
- 再次按 `p` 恢复自动显示行为

### 帮助提示

- 预览面板底部右侧显示 `p: Toggle │ J: Tree`（JSONB 内容时）
- 可滚动时显示 `↑↓: Scroll │ p: Toggle │ y: Copy`

## 边界情况与细节

### 空内容处理

- 单元格为空或 NULL 时，不显示预览面板

### 超长内容处理

- 单行超过预览面板宽度时自动换行
- 总行数超过最大高度时启用滚动
- 滚动指示器：显示 `▲` / `▼`

### JSON 格式化失败

- 检测为 JSONB 但格式化失败时，回退显示原始内容
- 不报错，静默处理

### 窗口大小变化

- 监听 `tea.WindowSizeMsg`
- 重新计算 `MaxHeight`（屏幕 1/3）和 `Width`
- 重新渲染内容换行

### 性能考虑

- 大 JSON（>100KB）格式化可能较慢
- 超大内容时截断显示前 N 行 + 提示 "Content truncated, press J for JSONB Tree Viewer"

### JSONB 提示

- 所有 JSON/JSONB 内容在预览面板底部显示提示：`J: Tree` 提醒用户可使用 JSONB Tree Viewer 查看完整结构
