package query

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rebeliceyang/lazypg/internal/models"
)

// Execute executes a SQL query and returns the results
func Execute(ctx context.Context, pool *pgxpool.Pool, sql string) models.QueryResult {
	start := time.Now()

	rows, err := pool.Query(ctx, sql)
	if err != nil {
		return models.QueryResult{
			Error:    err,
			Duration: time.Since(start),
		}
	}
	defer rows.Close()

	// Get column names
	fieldDescs := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		columns[i] = string(fd.Name)
	}

	// Get rows
	var result [][]string
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return models.QueryResult{
				Error:    err,
				Duration: time.Since(start),
			}
		}

		row := make([]string, len(values))
		for i, v := range values {
			if v == nil {
				row[i] = "NULL"
			} else {
				row[i] = fmt.Sprintf("%v", v)
			}
		}
		result = append(result, row)
	}

	// Check for errors from iteration
	if err := rows.Err(); err != nil {
		return models.QueryResult{
			Error:    err,
			Duration: time.Since(start),
		}
	}

	return models.QueryResult{
		Columns:      columns,
		Rows:         result,
		RowsAffected: int64(len(result)),
		Duration:     time.Since(start),
	}
}
