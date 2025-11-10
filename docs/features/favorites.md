# Favorites

lazypg provides a powerful favorites system for managing and organizing your frequently used SQL queries. Save queries with descriptions and tags, organize them efficiently, and execute them with a single keystroke.

## Overview

The favorites feature allows you to:

- Save frequently used queries for quick access
- Organize queries with names, descriptions, and tags
- Track usage statistics (count and last used time)
- Search through favorites by name, description, or tags
- Execute favorites directly from the dialog
- Export favorites to CSV or JSON for backup or sharing
- Store favorites persistently across sessions

## Quick Start

1. Open the favorites dialog using the command palette (Ctrl+P, then type "favorites")
2. Press `a` or `n` to add a new favorite
3. Fill in the name, description, query, and optional tags
4. Press Enter on the Tags field to save
5. Press Enter on a favorite to execute it

## Opening the Favorites Dialog

There are two ways to open the favorites dialog:

### Command Palette
1. Press `Ctrl+P` to open the command palette
2. Type "favorites" or "favorite queries"
3. Select "Favorites - Manage favorite queries"

### Search Tags
The favorites dialog can also be found by searching for:
- "favorites"
- "bookmarks"
- "saved"

## Managing Favorites

### Adding a New Favorite

1. Open the favorites dialog
2. Press `a` or `n` to enter Add mode
3. Fill in the following fields (use Tab to move between fields):
   - **Name** (required): A short, descriptive name
   - **Description** (optional): Detailed explanation of what the query does
   - **Query** (required): The SQL query to save
   - **Tags** (optional): Comma-separated tags for organization (e.g., "users, reports, analytics")
4. Press Enter on the Tags field to save the favorite

**Example:**
```
Name: Active Users
Description: Get all active users from the last 30 days
Query: SELECT * FROM users WHERE last_login > NOW() - INTERVAL '30 days'
Tags: users, active
```

**Validation Rules:**
- Name cannot be empty
- Name must be unique (case-insensitive)
- Query cannot be empty
- All whitespace is automatically trimmed

### Editing a Favorite

1. Navigate to the favorite using ↑/↓ or j/k
2. Press `e` to enter Edit mode
3. Modify any fields (Tab to move between fields)
4. Press Enter on the Tags field to save changes
5. Press Esc to cancel without saving

**Note:** The favorite's ID, connection, database, creation time, and usage statistics are preserved when editing.

### Deleting a Favorite

1. Navigate to the favorite using ↑/↓ or j/k
2. Press `d` or `x` to enter delete confirmation mode
3. Press `d` or `x` again to confirm deletion
4. Press Esc to cancel deletion

**Note:** Deletion requires two key presses to prevent accidental deletion. After the first press, you'll see a confirmation message.

### Executing a Favorite

1. Navigate to the favorite using ↑/↓ or j/k
2. Press Enter to execute the query

**Requirements:**
- You must be connected to a database before executing queries
- If not connected, you'll see an error message prompting you to connect

**Automatic Tracking:**
- Each execution increments the usage count
- The "last used" timestamp is updated
- These statistics help identify your most frequently used queries

### Searching Favorites

The favorites dialog includes built-in search functionality:

1. As you type, favorites are filtered in real-time
2. Search matches against:
   - Favorite name
   - Description
   - Tags

**Note:** Search is case-insensitive and matches partial strings.

## Keyboard Shortcuts

### List Mode (Default)

| Key | Action |
|-----|--------|
| `↑` or `k` | Move selection up |
| `↓` or `j` | Move selection down |
| `Enter` | Execute selected favorite |
| `a` or `n` | Add new favorite |
| `e` | Edit selected favorite |
| `d` or `x` | Delete selected favorite (press twice to confirm) |
| `Esc` or `q` | Close favorites dialog |

### Add/Edit Mode

| Key | Action |
|-----|--------|
| `Tab` | Move to next field |
| `Shift+Tab` | Move to previous field |
| `Backspace` | Delete character |
| `Enter` | Save favorite (when on Tags field) |
| `Esc` | Cancel and return to list mode |

