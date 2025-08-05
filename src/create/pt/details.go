// Step 8: Stop Details Collection
// This file gathers detailed information for unique stops identified in previous steps.
// It reads the stops.txt file to collect stop names, latitudes, and longitudes.
package pt

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (pto *PTOrchestrator) collectStopDetails() error {
	pto.logMessage("Starting Stop Details Collection")

	// State integration available for future TUI implementation
	if pto.tuiEnabled {
		// Future: set_state("PT_7")
	}

	// Read the trip mappings from Step 6 to get list of trip IDs
	tripMappingsPath := filepath.Join(pto.config.TempDir, TripsDir, TripMappingsFile)
	tripMappingsData, err := os.ReadFile(tripMappingsPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read trip mappings file: %v", err))
		return NewFileParsingError(TripMappingsFile, err.Error())
	}

	var tripMappings []TripMapping
	if err := json.Unmarshal(tripMappingsData, &tripMappings); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to parse trip mappings JSON: %v", err))
		return NewFileParsingError(TripMappingsFile, err.Error())
	}

	// Read individual stop sequence files created by Step 7
	stopSeqDir := filepath.Join(pto.config.TempDir, StopSequencesDir)
	var allStopSequences []StopSequence

	for _, tripMapping := range tripMappings {
		tripFileName := fmt.Sprintf("%s.json", tripMapping.TripID)
		tripFilePath := filepath.Join(stopSeqDir, tripFileName)

		tripFileData, err := os.ReadFile(tripFilePath)
		if err != nil {
			pto.logMessage(fmt.Sprintf("Failed to read stop sequence file for trip %s: %v", tripMapping.TripID, err))
			continue // Skip missing files but continue processing
		}

		var stopSequence StopSequence
		if err := json.Unmarshal(tripFileData, &stopSequence); err != nil {
			pto.logMessage(fmt.Sprintf("Failed to parse stop sequence JSON for trip %s: %v", tripMapping.TripID, err))
			continue // Skip problematic files but continue processing
		}

		allStopSequences = append(allStopSequences, stopSequence)
	}

	// Collect unique stop IDs from all sequences
	uniqueStopIDs := make(map[string]bool)
	for _, sequence := range allStopSequences {
		for _, stop := range sequence.Stops {
			uniqueStopIDs[stop.StopID] = true
		}
	}

	pto.logMessage(fmt.Sprintf("Collected %d unique stop IDs from %d trip sequences", len(uniqueStopIDs), len(allStopSequences)))

	if len(uniqueStopIDs) == 0 {
		pto.logMessage("No stop IDs found in stop sequences")
		return NewInvalidServiceError("no stop IDs available")
	}

	// Read stops.txt to collect stop details
	stopsPath := filepath.Join(pto.config.GTFSDirectory, StopsFile)
	stopsFile, err := os.Open(stopsPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to open stops.txt: %v", err))
		return NewFileParsingError(StopsFile, err.Error())
	}
	defer stopsFile.Close()

	stopsReader := csv.NewReader(stopsFile)

	// Read stops header
	stopsHeaders, err := stopsReader.Read()
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read stops.txt headers: %v", err))
		return NewFileParsingError(StopsFile, "invalid header row")
	}

	// Create stops header index map
	stopsHeaderIndex := make(map[string]int)
	for i, header := range stopsHeaders {
		// Strip UTF-8 BOM from header if present (fixes STM GTFS data)
		cleanHeader := strings.TrimPrefix(header, "\uFEFF")
		stopsHeaderIndex[cleanHeader] = i
	}

	// Verify required columns exist in stops.txt
	requiredStopsColumns := []string{"stop_id", "stop_name", "stop_lat", "stop_lon"}
	for _, col := range requiredStopsColumns {
		if _, exists := stopsHeaderIndex[col]; !exists {
			pto.logMessage(fmt.Sprintf("Missing required column in stops.txt: %s", col))
			return NewFileParsingError(StopsFile, fmt.Sprintf("missing column: %s", col))
		}
	}

	var stopDetails []StopDetails
	stopCounter := 0

	// Read stops data
	for {
		record, err := stopsReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			pto.logMessage(fmt.Sprintf("Error reading stops.txt row: %v", err))
			continue
		}

		stopID := record[stopsHeaderIndex["stop_id"]]

		// Only collect details for stops we identified in the sequences
		if uniqueStopIDs[stopID] {
			stopName := record[stopsHeaderIndex["stop_name"]]
			stopLatStr := record[stopsHeaderIndex["stop_lat"]]
			stopLonStr := record[stopsHeaderIndex["stop_lon"]]

			// Parse coordinates as floats
			var stopLat, stopLon float64
			if stopLatStr != "" {
				if parsedLat, parseErr := strconv.ParseFloat(stopLatStr, 64); parseErr == nil {
					stopLat = parsedLat
				} else {
					pto.logMessage(fmt.Sprintf("Warning: Invalid stop_lat for stop %s: %s", stopID, stopLatStr))
				}
			}
			if stopLonStr != "" {
				if parsedLon, parseErr := strconv.ParseFloat(stopLonStr, 64); parseErr == nil {
					stopLon = parsedLon
				} else {
					pto.logMessage(fmt.Sprintf("Warning: Invalid stop_lon for stop %s: %s", stopID, stopLonStr))
				}
			}

			details := StopDetails{
				StopID:   stopID,
				StopLat:  stopLat,
				StopLon:  stopLon,
				StopName: stopName,
			}

			stopDetails = append(stopDetails, details)
			stopCounter++

			// Update display counter for future TUI integration
			pto.stats.StopCounter = stopCounter

			// Update live display every 50 stops
			if stopCounter%50 == 0 && pto.tui != nil {
				pto.tui.UpdateCounter("stops", stopCounter)
			}

			// Log only every 100th stop to reduce I/O overhead
			if stopCounter%100 == 0 {
				pto.logMessage(fmt.Sprintf("Collected %d stop details (last: %s - %s)", stopCounter, stopID, stopName))
			}

			// Check for quit request every 500 stops to ensure responsiveness
			if stopCounter%500 == 0 && pto.checkQuitRequested() {
				pto.logMessage("Stop details collection aborted by user")
				return nil
			}
		}
	}

	if len(stopDetails) == 0 {
		pto.logMessage("No stop details found for stop sequences")
		return NewInvalidServiceError("no stop details found")
	}

	// Create output directory for stop details (in trips directory per documentation)
	stopDetailsDir := filepath.Join(pto.config.TempDir, TripsDir)
	if err := os.MkdirAll(stopDetailsDir, DirPermissions); err != nil {
		return NewDirectoryError(stopDetailsDir, "create")
	}

	// Create output file in trips directory: /temp/pt/03_trips/stop_details.json
	outputPath := filepath.Join(stopDetailsDir, StopDetailsFile)
	jsonData, err := json.MarshalIndent(stopDetails, "", "  ")
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to marshal stop details to JSON: %v", err))
		return NewFileParsingError(StopDetailsFile, err.Error())
	}

	if err := os.WriteFile(outputPath, jsonData, FilePermissions); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to write stop details file: %v", err))
		return NewFileParsingError(StopDetailsFile, err.Error())
	}

	// Final display update
	if pto.tui != nil {
		pto.tui.UpdateCounter("stops", stopCounter)
	}

	pto.logMessage(fmt.Sprintf("Stop details collection completed successfully. Collected details for %d stops", stopCounter))
	return nil
}
