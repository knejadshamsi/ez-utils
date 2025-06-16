package pt

// Package pt provides GTFS (General Transit Feed Specification) to XML conversion functionality.
// 
// This package implements a 9-step processing pipeline that converts GTFS data into 
// transit schedule XML format:
//
// Step 1: File Validation - Validates all required GTFS files exist and have proper format
// Step 2: Service Pattern Identification - Identifies service patterns matching the specified day
// Step 3: Service-Route Mapping - Maps services to their corresponding routes
// Step 4: Service Selection - Selects specific services for processing
// Step 5: Trip Collection - Collects all trips for selected services
// Step 6: Stop Sequence Collection - Collects stop sequences for each trip
// Step 7: Stop Details Collection - Collects detailed stop information
// Step 8: Generate Transit Schedule XML - Creates the final XML output
// Step 9: Clean Temporary Files - Removes temporary processing files (optional)
//
// The package is organized following the population module's architectural patterns:
// - processing/ contains the orchestrator and step implementations
// - types.go defines data structures for GTFS entities
// - errors.go provides custom error types for specific failure scenarios

import (
	"os"
	
	"ez-utils/src/create/pt/processing"
)

// ServiceDay re-exports the processing ServiceDay type
type ServiceDay = processing.ServiceDay

// Service day constants
const (
	ServiceDayWeekday   = processing.ServiceDayWeekday
	ServiceDayWeekend   = processing.ServiceDayWeekend
	ServiceDayMonday    = processing.ServiceDayMonday
	ServiceDayTuesday   = processing.ServiceDayTuesday
	ServiceDayWednesday = processing.ServiceDayWednesday
	ServiceDayThursday  = processing.ServiceDayThursday
	ServiceDayFriday    = processing.ServiceDayFriday
	ServiceDaySaturday  = processing.ServiceDaySaturday
	ServiceDaySunday    = processing.ServiceDaySunday
)

// PTOrchestrator wraps the processing orchestrator
type PTOrchestrator struct {
	*processing.PTOrchestrator
}

// NewPTOrchestrator creates a new PT processing orchestrator
func NewPTOrchestrator(gtfsDir string, serviceDay ServiceDay, cleanFlag bool) *PTOrchestrator {
	return &PTOrchestrator{
		PTOrchestrator: processing.NewPTOrchestrator(gtfsDir, serviceDay, cleanFlag),
	}
}

// IsQuitRequested exposes the quit check functionality
func (pto *PTOrchestrator) IsQuitRequested() bool {
	return pto.PTOrchestrator.IsQuitRequested()
}

// FileExists checks if a file exists and is not a directory
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}