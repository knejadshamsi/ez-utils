// Step 6: Stop Sequence Collection
// This file collects stop sequences for each trip from the stop_times.txt file.
// It groups stops by trip ID and stores them as individual JSON files for subsequent processing.
package pt

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	
)

func (pto *PTOrchestrator) collectStopSequences() error {
	pto.logMessage("Starting Stop Sequence Collection")
	
	// State integration available for future TUI implementation
	if pto.tuiEnabled {
		// Future: set_state("PT_6")
	}
	
	// Read the trip mappings from Step 5
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
	
	// Create a map of trip IDs for quick lookup
	tripIDs := make(map[string]bool)
	for _, trip := range tripMappings {
		tripIDs[trip.TripID] = true
	}
	
	if len(tripIDs) == 0 {
		pto.logMessage("No trips found for stop sequence collection")
		return NewInvalidServiceError("no trips available")
	}
	
	// Read stop_times.txt to collect stop sequences for each trip
	stopTimesPath := filepath.Join(pto.config.GTFSDirectory, StopTimesFile)
	stopTimesFile, err := os.Open(stopTimesPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to open stop_times.txt: %v", err))
		return NewFileParsingError(StopTimesFile, err.Error())
	}
	defer stopTimesFile.Close()
	
	stopTimesReader := csv.NewReader(stopTimesFile)
	
	// Read stop_times header
	stopTimesHeaders, err := stopTimesReader.Read()
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read stop_times.txt headers: %v", err))
		return NewFileParsingError(StopTimesFile, "invalid header row")
	}
	
	// Create stop_times header index map
	stopTimesHeaderIndex := make(map[string]int)
	for i, header := range stopTimesHeaders {
		stopTimesHeaderIndex[header] = i
	}
	
	// Verify required columns exist in stop_times.txt
	requiredStopTimesColumns := []string{"trip_id", "arrival_time", "departure_time", "stop_id", "stop_sequence"}
	for _, col := range requiredStopTimesColumns {
		if _, exists := stopTimesHeaderIndex[col]; !exists {
			pto.logMessage(fmt.Sprintf("Missing required column in stop_times.txt: %s", col))
			return NewFileParsingError(StopTimesFile, fmt.Sprintf("missing column: %s", col))
		}
	}
	
	// Group stops by trip ID
	tripStops := make(map[string][]Stop)
	sequenceCounter := 0
	
	// Read stop_times data
	for {
		record, err := stopTimesReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			pto.logMessage(fmt.Sprintf("Error reading stop_times.txt row: %v", err))
			continue
		}
		
		tripID := record[stopTimesHeaderIndex["trip_id"]]
		
		// Only process stops for trips we collected in Step 5
		if tripIDs[tripID] {
			stopID := record[stopTimesHeaderIndex["stop_id"]]
			arrivalTime := record[stopTimesHeaderIndex["arrival_time"]]
			departureTime := record[stopTimesHeaderIndex["departure_time"]]
			stopSequenceStr := record[stopTimesHeaderIndex["stop_sequence"]]
			
			// Parse stop sequence as integer
			stopSequence := 0
			if stopSequenceStr != "" {
				if parsedSeq, parseErr := strconv.Atoi(stopSequenceStr); parseErr == nil {
					stopSequence = parsedSeq
				} else {
					pto.logMessage(fmt.Sprintf("Warning: Invalid stop_sequence for trip %s, stop %s: %s", tripID, stopID, stopSequenceStr))
				}
			}
			
			stop := Stop{
				StopID:        stopID,
				ArrivalTime:   arrivalTime,
				DepartureTime: departureTime,
				StopSequence:  stopSequence,
			}
			
			tripStops[tripID] = append(tripStops[tripID], stop)
			sequenceCounter++
			
			// Update display counter for future TUI integration
			pto.stats.SequenceCounter = sequenceCounter
			
			// Log only every 1000th stop sequence to reduce I/O overhead
			if sequenceCounter%1000 == 0 {
				pto.logMessage(fmt.Sprintf("Collected %d stop sequences (last: trip %s, stop %s)", sequenceCounter, tripID, stopID))
			}
			
			// Check for quit request every 2000 sequences to ensure responsiveness
			if sequenceCounter%2000 == 0 && pto.checkQuitRequested() {
				pto.logMessage("Stop sequence collection aborted by user")
				return nil
			}
		}
	}
	
	if len(tripStops) == 0 {
		pto.logMessage("No stop sequences found for collected trips")
		return NewInvalidServiceError("no stop sequences found")
	}
	
	// Create output directory for stop sequences
	stopSeqDir := filepath.Join(pto.config.TempDir, StopSequencesDir)
	if err := os.MkdirAll(stopSeqDir, DirPermissions); err != nil {
		return NewDirectoryError(stopSeqDir, "create")
	}
	
	// Create individual JSON files for each trip as per documentation
	tripFileCount := 0
	for tripID, stops := range tripStops {
		stopSequence := StopSequence{
			TripID: tripID,
			Stops:  stops,
		}
		
		// Create individual file for this trip: /temp/pt/03_trips/stop_sequences/[trip_id].json
		tripFileName := fmt.Sprintf("%s.json", tripID)
		tripFilePath := filepath.Join(stopSeqDir, tripFileName)
		
		jsonData, err := json.MarshalIndent(stopSequence, "", "  ")
		if err != nil {
			pto.logMessage(fmt.Sprintf("Failed to marshal stop sequence for trip %s to JSON: %v", tripID, err))
			return NewFileParsingError(tripFileName, err.Error())
		}
		
		if err := os.WriteFile(tripFilePath, jsonData, FilePermissions); err != nil {
			pto.logMessage(fmt.Sprintf("Failed to write stop sequence file for trip %s: %v", tripID, err))
			return NewFileParsingError(tripFileName, err.Error())
		}
		
		tripFileCount++
		pto.logMessage(fmt.Sprintf("Created stop sequence file for trip %s with %d stops", tripID, len(stops)))
		
		// Check for quit request every 50 files to ensure responsiveness
		if tripFileCount%50 == 0 && pto.checkQuitRequested() {
			pto.logMessage("Stop sequence collection aborted by user")
			return nil
		}
	}
	
	pto.logMessage(fmt.Sprintf("Stop sequence collection completed successfully. Created %d individual trip files with %d total stop sequences", tripFileCount, sequenceCounter))
	return nil
}