### Field Order
1. Name
2. Description
3. Query
4. Tags

Press Tab to cycle through fields. When you reach Tags and press Enter, the favorite is saved.

## Command Palette Commands

The following favorites-related commands are available in the command palette (Ctrl+P):

### Manage Favorites
- **Command**: "Favorites"
- **Icon**: ⭐
- **Description**: Manage favorite queries
- **Tags**: favorites, bookmarks, saved

### Export to CSV
- **Command**: "Export Favorites to CSV"
- **Icon**: 📊
- **Description**: Export all favorites to CSV file
- **Tags**: export, favorites, csv
- **Output**: `~/.config/lazypg/favorites.csv`

### Export to JSON
- **Command**: "Export Favorites to JSON"
- **Icon**: 📦
- **Description**: Export all favorites to JSON file
- **Tags**: export, favorites, json
- **Output**: `~/.config/lazypg/favorites.json`

## Export and Import

### Exporting Favorites

#### CSV Export

1. Open command palette (Ctrl+P)
2. Type "export favorites csv"
3. Select "Export Favorites to CSV"
4. File is saved to `~/.config/lazypg/favorites.csv`

**CSV Format:**
```csv
id,name,description,query,tags,connection,database,created_at,updated_at,usage_count,last_used
550e8400-e29b-41d4-a716-446655440000,Active Users,Get all active users from the last 30 days,SELECT * FROM users WHERE last_login > NOW() - INTERVAL '30 days',"users,active",production,myapp,2025-11-10T10:00:00Z,2025-11-10T10:00:00Z,5,2025-11-10T14:30:00Z
```

#### JSON Export

1. Open command palette (Ctrl+P)
2. Type "export favorites json"
3. Select "Export Favorites to JSON"
4. File is saved to `~/.config/lazypg/favorites.json`

**JSON Format:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Active Users",
    "description": "Get all active users from the last 30 days",
    "query": "SELECT * FROM users WHERE last_login > NOW() - INTERVAL '30 days'",
    "tags": ["users", "active"],
    "connection": "production",
    "database": "myapp",
    "created_at": "2025-11-10T10:00:00Z",
    "updated_at": "2025-11-10T10:00:00Z",
    "usage_count": 5,
    "last_used": "2025-11-10T14:30:00Z"
  }
]
```

### Importing Favorites

To import favorites from another machine or backup:

1. Copy the `favorites.yaml` file to `~/.config/lazypg/favorites.yaml`
2. Restart lazypg
3. Your favorites will be automatically loaded

**Note:** You can also manually edit the YAML file to add or modify favorites.

## Storage Format

Favorites are stored in YAML format at `~/.config/lazypg/favorites.yaml`.

### YAML Structure

```yaml
- id: "550e8400-e29b-41d4-a716-446655440000"
  name: "Active Users"
  description: "Get all active users from the last 30 days"
  query: "SELECT * FROM users WHERE last_login > NOW() - INTERVAL '30 days'"
  tags:
    - users
    - active
  connection: "production"
  database: "myapp"
  created_at: 2025-11-10T10:00:00Z
  updated_at: 2025-11-10T10:00:00Z
  usage_count: 5
  last_used: 2025-11-10T14:30:00Z

- id: "660e8400-e29b-41d4-a716-446655440001"
  name: "Orders This Month"
  description: "All orders placed in the current month"
  query: "SELECT * FROM orders WHERE created_at >= date_trunc('month', CURRENT_DATE)"
  tags:
    - orders
    - reports
  connection: "production"
  database: "sales"
  created_at: 2025-11-10T11:00:00Z
  updated_at: 2025-11-10T11:00:00Z
  usage_count: 12
  last_used: 2025-11-10T15:00:00Z
