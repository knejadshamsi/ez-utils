package gui

import (
	"database/sql"
	"fmt"
	"strings"
)

// PT Query Constants
const (
	// SELECT queries
	selectAllPTStopsQuery     = "SELECT id, lng, lat, name, raw_xml FROM %s ORDER BY id"
	selectPTStopByIDQuery     = "SELECT id, lng, lat, name, raw_xml FROM %s WHERE id = ?"
	selectAllPTLinesQuery     = "SELECT id, mode, raw_xml FROM %s ORDER BY id"
	selectPTLineByIDQuery     = "SELECT id, mode, raw_xml FROM %s WHERE id = ?"
	selectPTLinesByModeQuery  = "SELECT id, mode, raw_xml FROM %s WHERE mode = ? ORDER BY id"
	selectPTRoutesByLineQuery = "SELECT id, line_id, raw_xml FROM %s WHERE line_id = ? ORDER BY id"
	selectPTRouteStopsQuery   = "SELECT route_id, stop_ref_id, stop_order, arrival_offset, departure_offset FROM %s WHERE route_id = ? ORDER BY stop_order"
	selectPTDeparturesQuery   = "SELECT id, route_id, departure_time FROM %s WHERE route_id = ? ORDER BY departure_time"
	
	// INSERT queries
	insertPTStopQuery      = "INSERT INTO %s (id, lng, lat, name, raw_xml) VALUES (?, ?, ?, ?, ?)"
	insertPTLineQuery      = "INSERT INTO %s (id, mode, raw_xml) VALUES (?, ?, ?)"
	insertPTRouteQuery     = "INSERT INTO %s (id, line_id, raw_xml) VALUES (?, ?, ?)"
	insertPTRouteStopQuery = "INSERT INTO %s (route_id, stop_ref_id, stop_order, arrival_offset, departure_offset) VALUES (?, ?, ?, ?, ?)"
	insertPTDepartureQuery = "INSERT INTO %s (id, route_id, departure_time) VALUES (?, ?, ?)"
	
	// UPDATE queries
	updatePTStopQuery = "UPDATE %s SET lng = ?, lat = ?, name = ?, raw_xml = ? WHERE id = ?"
	
	// DELETE queries
	deletePTStopQuery      = "DELETE FROM %s WHERE id = ?"
	deletePTLineQuery      = "DELETE FROM %s WHERE id = ?"
	deletePTRouteQuery     = "DELETE FROM %s WHERE id = ?"
	deletePTRouteStopQuery = "DELETE FROM %s WHERE route_id = ? AND stop_order = ?"
	deletePTDepartureQuery = "DELETE FROM %s WHERE id = ?"
	deletePTDeparturesByRouteQuery = "DELETE FROM %s WHERE route_id = ?"
	
	// COUNT queries
	countPTStopsQuery      = "SELECT COUNT(*) FROM %s"
	countPTLinesByModeQuery = "SELECT mode, COUNT(*) FROM %s GROUP BY mode"
	countPTRoutesQuery     = "SELECT COUNT(*) FROM %s WHERE line_id = ?"
	countPTDeparturesQuery = "SELECT COUNT(*) FROM %s WHERE route_id = ?"
	
	// TELEMETRY queries
	selectPTTelemetryQuery = `SELECT process_id, total_file_size, bytes_read, 
	          COALESCE(stops_extracted, 0), COALESCE(lines_extracted, 0), 
	          COALESCE(routes_extracted, 0), error_count, last_updated 
	          FROM process_telemetry WHERE process_id = ?`
)

// PT Error Messages
const (
	// SELECT errors
	selectAllPTStopsError     = "failed to query PT stops from table %s"
	selectPTStopByIDError     = "failed to query PT stop by ID from table %s"
	selectAllPTLinesError     = "failed to query PT lines from table %s"
	selectPTLineByIDError     = "failed to query PT line by ID from table %s"
	selectPTLinesByModeError  = "failed to query PT lines by mode from table %s"
	selectPTRoutesByLineError = "failed to query PT routes from table %s"
	selectPTRouteStopsError   = "failed to query PT route stops from table %s"
	selectPTDeparturesError   = "failed to query PT departures from table %s"
	
	// INSERT errors
	insertPTStopError      = "failed to insert PT stop into table %s"
	insertPTLineError      = "failed to insert PT line into table %s"
	insertPTRouteError     = "failed to insert PT route into table %s"
	insertPTRouteStopError = "failed to insert PT route stop into table %s"
	insertPTDepartureError = "failed to insert PT departure into table %s"
	
	// UPDATE errors
	updatePTStopError = "failed to update PT stop in table %s"
	
	// DELETE errors
	deletePTStopError      = "failed to delete PT stop from table %s"
	deletePTLineError      = "failed to delete PT line from table %s"
	deletePTRouteError     = "failed to delete PT route from table %s"
	deletePTRouteStopError = "failed to delete PT route stop from table %s"
	deletePTDepartureError = "failed to delete PT departure from table %s"
	deletePTDeparturesByRouteError = "failed to delete PT departures from table %s"
	
	// COUNT errors
	countPTStopsError      = "failed to count PT stops in table %s"
	countPTLinesByModeError = "failed to count PT lines by mode in table %s"
	countPTRoutesError     = "failed to count PT routes in table %s"
	countPTDeparturesError = "failed to count PT departures in table %s"
	
	// TELEMETRY errors
	selectPTTelemetryError = "failed to get PT telemetry for process %d"
)

