// Step 5: Trip Database Creation & Step 6: Trip Collection
// This file handles both database creation from trips.txt and querying trips by selected services.
// Step 5 creates SQLite database, Step 6 queries for selected service trips.
package pt

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// Step 6: Trip Collection via Database Query
func (pto *PTOrchestrator) collectTrips() error {
	pto.logMessage("Starting Trip Collection via Database Query")

	// State integration available for future TUI implementation
	if pto.tuiEnabled {
		// Future: set_state("PT_6")
	}

	// Read the selected services from Step 4 (same as before)
	selectedServicesPath := filepath.Join(pto.config.TempDir, ServicesDir, SelectedServicesFile)
	selectedServicesData, err := os.ReadFile(selectedServicesPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read selected services file: %v", err))
		return NewFileParsingError(SelectedServicesFile, err.Error())
	}

	var selectedServices []SelectedService
	if err := json.Unmarshal(selectedServicesData, &selectedServices); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to parse selected services JSON: %v", err))
		return NewFileParsingError(SelectedServicesFile, err.Error())
	}

	// Create a map of selected service IDs for query
	selectedServiceIDs := make(map[string]bool)
	serviceList := make([]string, 0)
	for _, service := range selectedServices {
		if service.Selected {
			selectedServiceIDs[service.ServiceID] = true
			serviceList = append(serviceList, service.ServiceID)
			pto.logMessage(fmt.Sprintf("Querying trips for service: %s (%d routes)", service.ServiceID, len(service.Routes)))
		}
	}

	if len(selectedServiceIDs) == 0 {
		pto.logMessage("No services selected for trip collection")
		return NewInvalidServiceError("no services selected")
	}

	pto.logMessage(fmt.Sprintf("Querying database for %d selected services", len(selectedServiceIDs)))

	// Open database
	db, err := pto.openTripDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	// Query trips from database
	tripMappings, err := pto.queryTripsFromDatabase(db, serviceList)
	if err != nil {
		return err
	}

	if len(tripMappings) == 0 {
		pto.logMessage("No trips found for selected services")
		return NewInvalidServiceError("no trips found for selected services")
	}

	// Create output directory for trip data (same as before)
	tripDir := filepath.Join(pto.config.TempDir, TripsDir)
	if err := os.MkdirAll(tripDir, DirPermissions); err != nil {
		return NewDirectoryError(tripDir, "create")
	}

	// Create output file (same format as before)
	outputPath := filepath.Join(tripDir, TripMappingsFile)
	jsonData, err := json.MarshalIndent(tripMappings, "", "  ")
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to marshal trip mappings to JSON: %v", err))
		return NewFileParsingError(TripMappingsFile, err.Error())
	}

	if err := os.WriteFile(outputPath, jsonData, FilePermissions); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to write trip mappings file: %v", err))
		return NewFileParsingError(TripMappingsFile, err.Error())
	}

	// Final live update with completion status
	if pto.tuiEnabled && pto.tui != nil {
		pto.tui.UpdateCounter("trips", fmt.Sprintf("%d", len(tripMappings)))
		pto.tui.UpdateCounter("progress", "100%")
	}

	pto.logMessage(fmt.Sprintf("Trip collection completed successfully. Collected %d trips for %s via database query",
		len(tripMappings), pto.config.TargetDate))
	return nil
}

// Step 5: Trip Database Creation
func (pto *PTOrchestrator) buildTripDatabase() error {
	pto.logMessage("Starting Trip Database Creation")

	// State integration available for future TUI implementation
	if pto.tuiEnabled {
		// Future: set_state("PT_5")
	}

	// Get database path
	dbPath := pto.getTripDatabasePath()

	// Remove existing database if present
	if _, err := os.Stat(dbPath); err == nil {
		if err := os.Remove(dbPath); err != nil {
			pto.logMessage(fmt.Sprintf("Failed to remove existing database: %v", err))
			return NewFileParsingError(TripDatabaseFile, err.Error())
		}
	}

	// Create database
	db, err := pto.openTripDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	// Create table and index
	if err := pto.createTripTable(db); err != nil {
		return err
	}

	// Import trips from CSV
	if err := pto.importTripsFromCSV(db); err != nil {
		return err
	}

	pto.logMessage("Trip database creation completed successfully")
	return nil
}

// Database helper methods
func (pto *PTOrchestrator) getTripDatabasePath() string {
	return filepath.Join(pto.config.TempDir, TripDatabaseFile)
}

func (pto *PTOrchestrator) openTripDatabase() (*sql.DB, error) {
	dbPath := pto.getTripDatabasePath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to open trip database: %v", err))
		return nil, NewFileParsingError(TripDatabaseFile, err.Error())
	}
	return db, nil
}

