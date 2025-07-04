package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

// InitDB initializes the database connection.
func InitDB(dataSourceName string) error {
	var err error
	if _, err := os.Stat(dataSourceName); os.IsNotExist(err) {
		file, err := os.Create(dataSourceName)
		if err != nil {
			return fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
	}

	db, err = sql.Open("sqlite", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	createStatusTableSQL := `
			CREATE TABLE IF NOT EXISTS processes (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				file_path TEXT NOT NULL,
				status TEXT NOT NULL,
				timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
			);`
	_, err = db.Exec(createStatusTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create status table: %w", err)
	}

	// Create telemetry table
	createTelemetryTableSQL := `
		CREATE TABLE IF NOT EXISTS process_telemetry (
			process_id INTEGER PRIMARY KEY,
			total_file_size INTEGER NOT NULL,
			bytes_read INTEGER DEFAULT 0,
			persons_extracted INTEGER DEFAULT 0,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (process_id) REFERENCES processes(id) ON DELETE CASCADE
		);`
	_, err = db.Exec(createTelemetryTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create telemetry table: %w", err)
	}
	
	// Add error_count column if it doesn't exist (migration)
	addErrorCountSQL := `
		ALTER TABLE process_telemetry ADD COLUMN error_count INTEGER DEFAULT 0;`
	_, err = db.Exec(addErrorCountSQL)
	// Ignore error if column already exists
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		log.Printf("Note: error_count column may already exist: %v", err)
	}

	return nil
}

// GetDB returns the initialized database connection.
func GetDB() (*sql.DB, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return db, nil
}

var ErrProcessNotFound = fmt.Errorf("process not found")

// GetProcessStatus retrieves the status of a process from the database.
func GetProcessStatus(processID int) (string, error) {
	var status string
	err := db.QueryRow("SELECT status FROM processes WHERE id = ?", processID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrProcessNotFound
		}
		return "", fmt.Errorf("failed to query process status for ID %d: %w", processID, err)
	}
	return status, nil
}

// Person defines the structure for the population data.
type Person struct {
	ID     string `json:"id"`
	Coords string `json:"coords"` // "x,y"
	RawXML string `json:"raw_xml"`
}

// PersonData holds the extracted information for a single person (used by processor)
type PersonData struct {
	ID     string
	Coords string // "x,y"
	RawXML string
}