// GetPTStops retrieves all stops for a process
func (db *Database) GetPTStops(processID int) ([]PTStop, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	query := fmt.Sprintf(selectAllPTStopsQuery, stopsTable)
	rows, err := db.queryRows(query, fmt.Sprintf(selectAllPTStopsError, stopsTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stops []PTStop
	for rows.Next() {
		var stop PTStop
		if err := rows.Scan(&stop.ID, &stop.Lng, &stop.Lat, &stop.Name, &stop.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan PT stop: %w", err)
		}
		stops = append(stops, stop)
	}
	
	return stops, rows.Err()
}


// GetPTStop retrieves a specific stop by ID
func (db *Database) GetPTStop(processID int, stopID string) (*PTStop, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	query := fmt.Sprintf(selectPTStopByIDQuery, stopsTable)
	var stop PTStop
	err := db.queryRow(query, fmt.Sprintf(selectPTStopByIDError, stopsTable), stopID).Scan(
		&stop.ID, &stop.Lng, &stop.Lat, &stop.Name, &stop.RawXML)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("PT stop not found: %s", stopID)
		}
		return nil, err
	}
	
	return &stop, nil
}

// GetPTLines retrieves all lines for a process
func (db *Database) GetPTLines(processID int) ([]PTLine, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	
	query := fmt.Sprintf(selectAllPTLinesQuery, linesTable)
	rows, err := db.queryRows(query, fmt.Sprintf(selectAllPTLinesError, linesTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var lines []PTLine
	for rows.Next() {
		var line PTLine
		if err := rows.Scan(&line.ID, &line.Mode, &line.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan PT line: %w", err)
		}
		lines = append(lines, line)
	}
	
	return lines, rows.Err()
}

// GetPTLine retrieves a specific line by ID
func (db *Database) GetPTLine(processID int, lineID string) (*PTLine, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	
	query := fmt.Sprintf(selectPTLineByIDQuery, linesTable)
	var line PTLine
	err := db.queryRow(query, fmt.Sprintf(selectPTLineByIDError, linesTable), lineID).Scan(
		&line.ID, &line.Mode, &line.RawXML)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("PT line not found: %s", lineID)
		}
		return nil, err
	}
	
	return &line, nil
}