func (pto *PTOrchestrator) createTripTable(db *sql.DB) error {
	// Create simple trips table with index
	createTableSQL := `
	CREATE TABLE trips (
		trip_id TEXT PRIMARY KEY,
		route_id TEXT NOT NULL,
		service_id TEXT NOT NULL
	);
	CREATE INDEX idx_service_id ON trips(service_id);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to create trip table: %v", err))
		return NewFileParsingError(TripDatabaseFile, fmt.Sprintf("table creation failed: %v", err))
	}

	return nil
}

func (pto *PTOrchestrator) importTripsFromCSV(db *sql.DB) error {
	tripsPath := filepath.Join(pto.config.GTFSDirectory, TripsFile)

	file, err := os.Open(tripsPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to open trips.txt: %v", err))
		return NewFileParsingError(TripsFile, err.Error())
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read headers
	headers, err := reader.Read()
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read trips.txt headers: %v", err))
		return NewFileParsingError(TripsFile, "invalid header row")
	}

	// Create header index map
	headerIndex := make(map[string]int)
	for i, header := range headers {
		// Strip UTF-8 BOM from header if present (fixes STM GTFS data)
		cleanHeader := strings.TrimPrefix(header, "\uFEFF")
		headerIndex[cleanHeader] = i
	}

	// Verify required columns exist
	requiredColumns := []string{"trip_id", "route_id", "service_id"}
	for _, col := range requiredColumns {
		if _, exists := headerIndex[col]; !exists {
			pto.logMessage(fmt.Sprintf("Missing required column in trips.txt: %s", col))
			return NewFileParsingError(TripsFile, fmt.Sprintf("missing column: %s", col))
		}
	}

	// Begin transaction for bulk insert
	tx, err := db.Begin()
	if err != nil {
		return NewFileParsingError(TripDatabaseFile, fmt.Sprintf("transaction failed: %v", err))
	}
	defer tx.Rollback()

	// Prepare insert statement
	stmt, err := tx.Prepare("INSERT INTO trips (trip_id, route_id, service_id) VALUES (?, ?, ?)")
	if err != nil {
		return NewFileParsingError(TripDatabaseFile, fmt.Sprintf("prepare failed: %v", err))
	}
	defer stmt.Close()

	rowCounter := 0
	batchSize := 1000

	// Process CSV data
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Skip invalid rows
		}

		tripID := record[headerIndex["trip_id"]]
		routeID := record[headerIndex["route_id"]]
		serviceID := record[headerIndex["service_id"]]

		// Insert record
		if _, err := stmt.Exec(tripID, routeID, serviceID); err != nil {
			continue // Skip problematic records
		}

		rowCounter++

		// Update statistics
		pto.stats.TripCounter = rowCounter

		// Commit batch every 1000 records
		if rowCounter%batchSize == 0 {
			if err := tx.Commit(); err != nil {
				return NewFileParsingError(TripDatabaseFile, fmt.Sprintf("batch commit failed: %v", err))
			}

			// Start new transaction
			tx, err = db.Begin()
			if err != nil {
				return NewFileParsingError(TripDatabaseFile, fmt.Sprintf("new transaction failed: %v", err))
			}
			stmt, err = tx.Prepare("INSERT INTO trips (trip_id, route_id, service_id) VALUES (?, ?, ?)")
			if err != nil {
				return NewFileParsingError(TripDatabaseFile, fmt.Sprintf("prepare failed: %v", err))
			}

			// Update TUI progress
			if pto.tuiEnabled && pto.tui != nil {
				pto.tui.UpdateCounter("rows", fmt.Sprintf("%d", rowCounter))
			}

			pto.logMessage(fmt.Sprintf("Imported %d trips to database", rowCounter))
		}

		// Check for quit request every 500 rows
		if rowCounter%500 == 0 && pto.checkQuitRequested() {
			pto.logMessage("Database import aborted by user")
			return nil
		}
	}

	// Final commit
	if err := tx.Commit(); err != nil {
		return NewFileParsingError(TripDatabaseFile, fmt.Sprintf("final commit failed: %v", err))
	}

	// Final TUI update
	if pto.tuiEnabled && pto.tui != nil {
		pto.tui.UpdateCounter("rows", fmt.Sprintf("%d", rowCounter))
		pto.tui.UpdateCounter("progress", "100%")
	}

	pto.logMessage(fmt.Sprintf("Database import completed: %d trips imported", rowCounter))
	return nil
}

func (pto *PTOrchestrator) queryTripsFromDatabase(db *sql.DB, serviceList []string) ([]TripMapping, error) {
	// Build parameterized query for selected services
	placeholders := make([]string, len(serviceList))
	params := make([]interface{}, len(serviceList))
	for i, serviceID := range serviceList {
		placeholders[i] = "?"
		params[i] = serviceID
	}

	query := fmt.Sprintf(`
		SELECT trip_id, route_id, service_id
		FROM trips
		WHERE service_id IN (%s)
		ORDER BY service_id, route_id, trip_id
	`, strings.Join(placeholders, ","))

	pto.logMessage(fmt.Sprintf("Executing database query for %d services", len(serviceList)))

	// Execute query
	rows, err := db.Query(query, params...)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Database query failed: %v", err))
		return nil, NewFileParsingError(TripDatabaseFile, fmt.Sprintf("query failed: %v", err))
	}
	defer rows.Close()

	var tripMappings []TripMapping
	tripCounter := 0

	// Process query results
	for rows.Next() {
		var tripID, routeID, serviceID string
		if err := rows.Scan(&tripID, &routeID, &serviceID); err != nil {
			continue // Skip problematic rows
		}

		tripMapping := TripMapping{
			TripID:  tripID,
			RouteID: routeID,
		}
		tripMappings = append(tripMappings, tripMapping)
		tripCounter++

		// Update statistics
		pto.stats.TripCounter = tripCounter

		// Log every 10th trip for single-day processing
		if tripCounter%10 == 0 {
			pto.logMessage(fmt.Sprintf("Queried %d trips for %s (last: %s for route %s, service %s)",
				tripCounter, pto.config.TargetDate, tripID, routeID, serviceID))
		}

		// Update TUI progress
		if pto.tuiEnabled && pto.tui != nil {
			if tripCounter%50 == 0 {
				pto.tui.UpdateCounter("trips", fmt.Sprintf("%d", tripCounter))
			}
		}
	}

	// Check for query errors
	if err := rows.Err(); err != nil {
		pto.logMessage(fmt.Sprintf("Query result processing failed: %v", err))
		return nil, NewFileParsingError(TripDatabaseFile, fmt.Sprintf("result processing failed: %v", err))
	}

	pto.logMessage(fmt.Sprintf("Database query completed: %d trips retrieved", tripCounter))
	return tripMappings, nil
}
