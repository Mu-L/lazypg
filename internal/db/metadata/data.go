package metadata

import (
	"context"
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

	// Get column order from information_schema
	columnsQuery := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position`

	columnRows, err := pool.Query(ctx, columnsQuery, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to get column order: %w", err)
	}

	var columns []string
	for _, row := range columnRows {
		if colName, ok := row["column_name"].(string); ok {
			columns = append(columns, colName)
		}
	}

	if len(columns) == 0 {
		return &TableData{
			Columns:   []string{},
			Rows:      [][]string{},
			TotalRows: totalRows,
		}, nil
	}

	// Query paginated data
	query := fmt.Sprintf("SELECT * FROM %s.%s LIMIT %d OFFSET %d", schema, table, limit, offset)
	rows, err := pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table data: %w", err)
	}

	if len(rows) == 0 {
		return &TableData{
			Columns:   []string{},
			Rows:      [][]string{},
			TotalRows: totalRows,
		}, nil
	}

	// Convert rows to string slices
	data := make([][]string, len(rows))
	for i, row := range rows {
		rowData := make([]string, len(columns))
		for j, col := range columns {
			val := row[col]
			if val == nil {
				rowData[j] = "NULL"
			} else {
				rowData[j] = fmt.Sprintf("%v", val)
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
