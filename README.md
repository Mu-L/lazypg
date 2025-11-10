# lazypg

A modern Terminal User Interface (TUI) client for PostgreSQL, inspired by lazygit.

## Status

🚧 **In Development** - Phase 5 (Interactive Filtering) Complete

### Completed Features

- ✅ Multi-panel layout (left navigation, right content)
- ✅ Configuration system (YAML-based)
- ✅ Theme support
- ✅ Help system with keyboard shortcuts
- ✅ Panel focus management
- ✅ Responsive layout
- ✅ PostgreSQL connection management
- ✅ Connection pooling with pgx v5
- ✅ Auto-discovery (port scan, environment, .pgpass)
- ✅ Connection dialog UI
- ✅ Basic metadata queries
- ✅ Navigation tree (databases, schemas, tables)
- ✅ Table data viewing with virtual scrolling
- ✅ Pagination and lazy loading
- ✅ Interactive data navigation
- ✅ Interactive filter builder with type-aware operators
- ✅ Quick filter from cell (Ctrl+F)
- ✅ SQL preview and validation

### In Progress

- 🔄 Structure/Indexes/Constraints tabs

## Installation

### From Source

```bash
git clone https://github.com/rebeliceyang/lazypg.git
cd lazypg
make build
# Binary will be in bin/lazypg
```

### Run

```bash
make run
# Or
./bin/lazypg
```

## Quick Start

1. **Launch**: Run `lazypg`
2. **Help**: Press `?` to see keyboard shortcuts
3. **Navigate**: Use `Tab` to switch between panels
4. **Quit**: Press `q` or `Ctrl+C`

## Filtering

Press `f` while viewing a table to open the interactive filter builder, or `Ctrl+F` to quickly filter by the selected cell.

See [docs/features/filtering.md](docs/features/filtering.md) for detailed documentation.

## Configuration

lazypg looks for configuration in:
- `~/.config/lazypg/config.yaml` (user config)
- `~/.config/lazypg/connections.yaml` (saved connections)
- `./config.yaml` (current directory)

See `config/default.yaml` for all available options.

Example config:

```yaml
ui:
  theme: "default"
  panel_width_ratio: 25
  mouse_enabled: true
```

Example connection config (`~/.config/lazypg/connections.yaml`):

```yaml
connections:
  - name: "Local Dev"
    host: localhost
    port: 5432
    database: mydb
    user: postgres
    ssl_mode: prefer

  - name: "Production"
    host: prod-db.example.com
    port: 5432
    database: prod_db
    user: app_user
    ssl_mode: require
```

## Development

See [DEVELOPMENT.md](docs/DEVELOPMENT.md) for development guide.

```bash
# Install dependencies
make deps

# Build
make build

# Run tests
make test

# Format code
make fmt
```

## Documentation

- [Design Document](docs/plans/2025-11-07-lazypg-design.md) - Complete design specification
- [Phase 1 Plan](docs/plans/2025-11-07-phase1-foundation.md) - Implementation plan
- [Development Guide](docs/DEVELOPMENT.md) - Development workflow

## Roadmap

### Phase 1: Foundation ✅
- Multi-panel layout
- Configuration system
- Theme support
- Help system

### Phase 2: Connection & Discovery ✅
- PostgreSQL connection management
- Connection pool with pgx
- Auto-discovery of local instances
- Connection manager UI
- Metadata queries

### Phase 3: Data Browsing ✅
- Navigation tree
- Table data view
- Virtual scrolling with pagination
- Interactive data navigation

### Phase 5: Interactive Filtering ✅
- Filter builder UI with type-aware operators
- SQL WHERE clause generation
- Quick filter from cell
- Filter preview and validation
- Clear filter functionality

### Phase 4: Command Palette & Query (Next)
- Command palette UI
- Query execution
- Result display

### Phase 6+
- JSONB support
- History and favorites
- Export/Import

See [design document](docs/plans/2025-11-07-lazypg-design.md) for complete roadmap.

## Key Features (Planned)

- 🎯 **Command Palette** - Unified entry point (like VS Code)
- ⌨️ **Keyboard-First** - Optimized for keyboard with mouse support
- 📊 **Virtual Scrolling** - Handle large datasets smoothly
- 🔍 **Interactive Filters** - Visual filter builder
- 📦 **JSONB Excellence** - Advanced JSONB path extraction and filtering
- 💾 **Query Management** - History and favorites
- 🎨 **Customizable** - Themes, keybindings, configs

## Contributing

Contributions welcome! Please read [DEVELOPMENT.md](docs/DEVELOPMENT.md) first.

## License

TBD

## Acknowledgments

- Inspired by [lazygit](https://github.com/jesseduffield/lazygit)
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Styled with [Lipgloss](https://github.com/charmbracelet/lipgloss)
