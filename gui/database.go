package gui

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

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
			lng REAL NOT NULL,
			lat REAL NOT NULL,
			raw_xml TEXT
		)`

	// Network tables with separate lng/lat columns
	createNetworkNodesTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			lng REAL NOT NULL,
			lat REAL NOT NULL,
			raw_xml TEXT
		)`

	// PT tables
	createPTLinesTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			name TEXT,
			mode TEXT
		)`

	createPTRoutesTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			line_id TEXT NOT NULL,
			name TEXT,
			FOREIGN KEY (line_id) REFERENCES %s_lines(id) ON DELETE CASCADE
		)`

	createPTStopsTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			stop_id TEXT PRIMARY KEY,
			stop_name TEXT,
			lat REAL,
			lng REAL,
			arrival_offset TEXT,
			departure_offset TEXT,
			stop_type TEXT,
			wheelchair_accessible TEXT,
			timing_point BOOLEAN
		)`

	createPTRouteStopsTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			link_id TEXT PRIMARY KEY,
			route_id TEXT NOT NULL,
			stop_id TEXT NOT NULL,
			sequence INTEGER,
			FOREIGN KEY (route_id) REFERENCES %s_routes(id) ON DELETE CASCADE,
			FOREIGN KEY (stop_id) REFERENCES %s_stops(stop_id) ON DELETE CASCADE
		)`

	createPTDeparturesTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			route_id TEXT NOT NULL,
			departure_time TEXT,
			vehicle_ref_id TEXT,
			FOREIGN KEY (route_id) REFERENCES %s_routes(id) ON DELETE CASCADE
		)`

	createNetworkLinksTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			from_node TEXT NOT NULL,
			to_node TEXT NOT NULL,
			raw_xml TEXT
		)`

	dropTableQuery = "DROP TABLE IF EXISTS %s"


	// Spatial indexes for performance
	createPopulationSpatialIndex = "CREATE INDEX IF NOT EXISTS idx_%s_spatial ON %s (lng, lat)"
	createNodesSpatialIndex      = "CREATE INDEX IF NOT EXISTS idx_%s_spatial ON %s (lng, lat)"
	createStopsSpatialIndex      = "CREATE INDEX IF NOT EXISTS idx_%s_spatial ON %s (lng, lat)"
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

	// Configure connection pool
	conn.SetMaxOpenConns(25) // Maximum number of open connections
	conn.SetMaxIdleConns(5)  // Maximum number of idle connections
	conn.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime of a connection

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
	tablePrefix := fmt.Sprintf("pt_%d", processID)
	
	// Create lines table
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	if err := db.execTableQuery(fmt.Sprintf(createPTLinesTableQuery, linesTable)); err != nil {
		return fmt.Errorf("failed to create PT lines table: %w", err)
	}
	
	// Create routes table
	routesTable := fmt.Sprintf("%s_routes", tablePrefix)
	if err := db.execTableQuery(fmt.Sprintf(createPTRoutesTableQuery, routesTable, tablePrefix)); err != nil {
		return fmt.Errorf("failed to create PT routes table: %w", err)
	}
	
	// Create stops table
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	if err := db.execTableQuery(fmt.Sprintf(createPTStopsTableQuery, stopsTable)); err != nil {
		return fmt.Errorf("failed to create PT stops table: %w", err)
	}
	
	// Create route_stops junction table
	routeStopsTable := fmt.Sprintf("%s_route_stops", tablePrefix)
	if err := db.execTableQuery(fmt.Sprintf(createPTRouteStopsTableQuery, routeStopsTable, tablePrefix, tablePrefix)); err != nil {
		return fmt.Errorf("failed to create PT route_stops table: %w", err)
	}
	
	// Create departures table
	departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
	if err := db.execTableQuery(fmt.Sprintf(createPTDeparturesTableQuery, departuresTable, tablePrefix)); err != nil {
		return fmt.Errorf("failed to create PT departures table: %w", err)
	}
	
	// Create spatial index for stops
	indexName := fmt.Sprintf("stops_%d", processID)
	if err := db.execTableQuery(fmt.Sprintf(createStopsSpatialIndex, indexName, stopsTable)); err != nil {
		return fmt.Errorf("failed to create stops spatial index: %w", err)
	}
	
	return nil
}

