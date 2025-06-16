package processing

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	
)

func (pto *PTOrchestrator) mapServiceRoutes() error {
	pto.logMessage("Starting Step 3: Service-Route Mapping")
	
	// Initialize live updates for route mapping
	if pto.tuiEnabled && pto.displayInstance != nil {
		_ = pto.displayInstance.SetLiveUpdate(2, "routes", "0")
		_ = pto.displayInstance.SetLiveUpdate(2, "mappings", "0")
	}
	
	// Read the service patterns from Step 2
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
	
	// Create a map of service IDs for quick lookup
	serviceIDs := make(map[string]bool)
	for _, pattern := range servicePatterns {
		serviceIDs[pattern.ServiceID] = true
	}
	
	// Read trips.txt to get route IDs for each service
	tripsPath := filepath.Join(pto.config.GTFSDirectory, TripsFile)
	tripsFile, err := os.Open(tripsPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to open trips.txt: %v", err))
		return NewFileParsingError(TripsFile, err.Error())
	}
	defer tripsFile.Close()
	
	tripsReader := csv.NewReader(tripsFile)
	
	// Read trips header
	tripsHeaders, err := tripsReader.Read()
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read trips.txt headers: %v", err))
		return NewFileParsingError(TripsFile, "invalid header row")
	}
	
	// Create trips header index map
	tripsHeaderIndex := make(map[string]int)
	for i, header := range tripsHeaders {
		tripsHeaderIndex[header] = i
	}
	
	// Verify required columns exist in trips.txt
	requiredTripsColumns := []string{"trip_id", "route_id", "service_id"}
	for _, col := range requiredTripsColumns {
		if _, exists := tripsHeaderIndex[col]; !exists {
			pto.logMessage(fmt.Sprintf("Missing required column in trips.txt: %s", col))
			return NewFileParsingError(TripsFile, fmt.Sprintf("missing column: %s", col))
		}
	}
	
	// Map service IDs to route IDs
	serviceToRoutes := make(map[string]map[string]bool)
	
	// Read trips data
	for {
		record, err := tripsReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			pto.logMessage(fmt.Sprintf("Error reading %s row: %v", TripsFile, err))
			continue
		}
		
		serviceID := record[tripsHeaderIndex["service_id"]]
		routeID := record[tripsHeaderIndex["route_id"]]
		
		// Only process trips for services we identified in Step 2
		if serviceIDs[serviceID] {
			if serviceToRoutes[serviceID] == nil {
				serviceToRoutes[serviceID] = make(map[string]bool)
			}
			serviceToRoutes[serviceID][routeID] = true
		}
	}
	
	// Read routes.txt to get route types
	routesPath := filepath.Join(pto.config.GTFSDirectory, RoutesFile)
	routesFile, err := os.Open(routesPath)
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to open routes.txt: %v", err))
		return NewFileParsingError(RoutesFile, err.Error())
	}
	defer routesFile.Close()
	
	routesReader := csv.NewReader(routesFile)
	
	// Read routes header
	routesHeaders, err := routesReader.Read()
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to read routes.txt headers: %v", err))
		return NewFileParsingError(RoutesFile, "invalid header row")
	}
	
	// Create routes header index map
	routesHeaderIndex := make(map[string]int)
	for i, header := range routesHeaders {
		routesHeaderIndex[header] = i
	}
	
	// Verify required columns exist in routes.txt
	requiredRoutesColumns := []string{"route_id", "route_type"}
	for _, col := range requiredRoutesColumns {
		if _, exists := routesHeaderIndex[col]; !exists {
			pto.logMessage(fmt.Sprintf("Missing required column in routes.txt: %s", col))
			return NewFileParsingError(RoutesFile, fmt.Sprintf("missing column: %s", col))
		}
	}
	
	// Map route IDs to route types
	routeTypes := make(map[string]int)
	
	// Read routes data
	for {
		record, err := routesReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			pto.logMessage(fmt.Sprintf("Error reading routes.txt row: %v", err))
			continue
		}
		
		routeID := record[routesHeaderIndex["route_id"]]
		routeTypeStr := record[routesHeaderIndex["route_type"]]
		
		// Parse route type as integer
		routeType := 0 // Default to 0 if parsing fails
		if routeTypeStr != "" {
			if parsedType, parseErr := strconv.Atoi(routeTypeStr); parseErr == nil {
				routeType = parsedType
			} else {
				pto.logMessage(fmt.Sprintf("Warning: Invalid route_type for route %s: %s", routeID, routeTypeStr))
			}
		}
		
		// Only process bus (type 3) and metro (type 1) routes per documentation
		if routeType == 1 || routeType == 3 {
			routeTypes[routeID] = routeType
		} else {
			pto.logMessage(fmt.Sprintf("Filtered out route %s with type %d (only bus=3 and metro=1 are processed)", routeID, routeType))
		}
	}
	
	// Build the service mappings
	var serviceMappings []ServiceMapping
	routeCounter := 0
	mappingCounter := 0
	batchUpdateCounter := 0
	
	for serviceID, routes := range serviceToRoutes {
		mapping := ServiceMapping{
			ServiceID: serviceID,
			Routes:    []Route{},
		}
		
		for routeID := range routes {
			// Check for quit request
			if pto.tuiEnabled && pto.displayInstance != nil && pto.displayInstance.IsQuitRequested() {
				return fmt.Errorf("Step 3 interrupted by user")
			}
			
			// Only include routes that passed transport type filtering
			if routeType, exists := routeTypes[routeID]; exists {
				route := Route{
					RouteID:   routeID,
					RouteType: strconv.Itoa(routeType),
				}
				mapping.Routes = append(mapping.Routes, route)
				routeCounter++
				batchUpdateCounter++
				
				// Update display counter and stats
				pto.stats.RouteCounter = routeCounter
				
				// Batched progress updates (every 10 routes) to reduce UI overhead
				if batchUpdateCounter >= 10 || routeCounter == 1 {
					if pto.tuiEnabled && pto.displayInstance != nil {
						_ = pto.displayInstance.SetLiveUpdate(2, "routes", fmt.Sprintf("%d", routeCounter))
						_ = pto.displayInstance.SetLiveUpdate(2, "mappings", fmt.Sprintf("%d", mappingCounter))
					}
					batchUpdateCounter = 0
				}
				
				transportType := "metro"
				if routeType == 3 {
					transportType = "bus"
				}
				pto.logMessage(fmt.Sprintf("Mapped %s route %s (type %d) to service %s", transportType, routeID, routeType, serviceID))
			}
		}
		
		// Only include services that have at least one valid route (bus or metro)
		if len(mapping.Routes) > 0 {
			serviceMappings = append(serviceMappings, mapping)
			mappingCounter++
			
			// Update progress with current mapping count
			if pto.tuiEnabled && pto.displayInstance != nil {
				_ = pto.displayInstance.SetLiveUpdate(2, "mappings", fmt.Sprintf("%d", mappingCounter))
			}
		} else {
			pto.logMessage(fmt.Sprintf("Excluded service %s (no valid bus or metro routes found)", serviceID))
		}
	}
	
	// Create output file
	outputPath := filepath.Join(pto.config.TempDir, ServicesDir, ServiceMappingFile)
	jsonData, err := json.MarshalIndent(serviceMappings, "", "  ")
	if err != nil {
		pto.logMessage(fmt.Sprintf("Failed to marshal service mappings to JSON: %v", err))
		return NewFileParsingError(ServiceMappingFile, err.Error())
	}
	
	if err := os.WriteFile(outputPath, jsonData, FilePermissions); err != nil {
		pto.logMessage(fmt.Sprintf("Failed to write service mappings file: %v", err))
		return NewFileParsingError(ServiceMappingFile, err.Error())
	}
	
	// Final live update with completion status
	if pto.tuiEnabled && pto.displayInstance != nil {
		_ = pto.displayInstance.SetLiveUpdate(2, "routes", fmt.Sprintf("%d", routeCounter))
		_ = pto.displayInstance.SetLiveUpdate(2, "mappings", fmt.Sprintf("%d service links", mappingCounter))
	}
	
	pto.logMessage(fmt.Sprintf("Service-route mapping completed successfully. Mapped %d routes", routeCounter))
	return nil
}