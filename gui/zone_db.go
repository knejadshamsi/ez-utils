package gui

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// Zone table queries
const (
	createZoneTableQuery = `
		CREATE TABLE IF NOT EXISTS zones (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			polygon_json TEXT NOT NULL,
			min_x REAL NOT NULL,
			min_y REAL NOT NULL,
			max_x REAL NOT NULL,
			max_y REAL NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	
	createZoneIndexQuery = `
		CREATE INDEX IF NOT EXISTS idx_zones_bbox ON zones(min_x, min_y, max_x, max_y)
	`
	
	insertZoneQuery = `
		INSERT INTO zones (id, name, polygon_json, min_x, min_y, max_x, max_y)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	updateZoneQuery = `
		UPDATE zones 
		SET name = ?, polygon_json = ?, min_x = ?, min_y = ?, max_x = ?, max_y = ?, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	
	deleteZoneQuery = `DELETE FROM zones WHERE id = ?`
	
	selectAllZonesQuery = `
		SELECT id, name, polygon_json, min_x, min_y, max_x, max_y 
		FROM zones 
		ORDER BY name
	`
	
	selectZoneByIDQuery = `
		SELECT id, name, polygon_json, min_x, min_y, max_x, max_y 
		FROM zones 
		WHERE id = ?
	`
	
	selectZonesContainingPointQuery = `
		SELECT id, name, polygon_json, min_x, min_y, max_x, max_y 
		FROM zones 
		WHERE min_x <= ? AND max_x >= ? AND min_y <= ? AND max_y >= ?
	`
	
	// Add zone_id column to population tables
	addZoneColumnQuery = `ALTER TABLE %s ADD COLUMN zone_id TEXT`
	
	// Update person's zone
	updatePersonZoneQuery = `UPDATE %s SET zone_id = ? WHERE id = ?`
	
	// Select persons by zone
	selectPersonsByZoneQuery = `
		SELECT id, coords, raw_xml 
		FROM %s 
		WHERE zone_id IN (%s) 
		ORDER BY id 
		LIMIT ? OFFSET ?
	`
	
	// Count persons by zone
	countPersonsByZoneQuery = `
		SELECT COUNT(*) 
		FROM %s 
		WHERE zone_id IN (%s)
	`
	
	// Get zone statistics
	selectZoneStatsQuery = `
		SELECT zone_id, COUNT(*) as person_count 
		FROM %s 
		WHERE zone_id IS NOT NULL 
		GROUP BY zone_id
	`
)

// CreateZoneTables creates the zone tables and indexes
func (db *Database) CreateZoneTables() error {
	// Create zones table
	if _, err := db.execQuery(createZoneTableQuery, "failed to create zones table"); err != nil {
		return err
	}
	
	// Create spatial index
	if _, err := db.execQuery(createZoneIndexQuery, "failed to create zone index"); err != nil {
		return err
	}
	
	return nil
}

// AddZoneColumnToPopulationTable adds zone_id column to a population table
func (db *Database) AddZoneColumnToPopulationTable(tableName string) error {
	// Check if column already exists
	var count int
	query := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM pragma_table_info('%s') 
		WHERE name = 'zone_id'
	`, tableName)
	
	err := db.queryRow(query, "failed to check column existence").Scan(&count)
	if err != nil {
		return err
	}
	
	if count == 0 {
		// Column doesn't exist, add it
		alterQuery := fmt.Sprintf(addZoneColumnQuery, tableName)
		if _, err := db.execQuery(alterQuery, "failed to add zone_id column"); err != nil {
			return err
		}
	}
	
	return nil
}

