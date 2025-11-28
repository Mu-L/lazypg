# Unified Command Palette Design

## Overview

Merge `Ctrl+K` (command palette) and `Ctrl+T` (table jump) into a single unified command palette using VSCode-style prefix modes.

## Prefix Modes

| Prefix | Mode | Content | Example |
|--------|------|---------|---------|
| (none) | Default | Commands + Tables/Views | `connect`, `users` |
| `>` | Commands | Only commands | `>refresh` |
| `@` | Tables | Only tables/views | `@public.users` |
| `#` | History | Query history | `#SELECT` |

## Interaction

1. Press `Ctrl+K` to open the palette
2. Type to search across commands and tables (default mode)
3. Type a prefix character to switch modes:
   - `>` filters to commands only
   - `@` filters to tables/views only
   - `#` filters to query history only
4. Placeholder text updates dynamically based on mode
5. Select with arrow keys, confirm with Enter, cancel with Esc

## Visual Design

Default mode:
```
┌──────────────────────────────────────────┐
│ Search commands and tables...█           │
├──────────────────────────────────────────┤
│ ▸ Connect to database                    │
│ ▸ Refresh                                │
│ ▦ public.users                           │
│ ◎ public.user_stats                      │
└──────────────────────────────────────────┘
```

Tables mode (after typing `@`):
```
┌──────────────────────────────────────────┐
│ @ Search tables and views...█            │
├──────────────────────────────────────────┤
│ ▦ public.users                           │
│ ▦ public.orders                          │
│ ◎ public.user_stats                      │
└──────────────────────────────────────────┘
```

## Implementation

- `CommandPalette` struct stores three data sources: Commands, Tables, History
- `parseInput()` detects prefix and sets mode
- `Filter()` uses mode to select which data sources to search
- `getPlaceholder()` returns mode-specific placeholder text
- Table selection syncs tree view position (expands ancestors, moves cursor)

## Keyboard Shortcuts

- `Ctrl+K`: Open unified command palette
- Removed: `Ctrl+T` (functionality merged into `@` prefix)
