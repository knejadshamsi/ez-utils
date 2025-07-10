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

	// PT table creation queries
	createPTStopsTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			x REAL NOT NULL,
			y REAL NOT NULL,
			name TEXT,
			raw_xml TEXT
		)`

	createPTLinesTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			mode TEXT,
			raw_xml TEXT
		)`

	createPTRoutesTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			line_id TEXT NOT NULL,
			raw_xml TEXT,
			FOREIGN KEY (line_id) REFERENCES %s(id) ON DELETE CASCADE
		)`

	createPTRouteStopsTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			route_id TEXT NOT NULL,
			stop_ref_id TEXT NOT NULL,
			stop_order INTEGER NOT NULL,
			arrival_offset TEXT,
			departure_offset TEXT,
			PRIMARY KEY (route_id, stop_order),
			FOREIGN KEY (route_id) REFERENCES %s(id) ON DELETE CASCADE,
			FOREIGN KEY (stop_ref_id) REFERENCES %s(id)
		)`

	createPTDeparturesTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			route_id TEXT NOT NULL,
			departure_time TEXT NOT NULL,
			FOREIGN KEY (route_id) REFERENCES %s(id) ON DELETE CASCADE
		)`
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

// AlterTelemetryTableForPT adds PT-specific columns to the telemetry table if they don't exist
func (db *Database) AlterTelemetryTableForPT() error {
	// Try to add PT-specific columns - ignore errors if columns already exist
	alterQueries := []string{
		"ALTER TABLE process_telemetry ADD COLUMN stops_extracted INTEGER DEFAULT 0",
		"ALTER TABLE process_telemetry ADD COLUMN lines_extracted INTEGER DEFAULT 0", 
		"ALTER TABLE process_telemetry ADD COLUMN routes_extracted INTEGER DEFAULT 0",
	}
	
	for _, query := range alterQueries {
		// Execute but ignore "duplicate column" errors
		_, _ = db.conn.Exec(query)
	}
	
	return nil
}

// CreatePTTables creates all necessary tables for PT data
func (db *Database) CreatePTTables(processID int) error {
	// Ensure telemetry table has PT columns
	if err := db.AlterTelemetryTableForPT(); err != nil {
		return fmt.Errorf("failed to update telemetry table: %w", err)
	}
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	
	// Create stops table
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	if err := db.execTableQuery(fmt.Sprintf(createPTStopsTableQuery, stopsTable)); err != nil {
		return fmt.Errorf("failed to create PT stops table: %w", err)
	}
	
	// Create lines table
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	if err := db.execTableQuery(fmt.Sprintf(createPTLinesTableQuery, linesTable)); err != nil {
		return fmt.Errorf("failed to create PT lines table: %w", err)
	}
	
	// Create routes table
	routesTable := fmt.Sprintf("%s_routes", tablePrefix)
	query := fmt.Sprintf(createPTRoutesTableQuery, routesTable, linesTable)
	if err := db.execTableQuery(query); err != nil {
		return fmt.Errorf("failed to create PT routes table: %w", err)
	}
	
	// Create route stops table
	routeStopsTable := fmt.Sprintf("%s_route_stops", tablePrefix)
	query = fmt.Sprintf(createPTRouteStopsTableQuery, routeStopsTable, routesTable, stopsTable)
	if err := db.execTableQuery(query); err != nil {
		return fmt.Errorf("failed to create PT route stops table: %w", err)
	}
	
	// Create departures table
	departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
	query = fmt.Sprintf(createPTDeparturesTableQuery, departuresTable, routesTable)
	if err := db.execTableQuery(query); err != nil {
		return fmt.Errorf("failed to create PT departures table: %w", err)
	}
	
	return nil
}

// DropPTTables drops all PT tables for a given process
func (db *Database) DropPTTables(processID int) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	tables := []string{
		fmt.Sprintf("%s_departures", tablePrefix),
		fmt.Sprintf("%s_route_stops", tablePrefix),
		fmt.Sprintf("%s_routes", tablePrefix),
		fmt.Sprintf("%s_lines", tablePrefix),
		fmt.Sprintf("%s_stops", tablePrefix),
	}
	
	for _, table := range tables {
		if err := db.execTableQuery(fmt.Sprintf(dropTableQuery, table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}
	
	return nil
}