```

### Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string (UUID) | Yes | Unique identifier (auto-generated) |
| `name` | string | Yes | Display name (must be unique, case-insensitive) |
| `description` | string | No | Detailed description |
| `query` | string | Yes | SQL query to execute |
| `tags` | array | No | Tags for organization and search |
| `connection` | string | No | Connection name (captured when favorite is created) |
| `database` | string | No | Database name (captured when favorite is created) |
| `created_at` | timestamp | Yes | Creation timestamp (auto-generated) |
| `updated_at` | timestamp | Yes | Last modification timestamp (auto-updated) |
| `usage_count` | integer | Yes | Number of times executed (starts at 0) |
| `last_used` | timestamp | No | Last execution timestamp |

### Storage Location

- **Default**: `~/.config/lazypg/favorites.yaml`
- **Custom**: Set via `LAZYPG_CONFIG_DIR` environment variable

The favorites file is automatically created when you save your first favorite. The directory is created with permissions `0755`, and the file is created with permissions `0644`.

## Tips and Best Practices

### Organizing Favorites

1. **Use Descriptive Names**: Make names clear and specific (e.g., "Active Users Last 30 Days" instead of "User Query")
2. **Add Descriptions**: Include context about when and why to use each query
3. **Tag Consistently**: Use a consistent tagging scheme (e.g., always tag by table name, then by purpose)
4. **Group Related Queries**: Use tags to group related queries (e.g., "reports", "analytics", "troubleshooting")

### Effective Tagging Strategies

**By Entity:**
```
tags: users, customers, accounts
```

**By Purpose:**
```
tags: reports, analytics, monitoring, troubleshooting
```

**By Frequency:**
```
tags: daily, weekly, monthly
```

**By Team:**
```
tags: sales, engineering, support
```

### Query Best Practices

1. **Use Parameters Wisely**: While favorites don't support query parameters yet, you can save queries with common WHERE clauses and modify them after execution
2. **Keep Queries Simple**: Complex queries with many parameters might be better suited for the query editor
3. **Comment Complex Logic**: Add SQL comments to explain non-obvious logic
4. **Test Before Saving**: Always test queries before adding them to favorites

### Backup and Sharing

1. **Regular Backups**: Export favorites regularly using CSV or JSON format
2. **Version Control**: Keep your `favorites.yaml` in version control (but exclude sensitive queries)
3. **Team Sharing**: Share favorites with team members by exporting to JSON and distributing
4. **Document Conventions**: If sharing with a team, document your tagging and naming conventions

### Performance Tips

1. **Limit Result Sets**: Add `LIMIT` clauses to queries that might return large result sets
2. **Use Indexes**: Ensure queries reference indexed columns for better performance
3. **Review Usage Stats**: Use usage_count and last_used to identify and optimize frequently used queries

## Use Cases

### Daily Operations

**Morning Health Check:**
```yaml
name: "Database Health Check"
query: |
  SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    n_live_tup AS rows
  FROM pg_stat_user_tables
  ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
  LIMIT 10
tags: [monitoring, daily, health]
```

**Active User Count:**
```yaml
name: "Active Users Today"
query: SELECT COUNT(*) FROM users WHERE last_seen > CURRENT_DATE
tags: [users, monitoring, daily]
```

### Reports

**Monthly Sales:**
```yaml
name: "Sales This Month"
query: |
  SELECT
    DATE(created_at) as date,
    COUNT(*) as orders,
    SUM(total_amount) as revenue
  FROM orders
  WHERE created_at >= date_trunc('month', CURRENT_DATE)
  GROUP BY DATE(created_at)
  ORDER BY date
tags: [sales, reports, monthly]
```

### Troubleshooting

**Find Slow Queries:**
```yaml
name: "Slow Queries Last Hour"
query: |
  SELECT
    query,
    calls,
    total_time,
    mean_time,
    max_time
  FROM pg_stat_statements
  WHERE max_time > 1000
  ORDER BY max_time DESC
  LIMIT 20
tags: [performance, troubleshooting]
```

**Lock Analysis:**
```yaml
name: "Current Locks"
query: |
  SELECT
    locktype,
    relation::regclass,
    mode,
    granted,
    pid
  FROM pg_locks
  WHERE NOT granted