// GetPTLinesByMode retrieves lines by transport mode
func (db *Database) GetPTLinesByMode(processID int, mode string) ([]PTLine, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	
	query := fmt.Sprintf(selectPTLinesByModeQuery, linesTable)
	rows, err := db.queryRows(query, fmt.Sprintf(selectPTLinesByModeError, linesTable), mode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var lines []PTLine
	for rows.Next() {
		var line PTLine
		if err := rows.Scan(&line.ID, &line.Mode, &line.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan PT line: %w", err)
		}
		lines = append(lines, line)
	}
	
	return lines, rows.Err()
}

// GetPTRoutes retrieves routes for a specific line
func (db *Database) GetPTRoutes(processID int, lineID string) ([]PTRoute, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	routesTable := fmt.Sprintf("%s_routes", tablePrefix)
	
	query := fmt.Sprintf(selectPTRoutesByLineQuery, routesTable)
	rows, err := db.queryRows(query, fmt.Sprintf(selectPTRoutesByLineError, routesTable), lineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var routes []PTRoute
	for rows.Next() {
		var route PTRoute
		if err := rows.Scan(&route.ID, &route.LineID, &route.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan PT route: %w", err)
		}
		routes = append(routes, route)
	}
	
	return routes, rows.Err()
}

// GetPTRouteStops retrieves stops for a specific route
func (db *Database) GetPTRouteStops(processID int, routeID string) ([]PTRouteStop, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	routeStopsTable := fmt.Sprintf("%s_route_stops", tablePrefix)
	
	query := fmt.Sprintf(selectPTRouteStopsQuery, routeStopsTable)
	rows, err := db.queryRows(query, fmt.Sprintf(selectPTRouteStopsError, routeStopsTable), routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var routeStops []PTRouteStop
	for rows.Next() {
		var rs PTRouteStop
		if err := rows.Scan(&rs.RouteID, &rs.StopRefID, &rs.StopOrder, 
			&rs.ArrivalOffset, &rs.DepartureOffset); err != nil {
			return nil, fmt.Errorf("failed to scan PT route stop: %w", err)
		}
		routeStops = append(routeStops, rs)
	}
	
	return routeStops, rows.Err()
}

// GetPTDepartures retrieves departures for a specific route
func (db *Database) GetPTDepartures(processID int, routeID string) ([]PTDeparture, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
	
	query := fmt.Sprintf(selectPTDeparturesQuery, departuresTable)
	rows, err := db.queryRows(query, fmt.Sprintf(selectPTDeparturesError, departuresTable), routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var departures []PTDeparture
	for rows.Next() {
		var dep PTDeparture
		if err := rows.Scan(&dep.ID, &dep.RouteID, &dep.DepartureTime); err != nil {
			return nil, fmt.Errorf("failed to scan PT departure: %w", err)
		}
		departures = append(departures, dep)
	}
	
	return departures, rows.Err()
}

// AddPTStop adds a new stop
func (db *Database) AddPTStop(processID int, stop PTStop) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	query := fmt.Sprintf(insertPTStopQuery, stopsTable)
	_, err := db.execQuery(query, fmt.Sprintf(insertPTStopError, stopsTable),
		stop.ID, stop.Lng, stop.Lat, stop.Name, stop.RawXML)
	return err
}

// UpdatePTStop updates an existing stop
func (db *Database) UpdatePTStop(processID int, stopID string, update PTStopUpdate) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	// Build raw XML
	rawXML := fmt.Sprintf(`<stopFacility id="%s" x="%.6f" y="%.6f" name="%s"/>`,
		stopID, update.Lng, update.Lat, escapeXML(update.Name))
	
	query := fmt.Sprintf(updatePTStopQuery, stopsTable)
	result, err := db.execQuery(query, fmt.Sprintf(updatePTStopError, stopsTable),
		update.Lng, update.Lat, update.Name, rawXML, stopID)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("stop not found: %s", stopID)
	}
	
	return nil
}

// DeletePTStop deletes a stop
func (db *Database) DeletePTStop(processID int, stopID string) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	
	query := fmt.Sprintf(deletePTStopQuery, stopsTable)
	result, err := db.execQuery(query, fmt.Sprintf(deletePTStopError, stopsTable), stopID)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("stop not found: %s", stopID)
	}
	
	return nil
}

// BatchUpdatePTStops updates multiple stops in a transaction
func (db *Database) BatchUpdatePTStops(processID int, updates map[string]PTStopUpdate) error {
	return db.WithTransaction(func(tx *sql.Tx) error {
		tablePrefix := fmt.Sprintf("pt_data_%d", processID)
		stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
		
		updateQuery := fmt.Sprintf(updatePTStopQuery, stopsTable)
		stmt, err := tx.Prepare(updateQuery)
		if err != nil {
			return fmt.Errorf("failed to prepare update statement: %w", err)
		}
		defer stmt.Close()
		
		for stopID, update := range updates {
			rawXML := fmt.Sprintf(`<stopFacility id="%s" x="%.6f" y="%.6f" name="%s"/>`,
				stopID, update.Lng, update.Lat, escapeXML(update.Name))
			
			if _, err := stmt.Exec(update.Lng, update.Lat, update.Name, rawXML, stopID); err != nil {
				return fmt.Errorf("failed to update stop %s: %w", stopID, err)
			}
		}
		
		return nil
	})
}

// DeletePTLine deletes a line (cascades to routes, route stops, and departures)
func (db *Database) DeletePTLine(processID int, lineID string) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	
	query := fmt.Sprintf(deletePTLineQuery, linesTable)
	result, err := db.execQuery(query, fmt.Sprintf(deletePTLineError, linesTable), lineID)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("line not found: %s", lineID)
	}
	
	return nil
}