// CreateTable creates a new table for population data if it doesn't already exist.
func CreateTable(tableName string) error {
	// IMPORTANT: The caller is responsible for sanitizing tableName to prevent SQL injection.
	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            id TEXT PRIMARY KEY,
            coords TEXT,
            raw_xml TEXT
        )`, tableName)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %w", tableName, err)
	}
	return nil
}

// GetPopulationData retrieves all person data from a specific table.
func GetPopulationData(tableName string) ([]Person, error) {
	// IMPORTANT: The caller is responsible for sanitizing tableName to prevent SQL injection.
	rows, err := db.Query(fmt.Sprintf("SELECT id, coords, raw_xml FROM %s", tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query population data from %s: %w", tableName, err)
	}
	defer rows.Close()

	var people []Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.Coords, &p.RawXML); err != nil {
			return nil, fmt.Errorf("failed to scan person row: %w", err)
		}
		people = append(people, p)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating population data rows: %w", err)
	}

	return people, nil
}

// UpdatePersonXML updates the raw_xml for a specific person in a given table.
func UpdatePersonXML(tableName, personID, rawXML string) error {
	// IMPORTANT: The caller is responsible for sanitizing tableName to prevent SQL injection.
	query := fmt.Sprintf("UPDATE %s SET raw_xml = ? WHERE id = ?", tableName)
	_, err := db.Exec(query, rawXML, personID)
	if err != nil {
		return fmt.Errorf("failed to update person xml in %s for person %s: %w", tableName, personID, err)
	}
	return nil
}

// UpdatePersonCoords updates the coords for a specific person in a given table.
func UpdatePersonCoords(tableName, personID, coords string) error {
	// IMPORTANT: The caller is responsible for sanitizing tableName to prevent SQL injection.
	query := fmt.Sprintf("UPDATE %s SET coords = ? WHERE id = ?", tableName)
	_, err := db.Exec(query, coords, personID)
	if err != nil {
		return fmt.Errorf("failed to update person coords in %s for person %s: %w", tableName, personID, err)
	}
	return nil
}

// DeletePerson removes a person's record from the specified table.
func DeletePerson(tableName, personID string) error {
	// IMPORTANT: The caller is responsible for sanitizing tableName to prevent SQL injection.
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	_, err := db.Exec(query, personID)
	if err != nil {
		return fmt.Errorf("failed to delete person in %s: %w", tableName, err)
	}
	return nil
}

// AddPerson adds a new person to the specified table.
func AddPerson(tableName string, personData map[string]interface{}) error {
	// IMPORTANT: The caller is responsible for sanitizing tableName to prevent SQL injection.
	var p Person

	if id, ok := personData["id"].(string); ok {
		p.ID = id
	} else {
		return fmt.Errorf("person id is required and must be a string")
	}

	// Coords can be provided as a "x,y" string or as separate x and y float64 values.
	if coords, ok := personData["coords"].(string); ok {
		p.Coords = coords
	} else {
		x, xOk := personData["home_x"].(float64)
		y, yOk := personData["home_y"].(float64)
		if xOk && yOk {
			p.Coords = fmt.Sprintf("%f,%f", x, y)
		}
	}

	// RawXML is optional.
	if rawXML, ok := personData["raw_xml"].(string); ok {
		p.RawXML = rawXML
	}

	// For backward compatibility, check for a 'plan' field. If it exists,
	// it will be wrapped in a basic person structure to form the raw_xml.
	if plan, ok := personData["plan"].(string); ok && p.RawXML == "" {
		p.RawXML = fmt.Sprintf(`<person id="%s"><plan>%s</plan></person>`, p.ID, plan)
	}

	query := fmt.Sprintf("INSERT INTO %s (id, coords, raw_xml) VALUES (?, ?, ?)", tableName)
	_, err := db.Exec(query, p.ID, p.Coords, p.RawXML)
	if err != nil {
		return fmt.Errorf("failed to add person to %s: %w", tableName, err)
	}
	return nil
}

// DropTable drops a table if it exists.
func DropTable(tableName string) error {
	// IMPORTANT: The caller is responsible for sanitizing tableName to prevent SQL injection.
	if !strings.HasPrefix(strings.ToLower(tableName), "population_data_") {
		return fmt.Errorf("table name must begin with 'population_data_' for safety")
	}
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop table %s: %w", tableName, err)
	}
	return nil
}

// Process defines the structure for a record from the 'processes' table.
type Process struct {
	ID          int    `json:"id"`
	FilePath    string `json:"file_path"`
	Status      string `json:"status"`
	Timestamp   string `json:"timestamp"`
	TableName   string `json:"table_name"`
	RecordCount int    `json:"record_count"`
}

// GetProcessesByFile retrieves all process records for a specific file path.
func GetProcessesByFile(filePath string) ([]Process, error) {
	rows, err := db.Query("SELECT id, file_path, status, timestamp FROM processes WHERE file_path = ? ORDER BY timestamp DESC", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to query processes for file %s: %w", filePath, err)
	}
	defer rows.Close()
	return scanProcesses(rows)
}

// DeleteProcess deletes a process record from the database.
func DeleteProcess(processID int) error {
	query := "DELETE FROM processes WHERE id = ?"
	_, err := db.Exec(query, processID)
	if err != nil {
		return fmt.Errorf("failed to delete process with ID %d: %w", processID, err)
	}
	return nil
}

// CreateProcess creates a new process record and returns its ID.
func CreateProcess(filePath string) (int, error) {
	query := "INSERT INTO processes (file_path, status) VALUES (?, ?)"
	result, err := db.Exec(query, filePath, "started")
	if err != nil {
		return 0, fmt.Errorf("failed to create process: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}
	return int(id), nil
}

// UpdateProcessStatus updates the status of an existing process.
func UpdateProcessStatus(processID int, status string) error {
	query := "UPDATE processes SET status = ? WHERE id = ?"
	_, err := db.Exec(query, status, processID)
	if err != nil {
		return fmt.Errorf("failed to update process status for ID %d: %w", processID, err)
	}
	return nil
}

// GetProcesses retrieves all process records.
func GetProcesses() ([]Process, error) {
	rows, err := db.Query("SELECT id, file_path, status, timestamp FROM processes ORDER BY timestamp DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query processes: %w", err)
	}
	defer rows.Close()
	return scanProcesses(rows)
}

// scanProcesses is a helper function to scan rows into a slice of Process structs.
func scanProcesses(rows *sql.Rows) ([]Process, error) {
	var processes []Process
	for rows.Next() {
		var p Process
		if err := rows.Scan(&p.ID, &p.FilePath, &p.Status, &p.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan process row: %w", err)
		}

		p.TableName = fmt.Sprintf("population_data_%d", p.ID)

		var count int
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", p.TableName)
		err := db.QueryRow(countQuery).Scan(&count)
		if err != nil {
			// This can happen if processing failed before the table was created.
			// Default to 0 in case of error.
			p.RecordCount = 0
		} else {
			p.RecordCount = count
		}

		processes = append(processes, p)
	}
	return processes, nil
}

// InitializeTelemetry creates an initial telemetry record for a process
func InitializeTelemetry(processID int, totalFileSize int64) error {
	query := `INSERT INTO process_telemetry (process_id, total_file_size, bytes_read, persons_extracted) 
		  VALUES (?, ?, 0, 0)`
	_, err := db.Exec(query, processID, totalFileSize)
	if err != nil {
		return fmt.Errorf("failed to initialize telemetry for process %d: %w", processID, err)
	}
	return nil
}

// UpdateTelemetry updates the telemetry data for a process
func UpdateTelemetry(processID int, bytesRead, personsExtracted, errorCount int64) error {
	query := `UPDATE process_telemetry 
		  SET bytes_read = ?, persons_extracted = ?, error_count = ?, last_updated = CURRENT_TIMESTAMP 
		  WHERE process_id = ?`
	_, err := db.Exec(query, bytesRead, personsExtracted, errorCount, processID)
	if err != nil {
		return fmt.Errorf("failed to update telemetry for process %d: %w", processID, err)
	}
	return nil
}

// ProcessTelemetry represents telemetry data for a process
type ProcessTelemetry struct {
	ProcessID        int       `json:"process_id"`
	TotalFileSize    int64     `json:"total_file_size"`
	BytesRead        int64     `json:"bytes_read"`
	PersonsExtracted int64     `json:"persons_extracted"`
	ErrorCount       int64     `json:"error_count"`
	LastUpdated      time.Time `json:"last_updated"`
}

// GetTelemetry retrieves telemetry data for a specific process
func GetTelemetry(processID int) (*ProcessTelemetry, error) {
	var telemetry ProcessTelemetry
	var lastUpdatedStr string
	query := `SELECT process_id, total_file_size, bytes_read, persons_extracted, 
		  COALESCE(error_count, 0) as error_count, last_updated 
		  FROM process_telemetry WHERE process_id = ?`
	err := db.QueryRow(query, processID).Scan(
		&telemetry.ProcessID,
		&telemetry.TotalFileSize,
		&telemetry.BytesRead,
		&telemetry.PersonsExtracted,
		&telemetry.ErrorCount,
		&lastUpdatedStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("telemetry not found for process %d", processID)
		}
		return nil, fmt.Errorf("failed to get telemetry for process %d: %w", processID, err)
	}
	
	// Parse the timestamp string into time.Time
	telemetry.LastUpdated, err = time.Parse("2006-01-02 15:04:05", lastUpdatedStr)
	if err != nil {
		// If parsing fails, use current time as fallback
		telemetry.LastUpdated = time.Now()
	}
	
	return &telemetry, nil
}