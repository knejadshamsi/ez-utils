// Step 1: File Validation
// This file contains the validation logic for GTFS files. It ensures that all required
// GTFS files exist, are not empty, and adhere to the expected CSV structure and data integrity rules.
package pt

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// validateFiles performs comprehensive validation of all required GTFS files
func (pto *PTOrchestrator) validateFiles() error {

	// Required GTFS files for PT processing
	requiredFiles := []string{
		CalendarFile,   // Contains service IDs and days of operation
		TripsFile,      // Links trips to routes and services
		StopTimesFile,  // Contains trip stop sequences and timing
		StopsFile,      // Contains stop location and name data
		RoutesFile,     // Contains route type information
	}

	// Validate each file one at a time
	for _, filename := range requiredFiles {
		// Update stats for display integration
		pto.stats.ValidationFile = filename
		
		// Update progress in display with live updates
		if pto.tuiEnabled && pto.displayInstance != nil {
			_ = pto.displayInstance.SetLiveUpdate(0, "file", filename)
			_ = pto.displayInstance.SetLiveUpdate(0, "status", "validating")
		}
		
		// File validation progress handled by TUI
		
		// Check if file exists
		filePath := filepath.Join(pto.config.GTFSDirectory, filename)
		if !FileExists(filePath) {
			pto.logMessage(fmt.Sprintf("Missing required file: %s", filename))
			return NewMissingFilesError(filename)
		}
		
		// Verify file is not empty
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			pto.logMessage(fmt.Sprintf("Error reading file info for %s: %v", filename, err))
			return NewFileParsingError(filename, "unable to read file information")
		}
		
		if fileInfo.Size() == 0 {
			pto.logMessage(fmt.Sprintf("File %s is empty", filename))
			// Update display with error status
			if pto.tuiEnabled && pto.displayInstance != nil {
				_ = pto.displayInstance.SetLiveUpdate(0, "status", "error")
			}
			return NewFileParsingError(filename, "file is empty")
		}
		
		// Update display with file size
		if pto.tuiEnabled && pto.displayInstance != nil {
			_ = pto.displayInstance.SetLiveUpdate(0, "size", fmt.Sprintf("%.1f KB", float64(fileInfo.Size())/1024))
			_ = pto.displayInstance.SetLiveUpdate(0, "status", "valid")
		}
		
		// Validate CSV format and required columns
		if err := pto.validateCSVStructure(filePath, filename); err != nil {
			return err
		}
	}
	pto.logMessage("File validation completed successfully")
	return nil
}

func (pto *PTOrchestrator) validateCSVStructure(filePath, filename string) error {
	file, err := os.Open(filePath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to open %s for validation: %v", filename, err))
		return NewFileParsingError(filename, "unable to open file for validation")
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields
	
	// Read header row
	headers, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			pto.logMessage(fmt.Sprintf("File %s contains no data", filename))
			return NewFileParsingError(filename, "file contains no data")
		}
		pto.logMessage(fmt.Sprintf("Failed to read headers from %s: %v", filename, err))
		return NewFileParsingError(filename, "invalid CSV format - cannot read headers")
	}
	
	// Validate required columns for each file type
	requiredColumns := map[string][]string{
		CalendarFile:   {"service_id", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"},
		TripsFile:      {"route_id", "service_id", "trip_id"},
		StopTimesFile:  {"trip_id", "arrival_time", "departure_time", "stop_id", "stop_sequence"},
		StopsFile:      {"stop_id", "stop_name", "stop_lat", "stop_lon"},
		RoutesFile:     {"route_id", "route_short_name", "route_long_name", "route_type"},
	}
	
	required, exists := requiredColumns[filename]
	if !exists {
		pto.logMessage(fmt.Sprintf("No validation rules defined for %s", filename))
		return nil // Skip validation for unknown files
	}
	
	// Check if all required columns exist
	headerMap := make(map[string]bool)
	for _, header := range headers {
		headerMap[header] = true
	}
	
	var missingColumns []string
	for _, col := range required {
		if !headerMap[col] {
			missingColumns = append(missingColumns, col)
		}
	}
	
	if len(missingColumns) > 0 {
		pto.logMessage(fmt.Sprintf("File %s missing required columns: %v", filename, missingColumns))
		return NewFileParsingError(filename, fmt.Sprintf("missing required columns: %v", missingColumns))
	}
	
	// Try to read at least one data row to verify CSV is parseable
	dataRow, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			pto.logMessage(fmt.Sprintf("File %s contains only headers, no data rows", filename))
			return NewFileParsingError(filename, "file contains only headers, no data rows")
		}
		pto.logMessage(fmt.Sprintf("Failed to read data row from %s: %v", filename, err))
		return NewFileParsingError(filename, "invalid CSV format - cannot read data rows")
	}
	
	// Validate data integrity for specific files
	if err := pto.validateDataIntegrity(filename, headers, dataRow); err != nil {
		return err
	}
	
	pto.logMessage(fmt.Sprintf("CSV validation passed for %s (%d columns, data rows present)", filename, len(headers)))
	return nil
}

