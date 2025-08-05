package pt

// Package pt provides GTFS (General Transit Feed Specification) to XML conversion functionality.
//
// This package implements a 10-step processing pipeline that converts GTFS data into
// transit schedule XML format:
//
// Step 1: File Validation (validate.go) - Validates all required GTFS files exist and have proper format
// Step 2: Service Pattern Identification (identify.go) - Identifies service patterns matching the specified day
// Step 3: Service-Route Mapping (map.go) - Maps services to their corresponding routes
// Step 4: Service Selection (select.go) - Selects specific services for processing
// Step 5: Trip Database Creation (collect.go) - Creates SQLite database from trips.txt for performance
// Step 6: Trip Collection (collect.go) - Queries selected trips from database
// Step 7: Stop Sequence Collection (sequence.go) - Collects stop sequences for each trip
// Step 8: Stop Details Collection (details.go) - Collects detailed stop information
// Step 9: Generate Transit Schedule XML (xml.go) - Creates the final XML output
// Step 10: Clean Temporary Files (cleanup.go) - Removes temporary processing files (optional)
//
// The package is organized following the population module's architectural patterns:
// - orchestrator.go contains the orchestrator and step implementations
// - types.go defines data structures for GTFS entities
// - errors.go provides custom error types for specific failure scenarios

import ()

// PTFacade wraps the processing orchestrator to provide a public API.
type PTFacade struct {
	*PTOrchestrator
}

// NewPTFacade creates a new PT processing orchestrator facade.
func NewPTFacade(gtfsDir string, serviceDay ServiceDay, cleanFlag bool) *PTFacade {
	return &PTFacade{
		PTOrchestrator: NewPTOrchestrator(gtfsDir, serviceDay, cleanFlag),
	}
}

// IsQuitRequested exposes the quit check functionality
func (f *PTFacade) IsQuitRequested() bool {
	return f.PTOrchestrator.IsQuitRequested()
}