tags: [troubleshooting, locks]
```

### Data Analysis

**User Growth:**
```yaml
name: "User Growth by Week"
query: |
  SELECT
    date_trunc('week', created_at) as week,
    COUNT(*) as new_users
  FROM users
  WHERE created_at >= NOW() - INTERVAL '3 months'
  GROUP BY week
  ORDER BY week
tags: [analytics, users, growth]
```

## Troubleshooting

### Favorites Not Loading

**Problem:** Favorites dialog is empty after adding favorites.

**Solutions:**
1. Check that `~/.config/lazypg/favorites.yaml` exists
2. Verify file permissions (should be readable: `0644`)
3. Check YAML syntax using a YAML validator
4. Look for error messages in the application log

### Cannot Add Favorite

**Problem:** Getting "favorite name cannot be empty" error.

**Solutions:**
1. Ensure the Name field is not empty
2. Check that you're not using only whitespace
3. Try a different name if there's a conflict

**Problem:** Getting "a favorite with the name 'X' already exists" error.

**Solutions:**
1. Favorite names must be unique (case-insensitive)
2. Use a different name or edit the existing favorite
3. Search for the existing favorite to see its full query

### Cannot Execute Favorite

**Problem:** "No Database Connection" error when executing.

**Solutions:**
1. Connect to a database first (press 'c' or use command palette)
2. Ensure the connection is active (check status bar)
3. Verify the database specified in the favorite exists

### Export Failed

**Problem:** Export operation fails with permissions error.

**Solutions:**
1. Check write permissions on `~/.config/lazypg/` directory
2. Ensure sufficient disk space
3. Try exporting to a custom location with write permissions

### YAML Parsing Error

**Problem:** Favorites file has syntax errors after manual editing.

**Solutions:**
1. Validate YAML syntax using an online validator
2. Check indentation (use spaces, not tabs)
3. Ensure strings with special characters are quoted
4. Restore from backup if available
5. Delete the file and recreate favorites through the UI

## Advanced Features

### Usage Statistics

Every time you execute a favorite, lazypg automatically tracks:
- **usage_count**: Total number of executions
- **last_used**: Timestamp of most recent execution

These statistics help you:
- Identify your most frequently used queries
- Find queries that might be candidates for optimization
- Clean up unused favorites

### Search Algorithm

The search functionality uses case-insensitive substring matching across:
1. Favorite name (highest priority)
2. Description (medium priority)
3. Tags (if any tag matches, the favorite is included)

Example: Searching for "user" will match:
- Name: "Active Users"
- Description: "Query to find user sessions"
- Tags: ["users", "accounts"]

### Connection Context

When you create a favorite, lazypg captures:
- **connection**: The name of the active connection
- **database**: The name of the active database

While these aren't currently used for automatic connection switching, they provide useful context about where the query was originally used and can help when organizing queries across multiple databases.

## Future Enhancements

The favorites feature is designed to be extensible. Potential future improvements include:

- **Query Parameters**: Support for parameterized queries with placeholders
- **Auto-Connect**: Automatically switch to the correct connection/database when executing a favorite
- **Query Templates**: Create reusable query templates with variable substitution
- **Smart Suggestions**: Suggest favorites based on current context (table, schema, etc.)
- **Import from Files**: Import favorites from CSV/JSON files
- **Favorite Collections**: Group favorites into collections or folders
- **Sharing**: Share favorites with team members via cloud sync
- **Execution History**: Track results and execution times for each favorite run

## Related Documentation

- [Filtering](filtering.md) - Learn about the interactive filter builder
- [JSONB Support](jsonb.md) - Working with JSONB data in queries
- [Command Palette](../plans/phase-4-command-palette-query.md) - Complete command palette documentation

## Support

If you encounter issues with the favorites feature:

1. Check this troubleshooting guide
2. Review the YAML file for syntax errors
3. Check file permissions
4. File a bug report with:
   - Steps to reproduce
   - Error messages
   - Contents of favorites.yaml (redact sensitive queries)
   - Operating system and lazypg version
