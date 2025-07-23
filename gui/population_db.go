package gui

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// Population table queries
const (
	selectAllFromPopulationQuery = `SELECT id, lng, lat, raw_xml FROM %s`
	selectPopulationByBboxQuery  = `SELECT id, lng, lat, raw_xml FROM %s WHERE lng IS NOT NULL AND lat IS NOT NULL`
	selectPersonByIDQuery        = `SELECT id, lng, lat, raw_xml FROM %s WHERE id = ?`
	selectPopulationPaginatedQuery = `SELECT id, lng, lat, raw_xml FROM %s ORDER BY id LIMIT ? OFFSET ?`
	countPopulationQuery = `SELECT COUNT(*) FROM %s`
	selectZonesWithCountsQuery = `SELECT lng, lat, COUNT(*) as count FROM %s WHERE lng IS NOT NULL AND lat IS NOT NULL GROUP BY lng, lat`
	updatePersonXMLQuery         = `UPDATE %s SET raw_xml = ? WHERE id = ?`
	updatePersonCoordsQuery      = `UPDATE %s SET lng = ?, lat = ? WHERE id = ?`
	updatePersonFullQuery        = `UPDATE %s SET lng = ?, lat = ?, raw_xml = ? WHERE id = ?`
	deletePersonQuery                   = `DELETE FROM %s WHERE id = ?`
	insertPersonQuery                   = `INSERT INTO %s (id, lng, lat, raw_xml) VALUES (?, ?, ?, ?)`
	populationInsertBatchPlaceholder    = "(?, ?, ?, ?)"
	insertBatchBaseQuery                = "INSERT INTO %s (id, lng, lat, raw_xml) VALUES "
)

// Population table error messages
const (
	selectAllFromPopulationError = "failed to query table %s"
	selectPopulationByBboxError  = "failed to query table %s"
	selectPersonByIDError        = "failed to query person in table %s"
	updatePersonXMLError         = "failed to update person XML in table %s"
	updatePersonCoordsError      = "failed to update person location in table %s"
	updatePersonFullError        = "failed to prepare update statement"
	deletePersonError            = "failed to delete person from table %s"
	insertPersonError            = "failed to insert person into table %s"
	insertBatchError             = "failed to insert batch"
	selectPopulationPaginatedError = "failed to query paginated population from table %s"
	countPopulationError         = "failed to count population in table %s"
	selectZonesWithCountsError   = "failed to query zones with counts from table %s"
)

