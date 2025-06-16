package processing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (pto *PTOrchestrator) generateXML() error {
	pto.logMessage("Starting Step 8: Generate MATSim Transit Schedule XML")
	
	// Initialize live updates for XML generation
	if pto.tuiEnabled && pto.displayInstance != nil {
		_ = pto.displayInstance.SetLiveUpdate(7, "progress", "0%")
		_ = pto.displayInstance.SetLiveUpdate(7, "size", "0 KB")
	}
	
	// 1. Read trip mappings to get list of trips and their routes
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
	
	// 2. Read stop details from trips directory
	stopDetailsPath := filepath.Join(pto.config.TempDir, TripsDir, StopDetailsFile)
	stopDetailsData, err := os.ReadFile(stopDetailsPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read stop details file: %v", err))
		return NewFileParsingError(StopDetailsFile, err.Error())
	}
	
	var stopDetails []StopDetails
	if err := json.Unmarshal(stopDetailsData, &stopDetails); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to parse stop details JSON: %v", err))
		return NewFileParsingError(StopDetailsFile, err.Error())
	}
	
	// 3. Read individual trip stop sequence files
	stopSeqDir := filepath.Join(pto.config.TempDir, StopSequencesDir)
	allTripSequences := make(map[string]StopSequence)
	
	for _, tripMapping := range tripMappings {
		tripFileName := fmt.Sprintf("%s.json", tripMapping.TripID)
		tripFilePath := filepath.Join(stopSeqDir, tripFileName)
		
		tripFileData, err := os.ReadFile(tripFilePath)
		if err != nil {
			pto.logMessage(fmt.Sprintf("Failed to read stop sequence file for trip %s: %v", tripMapping.TripID, err))
			continue
		}
		
		var stopSequence StopSequence
		if err := json.Unmarshal(tripFileData, &stopSequence); err != nil {
			pto.logMessage(fmt.Sprintf("Failed to parse stop sequence JSON for trip %s: %v", tripMapping.TripID, err))
			continue
		}
		
		allTripSequences[tripMapping.TripID] = stopSequence
	}
	
	// Create lookup maps
	stopDetailsMap := make(map[string]StopDetails)
	for _, details := range stopDetails {
		stopDetailsMap[details.StopID] = details
	}
	
	tripToRouteMap := make(map[string]string)
	for _, mapping := range tripMappings {
		tripToRouteMap[mapping.TripID] = mapping.RouteID
	}
	
	// Group trips by route_id for MATSim transitLines structure
	routeToTrips := make(map[string][]string)
	for _, mapping := range tripMappings {
		routeToTrips[mapping.RouteID] = append(routeToTrips[mapping.RouteID], mapping.TripID)
	}
	
	// Generate MATSim XML content
	var xmlBuffer bytes.Buffer
	
	// Write XML header and root element
	xmlBuffer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
`)
	xmlBuffer.WriteString(`<transitSchedule>
