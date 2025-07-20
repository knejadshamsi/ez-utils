package gui

import (
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// FilterSession represents a population filtering session
type FilterSession struct {
	SessionID   string
	TableName   string
	CreatedAt   time.Time
	mu          sync.RWMutex
}

// Global session storage
var (
	filterSessions = make(map[string]*FilterSession)
	sessionMutex   sync.RWMutex
)

// CreateFilterSession creates a new filter session with temporary tables
func (a *App) CreateFilterSession(tableName string) (string, error) {
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	
	// Create temp tables for this session
	err := a.db.CreateTempFilterTables(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to create temp tables: %w", err)
	}
	
	// Store session info
	session := &FilterSession{
		SessionID: sessionID,
		TableName: tableName,
		CreatedAt: time.Now(),
	}
	
	sessionMutex.Lock()
	filterSessions[sessionID] = session
	sessionMutex.Unlock()
	
	return sessionID, nil
}

// CreateTempFilterTables creates the temporary tables for a filter session
func (db *Database) CreateTempFilterTables(sessionID string) error {
	// Create zone-person mapping table
	createMappingQuery := fmt.Sprintf(`
		CREATE TEMP TABLE IF NOT EXISTS temp_person_zones_%s (
			zone_id TEXT,
			person_id TEXT,
			PRIMARY KEY (zone_id, person_id)
		)
	`, sessionID)
	
	// Create person cache table
	createCacheQuery := fmt.Sprintf(`
		CREATE TEMP TABLE IF NOT EXISTS temp_person_cache_%s (
			person_id TEXT PRIMARY KEY,
			coords TEXT,
			raw_xml TEXT
		)
	`, sessionID)
	
	// Execute in transaction
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	if _, err := tx.Exec(createMappingQuery); err != nil {
		return fmt.Errorf("failed to create mapping table: %w", err)
	}
	
	if _, err := tx.Exec(createCacheQuery); err != nil {
		return fmt.Errorf("failed to create cache table: %w", err)
	}
	
	return tx.Commit()
}

// AddZoneToSession adds a zone's population to the filter session
func (a *App) AddZoneToSession(sessionID string, zoneID string, polygon []Point) error {
	session, exists := filterSessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	// Calculate bounding box from polygon
	bbox := calculateBoundingBox(polygon)
	
	// Find persons within the zone
	persons, err := a.db.GetPersonsInPolygon(session.TableName, polygon, bbox)
	if err != nil {
		return fmt.Errorf("failed to get persons in zone: %w", err)
	}
	
	// Add to temp tables
	return a.db.AddPersonsToFilterSession(sessionID, zoneID, persons)
}

// calculateBoundingBox calculates the bounding box of a polygon
func calculateBoundingBox(polygon []Point) BoundingBox {
	if len(polygon) == 0 {
		return BoundingBox{}
	}
	
	minX, minY := polygon[0].X, polygon[0].Y
	maxX, maxY := polygon[0].X, polygon[0].Y
	
	for _, p := range polygon[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	
	// Convert to North/South/East/West format
	return BoundingBox{
		North: maxY,
		South: minY,
		East:  maxX,
		West:  minX,
	}
}

// AddPersonsToFilterSession adds persons to the filter session tables
func (db *Database) AddPersonsToFilterSession(sessionID string, zoneID string, persons []Person) error {
	if len(persons) == 0 {
		return nil
	}
	
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Prepare statements
	mappingStmt, err := tx.Prepare(fmt.Sprintf(
		`INSERT OR IGNORE INTO temp_person_zones_%s (zone_id, person_id) VALUES (?, ?)`,
		sessionID,
	))
	if err != nil {
		return err
	}
	defer mappingStmt.Close()
	
	cacheStmt, err := tx.Prepare(fmt.Sprintf(
		`INSERT OR IGNORE INTO temp_person_cache_%s (person_id, coords, raw_xml) VALUES (?, ?, ?)`,
		sessionID,
	))
	if err != nil {
		return err
	}
	defer cacheStmt.Close()
	
	// Insert data
	for _, person := range persons {
		// Add to mapping table
		if _, err := mappingStmt.Exec(zoneID, person.ID); err != nil {
			return fmt.Errorf("failed to insert mapping: %w", err)
		}
		
		// Add to cache table
		if _, err := cacheStmt.Exec(person.ID, person.Coords, person.RawXML); err != nil {
			return fmt.Errorf("failed to insert to cache: %w", err)
		}
	}
	
	return tx.Commit()
}

// RemoveZoneFromSession removes a zone from the filter session
func (a *App) RemoveZoneFromSession(sessionID string, zoneID string) error {
	if _, exists := filterSessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	return a.db.RemoveZoneFromFilterSession(sessionID, zoneID)
}

// RemoveZoneFromFilterSession removes zone data from temp tables
func (db *Database) RemoveZoneFromFilterSession(sessionID string, zoneID string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Delete zone mappings
	deleteMapping := fmt.Sprintf(
		`DELETE FROM temp_person_zones_%s WHERE zone_id = ?`,
		sessionID,
	)
	if _, err := tx.Exec(deleteMapping, zoneID); err != nil {
		return fmt.Errorf("failed to delete mappings: %w", err)
	}
	
	// Clean up persons not in any zone
	cleanupQuery := fmt.Sprintf(`
		DELETE FROM temp_person_cache_%s 
		WHERE person_id NOT IN (
			SELECT DISTINCT person_id FROM temp_person_zones_%s
		)
	`, sessionID, sessionID)
	
	if _, err := tx.Exec(cleanupQuery); err != nil {
		return fmt.Errorf("failed to cleanup orphaned persons: %w", err)
	}
	
	return tx.Commit()
}

// GetFilteredPopulationPage gets a page of filtered population data
func (a *App) GetFilteredPopulationPage(sessionID string, page int, pageSize int) (*PaginatedResponse, error) {
	if _, exists := filterSessions[sessionID]; !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	// Get total count
	totalCount, err := a.db.GetFilteredPopulationCount(sessionID)
	if err != nil {
		return nil, err
	}
	
	// Get page data
	persons, err := a.db.GetFilteredPopulationData(sessionID, page, pageSize)
	if err != nil {
		return nil, err
	}
	
	totalPages := (totalCount + pageSize - 1) / pageSize
	
	return &PaginatedResponse{
		Persons:      persons,
		TotalCount:   totalCount,
		CurrentPage:  page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
	}, nil
}

// GetFilteredPopulationCount gets the total count of filtered persons
func (db *Database) GetFilteredPopulationCount(sessionID string) (int, error) {
	query := fmt.Sprintf(
		`SELECT COUNT(DISTINCT person_id) FROM temp_person_zones_%s`,
		sessionID,
	)
	
	var count int
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count filtered population: %w", err)
	}
	
	return count, nil
}

// GetFilteredPopulationData retrieves paginated filtered population
func (db *Database) GetFilteredPopulationData(sessionID string, page int, pageSize int) ([]Person, error) {
	offset := (page - 1) * pageSize
	
	query := fmt.Sprintf(`
		SELECT 
			p.person_id,
			p.coords,
			p.raw_xml,
			GROUP_CONCAT(z.zone_id) as zones
		FROM temp_person_cache_%s p
		JOIN temp_person_zones_%s z ON p.person_id = z.person_id
		GROUP BY p.person_id
		ORDER BY p.person_id
		LIMIT ? OFFSET ?
	`, sessionID, sessionID)
	
	rows, err := db.conn.Query(query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query filtered population: %w", err)
	}
	defer rows.Close()
	
	var persons []Person
	for rows.Next() {
		var p Person
		var zones sql.NullString
		
		if err := rows.Scan(&p.ID, &p.Coords, &p.RawXML, &zones); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		// Note: zones are available but not included in Person struct
		// Could be added if needed for UI display
		persons = append(persons, p)
	}
	
	return persons, rows.Err()
}

// CloseFilterSession closes and cleans up a filter session
func (a *App) CloseFilterSession(sessionID string) error {
	sessionMutex.Lock()
	delete(filterSessions, sessionID)
	sessionMutex.Unlock()
	
	// Temp tables are automatically cleaned up when connection closes
	// But we can explicitly drop them if needed
	return a.db.DropTempFilterTables(sessionID)
}

// DropTempFilterTables explicitly drops temp tables for a session
func (db *Database) DropTempFilterTables(sessionID string) error {
	dropMapping := fmt.Sprintf(`DROP TABLE IF EXISTS temp_person_zones_%s`, sessionID)
	dropCache := fmt.Sprintf(`DROP TABLE IF EXISTS temp_person_cache_%s`, sessionID)
	
	if _, err := db.conn.Exec(dropMapping); err != nil {
		return fmt.Errorf("failed to drop mapping table: %w", err)
	}
	
	if _, err := db.conn.Exec(dropCache); err != nil {
		return fmt.Errorf("failed to drop cache table: %w", err)
	}
	
	return nil
}

// GetPersonsInPolygon finds persons within a polygon (with bbox pre-filter)
func (db *Database) GetPersonsInPolygon(tableName string, polygon []Point, bbox BoundingBox) ([]Person, error) {
	// First, get persons within bounding box
	query := fmt.Sprintf(`
		SELECT id, coords, raw_xml 
		FROM %s 
		WHERE coords IS NOT NULL AND coords != ''
	`, tableName)
	
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var persons []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.Coords, &p.RawXML); err != nil {
			continue
		}
		
		// Parse coordinates
		x, y, err := ParseCoordinates(p.Coords)
		if err != nil {
			continue
		}
		
		// Quick bbox check first
		if x < bbox.West || x > bbox.East || y < bbox.South || y > bbox.North {
			continue
		}
		
		// Detailed point-in-polygon check
		point := Point{X: x, Y: y}
		if IsPointInPolygon(point, polygon) {
			persons = append(persons, p)
		}
	}
	
	return persons, rows.Err()
}