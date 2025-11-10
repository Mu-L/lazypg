# JSONB Support

lazypg provides comprehensive JSONB support with multiple viewing modes and filtering capabilities.

## Detecting JSONB Columns

JSONB columns are automatically detected and displayed with a 📦 icon in table views. Values are truncated for display but can be expanded using the JSONB viewer.

## Opening the JSONB Viewer

1. Navigate to a cell containing JSONB data
2. Press `j` to open the JSONB viewer

The viewer has three modes:

### Mode 1: Formatted View

Press `1` to see pretty-printed JSON with proper indentation.

```json
{
  "user": {
    "name": "Alice",
    "age": 30,
    "address": {
      "city": "San Francisco"
    }
  }
}
```

### Mode 2: Tree View

Press `2` to see an interactive tree representation:

```
$ {3 keys}
  user {3 keys}
    name: "Alice"
    age: 30
    address {1 keys}
      city: "San Francisco"
```

Use ↑↓ to navigate through paths.

### Mode 3: Query Mode

Press `3` to see PostgreSQL query examples for the selected path:

```sql
-- Get JSONB value
data #> '{user,address,city}'

-- Get text value
data #>> '{user,address,city}'

-- Filter rows containing this path
data @> '{"user": {"address": {"city": "San Francisco"}}}'
```

## JSONB Filtering

The filter builder supports JSONB-specific operators:

### Contains (@>)

Filter rows where JSONB contains a specific value:

1. Press `f` to open filter builder
2. Select a JSONB column
3. Choose `@>` operator
4. Enter JSON value: `{"user": {"name": "Alice"}}`

### Contained By (<@)

Filter rows where JSONB is contained by a value:

1. Select JSONB column
2. Choose `<@` operator
3. Enter containing JSON value

### Has Key (?)

Filter rows where JSONB has a specific key:

1. Select JSONB column
2. Choose `?` operator
3. Enter key name: `"user"`

## Path Extraction

The tree view automatically extracts all JSON paths from nested structures. This makes it easy to:

- See the structure of complex JSON documents
- Find specific paths to use in queries
- Navigate deeply nested data

## Keyboard Shortcuts

- `j` - Open JSONB viewer (on JSONB cell)
- `1` - Switch to Formatted view
- `2` - Switch to Tree view
- `3` - Switch to Query view
- `↑`/`↓` - Navigate tree paths
- `Esc` - Close viewer

## Examples

### View User Profile

```
1. Run: SELECT * FROM users WHERE id = 123
2. Navigate to profile_data column
3. Press 'j' to open viewer
4. Press '2' for tree view
5. Navigate to $.profile.email
6. Press '3' to see query for that path
```

### Filter by Nested Value

```
1. Press 'f' to open filter builder
2. Column: profile_data
3. Operator: @>
4. Value: {"profile": {"city": "NYC"}}
5. Press Enter to apply
```
