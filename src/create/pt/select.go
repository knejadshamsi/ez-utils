// Step 4: Service Selection
// This file handles the selection of services based on the identified service patterns.
// It filters and selects all services that match the configured service day for further processing.
package pt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func (pto *PTOrchestrator) selectServices() error {
	pto.logMessage("Starting Service Selection")

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

	// Select ONE bus service + ONE metro service for single day processing
	busService := ""
	metroService := ""
	busRoutes := []Route{}
	metroRoutes := []Route{}

	// Find best bus and metro services
	for _, mapping := range serviceMappings {
		if !validServiceIDs[mapping.ServiceID] {
			continue
		}

		// Count bus and metro routes in this service
		busCount := 0
		metroCount := 0
		serviceBusRoutes := []Route{}
		serviceMetroRoutes := []Route{}

		for _, route := range mapping.Routes {
			if routeType, err := strconv.Atoi(route.RouteType); err == nil {
				if routeType == 3 { // Bus
					busCount++
					serviceBusRoutes = append(serviceBusRoutes, route)
				} else if routeType == 1 { // Metro
					metroCount++
					serviceMetroRoutes = append(serviceMetroRoutes, route)
				}
			}
		}

		// Pick service with most bus routes
		if busCount > 0 && (busService == "" || busCount > len(busRoutes)) {
			busService = mapping.ServiceID
			busRoutes = serviceBusRoutes
		}

		// Pick service with most metro routes
		if metroCount > 0 && (metroService == "" || metroCount > len(metroRoutes)) {
			metroService = mapping.ServiceID
			metroRoutes = serviceMetroRoutes
		}
	}

	if busService == "" && metroService == "" {
		pto.logMessage("No bus or metro services found for target date")
		return NewInvalidServiceError("no transit services available")
	}

	// Log the selection for future SelectedServicesByType format
	pto.logMessage(fmt.Sprintf("Selected for %s: Bus=%s, Metro=%s", pto.config.TargetDate, busService, metroService))

	// Also create legacy SelectedService array for compatibility
	var selectedServices []SelectedService
	if busService != "" {
		selectedServices = append(selectedServices, SelectedService{
			ServiceID: busService,
			Routes:    busRoutes,
			Selected:  true,
			TripCount: 0,
		})
	}
	if metroService != "" && metroService != busService {
		selectedServices = append(selectedServices, SelectedService{
			ServiceID: metroService,
			Routes:    metroRoutes,
			Selected:  true,
			TripCount: 0,
		})
	}

	// Log selected services
	if busService != "" {
		pto.logMessage(fmt.Sprintf("Selected bus service: %s (%d routes)", busService, len(busRoutes)))
	}
	if metroService != "" {
		pto.logMessage(fmt.Sprintf("Selected metro service: %s (%d routes)", metroService, len(metroRoutes)))
	}

	if len(selectedServices) == 0 {
		pto.logMessage("No services found matching the specified day")
		return NewInvalidServiceError(fmt.Sprintf("no services available for %s", pto.getServiceDayString()))
	}

	// Calculate totals
	totalBusRoutes := len(busRoutes)
	totalMetroRoutes := len(metroRoutes)
	totalRoutes := totalBusRoutes + totalMetroRoutes

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
		len(selectedServices), totalBusRoutes, totalMetroRoutes, totalRoutes, pto.config.TargetDate))
	return nil
}
