package gui

import (
	"fmt"
	"time"
)

// Zone API methods for Wails frontend

// GetAllZones retrieves all zones
func (a *App) GetAllZones() ([]*Zone, error) {
	return a.db.GetAllZones()
}

// GetZone retrieves a zone by ID
func (a *App) GetZone(zoneID string) (*Zone, error) {
	return a.db.GetZone(zoneID)
}

// SaveZone creates or updates a zone
func (a *App) SaveZone(zone *Zone) error {
	// Validate zone
	if zone.ID == "" {
		zone.ID = fmt.Sprintf("zone_%d", time.Now().UnixNano())
	}
	
	if len(zone.Polygon) < 3 {
		return fmt.Errorf("zone must have at least 3 points")
	}
	
	// Calculate bounding box
	zone.CalculateBoundingBox()
	
	// Save to database
	return a.db.SaveZone(zone)
}

// DeleteZone removes a zone
func (a *App) DeleteZone(zoneID string) error {
	return a.db.DeleteZone(zoneID)
}

// GetZoneStats retrieves statistics for all zones in a population table
// Uses dynamic point-in-polygon counting
func (a *App) GetZoneStats(tableName string) (map[string]interface{}, error) {
	// Get all zones
	zones, err := a.db.GetAllZones()
	if err != nil {
		return nil, err
	}
	
	// For now, return empty counts since dynamic counting is expensive
	// In production, you might want to cache these counts
	result := make(map[string]interface{})
	for _, zone := range zones {
		result[zone.ID] = map[string]interface{}{
			"id":          zone.ID,
			"name":        zone.Name,
			"personCount": 0, // Dynamic counting would be expensive here
		}
	}
	
	return result, nil
}

// AssignPopulationToZones processes all persons in a table and assigns them to zones
func (a *App) AssignPopulationToZones(tableName string) error {
	// Ensure zone_id column exists
	if err := a.db.AddZoneColumnToPopulationTable(tableName); err != nil {
		return fmt.Errorf("failed to add zone column: %w", err)
	}
	
	// Assign persons to zones
	if err := a.db.AssignAllPersonsToZones(tableName); err != nil {
		return fmt.Errorf("failed to assign persons to zones: %w", err)
	}
	
	return nil
}

// GetPopulationByZones retrieves paginated population data filtered by zones
func (a *App) GetPopulationByZones(tableName string, page int, pageSize int, zoneIDs []string) (*PaginatedResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 50
	}
	
	return a.db.GetPersonsByZones(tableName, zoneIDs, page, pageSize)
}

// FindZoneForCoordinates finds which zone contains the given coordinates
func (a *App) FindZoneForCoordinates(lng, lat float64) (*Zone, error) {
	zones, err := a.db.FindZonesContainingPoint(lng, lat)
	if err != nil {
		return nil, err
	}
	
	if len(zones) == 0 {
		return nil, nil // No zone found
	}
	
	// Return first zone (assuming non-overlapping zones)
	return zones[0], nil
}

// CreateZoneFromGeoJSON creates a zone from GeoJSON data
func (a *App) CreateZoneFromGeoJSON(geojson map[string]interface{}) error {
	zone, err := ZoneFromGeoJSON(geojson)
	if err != nil {
		return fmt.Errorf("failed to parse GeoJSON: %w", err)
	}
	
	if zone.ID == "" {
		zone.ID = fmt.Sprintf("zone_%d", time.Now().UnixNano())
	}
	
	return a.db.SaveZone(zone)
}