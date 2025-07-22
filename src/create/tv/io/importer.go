package io

import (
	"encoding/xml"
	"fmt"
	"os"
	"ez-utils/src/create/tv/core"
)

// ImportVehicles reads an existing transitVehicles.xml file
func ImportVehicles(filename string) ([]*core.VehicleType, []*core.Vehicle, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var vd VehicleDefinitions
	decoder := xml.NewDecoder(file)
	if err := decoder.Decode(&vd); err != nil {
		return nil, nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Convert to internal structures
	var vehicleTypes []*core.VehicleType
	var vehicles []*core.Vehicle

	// Import vehicle types
	for _, xmlType := range vd.VehicleTypes {
		vt := &core.VehicleType{
			ID:          xmlType.ID,
			Description: xmlType.Description,
			Capacity: core.VehicleCapacity{
				Seats:    xmlType.Capacity.Seats.Persons,
				Standing: xmlType.Capacity.StandingRoom.Persons,
			},
			Length:                  xmlType.Length,
			Width:                   xmlType.Width,
			AccessTime:              xmlType.AccessTime,
			EgressTime:              xmlType.EgressTime,
			DoorOperation:           xmlType.DoorOperation,
			PassengerCarEquivalents: xmlType.PassengerCarEquivalents,
		}
		vehicleTypes = append(vehicleTypes, vt)
	}

	// Import vehicles
	for _, xmlVehicle := range vd.Vehicles {
		v := &core.Vehicle{
			ID:     xmlVehicle.ID,
			TypeID: xmlVehicle.Type,
		}
		vehicles = append(vehicles, v)
	}

	return vehicleTypes, vehicles, nil
}

// MergeVehicles merges imported vehicles with existing ones
func MergeVehicles(existing, imported map[string]*core.Vehicle, strategy string) (map[string]*core.Vehicle, int, int) {
	merged := make(map[string]*core.Vehicle)
	updated := 0
	added := 0

	// Copy existing vehicles
	for id, v := range existing {
		merged[id] = v
	}

	// Merge imported vehicles
	for id, v := range imported {
		if _, exists := merged[id]; exists {
			if strategy == "overwrite" {
				merged[id] = v
				updated++
			}
			// If strategy is "skip", do nothing
		} else {
			merged[id] = v
			added++
		}
	}

	return merged, added, updated
}

// MergeVehicleTypes merges imported vehicle types with existing ones
func MergeVehicleTypes(existing, imported map[string]*core.VehicleType, strategy string) (map[string]*core.VehicleType, int, int) {
	merged := make(map[string]*core.VehicleType)
	updated := 0
	added := 0

	// Copy existing types
	for id, vt := range existing {
		merged[id] = vt
	}

	// Merge imported types
	for id, vt := range imported {
		if _, exists := merged[id]; exists {
			if strategy == "overwrite" {
				merged[id] = vt
				updated++
			}
			// If strategy is "skip", do nothing
		} else {
			merged[id] = vt
			added++
		}
	}

	return merged, added, updated
}