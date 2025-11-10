# lazypg Documentation Index

Complete documentation for lazypg, a modern Terminal User Interface (TUI) client for PostgreSQL.

## Getting Started

- [README](../README.md) - Project overview, installation, and quick start
- [Development Guide](DEVELOPMENT.md) - Setup development environment and workflow

## User Features

### Data Management
- [Interactive Filtering](features/filtering.md) - Build complex filters with type-aware operators
- [JSONB Support](features/jsonb.md) - View, navigate, and filter JSONB data
- [Query Favorites](features/favorites.md) - Save, organize, and execute frequently used queries

### Feature Comparison

| Feature | Status | Documentation |
|---------|--------|---------------|
| Interactive Filtering | ✅ Complete | [filtering.md](features/filtering.md) |
| JSONB Support | ✅ Complete | [jsonb.md](features/jsonb.md) |
| Query Favorites | ✅ Complete | [favorites.md](features/favorites.md) |
| Query History | 🔄 Planned | TBD |
| Query Editor | 🔄 Planned | TBD |
| Schema Viewer | 🔄 Planned | TBD |

## Design & Architecture

### Planning Documents
- [Overall Design](plans/2025-11-07-lazypg-design.md) - Complete design specification and vision
- [Phase 1: Foundation](plans/2025-11-07-phase1-foundation.md) - UI framework and configuration
- [Phase 2: Connection](plans/2025-11-07-phase2-connection.md) - PostgreSQL connection management
- [Phase 3: Data Browsing](plans/2025-11-07-phase3-data-browsing.md) - Navigation tree and table views
- [Phase 4: Command Palette](plans/phase-4-command-palette-query.md) - Command palette and query execution
- [Phase 5: Filtering](plans/2025-11-10-phase5-filtering.md) - Interactive filter builder
- [Phase 6: JSONB Support](plans/2025-11-10-phase6-jsonb-support.md) - JSONB detection and viewing
- [Phase 8: Favorites & Polish](plans/2025-11-10-phase8-favorites-polish.md) - Query favorites system

### UI/UX Documentation
- [UI/UX Design Specification](UI_UX_DESIGN_SPECIFICATION.md) - Complete UI/UX design specification
- [UI Design Summary](UI_DESIGN_SUMMARY.md) - Overview of UI design decisions
- [UI Design Index](UI_DESIGN_INDEX.md) - Index of UI components and patterns
- [UI Implementation Guide](UI_IMPLEMENTATION_GUIDE.md) - Implementation guidelines for UI components
- [UI Visual Examples](UI_VISUAL_EXAMPLES.md) - Visual examples of UI components
- [Tree View Visual Guide](TREE_VIEW_VISUAL_GUIDE.md) - Navigation tree implementation guide

## Testing

### Test Checklists
- [Phase 2 Checklist](testing/phase2-checklist.md) - Connection management testing
- [Phase 3 Checklist](PHASE3_TEST_CHECKLIST.md) - Data browsing testing

### Task Documentation
- [Task 4 Implementation](TASK_4_IMPLEMENTATION.md) - Implementation details for Task 4

## Quick Reference

### Keyboard Shortcuts

#### Global
- `Ctrl+P` - Open command palette
- `Tab` - Switch between panels
- `?` - Show help
- `q` - Quit (from main view)
- `Esc` - Close dialogs/cancel operations

#### Navigation
- `↑`/`↓` or `k`/`j` - Move selection up/down
- `←`/`→` or `h`/`l` - Move left/right (in data view)
- `PgUp`/`PgDn` - Page up/down
- `Home`/`End` - Jump to start/end

#### Data Browsing
- `Enter` - Expand/collapse tree node or open table
- `c` - Open connection dialog
- `r` - Refresh current view

#### Filtering
- `f` - Open filter builder
- `Ctrl+F` - Quick filter from cell
- `Ctrl+R` - Clear all filters

#### JSONB
- `j` - Open JSONB viewer (on JSONB cell)
- `1` - Formatted view
- `2` - Tree view
- `3` - Query view

#### Favorites
- Open command palette (`Ctrl+P`) and type "favorites"
- In favorites dialog:
  - `a`/`n` - Add new favorite
  - `e` - Edit selected favorite
  - `d`/`x` - Delete selected favorite (press twice to confirm)
  - `Enter` - Execute selected favorite

### Command Palette Commands

