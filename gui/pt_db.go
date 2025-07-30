package gui

import (
	"fmt"
	"gorm.io/gorm"
)

// getGormDB returns a GORM database connection
func (a *App) getGormDB() (*gorm.DB, error) {
	// Return the cached GORM instance
	if a.gormDB == nil {
		return nil, fmt.Errorf("GORM not initialized")
	}
	return a.gormDB, nil
}

func (a *App) GetPTLinesByMode(processID int, mode TransportMode) ([]Line, error) {
	db, err := a.getGormDB()
	if err != nil {
		return nil, err
	}
	
	var lines []Line
	tableName := fmt.Sprintf("pt_%d_lines", processID)
	err = db.Table(tableName).Where("type = ?", mode).Find(&lines).Error
	return lines, err
}


func (a *App) GetPTRoutesByLineID(processID int, lineID string) ([]Route, error) {
	db, err := a.getGormDB()
	if err != nil {
		return nil, err
	}
	
	var routes []Route
	tableName := fmt.Sprintf("pt_%d_routes", processID)
	err = db.Table(tableName).Where("line_id = ?", lineID).Find(&routes).Error
	return routes, err
}

func (a *App) GetPTStopsByRouteID(processID int, routeID string) ([]RouteStop, error) {
	db, err := a.getGormDB()
	if err != nil {
		return nil, err
	}
	
	var routeStops []RouteStop
	tableName := fmt.Sprintf("pt_%d_route_stops", processID)
	
	// Get route-stop junctions for this route
	err = db.Table(tableName).Where("route_id = ?", routeID).Order("sequence").Find(&routeStops).Error
	return routeStops, err
}

func (a *App) GetPTStops(processID int, stopIDs []string) ([]Stop, error) {
	db, err := a.getGormDB()
	if err != nil {
		return nil, err
	}
	
	var stops []Stop
	tableName := fmt.Sprintf("pt_%d_stops", processID)
	
	// Get stops by IDs
	err = db.Table(tableName).Where("stop_id IN ?", stopIDs).Find(&stops).Error
	return stops, err
}

func (a *App) GetPTDeparturesByRouteID(processID int, routeID string) ([]Departure, error) {
	db, err := a.getGormDB()
	if err != nil {
		return nil, err
	}
	
	var departures []Departure
	tableName := fmt.Sprintf("pt_%d_departures", processID)
	err = db.Table(tableName).Where("route_id = ?", routeID).Find(&departures).Error
	return departures, err
}


func (a *App) SavePTLines(processID int, lines []Line) error {
	db, err := a.getGormDB()
	if err != nil {
		return err
	}
	
	tableName := fmt.Sprintf("pt_%d_lines", processID)
	return db.Table(tableName).Save(&lines).Error
}

func (a *App) SavePTRoutes(processID int, routes []Route) error {
	db, err := a.getGormDB()
	if err != nil {
		return err
	}
	
	tableName := fmt.Sprintf("pt_%d_routes", processID)
	return db.Table(tableName).Save(&routes).Error
}

func (a *App) SavePTStops(processID int, stops []Stop) error {
	db, err := a.getGormDB()
	if err != nil {
		return err
	}
	
	tableName := fmt.Sprintf("pt_%d_stops", processID)
	return db.Table(tableName).Save(&stops).Error
}

func (a *App) SavePTRouteStops(processID int, routeStops []RouteStop) error {
	db, err := a.getGormDB()
	if err != nil {
		return err
	}
	
	tableName := fmt.Sprintf("pt_%d_route_stops", processID)
	return db.Table(tableName).Save(&routeStops).Error
}

func (a *App) SavePTDepartures(processID int, departures []Departure) error {
	db, err := a.getGormDB()
	if err != nil {
		return err
	}
	
	tableName := fmt.Sprintf("pt_%d_departures", processID)
	return db.Table(tableName).Save(&departures).Error
}

func (a *App) DeletePTLines(processID int, lineIDs []string) error {
	db, err := a.getGormDB()
	if err != nil {
		return err
	}
	
	tableName := fmt.Sprintf("pt_%d_lines", processID)
	return db.Table(tableName).Where("id IN ?", lineIDs).Delete(&Line{}).Error
}

func (a *App) DeletePTRoutes(processID int, routeIDs []string) error {
	db, err := a.getGormDB()
	if err != nil {
		return err
	}
	
	tableName := fmt.Sprintf("pt_%d_routes", processID)
	return db.Table(tableName).Where("id IN ?", routeIDs).Delete(&Route{}).Error
}

func (a *App) DeletePTStops(processID int, stopIDs []string) error {
	db, err := a.getGormDB()
	if err != nil {
		return err
	}
	
	tableName := fmt.Sprintf("pt_%d_stops", processID)
	return db.Table(tableName).Where("stop_id IN ?", stopIDs).Delete(&Stop{}).Error
}

func (a *App) DeletePTDepartures(processID int, departureIDs []string) error {
	db, err := a.getGormDB()
	if err != nil {
		return err
	}
	
	tableName := fmt.Sprintf("pt_%d_departures", processID)
	return db.Table(tableName).Where("id IN ?", departureIDs).Delete(&Departure{}).Error
}

func (a *App) GetPTStopsInBounds(processID int, mode string, bounds ViewportBounds) ([]Stop, error) {
	db, err := a.getGormDB()
	if err != nil {
		return nil, err
	}
	
	stopsTable := fmt.Sprintf("pt_%d_stops", processID)
	routeStopsTable := fmt.Sprintf("pt_%d_route_stops", processID)
	routesTable := fmt.Sprintf("pt_%d_routes", processID)
	linesTable := fmt.Sprintf("pt_%d_lines", processID)
	
	var stops []Stop
	
	// Get stops that belong to lines of the specified transport mode within bounds
	query := fmt.Sprintf(`
		SELECT DISTINCT s.*
		FROM %s s
		WHERE s.lat BETWEEN ? AND ?
		AND s.lng BETWEEN ? AND ?
		AND EXISTS (
			SELECT 1 FROM %s rs
			JOIN %s r ON rs.route_id = r.id
			JOIN %s l ON r.line_id = l.id
			WHERE rs.stop_id = s.stop_id
			AND l.type = ?
		)
	`, stopsTable, routeStopsTable, routesTable, linesTable)
	
	err = db.Raw(query, bounds.MinLat, bounds.MaxLat, bounds.MinLng, bounds.MaxLng, mode).Scan(&stops).Error
	return stops, err
}

func (a *App) RemovePTStopsFromRoute(processID int, routeID string, stopIDs []string) error {
	db, err := a.getGormDB()
	if err != nil {
		return err
	}
	
	return db.Transaction(func(tx *gorm.DB) error {
		tableName := fmt.Sprintf("pt_%d_route_stops", processID)
		
		// Delete the route-stop junctions
		if err := tx.Table(tableName).Where("route_id = ? AND stop_id IN ?", routeID, stopIDs).Delete(&RouteStop{}).Error; err != nil {
			return err
		}
		
		// Get remaining route-stops ordered by sequence
		var remainingRouteStops []RouteStop
		if err := tx.Table(tableName).Where("route_id = ?", routeID).Order("sequence").Find(&remainingRouteStops).Error; err != nil {
			return err
		}
		
		// Update sequences to remove gaps
		for i, rs := range remainingRouteStops {
			if rs.Sequence != i {
				if err := tx.Table(tableName).Where("route_id = ? AND stop_id = ?", routeID, rs.StopID).Update("sequence", i).Error; err != nil {
					return err
				}
			}
		}
		
		return nil
	})
}