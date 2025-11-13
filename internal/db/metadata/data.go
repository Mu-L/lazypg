package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rebeliceyang/lazypg/internal/db/connection"
)

// TableData represents paginated table data
type TableData struct {
	Columns   []string
	Rows      [][]string
	TotalRows int64
}

// QueryTableData fetches paginated table data
func QueryTableData(ctx context.Context, pool *connection.Pool, schema, table string, offset, limit int) (*TableData, error) {
	// First get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) as count FROM %s.%s", schema, table)
	countRow, err := pool.QueryRow(ctx, countQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to count rows: %w", err)
	}

	totalRows := int64(0)
	if count, ok := countRow["count"].(int64); ok {
		totalRows = count
	}

	// Query paginated data with columns in order
	query := fmt.Sprintf("SELECT * FROM %s.%s LIMIT %d OFFSET %d", schema, table, limit, offset)
	result, err := pool.QueryWithColumns(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table data: %w", err)
	}

	if len(result.Rows) == 0 {
		return &TableData{
			Columns:   result.Columns,
			Rows:      [][]string{},
			TotalRows: totalRows,
		}, nil
	}

	columns := result.Columns

	// Convert rows to string slices
	data := make([][]string, len(result.Rows))
	for i, row := range result.Rows {
		rowData := make([]string, len(columns))
		for j, col := range columns {
			val := row[col]
			if val == nil {
				rowData[j] = "NULL"
			} else {
				rowData[j] = convertValueToString(val)
			}
		}
		data[i] = rowData
	}

	return &TableData{
		Columns:   columns,
		Rows:      data,
		TotalRows: totalRows,
	}, nil
}

// convertValueToString converts a database value to string, handling JSONB properly
func convertValueToString(val interface{}) string {
	// Check if it's a map or slice (JSONB types)
	switch v := val.(type) {
	case map[string]interface{}, []interface{}:
		// Convert to JSON string
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(jsonBytes)
	case []byte:
		// Might be raw JSON bytes
		return string(v)
	default:
		return fmt.Sprintf("%v", val)
	}
}