| Command | Icon | Description | Tags |
|---------|------|-------------|------|
| Connect to Database | 🔌 | Open connection dialog | connection, database |
| Disconnect | 🔴 | Close current connection | disconnect, close |
| Refresh | 🔄 | Refresh current view | view, refresh, reload |
| Quick Query | ⚡ | Execute a quick SQL query | query, sql, execute |
| Query Editor | 📝 | Open full query editor | query, editor, write |
| Query History | 📜 | View query history | history, past |
| Favorites | ⭐ | Manage favorite queries | favorites, bookmarks |
| Export Favorites to CSV | 📊 | Export favorites to CSV | export, csv |
| Export Favorites to JSON | 📦 | Export favorites to JSON | export, json |
| Help | ❓ | Show keyboard shortcuts | help, shortcuts |
| Settings | ⚙️ | Open settings | config, preferences |

### Configuration Files

| File | Location | Purpose |
|------|----------|---------|
| `config.yaml` | `~/.config/lazypg/` | User configuration |
| `connections.yaml` | `~/.config/lazypg/` | Saved connections |
| `favorites.yaml` | `~/.config/lazypg/` | Query favorites |
| `history.yaml` | `~/.config/lazypg/` | Query history (planned) |

### Storage Formats

#### Favorites YAML
```yaml
- id: "uuid"
  name: "Query Name"
  description: "Description"
  query: "SELECT * FROM table"
  tags: ["tag1", "tag2"]
  connection: "connection-name"
  database: "database-name"
  created_at: 2025-11-10T10:00:00Z
  updated_at: 2025-11-10T10:00:00Z
  usage_count: 5
  last_used: 2025-11-10T14:30:00Z
```

#### Connections YAML
```yaml
connections:
  - name: "Local Dev"
    host: localhost
    port: 5432
    database: mydb
    user: postgres
    ssl_mode: prefer
```

## Contributing

### Code Style
- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `make fmt` before committing

### Testing
- Write tests for new features
- Run `make test` before submitting PRs
- Update test checklists for new features

### Documentation
- Update feature documentation for user-facing changes
- Update this index when adding new documentation
- Keep README.md in sync with major features

## Resources

### External Links
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver
- [lazygit](https://github.com/jesseduffield/lazygit) - Inspiration

### Related Projects
- [pgcli](https://www.pgcli.com/) - PostgreSQL CLI with auto-completion
- [usql](https://github.com/xo/usql) - Universal SQL CLI
- [psql](https://www.postgresql.org/docs/current/app-psql.html) - PostgreSQL interactive terminal

## Support

### Getting Help
1. Check this documentation index
2. Read the specific feature documentation
3. Review the design documents
4. Check the GitHub issues
5. File a bug report with reproduction steps

### Filing Bug Reports

Include:
- lazypg version
- Operating system and version
- PostgreSQL version
- Steps to reproduce
- Expected vs actual behavior
- Relevant log output
- Configuration files (redact sensitive data)

### Feature Requests

Include:
- Clear description of the feature
- Use cases and benefits
- Proposed UI/UX (if applicable)
- Any related features or dependencies

## Version History

### Phase 8 (Current) - Favorites & Polish
- Query favorites management
- Export to CSV/JSON
- Usage tracking
- YAML-based storage
- Search and filtering

### Phase 6 - JSONB Support
- JSONB detection and formatting
- Three-mode viewer
- Path extraction algorithm
- JSONB filtering operators

### Phase 5 - Interactive Filtering
- Filter builder UI
- Type-aware operators
- Quick filter from cell
- SQL preview

### Phase 3 - Data Browsing
- Navigation tree
- Table data viewing
- Virtual scrolling
- Pagination

### Phase 2 - Connection Management
- PostgreSQL connections
- Connection pooling
- Auto-discovery
- Connection dialog

### Phase 1 - Foundation
- Multi-panel layout
- Configuration system
- Theme support
- Help system

## Roadmap

### Near Term
- [ ] Query history
- [ ] Full query editor
- [ ] Result export (CSV/JSON)
- [ ] Schema viewer (structure/indexes/constraints)

### Medium Term
- [ ] Query parameterization
- [ ] Auto-complete for SQL
- [ ] Connection switching from favorites
- [ ] Favorite collections/folders

### Long Term
- [ ] Multiple result tabs
- [ ] Query performance analysis
- [ ] Database migration support
- [ ] Team collaboration features

## License

TBD

---

**Last Updated**: 2025-11-10
**Version**: Phase 8 (Favorites & Polish)