// GetPopulationData retrieves all person data from a specific table
func (db *Database) GetPopulationData(tableName string) ([]Person, error) {
	query := fmt.Sprintf(selectAllFromPopulationQuery, tableName)
	rows, err := db.queryRows(query, fmt.Sprintf(selectAllFromPopulationError, tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.Lng, &p.Lat, &p.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		persons = append(persons, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return persons, nil
}

// GetPopulationByBbox retrieves population data within a bounding box
func (db *Database) GetPopulationByBbox(tableName string, minLat, minLng, maxLat, maxLng float64) ([]Person, error) {
	query := fmt.Sprintf(selectPopulationByBboxQuery, tableName)
	rows, err := db.queryRows(query, fmt.Sprintf(selectPopulationByBboxError, tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.Lng, &p.Lat, &p.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Check if coordinates are within bounding box
		if p.Lng >= minLng && p.Lng <= maxLng && p.Lat >= minLat && p.Lat <= maxLat {
			persons = append(persons, p)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return persons, nil
}

// GetPerson retrieves a single person by ID from the specified table
func (db *Database) GetPerson(tableName, personID string) (*Person, error) {
	query := fmt.Sprintf(selectPersonByIDQuery, tableName)
	var p Person
	err := db.queryRow(query, fmt.Sprintf(selectPersonByIDError, tableName), personID).Scan(&p.ID, &p.Lng, &p.Lat, &p.RawXML)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("person with ID %s not found in table %s", personID, tableName)
		}
		return nil, fmt.Errorf("failed to query person: %w", err)
	}
	return &p, nil
}

// AddPerson adds a new person to a population table
func (db *Database) AddPerson(tableName string, id string, lng, lat float64, rawXML string) error {
	query := fmt.Sprintf(insertPersonQuery, tableName)
	if _, err := db.execQuery(query, fmt.Sprintf(insertPersonError, tableName), id, lng, lat, rawXML); err != nil {
		return fmt.Errorf("failed to add person: %w", err)
	}
	return nil
}

// DeletePerson deletes a person from a population table
func (db *Database) DeletePerson(tableName, personID string) error {
	query := fmt.Sprintf(deletePersonQuery, tableName)
	if _, err := db.execQuery(query, fmt.Sprintf(deletePersonError, tableName), personID); err != nil {
		return fmt.Errorf("failed to delete person: %w", err)
	}
	return nil
}

// UpdatePersonXML updates a person's XML data
func (db *Database) UpdatePersonXML(tableName, personID, rawXML string) error {
	query := fmt.Sprintf(updatePersonXMLQuery, tableName)
	if _, err := db.execQuery(query, fmt.Sprintf(updatePersonXMLError, tableName), rawXML, personID); err != nil {
		return fmt.Errorf("failed to update person XML: %w", err)
	}
	return nil
}

// UpdatePersonCoords updates a person's coordinates
func (db *Database) UpdatePersonCoords(tableName, personID string, lng, lat float64) error {
	query := fmt.Sprintf(updatePersonCoordsQuery, tableName)
	if _, err := db.execQuery(query, fmt.Sprintf(updatePersonCoordsError, tableName), lng, lat, personID); err != nil {
		return fmt.Errorf("failed to update person coordinates: %w", err)
	}
	return nil
}

// PaginatedResponse represents a paginated query response
type PaginatedResponse struct {
	Persons     []Person `json:"persons"`
	TotalCount  int      `json:"totalCount"`
	CurrentPage int      `json:"currentPage"`
	TotalPages  int      `json:"totalPages"`
	PageSize    int      `json:"pageSize"`
}

// ZoneCount represents a zone with its person count
type ZoneCount struct {
	ZoneID string `json:"zoneId"`
	Count  int    `json:"count"`
}

// GetPopulationPaginated retrieves paginated population data
// Uses dynamic zone filtering - if no zones selected, returns empty result
func (db *Database) GetPopulationPaginated(tableName string, page, pageSize int, zoneFilter []string) (*PaginatedResponse, error) {
	// Use dynamic point-in-polygon filtering
	return db.GetPopulationByZonesDynamic(tableName, zoneFilter, page, pageSize)
}

// GetZonesWithCounts retrieves all zones with their person counts
// DEPRECATED: Use GetZoneStats instead for proper zone management
func (db *Database) GetZonesWithCounts(tableName string) ([]ZoneCount, error) {
	// Get zone statistics
	stats, err := db.GetZoneStats(tableName)
	if err != nil {
		return nil, err
	}
	
	// Get all zones
	zones, err := db.GetAllZones()
	if err != nil {
		return nil, err
	}
	
	// Convert to ZoneCount format
	var zoneCounts []ZoneCount
	for _, zone := range zones {
		count := stats[zone.ID]
		zoneCounts = append(zoneCounts, ZoneCount{
			ZoneID: zone.ID,
			Count:  count,
		})
	}
	
	return zoneCounts, nil
}




// BatchUpdatePersons updates multiple persons in a single transaction
func (db *Database) BatchUpdatePersons(tableName string, updates []PersonUpdate) error {
	return db.WithTransaction(func(tx *sql.Tx) error {
		query := fmt.Sprintf(updatePersonFullQuery, tableName)
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("failed to prepare update statement: %w", err)
		}
		defer stmt.Close()

		for _, update := range updates {
			_, err = stmt.Exec(update.Lng, update.Lat, update.RawXML, update.ID)
			if err != nil {
				return fmt.Errorf("failed to update person %s: %w", update.ID, err)
			}
		}

		return nil
	})
}

// InsertBatch inserts a batch of PersonData into the specified table
func InsertBatch(tx *sql.Tx, tableName string, batch []PersonData) error {
	if len(batch) == 0 {
		return nil
	}

	query := fmt.Sprintf(insertBatchBaseQuery, tableName)
	var values []string
	var args []any
	for _, person := range batch {
		values = append(values, populationInsertBatchPlaceholder)
		args = append(args, person.ID, person.Lng, person.Lat, person.RawXML)
	}
	query += strings.Join(values, ", ")

	_, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert batch: %w", err)
	}

	return nil
}