func (pto *PTOrchestrator) validateDataIntegrity(filename string, headers, dataRow []string) error {
	// Create header index map
	headerIndex := make(map[string]int)
	for i, header := range headers {
		headerIndex[header] = i
	}
	
	// Validate specific data types and formats based on file
	switch filename {
	case CalendarFile:
		// Validate service_id is not empty
		if serviceIdx, exists := headerIndex["service_id"]; exists && serviceIdx < len(dataRow) {
			if dataRow[serviceIdx] == "" {
				pto.logMessage(fmt.Sprintf("Empty service_id found in %s", filename))
				return NewFileParsingError(filename, "empty service_id not allowed")
			}
		}
		
		// Validate day fields are 0 or 1
		dayFields := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
		for _, day := range dayFields {
			if dayIdx, exists := headerIndex[day]; exists && dayIdx < len(dataRow) {
				if dataRow[dayIdx] != "0" && dataRow[dayIdx] != "1" {
					pto.logMessage(fmt.Sprintf("Invalid day value '%s' for %s in %s", dataRow[dayIdx], day, filename))
					return NewFileParsingError(filename, fmt.Sprintf("day fields must be 0 or 1, found: %s", dataRow[dayIdx]))
				}
			}
		}
		
	case StopsFile:
		// Validate coordinates are numeric
		if latIdx, exists := headerIndex["stop_lat"]; exists && latIdx < len(dataRow) {
			if _, err := strconv.ParseFloat(dataRow[latIdx], 64); err != nil {
				pto.logMessage(fmt.Sprintf("Invalid latitude '%s' in %s", dataRow[latIdx], filename))
				return NewFileParsingError(filename, fmt.Sprintf("invalid latitude format: %s", dataRow[latIdx]))
			}
		}
		if lonIdx, exists := headerIndex["stop_lon"]; exists && lonIdx < len(dataRow) {
			if _, err := strconv.ParseFloat(dataRow[lonIdx], 64); err != nil {
				pto.logMessage(fmt.Sprintf("Invalid longitude '%s' in %s", dataRow[lonIdx], filename))
				return NewFileParsingError(filename, fmt.Sprintf("invalid longitude format: %s", dataRow[lonIdx]))
			}
		}
		
		// Validate stop_id and stop_name are not empty
		if stopIdIdx, exists := headerIndex["stop_id"]; exists && stopIdIdx < len(dataRow) {
			if dataRow[stopIdIdx] == "" {
				pto.logMessage(fmt.Sprintf("Empty stop_id found in %s", filename))
				return NewFileParsingError(filename, "empty stop_id not allowed")
			}
		}
		
	case StopTimesFile:
		// Validate stop_sequence is numeric
		if seqIdx, exists := headerIndex["stop_sequence"]; exists && seqIdx < len(dataRow) {
			if _, err := strconv.Atoi(dataRow[seqIdx]); err != nil {
				pto.logMessage(fmt.Sprintf("Invalid stop_sequence '%s' in %s", dataRow[seqIdx], filename))
				return NewFileParsingError(filename, fmt.Sprintf("stop_sequence must be numeric: %s", dataRow[seqIdx]))
			}
		}
		
		// Validate required IDs are not empty
		requiredIds := []string{"trip_id", "stop_id"}
		for _, idField := range requiredIds {
			if idIdx, exists := headerIndex[idField]; exists && idIdx < len(dataRow) {
				if dataRow[idIdx] == "" {
					pto.logMessage(fmt.Sprintf("Empty %s found in %s", idField, filename))
					return NewFileParsingError(filename, fmt.Sprintf("empty %s not allowed", idField))
				}
			}
		}
		
	case TripsFile:
		// Validate required IDs are not empty
		requiredIds := []string{"trip_id", "route_id", "service_id"}
		for _, idField := range requiredIds {
			if idIdx, exists := headerIndex[idField]; exists && idIdx < len(dataRow) {
				if dataRow[idIdx] == "" {
					pto.logMessage(fmt.Sprintf("Empty %s found in %s", idField, filename))
					return NewFileParsingError(filename, fmt.Sprintf("empty %s not allowed", idField))
				}
			}
		}
		
	case RoutesFile:
		// Validate route_type is numeric
		if typeIdx, exists := headerIndex["route_type"]; exists && typeIdx < len(dataRow) {
			if _, err := strconv.Atoi(dataRow[typeIdx]); err != nil {
				pto.logMessage(fmt.Sprintf("Invalid route_type '%s' in %s", dataRow[typeIdx], filename))
				return NewFileParsingError(filename, fmt.Sprintf("route_type must be numeric: %s", dataRow[typeIdx]))
			}
		}
		
		// Validate route_id is not empty
		if routeIdIdx, exists := headerIndex["route_id"]; exists && routeIdIdx < len(dataRow) {
			if dataRow[routeIdIdx] == "" {
				pto.logMessage(fmt.Sprintf("Empty route_id found in %s", filename))
				return NewFileParsingError(filename, "empty route_id not allowed")
			}
		}
	}
	
	return nil
}