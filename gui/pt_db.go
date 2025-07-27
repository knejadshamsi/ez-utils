package gui

// PT App Layer - Simple delegation to Database methods
// All SQL operations are handled in pt_database.go

// GetPTStops retrieves all stops (route-stop relationships) for a process
func (a *App) GetPTStops(processID int) ([]Stop, error) {
	stops, err := a.db.GetPTStops(processID)
	if err != nil {
		return make([]Stop, 0), err
	}
	if stops == nil {
		return make([]Stop, 0), nil
	}
	return stops, nil
}



// GetPTLines retrieves all lines for a process
func (a *App) GetPTLines(processID int) ([]Line, error) {
	lines, err := a.db.GetPTLines(processID)
	if err != nil {
		return make([]Line, 0), err
	}
	if lines == nil {
		return make([]Line, 0), nil
	}
	return lines, nil
}

// GetPTLine retrieves a specific line by ID
func (a *App) GetPTLine(processID int, lineID string) (*Line, error) {
	return a.db.GetPTLine(processID, lineID)
}

// GetPTLinesByMode retrieves lines by transport mode
func (a *App) GetPTLinesByMode(processID int, mode string) ([]Line, error) {
	lines, err := a.db.GetPTLinesByMode(processID, mode)
	if err != nil {
		return make([]Line, 0), err
	}
	if lines == nil {
		return make([]Line, 0), nil
	}
	return lines, nil
}

// GetPTStopsForRoute retrieves stops for a specific route
func (a *App) GetPTStopsForRoute(processID int, routeID string) ([]Stop, error) {
	stops, err := a.db.GetPTStopsForRoute(processID, routeID)
	if err != nil {
		return make([]Stop, 0), err
	}
	if stops == nil {
		return make([]Stop, 0), nil
	}
	return stops, nil
}

// GetPTDepartures retrieves departures for a specific route
func (a *App) GetPTDepartures(processID int, routeID string) ([]Departure, error) {
	departures, err := a.db.GetPTDepartures(processID, routeID)
	if err != nil {
		return make([]Departure, 0), err
	}
	if departures == nil {
		return make([]Departure, 0), nil
	}
	return departures, nil
}

// GetPTStatistics returns statistics for PT data
func (a *App) GetPTStatistics(processID int) (map[string]interface{}, error) {
	return a.db.GetPTStatistics(processID)
}

// GetPTRoutes retrieves routes for a specific line
func (a *App) GetPTRoutes(processID int, lineID string) ([]Route, error) {
	line, err := a.db.GetPTLine(processID, lineID)
	if err != nil {
		return make([]Route, 0), err
	}
	if line == nil {
		return make([]Route, 0), nil
	}
	return line.Routes, nil
}

// GetPTRouteStops is an alias for GetPTStopsForRoute for backward compatibility
func (a *App) GetPTRouteStops(processID int, routeID string) ([]Stop, error) {
	return a.GetPTStopsForRoute(processID, routeID)
}

