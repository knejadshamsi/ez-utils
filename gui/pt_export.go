package gui

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// buildTransitSchedule builds the complete transit schedule structure
func (e *PTExporter) buildTransitSchedule(processed *int, total int) (*TransitSchedule, error) {
	schedule := &TransitSchedule{}

	// Build stops
	transitStops, err := e.buildTransitStops(processed, total)
	if err != nil {
		return nil, fmt.Errorf("failed to build transit stops: %w", err)
	}
	schedule.TransitStops = transitStops

	// Build lines
	transitLines, err := e.buildTransitLines(processed, total)
	if err != nil {
		return nil, fmt.Errorf("failed to build transit lines: %w", err)
	}
	schedule.TransitLines = transitLines

	return schedule, nil
}

// NewPTExporter creates a new PT exporter
func NewPTExporter(app *App, processID int) *PTExporter {
	return &PTExporter{
		processID: processID,
		db:        app.db,
		app:       app,
	}
}

// ExportPTFile exports the PT data to a new XML file
func (a *App) ExportPTFile(processID int, outputPath string) error {
	exporter := NewPTExporter(a, processID)
	return exporter.Export(outputPath)
}

// Export performs the actual export
func (e *PTExporter) Export(outputPath string) error {
	// We'll count stops as we export them since we need to get all stops
	// from all routes to count them
	stopCount := 0
	uniqueStopIDs := make(map[string]bool)

	// Get all lines for all modes to count them
	modes := []TransportMode{"BUS", "METRO", "TRAM"}
	var allLines []Line
	for _, mode := range modes {
		lines, err := e.app.GetPTLinesByMode(e.processID, mode)
		if err == nil {
			allLines = append(allLines, lines...)
		}
	}
	lineCount := len(allLines)

	// Count routes and departures by iterating through lines
	routeCount := 0
	departureCount := 0
	for _, line := range allLines {
		routes, err := e.app.GetPTRoutesByLineID(e.processID, line.ID)
		if err == nil {
			routeCount += len(routes)
			for _, route := range routes {
				deps, err := e.app.GetPTDeparturesByRouteID(e.processID, route.ID)
				if err == nil {
					departureCount += len(deps)
				}
				// Also count stops for this route
				routeStops, err := e.app.GetPTStopsByRouteID(e.processID, route.ID)
				if err == nil {
					for _, rs := range routeStops {
						uniqueStopIDs[rs.StopID] = true
					}
				}
			}
		}
	}
	stopCount = len(uniqueStopIDs)

	totalElements := stopCount + lineCount + routeCount + departureCount
	processed := 0

	// Emit initial progress
	runtime.EventsEmit(e.app.ctx, "export:progress", map[string]interface{}{
		"current": 0,
		"total":   totalElements,
	})

	// Build the complete transit schedule structure
	schedule, err := e.buildTransitSchedule(&processed, totalElements)
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to build transit schedule: %w", err)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write XML declaration
	if _, err := file.WriteString(`<?xml version="1.0" encoding="utf-8"?>`); err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to write XML header: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Write DOCTYPE
	if _, err := file.WriteString(`<!DOCTYPE transitSchedule SYSTEM "http://www.matsim.org/files/dtd/transitSchedule_v1.dtd">`); err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to write DOCTYPE: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	// Create XML encoder
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "\t")

	// Encode the transit schedule
	if err := encoder.Encode(schedule); err != nil {
		runtime.EventsEmit(e.app.ctx, "export:error", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to encode transit schedule: %w", err)
	}

	// Add final newline
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	// Emit completion
	runtime.EventsEmit(e.app.ctx, "export:complete", map[string]interface{}{
		"total": totalElements,
	})

	return nil
}

func (e *PTExporter) buildTransitStops(processed *int, total int) ([]TransitStop, error) {
	// Get all stops by collecting from all routes
	uniqueStops := make(map[string]TransitStop)
	
	// Get all transport modes
	modes := []TransportMode{"BUS", "METRO", "TRAM"}
	for _, mode := range modes {
		lines, err := e.app.GetPTLinesByMode(e.processID, mode)
		if err != nil {
			continue
		}
		
		for _, line := range lines {
			routes, err := e.app.GetPTRoutesByLineID(e.processID, line.ID)
			if err != nil {
				continue
			}
			
			for _, route := range routes {
				routeStops, err := e.app.GetPTStopsByRouteID(e.processID, route.ID)
				if err != nil {
					continue
				}
				
				// Get actual stop details for each route stop
				var stopIDs []string
				for _, rs := range routeStops {
					stopIDs = append(stopIDs, rs.StopID)
				}
				
				if len(stopIDs) > 0 {
					stops, err := e.app.GetPTStops(e.processID, stopIDs)
					if err == nil {
						for _, stop := range stops {
							if _, exists := uniqueStops[stop.StopID]; !exists {
								uniqueStops[stop.StopID] = TransitStop{
									ID:   stop.StopID,
									Lng:  stop.Lng,
									Lat:  stop.Lat,
									Name: stop.StopName,
								}
							}
						}
					}
				}
			}
		}
	}


	// Build transit stops from unique map
	var transitStops []TransitStop
	for _, transitStop := range uniqueStops {
		transitStops = append(transitStops, transitStop)
		
		*processed++
		// Emit progress update
		runtime.EventsEmit(e.app.ctx, "export:progress", map[string]interface{}{
			"current": *processed,
			"total":   total,
		})
	}

	return transitStops, nil
}