// SaveZone inserts or updates a zone
func (db *Database) SaveZone(zone *Zone) error {
	// Serialize polygon to JSON
	polygonJSON, err := json.Marshal(zone.Polygon)
	if err != nil {
		return fmt.Errorf("failed to serialize polygon: %w", err)
	}
	
	// Check if zone exists
	var exists bool
	err = db.queryRow(
		"SELECT EXISTS(SELECT 1 FROM zones WHERE id = ?)",
		"failed to check zone existence",
		zone.ID,
	).Scan(&exists)
	if err != nil {
		return err
	}
	
	if exists {
		// Update existing zone
		_, err = db.execQuery(
			updateZoneQuery,
			"failed to update zone",
			zone.Name,
			string(polygonJSON),
			zone.BoundingBox.MinX,
			zone.BoundingBox.MinY,
			zone.BoundingBox.MaxX,
			zone.BoundingBox.MaxY,
			zone.ID,
		)
	} else {
		// Insert new zone
		_, err = db.execQuery(
			insertZoneQuery,
			"failed to insert zone",
			zone.ID,
			zone.Name,
			string(polygonJSON),
			zone.BoundingBox.MinX,
			zone.BoundingBox.MinY,
			zone.BoundingBox.MaxX,
			zone.BoundingBox.MaxY,
		)
	}
	
	return err
}

// DeleteZone removes a zone
func (db *Database) DeleteZone(zoneID string) error {
	_, err := db.execQuery(deleteZoneQuery, "failed to delete zone", zoneID)
	return err
}

// GetAllZones retrieves all zones
func (db *Database) GetAllZones() ([]*Zone, error) {
	rows, err := db.queryRows(selectAllZonesQuery, "failed to query zones")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	zones := make([]*Zone, 0)
	for rows.Next() {
		zone, err := scanZone(rows)
		if err != nil {
			return nil, err
		}
		zones = append(zones, zone)
	}
	
	return zones, rows.Err()
}

// GetZone retrieves a zone by ID
func (db *Database) GetZone(zoneID string) (*Zone, error) {
	var polygonJSON string
	zone := &Zone{}
	
	err := db.queryRow(
		selectZoneByIDQuery,
		"failed to query zone",
		zoneID,
	).Scan(
		&zone.ID,
		&zone.Name,
		&polygonJSON,
		&zone.BoundingBox.MinX,
		&zone.BoundingBox.MinY,
		&zone.BoundingBox.MaxX,
		&zone.BoundingBox.MaxY,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("zone not found: %s", zoneID)
	}
	if err != nil {
		return nil, err
	}
	
	// Deserialize polygon
	if err := json.Unmarshal([]byte(polygonJSON), &zone.Polygon); err != nil {
		return nil, fmt.Errorf("failed to deserialize polygon: %w", err)
	}
	
	return zone, nil
}

