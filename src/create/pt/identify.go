// Step 2: Service Pattern Identification
// This file focuses on identifying service patterns within the GTFS calendar data.
// It reads the calendar.txt file to determine which service IDs operate on the specified service day.
package pt

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	
)

func (pto *PTOrchestrator) identifyServicePatterns() error {
	pto.logMessage("Starting Service Pattern Identification")
	
	// Read calendar.txt file
	calendarPath := filepath.Join(pto.config.GTFSDirectory, CalendarFile)
	file, err := os.Open(calendarPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to open %s: %v", CalendarFile, err))
		return NewFileParsingError(CalendarFile, err.Error())
	}
	defer file.Close()
	
	// Parse CSV
	reader := csv.NewReader(file)
	
	// Read header row
	headers, err := reader.Read()
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read %s headers: %v", CalendarFile, err))
		return NewFileParsingError(CalendarFile, "invalid header row")
	}
	
	// Create header index map
	headerIndex := make(map[string]int)
	for i, header := range headers {
		headerIndex[header] = i
	}
	
	// Verify required columns exist
	requiredColumns := []string{"service_id", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	for _, col := range requiredColumns {
		if _, exists := headerIndex[col]; !exists {
			pto.logMessage(fmt.Sprintf("Missing required column in %s: %s", CalendarFile, col))
			return NewFileParsingError(CalendarFile, fmt.Sprintf("missing column: %s", col))
		}
	}
	
	var servicePatterns []ServicePattern
	serviceCounter := 0
	totalRows := 0
	
	// First pass: count total rows for accurate progress
	file.Seek(0, 0) // Reset to beginning
	tempReader := csv.NewReader(file)
	tempReader.Read() // Skip header
	for {
		if _, err := tempReader.Read(); err != nil {
			break
		}
		totalRows++
	}
	
	// Reset file for actual processing
	file.Seek(0, 0)
	reader = csv.NewReader(file)
	reader.Read() // Skip header again
	
	processedRows := 0
	
	// Read data rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			pto.logMessage(fmt.Sprintf("Error reading %s row: %v", CalendarFile, err))
			processedRows++
			continue // Skip problematic rows but continue processing
		}
		
		processedRows++
		
		// Parse service pattern
		pattern := ServicePattern{
			ServiceID: record[headerIndex["service_id"]],
			Monday:    record[headerIndex["monday"]] == "1",
			Tuesday:   record[headerIndex["tuesday"]] == "1",
			Wednesday: record[headerIndex["wednesday"]] == "1",
			Thursday:  record[headerIndex["thursday"]] == "1",
			Friday:    record[headerIndex["friday"]] == "1",
			Saturday:  record[headerIndex["saturday"]] == "1",
			Sunday:    record[headerIndex["sunday"]] == "1",
		}
		
		// Check if pattern matches the specified day type
		if pto.matchesServiceDay(pattern) {
			servicePatterns = append(servicePatterns, pattern)
			serviceCounter++
			
			// Update display counter and live updates
			pto.stats.ServiceCounter = serviceCounter
			
			// Log only every 10th matching service to reduce I/O overhead
			if serviceCounter%10 == 0 {
				pto.logMessage(fmt.Sprintf("Found %d matching service patterns (last: %s)", serviceCounter, pattern.ServiceID))
			}
		}
		
		// Update display with accurate progress based on file processing
		if pto.tuiEnabled && pto.displayInstance != nil {
			_ = pto.displayInstance.SetLiveUpdate(1, "services", fmt.Sprintf("%d", serviceCounter))
			if processedRows%50 == 0 { // Update progress every 50 rows
				progress := int(float64(processedRows) / float64(totalRows) * 100)
				if progress > 100 { progress = 100 }
				_ = pto.displayInstance.SetLiveUpdate(1, "progress", fmt.Sprintf("%d%%", progress))
			}
		}
	}
	
	if len(servicePatterns) == 0 {
		pto.logMessage("No matching service patterns found")
		return NewInvalidServiceError(pto.getServiceDayString())
	}
	
	// Create output directory and file
	outputDir := filepath.Join(pto.config.TempDir, ServicesDir)
	if err := CreateDirectoryIfNotExists(outputDir); err != nil {
		return NewDirectoryError(outputDir, "create")
	}
	outputPath := filepath.Join(outputDir, ServicePatternsFile)
	jsonData, err := json.MarshalIndent(servicePatterns, "", "  ")
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to marshal service patterns to JSON: %v", err))
		return NewFileParsingError(ServicePatternsFile, err.Error())
	}
	
	if err := os.WriteFile(outputPath, jsonData, FilePermissions); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to write service patterns file: %v", err))
		return NewFileParsingError(ServicePatternsFile, err.Error())
	}
	
	// Final live update with completion status
	if pto.tuiEnabled && pto.displayInstance != nil {
		_ = pto.displayInstance.SetLiveUpdate(1, "services", fmt.Sprintf("%d", serviceCounter))
		_ = pto.displayInstance.SetLiveUpdate(1, "progress", "100%")
	}
	
	pto.logMessage(fmt.Sprintf("Service pattern identification completed successfully. Found %d matching patterns", serviceCounter))
	return nil
}

// matchesServiceDay checks if a service pattern matches the configured service day
func (pto *PTOrchestrator) matchesServiceDay(pattern ServicePattern) bool {
	switch pto.config.ServiceDay {
	case ServiceDayWeekday:
		return pattern.Monday || pattern.Tuesday || pattern.Wednesday || pattern.Thursday || pattern.Friday
	case ServiceDayWeekend:
		return pattern.Saturday || pattern.Sunday
	case ServiceDayMonday:
		return pattern.Monday
	case ServiceDayTuesday:
		return pattern.Tuesday
	case ServiceDayWednesday:
		return pattern.Wednesday
	case ServiceDayThursday:
		return pattern.Thursday
	case ServiceDayFriday:
		return pattern.Friday
	case ServiceDaySaturday:
		return pattern.Saturday
	case ServiceDaySunday:
		return pattern.Sunday
	default:
		return false
	}
}

// getServiceDayString returns a string representation of the service day for error messages
func (pto *PTOrchestrator) getServiceDayString() string {
	switch pto.config.ServiceDay {
	case ServiceDayWeekday:
		return "weekday"
	case ServiceDayWeekend:
		return "weekend"
	case ServiceDayMonday:
		return "monday"
	case ServiceDayTuesday:
		return "tuesday"
	case ServiceDayWednesday:
		return "wednesday"
	case ServiceDayThursday:
		return "thursday"
	case ServiceDayFriday:
		return "friday"
	case ServiceDaySaturday:
		return "saturday"
	case ServiceDaySunday:
		return "sunday"
	default:
		return "unknown"
	}
}