// DropPTTables drops all PT tables for a given process
func (db *Database) DropPTTables(processID int) error {
	tablePrefix := fmt.Sprintf("pt_%d", processID)
	tables := []string{
		fmt.Sprintf("%s_departures", tablePrefix),
		fmt.Sprintf("%s_route_stops", tablePrefix),
		fmt.Sprintf("%s_routes", tablePrefix),
		fmt.Sprintf("%s_stops", tablePrefix),
		fmt.Sprintf("%s_lines", tablePrefix),
	}
	
	for _, table := range tables {
		if err := db.execTableQuery(fmt.Sprintf(dropTableQuery, table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}
	
	return nil
}

// Spatial query functions

// CountPopulationInTable counts total persons in a population table
func (db *Database) CountPopulationInTable(processId int) (int, error) {
	tableName := fmt.Sprintf("population_data_%d", processId)
	
	var count int
	err := db.queryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName),
		"failed to count population",
	).Scan(&count)
	
	return count, err
}

// GetAllPopulation retrieves all population data without limits
func (db *Database) GetAllPopulation(processId int) ([]PersonData, error) {
	tableName := fmt.Sprintf("population_data_%d", processId)
	
	query := fmt.Sprintf(`
		SELECT id, lng, lat, raw_xml 
		FROM %s 
		ORDER BY id`, tableName)
	
	rows, err := db.queryRows(query, "failed to query all population")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []PersonData
	for rows.Next() {
		var p PersonData
		if err := rows.Scan(&p.ID, &p.Lng, &p.Lat, &p.RawXML); err != nil {
			return nil, err
		}
		persons = append(persons, p)
	}
	
	return persons, nil
}

