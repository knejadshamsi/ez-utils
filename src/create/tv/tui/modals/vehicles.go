package modals

import (
	"fmt"

	"ez-utils/src/create/tv/core"
	"ez-utils/src/create/tv/tui/form"
	"ez-utils/src/create/tv/tui/interfaces"
)

// ShowEditVehicleModal creates and shows the edit vehicle modal
func ShowEditVehicleModal(m interfaces.ModelInterface, selectedVehicle *core.Vehicle) {
	if selectedVehicle == nil {
		m.SetErrorMsg("No vehicle selected")
		return
	}

	// Modal constants are defined in vehicle_types.go
	m.SetModal(ModalEditType)

	f := form.NewForm("Edit Vehicle", fmt.Sprintf("Editing: %s", selectedVehicle.ID), m.GetStyles())

	// Add Vehicle ID field with existing value
	f.AddFieldWithValue("Vehicle ID", selectedVehicle.ID, "e.g., bus_001", true, func(s string) error {
		if s == "" {
			return fmt.Errorf("ID is required")
		}
		// Allow same ID (editing existing) but check for duplicates with other vehicles
		if s != selectedVehicle.ID {
			for _, v := range m.GetManager().Vehicles {
				if v.ID == s {
					return fmt.Errorf("ID already exists")
				}
			}
		}
		return nil
	})

	// Get all vehicle type IDs for the select field
	// Put the current vehicle's type first so it's pre-selected
	var vehicleTypeOptions []string

	// Add current type first
	vehicleTypeOptions = append(vehicleTypeOptions, selectedVehicle.TypeID)

	// Add other types
	for _, vt := range m.GetManager().VehicleTypes {
		if vt.ID != selectedVehicle.TypeID {
			vehicleTypeOptions = append(vehicleTypeOptions, vt.ID)
		}
	}

	// Add Vehicle Type select field - it will automatically select the first option
	f.AddSelectField("Vehicle Type", vehicleTypeOptions, true)

	// Set submit handler
	f.SetSubmitHandler(func(values map[string]string) error {
		// Update the vehicle with new values
		selectedVehicle.ID = values["Vehicle ID"]
		selectedVehicle.TypeID = values["Vehicle Type"]

		// Mark that changes have been made
		m.GetManager().HasChanges = true
		m.SetStatusMsg(fmt.Sprintf("Updated vehicle: %s", selectedVehicle.ID))

		return nil
	})

	// Initialize form and focus first field
	f.Init()
	m.SetCurrentForm(f)
}
