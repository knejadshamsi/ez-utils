package gui

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Process query constants
const (
	deleteProcessQuery = `DELETE FROM processes WHERE process_id = ?`
	createProcessQuery = `INSERT INTO processes (file_path, status) VALUES (?, 'PENDING')`
	updateProcessStatusQuery = `UPDATE processes SET status = ? WHERE process_id = ?`
	initTelemetryQuery = `INSERT INTO process_telemetry (process_id, total_file_size, bytes_read, persons_extracted) 
		  VALUES (?, ?, 0, 0)`
	updateTelemetryQuery = `UPDATE process_telemetry 
		  SET bytes_read = ?, persons_extracted = ?, error_count = ?, last_updated = CURRENT_TIMESTAMP 
		  WHERE process_id = ?`
	selectAllProcessesQuery = `SELECT process_id, file_path, status, timestamp FROM processes ORDER BY timestamp DESC`
	selectProcessStatusQuery = `SELECT status FROM processes WHERE process_id = ?`
	selectProcessesByFileQuery = `SELECT process_id, file_path, status, timestamp FROM processes WHERE file_path = ? ORDER BY timestamp DESC`
	selectTelemetryQuery = `SELECT process_id, total_file_size, bytes_read, persons_extracted, 
		  COALESCE(error_count, 0) as error_count, last_updated 
		  FROM process_telemetry WHERE process_id = ?`
)

// Process error messages
const (
	deleteProcessError = "failed to delete process with ID %d"
	createProcessError = "failed to create process for file %s"
	updateProcessStatusError = "failed to update process status for ID %d"
	initTelemetryError = "failed to initialize telemetry for process %d"
	updateTelemetryError = "failed to update telemetry for process %d"
	selectAllProcessesError = "failed to query processes"
	selectProcessStatusError = "failed to query process status for ID %d"
	selectProcessesByFileError = "failed to query processes for file %s"
	selectTelemetryError = "failed to get telemetry for process %d"
)

// GetProcesses retrieves all process records from the database.
func (a *App) GetProcesses() ([]Process, error) {
	rows, err := a.db.queryRows(selectAllProcessesQuery, selectAllProcessesError)
	if err != nil {
		return make([]Process, 0), err
	}
	defer rows.Close()

	processes, err := scanProcesses(rows)
	if err != nil {
		return make([]Process, 0), err
	}
	if processes == nil {
		return make([]Process, 0), nil
	}
	return processes, nil
}

// CheckProcessingStatus retrieves the status of a specific process.
func (a *App) CheckProcessingStatus(processID int) (string, error) {
	var status string
	err := a.db.queryRow(selectProcessStatusQuery, fmt.Sprintf(selectProcessStatusError, processID), processID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrProcessNotFound
		}
		return "", fmt.Errorf("failed to query process status for ID %d: %w", processID, err)
	}
	return status, nil
}

// GetProcessesByFile retrieves all process records for a given file path.
func (a *App) GetProcessesByFile(filePath string) []Process {
	log.Printf("GetProcessesByFile called with filePath: %s", filePath)
	
	if a.db == nil {
		log.Printf("Database is nil!")
		// Return an initialized slice, not nil
		return make([]Process, 0)
	}
	
	rows, err := a.db.queryRows(selectProcessesByFileQuery, fmt.Sprintf(selectProcessesByFileError, filePath), filePath)
	if err != nil {
		log.Printf("Failed to query processes for file %s: %v", filePath, err)
		// Return an initialized slice, not nil
		return make([]Process, 0)
	}
	defer rows.Close()

	processes, err := scanProcesses(rows)
	if err != nil {
		log.Printf("Failed to scan processes for file %s: %v", filePath, err)
		// Return an initialized slice, not nil
		return make([]Process, 0)
	}

	// Ensure we never return nil, even if processes is nil
	if processes == nil {
		processes = make([]Process, 0)
	}

	log.Printf("GetProcessesByFile returning %d processes", len(processes))
	return processes
}

// deleteProcess deletes a process and its associated data (internal function for interpreter)
func (a *App) deleteProcess(processID int) error {
	// Drop all possible tables associated with this process
	// Population tables
	populationTable := fmt.Sprintf("population_data_%d", processID)
	_ = a.db.execTableQuery(fmt.Sprintf("DROP TABLE IF EXISTS %s", populationTable))
	
	// Network tables
	nodesTable := fmt.Sprintf("network_nodes_%d", processID)
	linksTable := fmt.Sprintf("network_links_%d", processID)
	_ = a.db.execTableQuery(fmt.Sprintf("DROP TABLE IF EXISTS %s", nodesTable))
	_ = a.db.execTableQuery(fmt.Sprintf("DROP TABLE IF EXISTS %s", linksTable))
	
	// PT tables
	ptPrefix := fmt.Sprintf("pt_data_%d", processID)
	_ = a.db.execTableQuery(fmt.Sprintf("DROP TABLE IF EXISTS %s_stops", ptPrefix))
	_ = a.db.execTableQuery(fmt.Sprintf("DROP TABLE IF EXISTS %s_lines", ptPrefix))
	_ = a.db.execTableQuery(fmt.Sprintf("DROP TABLE IF EXISTS %s_routes", ptPrefix))
	_ = a.db.execTableQuery(fmt.Sprintf("DROP TABLE IF EXISTS %s_route_stops", ptPrefix))
	_ = a.db.execTableQuery(fmt.Sprintf("DROP TABLE IF EXISTS %s_departures", ptPrefix))
	
	// Delete telemetry record
	_, _ = a.db.execQuery("DELETE FROM process_telemetry WHERE process_id = ?", fmt.Sprintf("failed to delete telemetry for process %d", processID), processID)
	
	// Finally, delete the process record
	if _, err := a.db.execQuery(deleteProcessQuery, fmt.Sprintf(deleteProcessError, processID), processID); err != nil {
		return err
	}
	return nil
}

// Database operations for processes

var ErrProcessNotFound = fmt.Errorf("process not found")





// scanProcesses is a helper function to scan rows into a slice of Process structs.
func scanProcesses(rows *sql.Rows) ([]Process, error) {
	var processes []Process
	for rows.Next() {
		var p Process
		var timestampStr string
		err := rows.Scan(&p.ProcessID, &p.FilePath, &p.Status, &timestampStr)
		if err != nil {
			return nil, fmt.Errorf("failed to scan process row: %w", err)
		}
		
		// Parse the timestamp
		p.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestampStr)
		if err != nil {
			// If parsing fails, use current time as fallback
			p.Timestamp = time.Now()
		}
		
		processes = append(processes, p)
	}
	return processes, nil
}


// Telemetry operations

// GetProcessTelemetry retrieves telemetry data for a specific process
func (a *App) GetProcessTelemetry(processID int) (*ProcessTelemetry, error) {
	var telemetry ProcessTelemetry
	var lastUpdatedStr string
	err := a.db.queryRow(selectTelemetryQuery, fmt.Sprintf(selectTelemetryError, processID), processID).Scan(
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

