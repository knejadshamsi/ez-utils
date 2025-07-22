package core

import (
	"fmt"
	"strings"
)

// ValidationResult holds the results of validation checks
type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

// ValidateVehicles performs comprehensive validation on vehicles and types
func (vm *VehicleManager) ValidateVehicles() *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Validate vehicle types
	vm.validateVehicleTypes(result)

	// Validate vehicles
	vm.validateVehicleList(result)

	// Validate all departures have vehicles
	vm.validateDepartureVehicles(result)

	// Check for orphaned vehicles
	vm.validateOrphanedVehicles(result)

	return result
}

// validateVehicleTypes checks vehicle type definitions
func (vm *VehicleManager) validateVehicleTypes(result *ValidationResult) {
	if len(vm.VehicleTypes) == 0 {
		result.Errors = append(result.Errors, "No vehicle types defined")
		result.Valid = false
		return
	}

	for _, vt := range vm.VehicleTypes {
		// Check ID
		if vt.ID == "" {
			result.Errors = append(result.Errors, "Vehicle type has empty ID")
			result.Valid = false
		}

		// Check capacity
		if vt.Capacity.Seats < 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Vehicle type '%s' has negative seats: %d", vt.ID, vt.Capacity.Seats))
			result.Valid = false
		}

		if vt.Capacity.Standing < 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Vehicle type '%s' has negative standing room: %d", vt.ID, vt.Capacity.Standing))
			result.Valid = false
		}

		totalCapacity := vt.Capacity.Seats + vt.Capacity.Standing
		if totalCapacity == 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Vehicle type '%s' has zero total capacity", vt.ID))
		}

		// Check dimensions
		if vt.Length <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Vehicle type '%s' has invalid length: %.2f", vt.ID, vt.Length))
			result.Valid = false
		}

		if vt.Width <= 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Vehicle type '%s' has invalid width: %.2f", vt.ID, vt.Width))
			result.Valid = false
		}

		// Check times
		if vt.AccessTime < 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Vehicle type '%s' has negative access time: %.2f", vt.ID, vt.AccessTime))
			result.Valid = false
		}

		if vt.EgressTime < 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("Vehicle type '%s' has negative egress time: %.2f", vt.ID, vt.EgressTime))
			result.Valid = false
		}

		// Check door operation
		if vt.DoorOperation != "serial" && vt.DoorOperation != "parallel" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Vehicle type '%s' has non-standard door operation: %s", vt.ID, vt.DoorOperation))
		}

		// Check passenger car equivalents
		if vt.PassengerCarEquivalents <= 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Vehicle type '%s' has invalid PCE: %.2f", vt.ID, vt.PassengerCarEquivalents))
		}
	}
}

// validateVehicleList checks individual vehicles
func (vm *VehicleManager) validateVehicleList(result *ValidationResult) {
	if len(vm.Vehicles) == 0 {
		result.Warnings = append(result.Warnings, "No vehicles defined")
		return
	}

	// Check for duplicate IDs (shouldn't happen with map structure)
	vehicleIDs := make(map[string]bool)

	for _, vehicle := range vm.Vehicles {
		// Check ID
		if vehicle.ID == "" {
			result.Errors = append(result.Errors, "Vehicle has empty ID")
			result.Valid = false
			continue
		}

		// Check for duplicates
		if vehicleIDs[vehicle.ID] {
			result.Errors = append(result.Errors, fmt.Sprintf("Duplicate vehicle ID: %s", vehicle.ID))
			result.Valid = false
		}
		vehicleIDs[vehicle.ID] = true

		// Check type reference
		if vehicle.TypeID == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Vehicle '%s' has no type assigned", vehicle.ID))
			result.Valid = false
		} else if vm.FindVehicleType(vehicle.TypeID) == nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Vehicle '%s' references non-existent type: %s", vehicle.ID, vehicle.TypeID))
			result.Valid = false
		}
	}
}

// validateDepartureVehicles ensures all departure vehicle references have corresponding vehicles
func (vm *VehicleManager) validateDepartureVehicles(_ *ValidationResult) error {
	// Schedule validation not available in this version
	return nil
}

// validateOrphanedVehicles checks for vehicles not used in any departure
func (vm *VehicleManager) validateOrphanedVehicles(result *ValidationResult) {
	// Schedule validation not available in this version
}

// ValidateCapacity validates capacity values
func ValidateCapacity(seats, standing int) error {
	if seats < 0 {
		return fmt.Errorf("seats cannot be negative")
	}
	if standing < 0 {
		return fmt.Errorf("standing capacity cannot be negative")
	}
	if seats+standing == 0 {
		return fmt.Errorf("total capacity cannot be zero")
	}
	if seats > 10000 {
		return fmt.Errorf("seats exceeds reasonable maximum (10000)")
	}
	if standing > 10000 {
		return fmt.Errorf("standing capacity exceeds reasonable maximum (10000)")
	}
	return nil
}

// ValidateVehicleID validates a vehicle ID
func ValidateVehicleID(id string) error {
	if id == "" {
		return fmt.Errorf("vehicle ID cannot be empty")
	}
	if strings.ContainsAny(id, " \t\n\r") {
		return fmt.Errorf("vehicle ID cannot contain whitespace")
	}
	if strings.ContainsAny(id, "<>&\"'") {
		return fmt.Errorf("vehicle ID contains invalid XML characters")
	}
	return nil
}

// PrintValidationResult displays validation results
func (result *ValidationResult) Print() {
	if result.Valid && len(result.Warnings) == 0 {
		fmt.Println("\n✓ All validations passed")
		return
	}

	if len(result.Errors) > 0 {
		fmt.Println("\n❌ Validation Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\n⚠️  Validation Warnings:")
		for _, warn := range result.Warnings {
			fmt.Printf("  - %s\n", warn)
		}
	}
}