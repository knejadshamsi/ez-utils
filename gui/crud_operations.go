package gui

import (
	"database/sql"
	"fmt"
)

// Error messages for table queries
var tableQueryErrors = map[string]string{
	createProcessesTableQuery:  "failed to create processes table",
	createTelemetryTableQuery:  "failed to create telemetry table",
	createPopulationTableQuery: "failed to create table %s",
	dropTableQuery:             "failed to drop table %s",
}

// execQuery is a generic database execution wrapper that handles all CRUD operations
func execQuery(query string, errorMsg string, args ...any) (sql.Result, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		if errorMsg != "" {
			return nil, fmt.Errorf(errorMsg+": %w", err)
		}
		// Generic error if no custom message provided
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	return result, nil
}

// queryRows is a generic wrapper for SELECT queries that return multiple rows
func queryRows(query string, errorMsg string, args ...any) (*sql.Rows, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		if errorMsg != "" {
			return nil, fmt.Errorf(errorMsg+": %w", err)
		}
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	return rows, nil
}

// queryRow is a generic wrapper for SELECT queries that return a single row
func queryRow(query string, errorMsg string, args ...any) *sql.Row {
	return db.QueryRow(query, args...)
}

// execTableQuery wraps execQuery for backward compatibility with table operations
func execTableQuery(query string, args ...any) error {
	// Look up error message for table queries
	errorMsg := ""
	if msg, ok := tableQueryErrors[query]; ok {
		if len(args) > 0 {
			errorMsg = fmt.Sprintf(msg, args[0])
		} else {
			errorMsg = msg
		}
	}
	_, err := execQuery(query, errorMsg, args...)
	return err
}