func (e *PTExporter) buildTransitLines(processed *int, total int) ([]TransitLine, error) {
	// Get all lines for all modes
	var allLines []Line
	modes := []TransportMode{"BUS", "METRO", "TRAM"}
	for _, mode := range modes {
		lines, err := e.app.GetPTLinesByMode(e.processID, mode)
		if err == nil {
			allLines = append(allLines, lines...)
		}
	}

	// Build transit lines
	var transitLines []TransitLine
	for _, line := range allLines {
		transitLine, err := e.buildTransitLine(&line, processed, total)
		if err != nil {
			return nil, fmt.Errorf("failed to build line %s: %w", line.ID, err)
		}
		transitLines = append(transitLines, *transitLine)
		
		*processed++
		// Emit progress update
		runtime.EventsEmit(e.app.ctx, "export:progress", map[string]interface{}{
			"current": *processed,
			"total":   total,
		})
	}

	return transitLines, nil
}

func (e *PTExporter) buildTransitLine(line *Line, processed *int, total int) (*TransitLine, error) {
	// Get routes for this line
	routes, err := e.app.GetPTRoutesByLineID(e.processID, line.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}

	// Build transit routes
	var transitRoutes []TransitRoute
	for _, route := range routes {
		transitRoute, err := e.buildTransitRoute(&route, line.Type, processed, total)
		if err != nil {
			return nil, fmt.Errorf("failed to build route %s: %w", route.ID, err)
		}
		transitRoutes = append(transitRoutes, *transitRoute)
		
		*processed++
		// Emit progress update
		runtime.EventsEmit(e.app.ctx, "export:progress", map[string]interface{}{
			"current": *processed,
			"total":   total,
		})
	}

	return &TransitLine{
		ID:            line.ID,
		TransitRoutes: transitRoutes,
	}, nil
}

func (e *PTExporter) buildTransitRoute(route *Route, transportMode string, processed *int, total int) (*TransitRoute, error) {
	// Build route profile
	routeProfile, err := e.buildRouteProfile(route.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to build route profile: %w", err)
	}

	// Build departures
	departures, err := e.buildDepartures(route.ID, processed, total)
	if err != nil {
		return nil, fmt.Errorf("failed to build departures: %w", err)
	}

	return &TransitRoute{
		ID:            route.ID,
		TransportMode: transportMode,
		RouteProfile:  *routeProfile,
		Departures:    departures,
	}, nil
}

func (e *PTExporter) buildRouteProfile(routeID string) (*RouteProfile, error) {
	// Get route stops
	routeStops, err := e.app.GetPTStopsByRouteID(e.processID, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get route stops: %w", err)
	}

	// Build route stops
	var stops []XMLRouteStop
	for _, rs := range routeStops {
		// Get stop details
		stopIDs := []string{rs.StopID}
		stopDetails, _ := e.app.GetPTStops(e.processID, stopIDs)
		
		arrivalOffset := ""
		departureOffset := ""
		if len(stopDetails) > 0 {
			arrivalOffset = stopDetails[0].ArrivalOffset
			departureOffset = stopDetails[0].DepartureOffset
		}
		
		stop := XMLRouteStop{
			RefID:           rs.StopID,
			ArrivalOffset:   arrivalOffset,
			DepartureOffset: departureOffset,
		}
		stops = append(stops, stop)
	}

	return &RouteProfile{
		Stops: stops,
	}, nil
}


func (e *PTExporter) buildDepartures(routeID string, processed *int, total int) ([]Departure, error) {
	// Get departures
	departures, err := e.app.GetPTDeparturesByRouteID(e.processID, routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get departures: %w", err)
	}

	// Track progress
	for range departures {
		*processed++
		// Emit progress update
		runtime.EventsEmit(e.app.ctx, "export:progress", map[string]interface{}{
			"current": *processed,
			"total":   total,
		})
	}

	return departures, nil
}

