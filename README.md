# lazypg

A modern Terminal User Interface (TUI) client for PostgreSQL, inspired by lazygit.

## Status

🚧 **In Development** - Phase 8 (Favorites & Polish) Complete

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
- ✅ JSONB detection and formatting
- ✅ Three-mode JSONB viewer (Formatted/Tree/Query)
- ✅ Path extraction algorithm
- ✅ JSONB filtering operators (@>, <@, ?)
- ✅ Query favorites management
- ✅ Favorites export (CSV/JSON)

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
2. **Connect**: Press `c` or use `Ctrl+P` to open connection dialog
3. **Navigate**: Use arrow keys or `hjkl` to move, `Tab` to switch panels
4. **Command Palette**: Press `Ctrl+P` for quick access to all features
5. **Help**: Press `?` to see keyboard shortcuts
6. **Quit**: Press `q` or `Ctrl+C`

## Features

### Interactive Filtering

Press `f` while viewing a table to open the interactive filter builder, or `Ctrl+F` to quickly filter by the selected cell.

See [docs/features/filtering.md](docs/features/filtering.md) for detailed documentation.

### JSONB Support

Press `j` on a cell containing JSONB data to open the JSONB viewer with three modes:
- **Formatted View** (1) - Pretty-printed JSON
- **Tree View** (2) - Interactive tree navigation
- **Query View** (3) - PostgreSQL query examples

The filter builder supports JSONB operators (@>, <@, ?) for advanced filtering.

See [docs/features/jsonb.md](docs/features/jsonb.md) for detailed documentation.

### Query Favorites

Save and manage your frequently used queries with the favorites system:
- **Add Favorites**: Save queries with descriptions and tags
- **Quick Execute**: Run saved queries with a single keystroke
- **Search**: Find favorites by name, description, or tags
- **Export**: Backup favorites to CSV or JSON formats
- **Usage Tracking**: Automatically track query execution statistics

**Quick Access:**
- Press `Ctrl+P` and type "favorites" to open the favorites dialog
- Press `a` to add a new favorite
- Press `e` to edit, `d` to delete (with confirmation)
- Press `Enter` to execute a favorite

See [docs/features/favorites.md](docs/features/favorites.md) for detailed documentation.

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

### User Documentation
- [Filtering](docs/features/filtering.md) - Interactive filter builder guide
- [JSONB Support](docs/features/jsonb.md) - Working with JSONB data
- [Favorites](docs/features/favorites.md) - Query favorites management

### Developer Documentation
- [Design Document](docs/plans/2025-11-07-lazypg-design.md) - Complete design specification
- [Development Guide](docs/DEVELOPMENT.md) - Development workflow
- [Documentation Index](docs/INDEX.md) - Complete documentation listing

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

### Phase 6: JSONB Support ✅
- JSONB detection and formatting
- Three-mode viewer (Formatted/Tree/Query)
- Path extraction algorithm
- JSONB filtering operators (@>, <@, ?)

### Phase 8: Favorites & Polish ✅
- Query favorites management
- Add, edit, delete favorites
- Search and filtering
- Export to CSV/JSON
- Usage tracking and statistics
- YAML-based persistent storage

### Phase 4: Command Palette & Query (Next)
- Command palette UI (partial - favorites integration complete)
- Query execution
- Result display

### Phase 7+
- Query history
- Additional export formats
- Structure/Indexes/Constraints tabs

See [design document](docs/plans/2025-11-07-lazypg-design.md) for complete roadmap.

## Key Features

- 🎯 **Command Palette** - Unified entry point (like VS Code)
- ⌨️ **Keyboard-First** - Optimized for keyboard with mouse support
- 📊 **Virtual Scrolling** - Handle large datasets smoothly
- 🔍 **Interactive Filters** - Visual filter builder with type-aware operators
- 📦 **JSONB Excellence** - Advanced JSONB path extraction and filtering
- ⭐ **Query Favorites** - Save, organize, and execute frequently used queries
- 💾 **Export Capabilities** - Export favorites and results to CSV/JSON
- 🎨 **Customizable** - Themes, keybindings, configs

## Contributing

Contributions welcome! Please read [DEVELOPMENT.md](docs/DEVELOPMENT.md) first.

## License

TBD

## Acknowledgments

- Inspired by [lazygit](https://github.com/jesseduffield/lazygit)
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Styled with [Lipgloss](https://github.com/charmbracelet/lipgloss)
