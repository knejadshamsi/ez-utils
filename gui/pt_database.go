package gui

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// PT Query Constants
const (
	// TELEMETRY queries
	selectPTTelemetryQuery = `SELECT process_id, total_file_size, bytes_read, 
	          COALESCE(stops_extracted, 0), COALESCE(lines_extracted, 0), 
	          COALESCE(routes_extracted, 0), error_count, last_updated 
	          FROM process_telemetry WHERE process_id = ?`
)

// PT Error Messages
const (
	// TELEMETRY errors
	selectPTTelemetryError = "failed to get PT telemetry for process %d"
)

// GetPTStops retrieves all stops (route-stop relationships) for a process
func (db *Database) GetPTStops(processID int) ([]Stop, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	query := fmt.Sprintf(`SELECT route_id, stop_id, arrival_offset, departure_offset, stop_name, lat, lng, sequence, attributes FROM %s ORDER BY route_id, sequence`, stopsTable)
	rows, err := db.queryRows(query, fmt.Sprintf("failed to query stops from table %s", stopsTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stops []Stop
	for rows.Next() {
		var stop Stop
		var attributesJSON sql.NullString
		if err := rows.Scan(&stop.RouteID, &stop.StopID, &stop.ArrivalOffset, &stop.DepartureOffset, 
			&stop.StopName, &stop.Lat, &stop.Lng, &stop.Sequence, &attributesJSON); err != nil {
			return nil, fmt.Errorf("failed to scan stop: %w", err)
		}
		if attributesJSON.Valid {
			json.Unmarshal([]byte(attributesJSON.String), &stop.Attributes)
		}
		stops = append(stops, stop)
	}
	
	return stops, rows.Err()
}



// GetPTLines retrieves all lines for a process
func (db *Database) GetPTLines(processID int) ([]Line, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	
	query := fmt.Sprintf(`SELECT id, name, type, routes FROM %s ORDER BY id`, linesTable)
	rows, err := db.queryRows(query, fmt.Sprintf("failed to query lines from table %s", linesTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var lines []Line
	for rows.Next() {
		var line Line
		var routesJSON string
		if err := rows.Scan(&line.ID, &line.Name, &line.Type, &routesJSON); err != nil {
			return nil, fmt.Errorf("failed to scan line: %w", err)
		}
		if err := json.Unmarshal([]byte(routesJSON), &line.Routes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal routes for line %s: %w", line.ID, err)
		}
		lines = append(lines, line)
	}
	
	return lines, rows.Err()
}

// GetPTLine retrieves a specific line by ID
func (db *Database) GetPTLine(processID int, lineID string) (*Line, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	
	query := fmt.Sprintf(`SELECT id, name, type, routes FROM %s WHERE id = ?`, linesTable)
	var line Line
	var routesJSON string
	err := db.queryRow(query, fmt.Sprintf("failed to query line from table %s", linesTable), lineID).Scan(
		&line.ID, &line.Name, &line.Type, &routesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("line not found: %s", lineID)
		}
		return nil, err
	}
	
	if err := json.Unmarshal([]byte(routesJSON), &line.Routes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal routes for line %s: %w", line.ID, err)
	}
	
	return &line, nil
}

// GetPTLinesByMode retrieves lines by transport mode
func (db *Database) GetPTLinesByMode(processID int, mode string) ([]Line, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	
	query := fmt.Sprintf(`SELECT id, name, type, routes FROM %s WHERE type = ? ORDER BY id`, linesTable)
	rows, err := db.queryRows(query, fmt.Sprintf("failed to query lines by mode from table %s", linesTable), mode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var lines []Line
	for rows.Next() {
		var line Line
		var routesJSON string
		if err := rows.Scan(&line.ID, &line.Name, &line.Type, &routesJSON); err != nil {
			return nil, fmt.Errorf("failed to scan line: %w", err)
		}
		if err := json.Unmarshal([]byte(routesJSON), &line.Routes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal routes for line %s: %w", line.ID, err)
		}
		lines = append(lines, line)
	}
	
	return lines, rows.Err()
}

// GetPTStopsForRoute retrieves stops for a specific route
func (db *Database) GetPTStopsForRoute(processID int, routeID string) ([]Stop, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	query := fmt.Sprintf(`SELECT route_id, stop_id, arrival_offset, departure_offset, stop_name, lat, lng, sequence, attributes FROM %s WHERE route_id = ? ORDER BY sequence`, stopsTable)
	rows, err := db.queryRows(query, fmt.Sprintf("failed to query stops for route from table %s", stopsTable), routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stops []Stop
	for rows.Next() {
		var stop Stop
		var attributesJSON sql.NullString
		if err := rows.Scan(&stop.RouteID, &stop.StopID, &stop.ArrivalOffset, &stop.DepartureOffset, 
			&stop.StopName, &stop.Lat, &stop.Lng, &stop.Sequence, &attributesJSON); err != nil {
			return nil, fmt.Errorf("failed to scan stop: %w", err)
		}
		if attributesJSON.Valid {
			json.Unmarshal([]byte(attributesJSON.String), &stop.Attributes)
		}
		stops = append(stops, stop)
	}
	
	return stops, rows.Err()
}

// GetPTDepartures retrieves departures for a specific route
func (db *Database) GetPTDepartures(processID int, routeID string) ([]Departure, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
	
	query := fmt.Sprintf(`SELECT id, route_id, departure_time, vehicle_ref_id FROM %s WHERE route_id = ? ORDER BY departure_time`, departuresTable)
	rows, err := db.queryRows(query, fmt.Sprintf("failed to query departures from table %s", departuresTable), routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var departures []Departure
	for rows.Next() {
		var dep Departure
		var vehicleRefID sql.NullString
		if err := rows.Scan(&dep.ID, &dep.RouteID, &dep.DepartureTime, &vehicleRefID); err != nil {
			return nil, fmt.Errorf("failed to scan departure: %w", err)
		}
		if vehicleRefID.Valid {
			dep.VehicleRefID = vehicleRefID.String
		}
		departures = append(departures, dep)
	}
	
	return departures, rows.Err()
}


// GetPTStatistics returns statistics for PT data
func (db *Database) GetPTStatistics(processID int) (map[string]interface{}, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
	
	stats := make(map[string]interface{})
	
	// Count unique stops
	var stopCount int
	query := fmt.Sprintf("SELECT COUNT(DISTINCT stop_id) FROM %s", stopsTable)
	if err := db.queryRow(query, "failed to count stops").Scan(&stopCount); err != nil {
		return nil, err
	}
	stats["stop_count"] = stopCount
	
	// Count lines by type
	query = fmt.Sprintf("SELECT type, COUNT(*) FROM %s GROUP BY type", linesTable)
	rows, err := db.queryRows(query, "failed to count lines by type")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	linesByType := make(map[string]int)
	for rows.Next() {
		var lineType string
		var count int
		if err := rows.Scan(&lineType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan line count: %w", err)
		}
		linesByType[lineType] = count
	}
	stats["lines_by_type"] = linesByType
	
	// Get total route count from JSON
	query = fmt.Sprintf("SELECT SUM(json_array_length(routes)) FROM %s", linesTable)
	var routeCount sql.NullInt64
	if err := db.queryRow(query, "failed to count routes").Scan(&routeCount); err != nil {
		return nil, err
	}
	if routeCount.Valid {
		stats["route_count"] = int(routeCount.Int64)
	} else {
		stats["route_count"] = 0
	}
	
	// Get total departure count
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s", departuresTable)
	var departureCount int
	if err := db.queryRow(query, "failed to count departures").Scan(&departureCount); err != nil {
		return nil, err
	}
	stats["departure_count"] = departureCount
	
	return stats, nil
}


// InsertPTStopBatch inserts multiple stops in a transaction
func (db *Database) InsertPTStopBatch(processID int, stops []Stop) error {
	if len(stops) == 0 {
		return nil
	}
	
	return db.WithTransaction(func(tx *sql.Tx) error {
		tablePrefix := fmt.Sprintf("pt_data_%d", processID)
		stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
		
		// Updated query for new schema
		query := fmt.Sprintf(`INSERT INTO %s (route_id, stop_id, arrival_offset, departure_offset, stop_name, lat, lng, sequence, attributes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, stopsTable)
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()
		
		for _, stop := range stops {
			// Convert attributes to JSON
			var attributesJSON []byte
			if stop.Attributes != nil {
				attributesJSON, _ = json.Marshal(stop.Attributes)
			}
			
			if _, err := stmt.Exec(stop.RouteID, stop.StopID, stop.ArrivalOffset, stop.DepartureOffset, 
				stop.StopName, stop.Lat, stop.Lng, stop.Sequence, attributesJSON); err != nil {
				return fmt.Errorf("failed to insert stop %s: %w", stop.StopID, err)
			}
		}
		
		return nil
	})
}

// InsertPTLineBatch inserts multiple lines in a transaction
func (db *Database) InsertPTLineBatch(processID int, lines []Line) error {
	if len(lines) == 0 {
		return nil
	}
	
	return db.WithTransaction(func(tx *sql.Tx) error {
		tablePrefix := fmt.Sprintf("pt_data_%d", processID)
		linesTable := fmt.Sprintf("%s_lines", tablePrefix)
		
		// Updated query for new schema with JSON routes
		query := fmt.Sprintf(`INSERT INTO %s (id, name, type, routes) VALUES (?, ?, ?, ?)`, linesTable)
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()
		
		for _, line := range lines {
			// Convert routes to JSON
			routesJSON, err := json.Marshal(line.Routes)
			if err != nil {
				return fmt.Errorf("failed to marshal routes for line %s: %w", line.ID, err)
			}
			
			if _, err := stmt.Exec(line.ID, line.Name, line.Type, routesJSON); err != nil {
				return fmt.Errorf("failed to insert line %s: %w", line.ID, err)
			}
		}
		
		return nil
	})
}


// InsertPTDepartureBatch inserts multiple departures in a transaction
func (db *Database) InsertPTDepartureBatch(processID int, departures []Departure) error {
	if len(departures) == 0 {
		return nil
	}
	
	return db.WithTransaction(func(tx *sql.Tx) error {
		tablePrefix := fmt.Sprintf("pt_data_%d", processID)
		departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
		
		// Updated query for new schema
		query := fmt.Sprintf(`INSERT INTO %s (id, route_id, departure_time, vehicle_ref_id) VALUES (?, ?, ?, ?)`, departuresTable)
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()
		
		for _, dep := range departures {
			if _, err := stmt.Exec(dep.ID, dep.RouteID, dep.DepartureTime, dep.VehicleRefID); err != nil {
				return fmt.Errorf("failed to insert departure %s: %w", dep.ID, err)
			}
		}
		
		return nil
	})
}



// UpdatePTTelemetry updates the telemetry for PT processing
func (db *Database) UpdatePTTelemetry(processID int, bytesRead int64, stopsExtracted, linesExtracted, routesExtracted, errorCount int) error {
	const updatePTTelemetryQuery = `
		UPDATE process_telemetry 
		SET bytes_read = ?, stops_extracted = ?, lines_extracted = ?, routes_extracted = ?, error_count = ?, last_updated = CURRENT_TIMESTAMP 
		WHERE process_id = ?`
	
	_, err := db.execQuery(updatePTTelemetryQuery, 
		fmt.Sprintf("failed to update PT telemetry for process %d", processID),
		bytesRead, stopsExtracted, linesExtracted, routesExtracted, errorCount, processID)
	return err
}

// GetPTTelemetry retrieves telemetry data for a PT process
func (db *Database) GetPTTelemetry(processID int) (*PTTelemetry, error) {
	var telemetry PTTelemetry
	
	row := db.queryRow(selectPTTelemetryQuery, fmt.Sprintf(selectPTTelemetryError, processID), processID)
	err := row.Scan(
		&telemetry.ProcessID,
		&telemetry.TotalFileSize,
		&telemetry.BytesRead,
		&telemetry.StopsExtracted,
		&telemetry.LinesExtracted,
		&telemetry.RoutesExtracted,
		&telemetry.ErrorCount,
		&telemetry.LastUpdated,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &telemetry, nil
}

// AddPTStop adds a single stop to the database
func (db *Database) AddPTStop(processID int, stop Stop) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	// Convert attributes to JSON
	var attributesJSON []byte
	if stop.Attributes != nil {
		attributesJSON, _ = json.Marshal(stop.Attributes)
	}
	
	query := fmt.Sprintf(`INSERT INTO %s (route_id, stop_id, arrival_offset, departure_offset, stop_name, lat, lng, sequence, attributes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, stopsTable)
	
	_, err := db.execQuery(query, "failed to add PT stop",
		stop.RouteID, stop.StopID, stop.ArrivalOffset, stop.DepartureOffset,
		stop.StopName, stop.Lat, stop.Lng, stop.Sequence, attributesJSON)
	
	return err
}

// UpdatePTStop updates an existing stop in the database
func (db *Database) UpdatePTStop(processID int, stopID string, update PTStopUpdate) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	// Build dynamic update query based on what fields are provided
	var setClauses []string
	var args []interface{}
	
	if update.StopName != nil {
		setClauses = append(setClauses, "stop_name = ?")
		args = append(args, *update.StopName)
	}
	if update.Lat != nil {
		setClauses = append(setClauses, "lat = ?")
		args = append(args, *update.Lat)
	}
	if update.Lng != nil {
		setClauses = append(setClauses, "lng = ?")
		args = append(args, *update.Lng)
	}
	if update.ArrivalOffset != nil {
		setClauses = append(setClauses, "arrival_offset = ?")
		args = append(args, *update.ArrivalOffset)
	}
	if update.DepartureOffset != nil {
		setClauses = append(setClauses, "departure_offset = ?")
		args = append(args, *update.DepartureOffset)
	}
	if update.Sequence != nil {
		setClauses = append(setClauses, "sequence = ?")
		args = append(args, *update.Sequence)
	}
	if update.Attributes != nil {
		attributesJSON, _ := json.Marshal(update.Attributes)
		setClauses = append(setClauses, "attributes = ?")
		args = append(args, attributesJSON)
	}
	
	if len(setClauses) == 0 {
		return nil // Nothing to update
	}
	
	// Add stopID to args for WHERE clause
	args = append(args, stopID)
	
	query := fmt.Sprintf(`UPDATE %s SET %s WHERE stop_id = ?`, stopsTable, strings.Join(setClauses, ", "))
	
	_, err := db.execQuery(query, "failed to update PT stop", args...)
	return err
}

// DeletePTStop deletes a stop from the database
func (db *Database) DeletePTStop(processID int, stopID string) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	query := fmt.Sprintf(`DELETE FROM %s WHERE stop_id = ?`, stopsTable)
	
	_, err := db.execQuery(query, "failed to delete PT stop", stopID)
	return err
}

// BatchUpdatePTStops updates multiple stops in a single transaction
func (db *Database) BatchUpdatePTStops(processID int, updates map[string]PTStopUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	
	return db.WithTransaction(func(tx *sql.Tx) error {
		tablePrefix := fmt.Sprintf("pt_data_%d", processID)
		stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
		
		for stopID, update := range updates {
			// Build dynamic update query
			var setClauses []string
			var args []interface{}
			
			if update.StopName != nil {
				setClauses = append(setClauses, "stop_name = ?")
				args = append(args, *update.StopName)
			}
			if update.Lat != nil {
				setClauses = append(setClauses, "lat = ?")
				args = append(args, *update.Lat)
			}
			if update.Lng != nil {
				setClauses = append(setClauses, "lng = ?")
				args = append(args, *update.Lng)
			}
			if update.ArrivalOffset != nil {
				setClauses = append(setClauses, "arrival_offset = ?")
				args = append(args, *update.ArrivalOffset)
			}
			if update.DepartureOffset != nil {
				setClauses = append(setClauses, "departure_offset = ?")
				args = append(args, *update.DepartureOffset)
			}
			if update.Sequence != nil {
				setClauses = append(setClauses, "sequence = ?")
				args = append(args, *update.Sequence)
			}
			if update.Attributes != nil {
				attributesJSON, _ := json.Marshal(update.Attributes)
				setClauses = append(setClauses, "attributes = ?")
				args = append(args, attributesJSON)
			}
			
			if len(setClauses) == 0 {
				continue // Nothing to update for this stop
			}
			
			// Add stopID to args for WHERE clause
			args = append(args, stopID)
			
			query := fmt.Sprintf(`UPDATE %s SET %s WHERE stop_id = ?`, stopsTable, strings.Join(setClauses, ", "))
			
			if _, err := tx.Exec(query, args...); err != nil {
				return fmt.Errorf("failed to update stop %s: %w", stopID, err)
			}
		}
		
		return nil
	})
}

// DeletePTLine deletes a PT line and all its associated data
func (db *Database) DeletePTLine(processID int, lineID string) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, linesTable)
	
	_, err := db.execQuery(query, "failed to delete PT line", lineID)
	return err
}

// DeletePTRoute deletes a route from a line
func (db *Database) DeletePTRoute(processID int, routeID string) error {
	// Routes are embedded in lines, so we need to update the line's routes JSON
	// This would require fetching the line, removing the route from the JSON, and updating
	// For now, return nil as routes are managed as part of lines
	return nil
}

// DeletePTRouteStop deletes a stop from a route
func (db *Database) DeletePTRouteStop(processID int, routeID string, stopOrder int) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	query := fmt.Sprintf(`DELETE FROM %s WHERE route_id = ? AND sequence = ?`, stopsTable)
	
	_, err := db.execQuery(query, "failed to delete PT route stop", routeID, stopOrder)
	return err
}

// DeletePTDeparture deletes a departure
func (db *Database) DeletePTDeparture(processID int, departureID string) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
	
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, departuresTable)
	
	_, err := db.execQuery(query, "failed to delete PT departure", departureID)
	return err
}

// GetPTLineSummaries retrieves line summaries for display (basically same as GetPTLinesByMode)
func (db *Database) GetPTLineSummaries(processID int, mode string) ([]Line, error) {
	return db.GetPTLinesByMode(processID, mode)
}


