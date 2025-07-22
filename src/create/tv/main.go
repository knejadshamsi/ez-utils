package tv

import (
	"ez-utils/src/create/tv/core"
	"ez-utils/src/create/tv/io"
	"ez-utils/src/create/tv/tui"
	"fmt"
	"os"
)

// Run executes the transit vehicles creation command
func Run(args []string) error {
	if len(args) < 1 {
		fmt.Println("Error: Please provide a transit vehicles XML file")
		fmt.Println("Usage: ez-utils create tv <transitVehicles.xml>")
		return fmt.Errorf("missing transit vehicles file")
	}

	vehiclesFile := args[0]

	// Create the transit vehicles manager
	manager := core.NewVehicleManager(vehiclesFile)

	// Load data from file
	if err := loadFromFile(manager); err != nil {
		// If it's not a vehicles file, try as schedule file
		if scheduleInfo, scheduleErr := io.ParseTransitSchedule(manager.VehiclesFile); scheduleErr == nil {
			fmt.Printf("Detected transit schedule file with %d vehicle references\n", len(scheduleInfo.VehicleIDs))
			// Create a new vehicles file name
			manager.VehiclesFile = "transitVehicles.xml"
			fmt.Printf("Will create/edit vehicles in: %s\n", manager.VehiclesFile)

			// Try to load existing vehicles file if it exists
			if loadErr := loadFromFile(manager); loadErr != nil {
				fmt.Printf("Starting with empty vehicle database\n")
			} else {
				fmt.Printf("Loaded existing %d vehicle types and %d vehicles\n",
					len(manager.VehicleTypes), len(manager.Vehicles))
			}
		} else {
			// File doesn't exist - start empty
			fmt.Printf("Starting with empty vehicle database (file will be created on save)\n")
		}
	} else {
		fmt.Printf("Loaded %d vehicle types and %d vehicles from %s\n",
			len(manager.VehicleTypes), len(manager.Vehicles), manager.VehiclesFile)
	}

	// Run the interactive TUI
	return tui.RunBubbleTeaUI(manager)
}

// loadFromFile loads vehicles from the XML file
func loadFromFile(manager *core.VehicleManager) error {
	// Check if file exists
	if _, err := os.Stat(manager.VehiclesFile); os.IsNotExist(err) {
		return err
	}

	// Import the vehicles
	types, vehicles, err := io.ImportVehicles(manager.VehiclesFile)
	if err != nil {
		return fmt.Errorf("failed to load vehicles: %w", err)
	}

	manager.VehicleTypes = types
	manager.Vehicles = vehicles
	manager.HasChanges = false

	return nil
}

// SaveToFile saves vehicles to the XML file - can be called from TUI
func SaveToFile(manager *core.VehicleManager) error {
	return io.SaveToFile(manager, manager.VehiclesFile)
}
