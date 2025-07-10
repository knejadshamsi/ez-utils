package gui

// PT App Layer - Simple delegation to Database methods
// All SQL operations are handled in pt_database.go

// GetPTStops retrieves all stops for a process
func (a *App) GetPTStops(processID int) ([]PTStop, error) {
	return a.db.GetPTStops(processID)
}

// GetPTStopsByBbox retrieves stops within a bounding box
func (a *App) GetPTStopsByBbox(processID int, bbox BoundingBox) ([]PTStop, error) {
	return a.db.GetPTStopsByBbox(processID, bbox)
}

// GetPTStop retrieves a specific stop by ID
func (a *App) GetPTStop(processID int, stopID string) (*PTStop, error) {
	return a.db.GetPTStop(processID, stopID)
}

// GetPTLines retrieves all lines for a process
func (a *App) GetPTLines(processID int) ([]PTLine, error) {
	return a.db.GetPTLines(processID)
}

// GetPTLine retrieves a specific line by ID
func (a *App) GetPTLine(processID int, lineID string) (*PTLine, error) {
	return a.db.GetPTLine(processID, lineID)
}

// GetPTLinesByMode retrieves lines by transport mode
func (a *App) GetPTLinesByMode(processID int, mode string) ([]PTLine, error) {
	return a.db.GetPTLinesByMode(processID, mode)
}

// GetPTRoutes retrieves routes for a specific line
func (a *App) GetPTRoutes(processID int, lineID string) ([]PTRoute, error) {
	return a.db.GetPTRoutes(processID, lineID)
}

// GetPTRouteStops retrieves stops for a specific route
func (a *App) GetPTRouteStops(processID int, routeID string) ([]PTRouteStop, error) {
	return a.db.GetPTRouteStops(processID, routeID)
}

// GetPTDepartures retrieves departures for a specific route
func (a *App) GetPTDepartures(processID int, routeID string) ([]PTDeparture, error) {
	return a.db.GetPTDepartures(processID, routeID)
}

// GetPTStatistics returns statistics for PT data
func (a *App) GetPTStatistics(processID int) (map[string]interface{}, error) {
	return a.db.GetPTStatistics(processID)
}