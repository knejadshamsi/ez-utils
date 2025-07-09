package gui

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

// Table creation queries
const (
	createProcessesTableQuery = `
		CREATE TABLE IF NOT EXISTS processes (
			process_id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_path TEXT NOT NULL,
			status TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`

	createTelemetryTableQuery = `
		CREATE TABLE IF NOT EXISTS process_telemetry (
			process_id INTEGER PRIMARY KEY,
			total_file_size INTEGER NOT NULL,
			bytes_read INTEGER DEFAULT 0,
			persons_extracted INTEGER DEFAULT 0,
			error_count INTEGER DEFAULT 0,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (process_id) REFERENCES processes(process_id) ON DELETE CASCADE
		)`

	createPopulationTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			coords TEXT,
			raw_xml TEXT
		)`

	dropTableQuery = "DROP TABLE IF EXISTS %s"
)

// Error messages for table queries
var tableQueryErrors = map[string]string{
	createProcessesTableQuery:  "failed to create processes table",
	createTelemetryTableQuery:  "failed to create telemetry table",
	createPopulationTableQuery: "failed to create table %s",
	dropTableQuery:             "failed to drop table %s",
}

// Database encapsulates the database connection
type Database struct {
	conn *sql.DB
}

// NewDatabase initializes a new database connection.
func NewDatabase(dataSourceName string) (*Database, error) {
	var err error
	if _, err := os.Stat(dataSourceName); os.IsNotExist(err) {
		file, err := os.Create(dataSourceName)
		if err != nil {
			return nil, fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
	}

	conn, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db := &Database{conn: conn}

	// Create base tables
	if err := db.execTableQuery(createProcessesTableQuery); err != nil { return nil, err }
	if err := db.execTableQuery(createTelemetryTableQuery); err != nil { return nil, err }

	return db, nil
}

// GetConn returns the underlying database connection.
func (db *Database) GetConn() *sql.DB {
	return db.conn
}


// Close closes the database connection
func (db *Database) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// WithTransaction helper for batch operations
func (db *Database) WithTransaction(fn func(*sql.Tx) error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Failed to rollback transaction: %v", rbErr)
			}
		}
	}()
	
	if err = fn(tx); err != nil {
		return err
	}
	
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// execQuery is a generic database execution wrapper that handles all CRUD operations
func (db *Database) execQuery(query string, errorMsg string, args ...any) (sql.Result, error) {
	result, err := db.conn.Exec(query, args...)
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
func (db *Database) queryRows(query string, errorMsg string, args ...any) (*sql.Rows, error) {
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		if errorMsg != "" {
			return nil, fmt.Errorf(errorMsg+": %w", err)
		}
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	return rows, nil
}

// queryRow is a generic wrapper for SELECT queries that return a single row
func (db *Database) queryRow(query string, errorMsg string, args ...any) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// execTableQuery wraps execQuery for backward compatibility with table operations
func (db *Database) execTableQuery(query string, args ...any) error {
	// Look up error message for table queries
	errorMsg := ""
	if msg, ok := tableQueryErrors[query]; ok {
		if len(args) > 0 {
			errorMsg = fmt.Sprintf(msg, args[0])
		} else {
			errorMsg = msg
		}
	}
	_, err := db.execQuery(query, errorMsg, args...)
	return err
}