// DeletePTRoute deletes a route (cascades to route stops and departures)
func (db *Database) DeletePTRoute(processID int, routeID string) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	routesTable := fmt.Sprintf("%s_routes", tablePrefix)
	
	query := fmt.Sprintf(deletePTRouteQuery, routesTable)
	result, err := db.execQuery(query, fmt.Sprintf(deletePTRouteError, routesTable), routeID)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("route not found: %s", routeID)
	}
	
	return nil
}

// DeletePTRouteStop deletes a route stop
func (db *Database) DeletePTRouteStop(processID int, routeID string, stopOrder int) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	routeStopsTable := fmt.Sprintf("%s_route_stops", tablePrefix)
	
	query := fmt.Sprintf(deletePTRouteStopQuery, routeStopsTable)
	result, err := db.execQuery(query, fmt.Sprintf(deletePTRouteStopError, routeStopsTable), 
		routeID, stopOrder)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("route stop not found: route %s, order %d", routeID, stopOrder)
	}
	
	return nil
}

// DeletePTDeparture deletes a departure
func (db *Database) DeletePTDeparture(processID int, departureID string) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
	
	query := fmt.Sprintf(deletePTDepartureQuery, departuresTable)
	result, err := db.execQuery(query, fmt.Sprintf(deletePTDepartureError, departuresTable), departureID)
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("departure not found: %s", departureID)
	}
	
	return nil
}

