# Development Guide

## Prerequisites

- Go 1.21 or higher
- Make (optional, but recommended)

## Getting Started

### Clone and Setup

```bash
git clone https://github.com/rebeliceyang/lazypg.git
cd lazypg
make deps
```

### Build

```bash
make build
# Binary will be in bin/lazypg
```

### Run

```bash
make run
# Or directly
go run cmd/lazypg/main.go
```

## Project Structure

```
lazypg/
├── cmd/lazypg/          # Application entry point
├── internal/            # Internal packages
│   ├── app/            # Main application logic
│   ├── config/         # Configuration management
│   ├── models/         # Data models
│   └── ui/             # UI components
│       ├── components/ # Reusable UI components
│       ├── help/       # Help system
│       └── theme/      # Theme definitions
├── config/             # Default configuration
└── docs/               # Documentation
```

## Architecture

lazypg uses the Bubble Tea framework which follows the Elm architecture:

- **Model**: Application state (AppState, panels, config)
- **Update**: Message handling (keyboard, mouse, events)
- **View**: UI rendering (returns string to display)

### Key Concepts

1. **Panels**: Reusable UI components (left nav, right content)
2. **Themes**: Color schemes loaded from config
3. **Config**: YAML-based configuration with defaults
4. **View Modes**: Different modes (Normal, Help) with different key handling

## Configuration

Config files are loaded in this priority:
1. `~/.config/lazypg/config.yaml` (user config)
2. `./config.yaml` (current directory)
3. `./config/default.yaml` (fallback defaults)

See `config/default.yaml` for all available options.

## Development Workflow

### Make Changes

1. Create a new branch
2. Make your changes
3. Build and test: `make build && make test`
4. Format code: `make fmt`
5. Commit with descriptive message

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test -v ./internal/app/...
```

### Adding New Features

1. Update models in `internal/models/` if needed
2. Create UI components in `internal/ui/components/`
3. Add message handling in `internal/app/app.go`
4. Update View() to render new UI
5. Add tests
6. Update documentation

## Debugging

### Debug Prints

Bubble Tea captures stdout/stderr, so use a log file:

```go
import "log"

// In main.go
f, _ := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
defer f.Close()
log.SetOutput(f)

// Anywhere
log.Printf("Debug: %v", someValue)
```

### Common Issues

**Rendering issues**: Check terminal size in WindowSizeMsg
**Key bindings not working**: Verify ViewMode handling
**Config not loading**: Check file paths and permissions

## Code Style

- Use `go fmt` for formatting
- Follow Go best practices
- Keep functions small and focused
- Comment exported functions and types
- Use meaningful variable names

## Git Workflow

Commit messages should follow conventional commits:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `chore:` Maintenance tasks
- `refactor:` Code refactoring

## Resources

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
