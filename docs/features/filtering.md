# Interactive Filtering

lazypg provides a powerful interactive filter builder that generates SQL WHERE clauses.

## Opening the Filter Builder

- Press `f` while viewing a table to open the filter builder
- Press `Ctrl+F` on a cell to quickly filter by that cell's value

## Building Filters

1. Press `a` or `n` to add a new condition
2. Type the column name and press Enter
3. Select an operator using ↑↓ and press Enter
4. Type the value and press Enter
5. Repeat to add more conditions
6. Press Enter to apply the filter

## Operators by Type

### Numeric (int, numeric, real, double)
- `=`, `!=`, `>`, `>=`, `<`, `<=`
- `IS NULL`, `IS NOT NULL`

### Text (char, varchar, text)
- `=`, `!=`
- `LIKE`, `ILIKE` (case-insensitive)
- `IS NULL`, `IS NOT NULL`

### JSONB
- `=`, `!=`
- `@>` (contains)
- `<@` (contained by)
- `?` (has key)
- `IS NULL`, `IS NOT NULL`

### Arrays
- `=`, `!=`
- `&&` (overlap)
- `@>` (contains)
- `<@` (contained by)
- `IS NULL`, `IS NOT NULL`

### Boolean
- `=`, `!=`
- `IS NULL`, `IS NOT NULL`

### Date/Time
- `=`, `!=`, `>`, `>=`, `<`, `<=`
- `IS NULL`, `IS NOT NULL`

## Managing Filters

- Press `d` or `x` to delete a condition
- Press `Ctrl+R` to clear all filters
- Press `Esc` to close without applying

## SQL Preview

The filter builder shows a live SQL preview at the bottom, so you can see exactly what query will be executed.

## Multiple Conditions

All conditions are combined with AND logic. The filter will only show rows that match ALL conditions.

## Examples

### Find users named "Alice"
1. Press `f`
2. Press `a`
3. Type "name", Enter
4. Select `=`, Enter
5. Type "Alice", Enter
6. Press Enter to apply

### Find recent orders (last 7 days)
1. Press `f`
2. Press `a`
3. Type "created_at", Enter
4. Select `>`, Enter
5. Type "2024-01-01", Enter
6. Press Enter to apply

### Quick filter from cell
1. Navigate to a cell with Ctrl+F
2. Filter is automatically applied
