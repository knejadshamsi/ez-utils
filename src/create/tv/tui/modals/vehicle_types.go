package modals

import (
	"fmt"
	"strconv"

	"ez-utils/src/create/tv/core"
	"ez-utils/src/create/tv/tui/form"
	"ez-utils/src/create/tv/tui/interfaces"
)

// Modal states - will be moved to shared constants
const (
	ModalAddType = iota + 1
	ModalEditType
)

// ShowAddTypeModal creates and shows the add vehicle type modal
func ShowAddTypeModal(m interfaces.ModelInterface) {
	m.SetModal(ModalAddType)

	f := form.NewForm("Add Vehicle Type", "Create a new vehicle type", m.GetStyles())

	// Add fields
	f.AddField("Vehicle Type ID", "e.g., bus_standard", true, func(s string) error {
		if s == "" {
			return fmt.Errorf("ID is required")
		}
		if m.GetManager().FindVehicleType(s) != nil {
			return fmt.Errorf("ID already exists")
		}
		return nil
	})

	f.AddField("Description", "e.g., Standard city bus", false, nil)
	f.AddCapacityField("Seats", "30", true)
	f.AddCapacityField("Standing Room", "40", true)
	f.AddFloatField("Length (meters)", "12.0", true)
	f.AddFloatField("Width (meters)", "2.5", true)

	// Set submit handler
	f.SetSubmitHandler(func(values map[string]string) error {
		// Extract and validate the ID
		id := values["Vehicle Type ID"]
		if id == "" {
			return fmt.Errorf("Vehicle Type ID is required")
		}

		// Double-check if ID already exists (in case validation was bypassed)
		if m.GetManager().FindVehicleType(id) != nil {
			return fmt.Errorf("Vehicle type with ID '%s' already exists", id)
		}

		// Parse numeric values
		seats, err := strconv.Atoi(values["Seats"])
		if err != nil {
			return fmt.Errorf("Invalid seats value: %v", err)
		}

		standing, err := strconv.Atoi(values["Standing Room"])
		if err != nil {
			return fmt.Errorf("Invalid standing room value: %v", err)
		}

		length, err := strconv.ParseFloat(values["Length (meters)"], 64)
		if err != nil {
			return fmt.Errorf("Invalid length value: %v", err)
		}

		width, err := strconv.ParseFloat(values["Width (meters)"], 64)
		if err != nil {
			return fmt.Errorf("Invalid width value: %v", err)
		}

		// Create the vehicle type
		vt := &core.VehicleType{
			ID:          id,
			Description: values["Description"],
			Capacity: core.VehicleCapacity{
				Seats:    seats,
				Standing: standing,
			},
			Length: length,
			Width:  width,
		}

		m.GetManager().AddVehicleType(vt)
		m.SetStatusMsg(fmt.Sprintf("Added vehicle type: %s", vt.ID))

		return nil
	})

	// Initialize form and focus first field
	f.Init()
	m.SetCurrentForm(f)
}

// ShowEditTypeModal creates and shows the edit vehicle type modal
func ShowEditTypeModal(m interfaces.ModelInterface, selectedType *core.VehicleType) {
	if selectedType == nil {
		m.SetErrorMsg("No vehicle type selected")
		return
	}

	m.SetModal(ModalEditType)

	f := form.NewForm("Edit Vehicle Type", fmt.Sprintf("Editing: %s", selectedType.ID), m.GetStyles())

	// Add fields with existing values
	f.AddFieldWithValue("Vehicle Type ID", selectedType.ID, "e.g., bus_standard", true, func(s string) error {
		if s == "" {
			return fmt.Errorf("ID is required")
		}
		// Allow same ID (editing existing)
		if s != selectedType.ID && m.GetManager().FindVehicleType(s) != nil {
			return fmt.Errorf("ID already exists")
		}
		return nil
	})

	f.AddFieldWithValue("Description", selectedType.Description, "e.g., Standard city bus", false, nil)
	f.AddCapacityFieldWithValue("Seats", fmt.Sprintf("%d", selectedType.Capacity.Seats), "30", true)
	f.AddCapacityFieldWithValue("Standing Room", fmt.Sprintf("%d", selectedType.Capacity.Standing), "40", true)
	f.AddFieldWithValue("Length (meters)", fmt.Sprintf("%.1f", selectedType.Length), "12.0", true, nil)
	f.AddFieldWithValue("Width (meters)", fmt.Sprintf("%.1f", selectedType.Width), "2.5", true, nil)

	// Set submit handler
	f.SetSubmitHandler(func(values map[string]string) error {
		// Parse numeric values
		seats, err := strconv.Atoi(values["Seats"])
		if err != nil {
			return fmt.Errorf("Invalid seats value: %v", err)
		}

		standing, err := strconv.Atoi(values["Standing Room"])
		if err != nil {
			return fmt.Errorf("Invalid standing room value: %v", err)
		}

		length, err := strconv.ParseFloat(values["Length (meters)"], 64)
		if err != nil {
			return fmt.Errorf("Invalid length value: %v", err)
		}

		width, err := strconv.ParseFloat(values["Width (meters)"], 64)
		if err != nil {
			return fmt.Errorf("Invalid width value: %v", err)
		}

		// Update the vehicle type
		selectedType.ID = values["Vehicle Type ID"]
		selectedType.Description = values["Description"]
		selectedType.Capacity.Seats = seats
		selectedType.Capacity.Standing = standing
		selectedType.Length = length
		selectedType.Width = width

		m.GetManager().HasChanges = true
		m.SetStatusMsg(fmt.Sprintf("Updated vehicle type: %s", selectedType.ID))

		return nil
	})

	// Initialize form and focus first field
	f.Init()
	m.SetCurrentForm(f)
}
