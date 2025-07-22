package core

// FindVehicleType finds a vehicle type by ID
func (vm *VehicleManager) FindVehicleType(id string) *VehicleType {
	for _, vt := range vm.VehicleTypes {
		if vt.ID == id {
			return vt
		}
	}
	return nil
}

// FindVehicle finds a vehicle by ID
func (vm *VehicleManager) FindVehicle(id string) *Vehicle {
	for _, v := range vm.Vehicles {
		if v.ID == id {
			return v
		}
	}
	return nil
}

// AddVehicleType adds a new vehicle type
func (vm *VehicleManager) AddVehicleType(vt *VehicleType) {
	vm.VehicleTypes = append(vm.VehicleTypes, vt)
	vm.HasChanges = true
}

// AddVehicle adds a new vehicle
func (vm *VehicleManager) AddVehicle(v *Vehicle) {
	vm.Vehicles = append(vm.Vehicles, v)
	vm.HasChanges = true
}

// RemoveVehicleType removes a vehicle type by ID
func (vm *VehicleManager) RemoveVehicleType(id string) {
	for i, vt := range vm.VehicleTypes {
		if vt.ID == id {
			vm.VehicleTypes = append(vm.VehicleTypes[:i], vm.VehicleTypes[i+1:]...)
			vm.HasChanges = true
			return
		}
	}
}

// RemoveVehicle removes a vehicle by ID
func (vm *VehicleManager) RemoveVehicle(id string) {
	for i, v := range vm.Vehicles {
		if v.ID == id {
			vm.Vehicles = append(vm.Vehicles[:i], vm.Vehicles[i+1:]...)
			vm.HasChanges = true
			return
		}
	}
}