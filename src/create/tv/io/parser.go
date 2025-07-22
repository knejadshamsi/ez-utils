package io

import (
	"encoding/xml"
	"fmt"
	"os"
)

// TransitSchedule represents the root element of a MATSim transit schedule
type TransitSchedule struct {
	XMLName    xml.Name    `xml:"transitSchedule"`
	TransitLines []TransitLine `xml:"transitLine"`
}

// TransitLine represents a transit line in the schedule
type TransitLine struct {
	ID         string      `xml:"id,attr"`
	TransitRoutes []TransitRoute `xml:"transitRoute"`
}

// TransitRoute represents a route within a transit line
type TransitRoute struct {
	ID          string      `xml:"id,attr"`
	TransportMode string    `xml:"transportMode"`
	Departures  []Departure `xml:"departures>departure"`
}

// Departure represents a vehicle departure
type Departure struct {
	ID        string `xml:"id,attr"`
	VehicleRefID string `xml:"vehicleRefId,attr"`
	DepartureTime string `xml:"departureTime,attr"`
}

// ScheduleInfo contains extracted information from the schedule
type ScheduleInfo struct {
	VehicleIDs      map[string]bool           // Set of unique vehicle IDs
	VehicleModes    map[string]string         // Vehicle ID -> Mode mapping
	LineVehicles    map[string][]string       // Line ID -> Vehicle IDs
	ModeVehicles    map[string][]string       // Mode -> Vehicle IDs
}

// ParseTransitSchedule reads and parses a MATSim transit schedule XML file
func ParseTransitSchedule(filename string) (*ScheduleInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var schedule TransitSchedule
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&schedule); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	info := &ScheduleInfo{
		VehicleIDs:   make(map[string]bool),
		VehicleModes: make(map[string]string),
		LineVehicles: make(map[string][]string),
		ModeVehicles: make(map[string][]string),
	}

	// Extract information from the schedule
	for _, line := range schedule.TransitLines {
		var lineVehicles []string

		for _, route := range line.TransitRoutes {
			mode := route.TransportMode
			if mode == "" {
				mode = "bus" // Default to bus if not specified
			}

			for _, departure := range route.Departures {
				vehicleID := departure.VehicleRefID
				if vehicleID != "" {
					info.VehicleIDs[vehicleID] = true
					info.VehicleModes[vehicleID] = mode
					
					if !contains(lineVehicles, vehicleID) {
						lineVehicles = append(lineVehicles, vehicleID)
					}
					
					if !contains(info.ModeVehicles[mode], vehicleID) {
						info.ModeVehicles[mode] = append(info.ModeVehicles[mode], vehicleID)
					}
				}
			}
		}

		if len(lineVehicles) > 0 {
			info.LineVehicles[line.ID] = lineVehicles
		}
	}

	return info, nil
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}