`)
	
	// Generate transitStops section
	xmlBuffer.WriteString("\t<transitStops>\n")
	pto.logMessage("Generating transitStops section...")
	
	transitStopCounter := 0
	processedStops := make(map[string]bool)
	
	// Generate stopFacility elements for all unique stops
	for _, details := range stopDetails {
		if !processedStops[details.StopID] {
			xmlBuffer.WriteString(fmt.Sprintf("\t\t<stopFacility id=\"%s\" x=\"%.6f\" y=\"%.6f\" name=\"%s\"/>\n",
				details.StopID, details.StopLon, details.StopLat, escapeXML(details.StopName)))
			processedStops[details.StopID] = true
			transitStopCounter++
			
			// Update display every 50 stops
			if transitStopCounter%50 == 0 && pto.tuiEnabled && pto.displayInstance != nil {
				_ = pto.displayInstance.SetLiveUpdate(7, "progress", fmt.Sprintf("Stop %d", transitStopCounter))
			}
		}
	}
	
	xmlBuffer.WriteString("\t</transitStops>\n")
	pto.logMessage(fmt.Sprintf("Generated %d stopFacility elements", transitStopCounter))
	
	// Generate transitLines section
	xmlBuffer.WriteString("\t<transitLines>\n")
	pto.logMessage("Generating transitLines section...")
	
	routeCounter := 0
	totalRoutes := len(routeToTrips)
	
	// Process each route (group trips by route_id)
	for routeID, tripIDs := range routeToTrips {
		xmlBuffer.WriteString(fmt.Sprintf("\t\t<transitLine id=\"%s\">\n", routeID))
		
		// Process each trip in this route as a transitRoute
		for tripIndex, tripID := range tripIDs {
			tripSequence, exists := allTripSequences[tripID]
			if !exists {
				pto.logMessage(fmt.Sprintf("Warning: Stop sequence not found for trip %s", tripID))
				continue
			}
			
			xmlBuffer.WriteString(fmt.Sprintf("\t\t\t<transitRoute id=\"%s\">\n", tripID))
			
			// Generate routeProfile with offset calculations
			xmlBuffer.WriteString("\t\t\t\t<routeProfile>\n")
			
			// Sort stops by sequence number
			sort.Slice(tripSequence.Stops, func(i, j int) bool {
				return tripSequence.Stops[i].StopSequence < tripSequence.Stops[j].StopSequence
			})
			
			// Calculate offsets from first stop departure time
			var firstDepartureTime time.Time
			if len(tripSequence.Stops) > 0 {
				firstDepartureTime, _ = parseTime(tripSequence.Stops[0].DepartureTime)
			}
			
			for _, stop := range tripSequence.Stops {
				arrivalTime, _ := parseTime(stop.ArrivalTime)
				departureTime, _ := parseTime(stop.DepartureTime)
				
				arrivalOffset := formatDuration(arrivalTime.Sub(firstDepartureTime))
				departureOffset := formatDuration(departureTime.Sub(firstDepartureTime))
				
				xmlBuffer.WriteString(fmt.Sprintf("\t\t\t\t\t<stop refId=\"%s\" arrivalOffset=\"%s\" departureOffset=\"%s\"/>\n",
					stop.StopID, arrivalOffset, departureOffset))
			}
			
			xmlBuffer.WriteString("\t\t\t\t</routeProfile>\n")
			
			// Generate departures section
			xmlBuffer.WriteString("\t\t\t\t<departures>\n")
			if len(tripSequence.Stops) > 0 {
				xmlBuffer.WriteString(fmt.Sprintf("\t\t\t\t\t<departure id=\"%d\" departureTime=\"%s\"/>\n",
					tripIndex+1, tripSequence.Stops[0].DepartureTime))
			}
			xmlBuffer.WriteString("\t\t\t\t</departures>\n")
			
			xmlBuffer.WriteString("\t\t\t</transitRoute>\n")
		}
		
		xmlBuffer.WriteString("\t\t</transitLine>\n")
		routeCounter++
		
		// Update progress and check for quit
		if routeCounter%10 == 0 && pto.tuiEnabled && pto.displayInstance != nil {
			progress := int(float64(routeCounter) / float64(totalRoutes) * 100)
			_ = pto.displayInstance.SetLiveUpdate(7, "progress", fmt.Sprintf("%d%%", progress))
			currentSize := float64(xmlBuffer.Len()) / 1024
			_ = pto.displayInstance.SetLiveUpdate(7, "size", fmt.Sprintf("%.1f KB", currentSize))
		}
		
		if routeCounter%20 == 0 && pto.checkQuitRequested() {
			pto.logMessage("XML generation aborted by user")
			return nil
		}
		
		if routeCounter%50 == 0 {
			pto.logMessage(fmt.Sprintf("Generated XML for %d/%d routes", routeCounter, totalRoutes))
		}
	}
	
	xmlBuffer.WriteString("\t</transitLines>\n")
	xmlBuffer.WriteString("</transitSchedule>")
	
	// Write XML to output file
	outputPath := filepath.Join(pto.config.OutputDir, TransitScheduleFile)
	if err := os.WriteFile(outputPath, xmlBuffer.Bytes(), FilePermissions); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to write XML file: %v", err))
		return NewFileParsingError(TransitScheduleFile, err.Error())
	}
	
	// Final live update with completion status
	if pto.tuiEnabled && pto.displayInstance != nil {
		_ = pto.displayInstance.SetLiveUpdate(7, "progress", "100%")
		finalSize := float64(xmlBuffer.Len()) / 1024 // Size in KB
		_ = pto.displayInstance.SetLiveUpdate(7, "size", fmt.Sprintf("%.1f KB", finalSize))
	}
	
	pto.logMessage(fmt.Sprintf("MATSim XML generation completed successfully. Created %s with %d routes, %d trips and %d transit stops", outputPath, routeCounter, len(allTripSequences), transitStopCounter))
	return nil
}

// escapeXML escapes special XML characters in strings
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// parseTime parses GTFS time format (HH:MM:SS, can exceed 24 hours)
func parseTime(timeStr string) (time.Time, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid time format: %s", timeStr)
	}
	
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, err
	}
	
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, err
	}
	
	seconds, err := strconv.Atoi(parts[2])
	if err != nil {
		return time.Time{}, err
	}
	
	// Handle times beyond 24 hours (e.g., 25:30:00 for next day)
	baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	totalSeconds := hours*3600 + minutes*60 + seconds
	return baseTime.Add(time.Duration(totalSeconds) * time.Second), nil
}

// formatDuration formats a duration as HH:MM:SS for MATSim offsets
func formatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}