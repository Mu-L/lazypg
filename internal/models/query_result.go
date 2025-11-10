package models

import "time"

// QueryResult represents the result of a SQL query execution
type QueryResult struct {
	Columns      []string
	Rows         [][]string
	RowsAffected int64
	Duration     time.Duration
	Error        error
}
