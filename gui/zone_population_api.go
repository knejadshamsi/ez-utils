package gui

import (
	"fmt"
	"math"
)

// GetPersonsInPolygon returns persons within the given polygon coordinates
func (a *App) GetPersonsInPolygon(tableName string, polygon []Point, page int, pageSize int) (*PaginatedResponse, error) {
	if len(polygon) < 3 {
		return nil, fmt.Errorf("polygon must have at least 3 points")
	}
	
	// Calculate bounding box
	minX, maxX := polygon[0].X, polygon[0].X
	minY, maxY := polygon[0].Y, polygon[0].Y
	for _, p := range polygon[1:] {
		if p.X < minX { minX = p.X }
		if p.X > maxX { maxX = p.X }
		if p.Y < minY { minY = p.Y }
		if p.Y > maxY { maxY = p.Y }
	}
	
	// Query persons within bounding box
	query := fmt.Sprintf(`
		SELECT id, lng, lat, raw_xml 
		FROM %s 
		WHERE lng BETWEEN ? AND ?
		AND lat BETWEEN ? AND ?
	`, tableName)
	
	rows, err := a.db.queryRows(query, "failed to query persons", minX, maxX, minY, maxY)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	// Filter by polygon
	var persons []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.Lng, &p.Lat, &p.RawXML); err != nil {
			continue
		}
		
		// Check if point is in polygon
		if pointInPolygon(p.Lng, p.Lat, polygon) {
			persons = append(persons, p)
		}
	}
	
	// Paginate results
	totalCount := len(persons)
	totalPages := (totalCount + pageSize - 1) / pageSize
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}
	
	var paginatedPersons []Person
	if start < totalCount {
		paginatedPersons = persons[start:end]
	}
	
	return &PaginatedResponse{
		Persons:     paginatedPersons,
		TotalCount:  totalCount,
		CurrentPage: page,
		TotalPages:  totalPages,
		PageSize:    pageSize,
	}, nil
}

// pointInPolygon checks if a point is inside a polygon using ray casting
func pointInPolygon(x, y float64, polygon []Point) bool {
	inside := false
	n := len(polygon)
	p1 := polygon[0]
	
	for i := 1; i <= n; i++ {
		p2 := polygon[i%n]
		if y > math.Min(p1.Y, p2.Y) && y <= math.Max(p1.Y, p2.Y) {
			if x <= math.Max(p1.X, p2.X) {
				if p1.Y != p2.Y {
					xinters := (y-p1.Y)*(p2.X-p1.X)/(p2.Y-p1.Y) + p1.X
					if p1.X == p2.X || x <= xinters {
						inside = !inside
					}
				}
			}
		}
		p1 = p2
	}
	return inside
}