// ExportPTSubset exports a subset of PT data based on a bounding box
func (a *App) ExportPTSubset(processID int, bbox BoundingBox, outputPath string) error {
	exporter := NewPTExporter(a, processID)
	return exporter.ExportSubset(bbox, outputPath)
}

// ExportSubset exports only stops and related lines within a bounding box
func (e *PTExporter) ExportSubset(bbox BoundingBox, outputPath string) error {
	// Get stops within bounding box using viewport bounds
	viewport := ViewportBounds{
		MinLat: bbox.South,
		MaxLat: bbox.North,
		MinLng: bbox.West,
		MaxLng: bbox.East,
	}
	
	// Get stops for all modes within bounds
	var allStops []Stop
	modes := []string{"BUS", "METRO", "TRAM"}
	for _, mode := range modes {
		stops, err := e.app.GetPTStopsInBounds(e.processID, mode, viewport)
		if err == nil {
			allStops = append(allStops, stops...)
		}
	}
	
	// Filter stops by bounding box
	var stops []Stop
	for _, stop := range allStops {
		if stop.Lng >= bbox.West && stop.Lng <= bbox.East && stop.Lat >= bbox.South && stop.Lat <= bbox.North {
			stops = append(stops, stop)
		}
	}

	// Create a set of stop IDs for quick lookup
	stopIDSet := make(map[string]bool)
	for _, stop := range stops {
		stopIDSet[stop.StopID] = true
	}

	// Find lines and routes that use these stops
	relevantLines := make(map[string]bool)
	relevantRoutes := make(map[string]bool)

	// Get all lines for all modes
	var allLines []Line
	for _, mode := range []TransportMode{"BUS", "METRO", "TRAM"} {
		lines, err := e.app.GetPTLinesByMode(e.processID, mode)
		if err == nil {
			allLines = append(allLines, lines...)
		}
	}

	// Check each line's routes
	for _, line := range allLines {
		routes, err := e.app.GetPTRoutesByLineID(e.processID, line.ID)
		if err != nil {
			continue
		}

		for _, route := range routes {
			routeStops, err := e.app.GetPTStopsByRouteID(e.processID, route.ID)
			if err != nil {
				continue
			}

			// Check if any stop in the route is in our bbox
			for _, rs := range routeStops {
				if stopIDSet[rs.StopID] {
					relevantLines[line.ID] = true
					relevantRoutes[route.ID] = true
					break
				}
			}
		}
	}

	// Build transit schedule for subset
	schedule := &TransitSchedule{}

	// Build transit stops from filtered stops
	var transitStops []TransitStop
	for _, stop := range stops {
		transitStop := TransitStop{
			ID:   stop.StopID,
			Lng:  stop.Lng,
			Lat:  stop.Lat,
			Name: stop.StopName,
		}
		transitStops = append(transitStops, transitStop)
	}
	schedule.TransitStops = transitStops

	// Build transit lines for subset
	var transitLines []TransitLine
	for _, line := range allLines {
		if relevantLines[line.ID] {
			transitLine, err := e.buildTransitLineSubset(&line, relevantRoutes)
			if err != nil {
				return fmt.Errorf("failed to build line subset %s: %w", line.ID, err)
			}
			transitLines = append(transitLines, *transitLine)
		}
	}
	schedule.TransitLines = transitLines

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write XML declaration
	if _, err := file.WriteString(`<?xml version="1.0" encoding="utf-8"?>`); err != nil {
		return err
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	// Write DOCTYPE
	if _, err := file.WriteString(`<!DOCTYPE transitSchedule SYSTEM "http://www.matsim.org/files/dtd/transitSchedule_v1.dtd">`); err != nil {
		return err
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	// Create XML encoder
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "\t")

	// Encode the transit schedule
	if err := encoder.Encode(schedule); err != nil {
		return fmt.Errorf("failed to encode transit schedule: %w", err)
	}

	// Add final newline
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	return nil
}

func (e *PTExporter) buildTransitLineSubset(line *Line, relevantRoutes map[string]bool) (*TransitLine, error) {
	// Get routes for this line
	routes, err := e.app.GetPTRoutesByLineID(e.processID, line.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}

	// Build only relevant routes
	var transitRoutes []TransitRoute
	dummyProcessed := 0
	for _, route := range routes {
		if relevantRoutes[route.ID] {
			transitRoute, err := e.buildTransitRoute(&route, line.Type, &dummyProcessed, 0)
			if err != nil {
				return nil, fmt.Errorf("failed to build route %s: %w", route.ID, err)
			}
			transitRoutes = append(transitRoutes, *transitRoute)
		}
	}

	return &TransitLine{
		ID:            line.ID,
		TransitRoutes: transitRoutes,
	}, nil
}