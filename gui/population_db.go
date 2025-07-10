package gui

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// Population table queries
const (
	selectAllFromPopulationQuery = `SELECT id, coords, raw_xml FROM %s`
	selectPopulationByBboxQuery  = `SELECT id, coords, raw_xml FROM %s WHERE coords IS NOT NULL AND coords != ''`
	selectPersonByIDQuery        = `SELECT id, coords, raw_xml FROM %s WHERE id = ?`
	updatePersonXMLQuery         = `UPDATE %s SET raw_xml = ? WHERE id = ?`
	updatePersonCoordsQuery      = `UPDATE %s SET coords = ? WHERE id = ?`
	updatePersonFullQuery        = `UPDATE %s SET coords = ?, raw_xml = ? WHERE id = ?`
	deletePersonQuery                   = `DELETE FROM %s WHERE id = ?`
	insertPersonQuery                   = `INSERT INTO %s (id, coords, raw_xml) VALUES (?, ?, ?)`
	populationInsertBatchPlaceholder    = "(?, ?, ?)"
	insertBatchBaseQuery                = "INSERT INTO %s (id, coords, raw_xml) VALUES "
)

// Population table error messages
const (
	selectAllFromPopulationError = "failed to query table %s"
	selectPopulationByBboxError  = "failed to query table %s"
	selectPersonByIDError        = "failed to query person in table %s"
	updatePersonXMLError         = "failed to update person XML in table %s"
	updatePersonCoordsError      = "failed to update person coords in table %s"
	updatePersonFullError        = "failed to prepare update statement"
	deletePersonError            = "failed to delete person from table %s"
	insertPersonError            = "failed to insert person into table %s"
	insertBatchError             = "failed to insert batch"
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
		if err := rows.Scan(&p.ID, &p.Coords, &p.RawXML); err != nil {
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
		if err := rows.Scan(&p.ID, &p.Coords, &p.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse coordinates and check if within bbox
		parts := strings.Split(p.Coords, ",")
		if len(parts) == 2 {
			x, err1 := strconv.ParseFloat(parts[0], 64)
			y, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				// Check if coordinates are within bounding box
				if x >= minLng && x <= maxLng && y >= minLat && y <= maxLat {
					persons = append(persons, p)
				}
			}
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
	err := db.queryRow(query, fmt.Sprintf(selectPersonByIDError, tableName), personID).Scan(&p.ID, &p.Coords, &p.RawXML)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("person with ID %s not found in table %s", personID, tableName)
		}
		return nil, fmt.Errorf("failed to query person: %w", err)
	}
	return &p, nil
}

// AddPerson adds a new person to a population table
func (db *Database) AddPerson(tableName string, id, coords, rawXML string) error {
	query := fmt.Sprintf(insertPersonQuery, tableName)
	if _, err := db.execQuery(query, fmt.Sprintf(insertPersonError, tableName), id, coords, rawXML); err != nil {
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
func (db *Database) UpdatePersonCoords(tableName, personID, coords string) error {
	query := fmt.Sprintf(updatePersonCoordsQuery, tableName)
	if _, err := db.execQuery(query, fmt.Sprintf(updatePersonCoordsError, tableName), coords, personID); err != nil {
		return fmt.Errorf("failed to update person coordinates: %w", err)
	}
	return nil
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
			_, err := stmt.Exec(update.Coords, update.RawXML, update.ID)
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
		args = append(args, person.ID, person.Coords, person.RawXML)
	}
	query += strings.Join(values, ", ")

	_, err := tx.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert batch: %w", err)
	}

	return nil
}
