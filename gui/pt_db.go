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

// AddPTStop adds a new stop
func (a *App) AddPTStop(processID int, stop PTStop) error {
	return a.db.AddPTStop(processID, stop)
}

// UpdatePTStop updates an existing stop
func (a *App) UpdatePTStop(processID int, stopID string, update PTStopUpdate) error {
	return a.db.UpdatePTStop(processID, stopID, update)
}

// BatchUpdatePTStops updates multiple stops in a transaction
func (a *App) BatchUpdatePTStops(processID int, updates map[string]PTStopUpdate) error {
	return a.db.BatchUpdatePTStops(processID, updates)
}

// DeletePTStop deletes a stop
func (a *App) DeletePTStop(processID int, stopID string) error {
	return a.db.DeletePTStop(processID, stopID)
}

// DeletePTLine deletes a line (cascades to routes, route stops, and departures)
func (a *App) DeletePTLine(processID int, lineID string) error {
	return a.db.DeletePTLine(processID, lineID)
}

// DeletePTRoute deletes a route (cascades to route stops and departures)
func (a *App) DeletePTRoute(processID int, routeID string) error {
	return a.db.DeletePTRoute(processID, routeID)
}

// DeletePTRouteStop deletes a route stop
func (a *App) DeletePTRouteStop(processID int, routeID string, stopOrder int) error {
	return a.db.DeletePTRouteStop(processID, routeID, stopOrder)
}

// DeletePTDeparture deletes a departure
func (a *App) DeletePTDeparture(processID int, departureID string) error {
	return a.db.DeletePTDeparture(processID, departureID)
}

// GetPTStatistics returns statistics for PT data
func (a *App) GetPTStatistics(processID int) (map[string]interface{}, error) {
	return a.db.GetPTStatistics(processID)
}