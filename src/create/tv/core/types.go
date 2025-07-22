package core

// VehicleCapacity represents the passenger capacity of a vehicle
type VehicleCapacity struct {
	Seats    int `json:"seats"`
	Standing int `json:"standing"`
}

// VehicleType represents a type/class of vehicle with its specifications
type VehicleType struct {
	ID                      string          `json:"id"`
	Description             string          `json:"description"`
	Capacity                VehicleCapacity `json:"capacity"`
	Length                  float64         `json:"length"`
	Width                   float64         `json:"width"`
	AccessTime              float64         `json:"access_time"`
	EgressTime              float64         `json:"egress_time"`
	DoorOperation           string          `json:"door_operation"`
	PassengerCarEquivalents float64         `json:"passenger_car_equivalents"`
}

// Vehicle represents an individual vehicle instance
type Vehicle struct {
	ID     string `json:"id"`
	TypeID string `json:"type_id"`
}