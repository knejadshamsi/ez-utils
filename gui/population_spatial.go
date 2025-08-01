package gui

import (
	"fmt"
	"math"
)

// GetPopulationByZonesDynamic retrieves persons within selected zones using dynamic point-in-polygon
func (db *Database) GetPopulationByZonesDynamic(tableName string, zoneIDs []string, page, pageSize int) (*PaginatedResponse, error) {
	if len(zoneIDs) == 0 {
		// No zones selected, return empty result
		return &PaginatedResponse{
			Persons:     []Person{},
			TotalCount:  0,
			CurrentPage: page,
			TotalPages:  0,
			PageSize:    pageSize,
		}, nil
	}
	
	// Get the zones by their IDs
	zones := make([]*Zone, 0, len(zoneIDs))
	for _, zoneID := range zoneIDs {
		zone, err := db.GetZone(zoneID)
		if err != nil {
			continue // Skip zones that can't be loaded
		}
		zones = append(zones, zone)
	}
	
	if len(zones) == 0 {
		// No valid zones found
		return &PaginatedResponse{
			Persons:     []Person{},
			TotalCount:  0,
			CurrentPage: page,
			TotalPages:  0,
			PageSize:    pageSize,
		}, nil
	}
	
	// Calculate combined bounding box of all zones
	minX, minY := zones[0].BoundingBox.MinX, zones[0].BoundingBox.MinY
	maxX, maxY := zones[0].BoundingBox.MaxX, zones[0].BoundingBox.MaxY
	
	for _, zone := range zones[1:] {
		minX = math.Min(minX, zone.BoundingBox.MinX)
		minY = math.Min(minY, zone.BoundingBox.MinY)
		maxX = math.Max(maxX, zone.BoundingBox.MaxX)
		maxY = math.Max(maxY, zone.BoundingBox.MaxY)
	}
	
	// Query persons within the bounding box
	bboxQuery := fmt.Sprintf(`
		SELECT id, lng, lat, raw_xml 
		FROM %s 
		WHERE lng IS NOT NULL AND lat IS NOT NULL
		ORDER BY id
	`, tableName)
	
	rows, err := db.queryRows(bboxQuery, "failed to query persons by bbox")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	// Filter persons using point-in-polygon
	var filteredPersons []Person
	for rows.Next() {
		var p Person
		var lng, lat float64
		if err := rows.Scan(&p.ID, &lng, &lat, &p.RawXML); err != nil {
			continue
		}
		
		x, y := lng, lat
		
		// Quick bounding box check
		if x < minX || x > maxX || y < minY || y > maxY {
			continue
		}
		
		// Check if point is in any of the selected zones
		for _, zone := range zones {
			if zone.ContainsPoint(x, y) {
				filteredPersons = append(filteredPersons, p)
				break // Point found in at least one zone
			}
		}
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	// Apply pagination to filtered results
	totalCount := len(filteredPersons)
	totalPages := (totalCount + pageSize - 1) / pageSize
	
	// Calculate pagination bounds
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}
	
	var pagePersons []Person
	if start < totalCount {
		pagePersons = filteredPersons[start:end]
	}
	
	return &PaginatedResponse{
		Persons:     pagePersons,
		TotalCount:  totalCount,
		CurrentPage: page,
		TotalPages:  totalPages,
		PageSize:    pageSize,
	}, nil
}

// GetPopulationByBoundingBox retrieves persons within a bounding box with pagination
func (db *Database) GetPopulationByBoundingBox(tableName string, minX, minY, maxX, maxY float64, page, pageSize int) (*PaginatedResponse, error) {
	// For SQLite without spatial extensions, we need to parse coordinates in application
	query := fmt.Sprintf(`
		SELECT id, lng, lat, raw_xml 
		FROM %s 
		WHERE lng IS NOT NULL AND lat IS NOT NULL
		ORDER BY id
	`, tableName)
	
	rows, err := db.queryRows(query, "failed to query all persons")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	// Filter by bounding box in application
	var filteredPersons []Person
	for rows.Next() {
		var p Person
		var lng, lat float64
		if err := rows.Scan(&p.ID, &lng, &lat, &p.RawXML); err != nil {
			continue
		}
		
		x, y := lng, lat
		
		// Check if within bounding box
		if x >= minX && x <= maxX && y >= minY && y <= maxY {
			filteredPersons = append(filteredPersons, p)
		}
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	// Apply pagination
	totalCount := len(filteredPersons)
	totalPages := (totalCount + pageSize - 1) / pageSize
	
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}
	
	var pagePersons []Person
	if start < totalCount {
		pagePersons = filteredPersons[start:end]
	}
	
	return &PaginatedResponse{
		Persons:     pagePersons,
		TotalCount:  totalCount,
		CurrentPage: page,
		TotalPages:  totalPages,
		PageSize:    pageSize,
	}, nil
}