// FindZonesContainingPoint finds all zones that contain the given point
func (db *Database) FindZonesContainingPoint(x, y float64) ([]*Zone, error) {
	// First, use bounding box index to narrow down candidates
	rows, err := db.queryRows(
		selectZonesContainingPointQuery,
		"failed to query zones by bbox",
		x, x, y, y,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var zones []*Zone
	for rows.Next() {
		zone, err := scanZone(rows)
		if err != nil {
			return nil, err
		}
		
		// Precise point-in-polygon check
		if zone.ContainsPoint(x, y) {
			zones = append(zones, zone)
		}
	}
	
	return zones, rows.Err()
}

// scanZone scans a zone from a database row
func scanZone(rows *sql.Rows) (*Zone, error) {
	var polygonJSON string
	zone := &Zone{}
	
	err := rows.Scan(
		&zone.ID,
		&zone.Name,
		&polygonJSON,
		&zone.BoundingBox.MinX,
		&zone.BoundingBox.MinY,
		&zone.BoundingBox.MaxX,
		&zone.BoundingBox.MaxY,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan zone: %w", err)
	}
	
	// Deserialize polygon
	if err := json.Unmarshal([]byte(polygonJSON), &zone.Polygon); err != nil {
		return nil, fmt.Errorf("failed to deserialize polygon: %w", err)
	}
	
	return zone, nil
}

// AssignPersonToZone updates a person's zone_id
func (db *Database) AssignPersonToZone(tableName, personID, zoneID string) error {
	query := fmt.Sprintf(updatePersonZoneQuery, tableName)
	_, err := db.execQuery(query, "failed to assign person to zone", zoneID, personID)
	return err
}

// GetPersonsByZones retrieves persons filtered by zone IDs with pagination
func (db *Database) GetPersonsByZones(tableName string, zoneIDs []string, page, pageSize int) (*PaginatedResponse, error) {
	if len(zoneIDs) == 0 {
		// If no zones specified, return empty result
		return &PaginatedResponse{
			Persons:     []Person{},
			TotalCount:  0,
			CurrentPage: page,
			TotalPages:  0,
			PageSize:    pageSize,
		}, nil
	}
	
	// Build placeholders for zone IDs
	placeholders := make([]string, len(zoneIDs))
	args := make([]interface{}, 0, len(zoneIDs)+2)
	
	for i, zoneID := range zoneIDs {
		placeholders[i] = "?"
		args = append(args, zoneID)
	}
	
	// Count total persons in zones
	countQuery := fmt.Sprintf(countPersonsByZoneQuery, tableName, strings.Join(placeholders, ","))
	var totalCount int
	err := db.queryRow(countQuery, "failed to count persons by zone", args...).Scan(&totalCount)
	if err != nil {
		return nil, err
	}
	
	// Calculate pagination
	totalPages := (totalCount + pageSize - 1) / pageSize
	offset := (page - 1) * pageSize
	
	// Query persons
	selectQuery := fmt.Sprintf(selectPersonsByZoneQuery, tableName, strings.Join(placeholders, ","))
	args = append(args, pageSize, offset)
	
	rows, err := db.queryRows(selectQuery, "failed to query persons by zone", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var persons []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.Coords, &p.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan person: %w", err)
		}
		persons = append(persons, p)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return &PaginatedResponse{
		Persons:     persons,
		TotalCount:  totalCount,
		CurrentPage: page,
		TotalPages:  totalPages,
		PageSize:    pageSize,
	}, nil
}

// GetZoneStats retrieves zone statistics for a population table
func (db *Database) GetZoneStats(tableName string) (map[string]int, error) {
	query := fmt.Sprintf(selectZoneStatsQuery, tableName)
	rows, err := db.queryRows(query, "failed to query zone stats")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	stats := make(map[string]int)
	for rows.Next() {
		var zoneID string
		var count int
		if err := rows.Scan(&zoneID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan zone stats: %w", err)
		}
		stats[zoneID] = count
	}
	
	return stats, rows.Err()
}

// AssignAllPersonsToZones processes all persons in a table and assigns them to zones
func (db *Database) AssignAllPersonsToZones(tableName string) error {
	// Get all zones
	zones, err := db.GetAllZones()
	if err != nil {
		return fmt.Errorf("failed to get zones: %w", err)
	}
	
	if len(zones) == 0 {
		return nil // No zones to assign
	}
	
	// Process persons in batches
	const batchSize = 1000
	offset := 0
	
	for {
		// Get batch of persons
		query := fmt.Sprintf("SELECT id, coords FROM %s LIMIT ? OFFSET ?", tableName)
		rows, err := db.queryRows(query, "failed to query persons", batchSize, offset)
		if err != nil {
			return err
		}
		
		var updates []struct {
			personID string
			zoneID   string
		}
		
		for rows.Next() {
			var personID, coords string
			if err := rows.Scan(&personID, &coords); err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan person: %w", err)
			}
			
			// Parse coordinates
			x, y, err := ParseCoordinates(coords)
			if err != nil {
				continue // Skip invalid coordinates
			}
			
			// Find containing zone
			for _, zone := range zones {
				if zone.ContainsPoint(x, y) {
					updates = append(updates, struct {
						personID string
						zoneID   string
					}{personID, zone.ID})
					break // Assuming non-overlapping zones
				}
			}
		}
		rows.Close()
		
		// Apply updates in a transaction
		if len(updates) > 0 {
			err = db.WithTransaction(func(tx *sql.Tx) error {
				stmt, err := tx.Prepare(fmt.Sprintf(updatePersonZoneQuery, tableName))
				if err != nil {
					return err
				}
				defer stmt.Close()
				
				for _, update := range updates {
					if _, err := stmt.Exec(update.zoneID, update.personID); err != nil {
						return err
					}
				}
				
				return nil
			})
			
			if err != nil {
				return fmt.Errorf("failed to update person zones: %w", err)
			}
		}
		
		// Check if we're done
		if len(updates) < batchSize {
			break
		}
		
		offset += batchSize
	}
	
	return nil
}