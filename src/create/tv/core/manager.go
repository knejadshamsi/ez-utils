package core


// VehicleManager manages the vehicle creation process
type VehicleManager struct {
	VehiclesFile string
	
	// Vehicle data being managed
	VehicleTypes []*VehicleType
	Vehicles     []*Vehicle
	
	// Track if file has unsaved changes
	HasChanges bool
}

// NewVehicleManager creates a new vehicle manager instance
func NewVehicleManager(vehiclesFile string) *VehicleManager {
	return &VehicleManager{
		VehiclesFile: vehiclesFile,
		VehicleTypes: []*VehicleType{},
		Vehicles:     []*Vehicle{},
		HasChanges:   false,
	}
}