// GetPopulationInBounds retrieves population data within viewport bounds with randomization
func (db *Database) GetPopulationInBounds(processId int, viewport ViewportBounds, randomFactor float64, maxElements int) ([]PersonData, error) {
	tableName := fmt.Sprintf("population_data_%d", processId)
	
	query := fmt.Sprintf(`
		SELECT id, lng, lat, raw_xml 
		FROM %s 
		WHERE lat BETWEEN ? AND ? AND lng BETWEEN ? AND ?
		ORDER BY RANDOM()
		LIMIT ?`, tableName)
	
	// Calculate actual limit based on random factor
	estimatedCount := db.estimateRowsInBounds(tableName, viewport)
	actualLimit := int(float64(estimatedCount) * randomFactor)
	if actualLimit > maxElements {
		actualLimit = maxElements
	}
	if actualLimit < 1 {
		actualLimit = 1
	}
	
	rows, err := db.queryRows(query, "failed to query population in bounds",
		viewport.MinLat, viewport.MaxLat, viewport.MinLng, viewport.MaxLng, actualLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []PersonData
	for rows.Next() {
		var p PersonData
		if err := rows.Scan(&p.ID, &p.Lng, &p.Lat, &p.RawXML); err != nil {
			return nil, err
		}
		persons = append(persons, p)
	}
	
	return persons, nil
}

// GetNodesInBounds retrieves network nodes within viewport bounds
func (db *Database) GetNodesInBounds(processId int, viewport ViewportBounds, randomFactor float64, maxElements int) ([]NodeData, error) {
	tableName := fmt.Sprintf("network_%d_nodes", processId)
	
	query := fmt.Sprintf(`
		SELECT id, lng, lat, raw_xml 
		FROM %s 
		WHERE lat BETWEEN ? AND ? AND lng BETWEEN ? AND ?
		ORDER BY RANDOM()
		LIMIT ?`, tableName)
	
	estimatedCount := db.estimateRowsInBounds(tableName, viewport)
	actualLimit := int(float64(estimatedCount) * randomFactor)
	if actualLimit > maxElements {
		actualLimit = maxElements
	}
	if actualLimit < 1 {
		actualLimit = 1
	}
	
	rows, err := db.queryRows(query, "failed to query nodes in bounds",
		viewport.MinLat, viewport.MaxLat, viewport.MinLng, viewport.MaxLng, actualLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []NodeData
	for rows.Next() {
		var n NodeData
		if err := rows.Scan(&n.ID, &n.Lng, &n.Lat, &n.RawXML); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	
	return nodes, nil
}

// GetLinksForNodesInBounds retrieves network links where at least one node is in bounds
func (db *Database) GetLinksForNodesInBounds(processId int, viewport ViewportBounds, randomFactor float64, maxElements int) ([]LinkData, error) {
	linksTable := fmt.Sprintf("NETWORK_%d_links", processId)
	nodesTable := fmt.Sprintf("NETWORK_%d_nodes", processId)
	
	// Get links where at least one node is in bounds
	query := fmt.Sprintf(`
		SELECT DISTINCT l.id, l.from_node, l.to_node, l.raw_xml
		FROM %s l
		WHERE l.from_node IN (
			SELECT id FROM %s WHERE lat BETWEEN ? AND ? AND lng BETWEEN ? AND ?
		) OR l.to_node IN (
			SELECT id FROM %s WHERE lat BETWEEN ? AND ? AND lng BETWEEN ? AND ?
		)
		ORDER BY RANDOM()
		LIMIT ?`, linksTable, nodesTable, nodesTable)
	
	rows, err := db.queryRows(query, "failed to query links in bounds",
		viewport.MinLat, viewport.MaxLat, viewport.MinLng, viewport.MaxLng,
		viewport.MinLat, viewport.MaxLat, viewport.MinLng, viewport.MaxLng,
		maxElements)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []LinkData
	for rows.Next() {
		var l LinkData
		if err := rows.Scan(&l.ID, &l.FromNode, &l.ToNode, &l.RawXML); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	
	return links, nil
}

// GetPTLinesInBounds retrieves PT lines that have stops within viewport bounds
func (db *Database) GetPTLinesInBounds(processId int, viewport ViewportBounds, mode string, randomFactor float64, maxElements int) ([]Line, error) {
	linesTable := fmt.Sprintf("pt_%d_lines", processId)
	stopsTable := fmt.Sprintf("pt_%d_stops", processId)
	
	// Get lines that have stops in the viewport bounds
	// Routes are now in a separate table
	routesTable := fmt.Sprintf("pt_%d_routes", processId)
	routeStopsTable := fmt.Sprintf("pt_%d_route_stops", processId)
	
	query := fmt.Sprintf(`
		SELECT DISTINCT l.id, l.name, l.mode
		FROM %s l
		WHERE l.mode = ? AND EXISTS (
			SELECT 1 FROM %s r
			JOIN %s rs ON r.id = rs.route_id
			JOIN %s s ON rs.stop_id = s.stop_id
			WHERE r.line_id = l.id
			AND s.lat BETWEEN ? AND ? 
			AND s.lng BETWEEN ? AND ?
		)
		ORDER BY RANDOM()
		LIMIT ?`, linesTable, routesTable, routeStopsTable, stopsTable)
	
	rows, err := db.queryRows(query, "failed to query PT lines in bounds",
		mode, viewport.MinLat, viewport.MaxLat, viewport.MinLng, viewport.MaxLng, maxElements)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []Line
	for rows.Next() {
		var l Line
		if err := rows.Scan(&l.ID, &l.Name, &l.Mode); err != nil {
			return nil, err
		}
		lines = append(lines, l)
	}
	
	return lines, nil
}

// GetPTLineSummaries retrieves all PT lines for a given mode
func (db *Database) GetPTLineSummaries(processId int, mode string) ([]Line, error) {
	linesTable := fmt.Sprintf("pt_%d_lines", processId)
	
	query := fmt.Sprintf(`
		SELECT id, name, mode
		FROM %s
		WHERE mode = ?
		ORDER BY name`, linesTable)
	
	rows, err := db.queryRows(query, "failed to query PT line summaries", mode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []Line
	for rows.Next() {
		var l Line
		if err := rows.Scan(&l.ID, &l.Name, &l.Mode); err != nil {
			return nil, err
		}
		lines = append(lines, l)
	}
	
	return lines, nil
}

// Helper function to estimate rows in bounds for random sampling
func (db *Database) estimateRowsInBounds(tableName string, viewport ViewportBounds) int {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE lat BETWEEN ? AND ? AND lng BETWEEN ? AND ?", tableName)
	var count int
	err := db.conn.QueryRow(query, viewport.MinLat, viewport.MaxLat, viewport.MinLng, viewport.MaxLng).Scan(&count)
	if err != nil {
		return 1000 // Fallback estimate
	}
	return count
}

// CreateNetworkTables creates network tables for a given process
func (db *Database) CreateNetworkTables(processID int) error {
	// Create nodes table
	nodesTable := fmt.Sprintf("network_%d_nodes", processID)
	if err := db.execTableQuery(fmt.Sprintf(createNetworkNodesTableQuery, nodesTable)); err != nil {
		return fmt.Errorf("failed to create network nodes table: %w", err)
	}
	
	// Create spatial index for nodes
	indexName := fmt.Sprintf("network_%d_nodes", processID)
	if err := db.execTableQuery(fmt.Sprintf(createNodesSpatialIndex, indexName, nodesTable)); err != nil {
		return fmt.Errorf("failed to create nodes spatial index: %w", err)
	}
	
	// Create links table
	linksTable := fmt.Sprintf("network_%d_links", processID)
	if err := db.execTableQuery(fmt.Sprintf(createNetworkLinksTableQuery, linksTable)); err != nil {
		return fmt.Errorf("failed to create network links table: %w", err)
	}
	
	return nil
}

