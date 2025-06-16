package processing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func (pto *PTOrchestrator) selectServices() error {
	pto.logMessage("Starting Step 4: Service Selection")
	
	// State integration available for future TUI implementation
	if pto.tuiEnabled {
		// Future: set_state("PT_4")
	}
	
	// Read the service mapping from Step 3
	serviceMappingPath := filepath.Join(pto.config.TempDir, ServicesDir, ServiceMappingFile)
	serviceMappingData, err := os.ReadFile(serviceMappingPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read service mapping file: %v", err))
		return NewFileParsingError(ServiceMappingFile, err.Error())
	}
	
	var serviceMappings []ServiceMapping
	if err := json.Unmarshal(serviceMappingData, &serviceMappings); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to parse service mapping JSON: %v", err))
		return NewFileParsingError(ServiceMappingFile, err.Error())
	}
	
	if len(serviceMappings) == 0 {
		pto.logMessage("No service mappings found")
		return NewInvalidServiceError("no services available for selection")
	}
	
	// Read service patterns from Step 2 to filter services by day
	servicePatternsPath := filepath.Join(pto.config.TempDir, ServicesDir, ServicePatternsFile)
	servicePatternsData, err := os.ReadFile(servicePatternsPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read service patterns file: %v", err))
		return NewFileParsingError(ServicePatternsFile, err.Error())
	}
	
	var servicePatterns []ServicePattern
	if err := json.Unmarshal(servicePatternsData, &servicePatterns); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to parse service patterns JSON: %v", err))
		return NewFileParsingError(ServicePatternsFile, err.Error())
	}
	
	// Create a map of valid service IDs (those that match the service day)
	validServiceIDs := make(map[string]bool)
	for _, pattern := range servicePatterns {
		validServiceIDs[pattern.ServiceID] = true
	}
	
	// Select ALL services that match the configured service day
	// This ensures we process the complete transit system for the specified day
	var selectedServices []SelectedService
	selectedCount := 0
	totalServices := len(serviceMappings)
	
	for _, mapping := range serviceMappings {
		// Only process services that match the configured service day
		if !validServiceIDs[mapping.ServiceID] {
			pto.logMessage(fmt.Sprintf("Filtered out service %s (doesn't match %s)", mapping.ServiceID, pto.getServiceDayString()))
			continue
		}
		
		// Select this service - we want ALL services for the day, not random selection
		selected := SelectedService{
			ServiceID: mapping.ServiceID,
			Routes:    mapping.Routes,
			Selected:  true, // Select ALL valid services
			TripCount: 0,    // Will be populated in Step 5
		}
		
		selectedServices = append(selectedServices, selected)
		selectedCount++
		
		// Log service selection with route breakdown
		busRoutes := 0
		metroRoutes := 0
		for _, route := range mapping.Routes {
			if routeType, err := strconv.Atoi(route.RouteType); err == nil {
				if routeType == 3 {
					busRoutes++
				} else if routeType == 1 {
					metroRoutes++
				}
			}
		}
		
		pto.logMessage(fmt.Sprintf("Selected service %s: %d bus routes, %d metro routes (total: %d)", 
			mapping.ServiceID, busRoutes, metroRoutes, len(mapping.Routes)))
		
		// Update display progress
		if pto.tuiEnabled && pto.displayInstance != nil {
			_ = pto.displayInstance.SetLiveUpdate(3, "selected", fmt.Sprintf("%d", selectedCount))
			_ = pto.displayInstance.SetLiveUpdate(3, "total", fmt.Sprintf("%d", totalServices))
		}
	}
	
	if len(selectedServices) == 0 {
		pto.logMessage("No services found matching the specified day")
		return NewInvalidServiceError(fmt.Sprintf("no services available for %s", pto.getServiceDayString()))
	}
	
	// Calculate total routes across all selected services
	totalRoutes := 0
	totalBusRoutes := 0
	totalMetroRoutes := 0
	
	for _, service := range selectedServices {
		totalRoutes += len(service.Routes)
		for _, route := range service.Routes {
			if routeType, err := strconv.Atoi(route.RouteType); err == nil {
				if routeType == 3 {
					totalBusRoutes++
				} else if routeType == 1 {
					totalMetroRoutes++
				}
			}
		}
	}
	
	// Create output file
	outputPath := filepath.Join(pto.config.TempDir, ServicesDir, SelectedServicesFile)
	jsonData, err := json.MarshalIndent(selectedServices, "", "  ")
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to marshal selected services to JSON: %v", err))
		return NewFileParsingError(SelectedServicesFile, err.Error())
	}
	
	if err := os.WriteFile(outputPath, jsonData, FilePermissions); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to write selected services file: %v", err))
		return NewFileParsingError(SelectedServicesFile, err.Error())
	}
	
	pto.logMessage(fmt.Sprintf("Service selection completed successfully. Selected %d services (%d bus routes, %d metro routes, %d total routes) for %s", 
		len(selectedServices), totalBusRoutes, totalMetroRoutes, totalRoutes, pto.getServiceDayString()))
	return nil
}