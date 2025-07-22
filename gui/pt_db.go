package gui

// PT App Layer - Simple delegation to Database methods
// All SQL operations are handled in pt_database.go

// GetPTStops retrieves all stops for a process
func (a *App) GetPTStops(processID int) ([]PTStop, error) {
	stops, err := a.db.GetPTStops(processID)
	if err != nil {
		return make([]PTStop, 0), err
	}
	if stops == nil {
		return make([]PTStop, 0), nil
	}
	return stops, nil
}


// GetPTStop retrieves a specific stop by ID
func (a *App) GetPTStop(processID int, stopID string) (*PTStop, error) {
	return a.db.GetPTStop(processID, stopID)
}

// GetPTLines retrieves all lines for a process
func (a *App) GetPTLines(processID int) ([]PTLine, error) {
	lines, err := a.db.GetPTLines(processID)
	if err != nil {
		return make([]PTLine, 0), err
	}
	if lines == nil {
		return make([]PTLine, 0), nil
	}
	return lines, nil
}

// GetPTLine retrieves a specific line by ID
func (a *App) GetPTLine(processID int, lineID string) (*PTLine, error) {
	return a.db.GetPTLine(processID, lineID)
}

// GetPTLinesByMode retrieves lines by transport mode
func (a *App) GetPTLinesByMode(processID int, mode string) ([]PTLine, error) {
	lines, err := a.db.GetPTLinesByMode(processID, mode)
	if err != nil {
		return make([]PTLine, 0), err
	}
	if lines == nil {
		return make([]PTLine, 0), nil
	}
	return lines, nil
}

// GetPTRoutes retrieves routes for a specific line
func (a *App) GetPTRoutes(processID int, lineID string) ([]PTRoute, error) {
	routes, err := a.db.GetPTRoutes(processID, lineID)
	if err != nil {
		return make([]PTRoute, 0), err
	}
	if routes == nil {
		return make([]PTRoute, 0), nil
	}
	return routes, nil
}

// GetPTRouteStops retrieves stops for a specific route
func (a *App) GetPTRouteStops(processID int, routeID string) ([]PTRouteStop, error) {
	stops, err := a.db.GetPTRouteStops(processID, routeID)
	if err != nil {
		return make([]PTRouteStop, 0), err
	}
	if stops == nil {
		return make([]PTRouteStop, 0), nil
	}
	return stops, nil
}

// GetPTDepartures retrieves departures for a specific route
func (a *App) GetPTDepartures(processID int, routeID string) ([]PTDeparture, error) {
	departures, err := a.db.GetPTDepartures(processID, routeID)
	if err != nil {
		return make([]PTDeparture, 0), err
	}
	if departures == nil {
		return make([]PTDeparture, 0), nil
	}
	return departures, nil
}

// GetPTStatistics returns statistics for PT data
func (a *App) GetPTStatistics(processID int) (map[string]interface{}, error) {
	return a.db.GetPTStatistics(processID)
}

// GetPTLineSummaries retrieves line summaries with route and departure counts for a specific mode
func (a *App) GetPTLineSummaries(processID int, mode string) ([]PTLineSummary, error) {
	summaries, err := a.db.GetPTLineSummaries(processID, mode)
	if err != nil {
		return make([]PTLineSummary, 0), err
	}
	if summaries == nil {
		return make([]PTLineSummary, 0), nil
	}
	return summaries, nil
}