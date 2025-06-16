package processing

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	
)

func (pto *PTOrchestrator) collectTrips() error {
	pto.logMessage("Starting Step 5: Trip Collection")
	
	// State integration available for future TUI implementation
	if pto.tuiEnabled {
		// Future: set_state("PT_5")
	}
	
	// Read the selected services from Step 4 (ALL services format)
	selectedServicesPath := filepath.Join(pto.config.TempDir, ServicesDir, SelectedServicesFile)
	selectedServicesData, err := os.ReadFile(selectedServicesPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read selected services file: %v", err))
		return NewFileParsingError(SelectedServicesFile, err.Error())
	}
	
	var selectedServices []SelectedService
	if err := json.Unmarshal(selectedServicesData, &selectedServices); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to parse selected services JSON: %v", err))
		return NewFileParsingError(SelectedServicesFile, err.Error())
	}
	
	// Create a map of selected service IDs for quick lookup
	selectedServiceIDs := make(map[string]bool)
	for _, service := range selectedServices {
		if service.Selected {
			selectedServiceIDs[service.ServiceID] = true
			pto.logMessage(fmt.Sprintf("Processing trips for service: %s (%d routes)", service.ServiceID, len(service.Routes)))
		}
	}
	
	if len(selectedServiceIDs) == 0 {
		pto.logMessage("No services selected for trip collection")
		return NewInvalidServiceError("no services selected")
	}
	
	pto.logMessage(fmt.Sprintf("Processing trips for %d selected services (complete transit system)", len(selectedServiceIDs)))
	
	// Read trips.txt to collect all trips for selected services
	tripsPath := filepath.Join(pto.config.GTFSDirectory, TripsFile)
	tripsFile, err := os.Open(tripsPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to open trips.txt: %v", err))
		return NewFileParsingError(TripsFile, err.Error())
	}
	defer tripsFile.Close()
	
	tripsReader := csv.NewReader(tripsFile)
	
	// Read trips header
	tripsHeaders, err := tripsReader.Read()
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read trips.txt headers: %v", err))
		return NewFileParsingError(TripsFile, "invalid header row")
	}
	
	// Create trips header index map
	tripsHeaderIndex := make(map[string]int)
	for i, header := range tripsHeaders {
		tripsHeaderIndex[header] = i
	}
	
	// Verify required columns exist in trips.txt
	requiredTripsColumns := []string{"trip_id", "route_id", "service_id"}
	for _, col := range requiredTripsColumns {
		if _, exists := tripsHeaderIndex[col]; !exists {
			pto.logMessage(fmt.Sprintf("Missing required column in trips.txt: %s", col))
			return NewFileParsingError(TripsFile, fmt.Sprintf("missing column: %s", col))
		}
	}
	
	var tripMappings []TripMapping
	tripCounter := 0
	totalRows := 0
	
	// First pass: count total rows for accurate progress
	tripsFile.Seek(0, 0) // Reset to beginning
	tempReader := csv.NewReader(tripsFile)
	tempReader.Read() // Skip header
	for {
		if _, err := tempReader.Read(); err != nil {
			break
		}
		totalRows++
	}
	
	// Reset file for actual processing
	tripsFile.Seek(0, 0)
	tripsReader = csv.NewReader(tripsFile)
	tripsReader.Read() // Skip header again
	
	processedRows := 0
	
	// Read trips data
	for {
		record, err := tripsReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			pto.logMessage(fmt.Sprintf("Error reading %s row: %v", TripsFile, err))
			processedRows++
			continue
		}
		
		processedRows++
		serviceID := record[tripsHeaderIndex["service_id"]]
		tripID := record[tripsHeaderIndex["trip_id"]]
		routeID := record[tripsHeaderIndex["route_id"]]
		
		// Only collect trips for selected services
		if selectedServiceIDs[serviceID] {
			tripMapping := TripMapping{
				TripID:  tripID,
				RouteID: routeID,
			}
			tripMappings = append(tripMappings, tripMapping)
			tripCounter++
			
			// Update display counter and live updates
			pto.stats.TripCounter = tripCounter
			
			// Log only every 100th trip to reduce I/O overhead
			if tripCounter%100 == 0 {
				pto.logMessage(fmt.Sprintf("Collected %d trips (last: %s for route %s)", tripCounter, tripID, routeID))
			}
		}
		
		// Update display with accurate progress based on file processing
		if pto.tuiEnabled && pto.displayInstance != nil {
			_ = pto.displayInstance.SetLiveUpdate(4, "trips", fmt.Sprintf("%d", tripCounter))
			if processedRows%100 == 0 { // Update progress every 100 rows
				progress := int(float64(processedRows) / float64(totalRows) * 100)
				if progress > 100 { progress = 100 }
				_ = pto.displayInstance.SetLiveUpdate(4, "progress", fmt.Sprintf("%d%%", progress))
			}
		}
		
		// Check for quit request every 500 rows to ensure responsiveness
		if processedRows%500 == 0 && pto.checkQuitRequested() {
			pto.logMessage("Trip collection aborted by user")
			return nil
		}
	}
	
	if len(tripMappings) == 0 {
		pto.logMessage("No trips found for selected services")
		return NewInvalidServiceError("no trips found for selected services")
	}
	
	// Create output directory for trip data
	tripDir := filepath.Join(pto.config.TempDir, TripsDir)
	if err := os.MkdirAll(tripDir, DirPermissions); err != nil {
		return NewDirectoryError(tripDir, "create")
	}
	
	// Create output file
	outputPath := filepath.Join(tripDir, TripMappingsFile)
	jsonData, err := json.MarshalIndent(tripMappings, "", "  ")
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to marshal trip mappings to JSON: %v", err))
		return NewFileParsingError(TripMappingsFile, err.Error())
	}
	
	if err := os.WriteFile(outputPath, jsonData, FilePermissions); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to write trip mappings file: %v", err))
		return NewFileParsingError(TripMappingsFile, err.Error())
	}
	
	// Final live update with completion status
	if pto.tuiEnabled && pto.displayInstance != nil {
		_ = pto.displayInstance.SetLiveUpdate(4, "trips", fmt.Sprintf("%d", tripCounter))
		_ = pto.displayInstance.SetLiveUpdate(4, "progress", "100%")
	}
	
	pto.logMessage(fmt.Sprintf("Trip collection completed successfully. Collected %d trips for %d services (complete transit system)", 
		tripCounter, len(selectedServiceIDs)))
	return nil
}