// GetPTStatistics returns statistics for PT data
func (db *Database) GetPTStatistics(processID int) (map[string]interface{}, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	routesTable := fmt.Sprintf("%s_routes", tablePrefix)
	departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
	
	stats := make(map[string]interface{})
	
	// Count stops
	var stopCount int
	query := fmt.Sprintf(countPTStopsQuery, stopsTable)
	if err := db.queryRow(query, fmt.Sprintf(countPTStopsError, stopsTable)).Scan(&stopCount); err != nil {
		return nil, err
	}
	stats["stop_count"] = stopCount
	
	// Count lines by mode
	query = fmt.Sprintf(countPTLinesByModeQuery, linesTable)
	rows, err := db.queryRows(query, fmt.Sprintf(countPTLinesByModeError, linesTable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	linesByMode := make(map[string]int)
	for rows.Next() {
		var mode string
		var count int
		if err := rows.Scan(&mode, &count); err != nil {
			return nil, fmt.Errorf("failed to scan line count: %w", err)
		}
		linesByMode[mode] = count
	}
	stats["lines_by_mode"] = linesByMode
	
	// Get total route count by counting all routes
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s", routesTable)
	var routeCount int
	if err := db.queryRow(query, fmt.Sprintf("failed to count all routes in table %s", routesTable)).Scan(&routeCount); err != nil {
		return nil, err
	}
	stats["route_count"] = routeCount
	
	// Get total departure count
	query = fmt.Sprintf("SELECT COUNT(*) FROM %s", departuresTable)
	var departureCount int
	if err := db.queryRow(query, fmt.Sprintf("failed to count all departures in table %s", departuresTable)).Scan(&departureCount); err != nil {
		return nil, err
	}
	stats["departure_count"] = departureCount
	
	return stats, nil
}

// GetPTLineSummaries retrieves line summaries with route and departure counts for a specific mode
func (db *Database) GetPTLineSummaries(processID int, mode string) ([]PTLineSummary, error) {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	linesTable := fmt.Sprintf("%s_lines", tablePrefix)
	routesTable := fmt.Sprintf("%s_routes", tablePrefix)
	departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
	
	// Query to get line summaries with counts, filtered by mode
	query := fmt.Sprintf(`
		SELECT 
			l.id,
			COALESCE(SUBSTR(l.id, INSTR(l.id, '_') + 1), l.id) as name,
			l.mode,
			COUNT(DISTINCT r.id) as route_count,
			COUNT(DISTINCT d.id) as departure_count
		FROM %s l
		LEFT JOIN %s r ON r.line_id = l.id
		LEFT JOIN %s d ON d.route_id = r.id
		WHERE l.mode = ?
		GROUP BY l.id, l.mode
		ORDER BY l.id
	`, linesTable, routesTable, departuresTable)
	
	rows, err := db.queryRows(query, "failed to query PT line summaries", mode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var summaries []PTLineSummary
	for rows.Next() {
		var summary PTLineSummary
		if err := rows.Scan(&summary.ID, &summary.Name, &summary.Mode, 
			&summary.RouteCount, &summary.DepartureCount); err != nil {
			return nil, fmt.Errorf("failed to scan PT line summary: %w", err)
		}
		summaries = append(summaries, summary)
	}
	
	return summaries, rows.Err()
}

// InsertPTStopBatch inserts multiple stops in a transaction
func (db *Database) InsertPTStopBatch(processID int, stops []PTStopData) error {
	if len(stops) == 0 {
		return nil
	}
	
	return db.WithTransaction(func(tx *sql.Tx) error {
		tablePrefix := fmt.Sprintf("pt_data_%d", processID)
		stopsTable := fmt.Sprintf("%s_stops", tablePrefix)
		
		query := fmt.Sprintf(insertPTStopQuery, stopsTable)
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()
		
		for _, stop := range stops {
			if _, err := stmt.Exec(stop.ID, stop.Lng, stop.Lat, stop.Name, stop.RawXML); err != nil {
				return fmt.Errorf("failed to insert stop %s: %w", stop.ID, err)
			}
		}
		
		return nil
	})
}

// InsertPTLineBatch inserts multiple lines in a transaction
func (db *Database) InsertPTLineBatch(processID int, lines []PTLineData) error {
	if len(lines) == 0 {
		return nil
	}
	
	return db.WithTransaction(func(tx *sql.Tx) error {
		tablePrefix := fmt.Sprintf("pt_data_%d", processID)
		linesTable := fmt.Sprintf("%s_lines", tablePrefix)
		
		query := fmt.Sprintf(insertPTLineQuery, linesTable)
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()
		
		for _, line := range lines {
			if _, err := stmt.Exec(line.ID, line.Mode, line.RawXML); err != nil {
				return fmt.Errorf("failed to insert line %s: %w", line.ID, err)
			}
		}
		
		return nil
	})
}

// InsertPTRouteBatch inserts multiple routes in a transaction
func (db *Database) InsertPTRouteBatch(processID int, routes []PTRouteData) error {
	if len(routes) == 0 {
		return nil
	}
	
	return db.WithTransaction(func(tx *sql.Tx) error {
		tablePrefix := fmt.Sprintf("pt_data_%d", processID)
		routesTable := fmt.Sprintf("%s_routes", tablePrefix)
		
		query := fmt.Sprintf(insertPTRouteQuery, routesTable)
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()
		
		for _, route := range routes {
			if _, err := stmt.Exec(route.ID, route.LineID, route.RawXML); err != nil {
				return fmt.Errorf("failed to insert route %s: %w", route.ID, err)
			}
		}
		
		return nil
	})
}

// InsertPTRouteStopBatch inserts multiple route stops in a transaction
func (db *Database) InsertPTRouteStopBatch(processID int, routeStops []PTRouteStopData) error {
	if len(routeStops) == 0 {
		return nil
	}
	
	return db.WithTransaction(func(tx *sql.Tx) error {
		tablePrefix := fmt.Sprintf("pt_data_%d", processID)
		routeStopsTable := fmt.Sprintf("%s_route_stops", tablePrefix)
		
		query := fmt.Sprintf(insertPTRouteStopQuery, routeStopsTable)
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()
		
		for _, rs := range routeStops {
			if _, err := stmt.Exec(rs.RouteID, rs.StopRefID, rs.StopOrder, 
				rs.ArrivalOffset, rs.DepartureOffset); err != nil {
				return fmt.Errorf("failed to insert route stop: %w", err)
			}
		}
		
		return nil
	})
}

// InsertPTDepartureBatch inserts multiple departures in a transaction
func (db *Database) InsertPTDepartureBatch(processID int, departures []PTDepartureData) error {
	if len(departures) == 0 {
		return nil
	}
	
	return db.WithTransaction(func(tx *sql.Tx) error {
		tablePrefix := fmt.Sprintf("pt_data_%d", processID)
		departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
		
		query := fmt.Sprintf(insertPTDepartureQuery, departuresTable)
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %w", err)
		}
		defer stmt.Close()
		
		for _, dep := range departures {
			if _, err := stmt.Exec(dep.ID, dep.RouteID, dep.DepartureTime); err != nil {
				return fmt.Errorf("failed to insert departure %s: %w", dep.ID, err)
			}
		}
		
		return nil
	})
}

// DeletePTDeparturesByRoute deletes all departures for a route
func (db *Database) DeletePTDeparturesByRoute(processID int, routeID string) error {
	tablePrefix := fmt.Sprintf("pt_data_%d", processID)
	departuresTable := fmt.Sprintf("%s_departures", tablePrefix)
	
	query := fmt.Sprintf(deletePTDeparturesByRouteQuery, departuresTable)
	_, err := db.execQuery(query, fmt.Sprintf(deletePTDeparturesByRouteError, departuresTable), routeID)
	return err
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

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(s)
}

