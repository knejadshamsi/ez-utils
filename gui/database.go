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

	createNetworkLinksTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			from_node TEXT NOT NULL,
			to_node TEXT NOT NULL,
			raw_xml TEXT
		)`

	dropTableQuery = "DROP TABLE IF EXISTS %s"

	// PT table creation queries
	createPTStopsTableQuery = `
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			lng REAL NOT NULL,
			lat REAL NOT NULL,
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
	
	// Create spatial index for stops
	indexName := fmt.Sprintf("stops_%d", processID)
	if err := db.execTableQuery(fmt.Sprintf(createStopsSpatialIndex, indexName, stopsTable)); err != nil {
		return fmt.Errorf("failed to create stops spatial index: %w", err)
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

// Spatial query functions

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
	tableName := fmt.Sprintf("NETWORK_%d_nodes", processId)
	
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
func (db *Database) GetPTLinesInBounds(processId int, viewport ViewportBounds, mode string, randomFactor float64, maxElements int) ([]PTLineData, error) {
	linesTable := fmt.Sprintf("pt_data_%d_lines", processId)
	stopsTable := fmt.Sprintf("pt_data_%d_stops", processId)
	routesTable := fmt.Sprintf("pt_data_%d_routes", processId)
	routeStopsTable := fmt.Sprintf("pt_data_%d_route_stops", processId)
	
	// Get lines that have stops in the viewport bounds
	query := fmt.Sprintf(`
		SELECT DISTINCT l.id, l.mode, l.raw_xml
		FROM %s l
		WHERE l.mode = ? AND l.id IN (
			SELECT r.line_id FROM %s r
			WHERE r.id IN (
				SELECT rs.route_id FROM %s rs
				WHERE rs.stop_ref_id IN (
					SELECT s.id FROM %s s
					WHERE s.lat BETWEEN ? AND ? AND s.lng BETWEEN ? AND ?
				)
			)
		)
		ORDER BY RANDOM()
		LIMIT ?`, linesTable, routesTable, routeStopsTable, stopsTable)
	
	rows, err := db.queryRows(query, "failed to query PT lines in bounds",
		mode, viewport.MinLat, viewport.MaxLat, viewport.MinLng, viewport.MaxLng, maxElements)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []PTLineData
	for rows.Next() {
		var l PTLineData
		if err := rows.Scan(&l.ID, &l.Mode, &l.RawXML); err != nil {
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
	nodesTable := fmt.Sprintf("NETWORK_%d_nodes", processID)
	if err := db.execTableQuery(fmt.Sprintf(createNetworkNodesTableQuery, nodesTable)); err != nil {
		return fmt.Errorf("failed to create network nodes table: %w", err)
	}
	
	// Create spatial index for nodes
	indexName := fmt.Sprintf("NETWORK_%d_nodes", processID)
	if err := db.execTableQuery(fmt.Sprintf(createNodesSpatialIndex, indexName, nodesTable)); err != nil {
		return fmt.Errorf("failed to create nodes spatial index: %w", err)
	}
	
	// Create links table
	linksTable := fmt.Sprintf("NETWORK_%d_links", processID)
	if err := db.execTableQuery(fmt.Sprintf(createNetworkLinksTableQuery, linksTable)); err != nil {
		return fmt.Errorf("failed to create network links table: %w", err)
	}
	
	return nil
}