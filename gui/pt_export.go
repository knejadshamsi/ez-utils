package gui

import (
	"fmt"
	"os"
)


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
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write XML header
	if _, err := file.WriteString(`<?xml version="1.0" encoding="utf-8"?>`); err != nil {
		return fmt.Errorf("failed to write XML header: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	// Write DOCTYPE if needed (MATSim specific)
	if _, err := file.WriteString(`<!DOCTYPE transitSchedule SYSTEM "http://www.matsim.org/files/dtd/transitSchedule_v1.dtd">`); err != nil {
		return fmt.Errorf("failed to write DOCTYPE: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	// Start root element
	if _, err := file.WriteString(`<transitSchedule>`); err != nil {
		return fmt.Errorf("failed to write root element: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	// Export transit stops
	if err := e.exportTransitStops(file); err != nil {
		return fmt.Errorf("failed to export transit stops: %w", err)
	}

	// Export transit lines
	if err := e.exportTransitLines(file); err != nil {
		return fmt.Errorf("failed to export transit lines: %w", err)
	}

	// Close root element
	if _, err := file.WriteString(`</transitSchedule>`); err != nil {
		return fmt.Errorf("failed to close root element: %w", err)
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	return nil
}

func (e *PTExporter) exportTransitStops(file *os.File) error {
	// Write transitStops opening tag
	if _, err := file.WriteString("\t<transitStops>\n"); err != nil {
		return err
	}

	// Get all stops
	stops, err := e.app.GetPTStops(e.processID)
	if err != nil {
		return fmt.Errorf("failed to get stops: %w", err)
	}

	// Write each stop
	for _, stop := range stops {
		// Use raw XML if available, otherwise construct it
		if stop.RawXML != "" {
			if _, err := file.WriteString("\t\t" + stop.RawXML + "\n"); err != nil {
				return err
			}
		} else {
			// Construct XML manually
			stopXML := fmt.Sprintf(`<stopFacility id="%s" x="%.6f" y="%.6f"`,
				escapeXML(stop.ID), stop.X, stop.Y)
			
			if stop.Name != "" {
				stopXML += fmt.Sprintf(` name="%s"`, escapeXML(stop.Name))
			}
			stopXML += "/>"
			
			if _, err := file.WriteString("\t\t" + stopXML + "\n"); err != nil {
				return err
			}
		}
	}

	// Write transitStops closing tag
	if _, err := file.WriteString("\t</transitStops>\n"); err != nil {
		return err
	}

	return nil
}

func (e *PTExporter) exportTransitLines(file *os.File) error {
	// Write transitLines opening tag
	if _, err := file.WriteString("\t<transitLines>\n"); err != nil {
		return err
	}

	// Get all lines
	lines, err := e.app.GetPTLines(e.processID)
	if err != nil {
		return fmt.Errorf("failed to get lines: %w", err)
	}

	// Process each line
	for _, line := range lines {
		if err := e.exportTransitLine(file, &line); err != nil {
			return fmt.Errorf("failed to export line %s: %w", line.ID, err)
		}
	}

	// Write transitLines closing tag
	if _, err := file.WriteString("\t</transitLines>\n"); err != nil {
		return err
	}

	return nil
}

func (e *PTExporter) exportTransitLine(file *os.File, line *PTLine) error {
	// Write line opening tag
	if _, err := file.WriteString(fmt.Sprintf("\t\t<transitLine id=\"%s\">\n", escapeXML(line.ID))); err != nil {
		return err
	}

	// Get routes for this line
	routes, err := e.app.GetPTRoutes(e.processID, line.ID)
	if err != nil {
		return fmt.Errorf("failed to get routes: %w", err)
	}

	// Export each route
	for _, route := range routes {
		if err := e.exportTransitRoute(file, &route); err != nil {
			return fmt.Errorf("failed to export route %s: %w", route.ID, err)
		}
	}

	// Write line closing tag
	if _, err := file.WriteString("\t\t</transitLine>\n"); err != nil {
		return err
	}

	return nil
}

func (e *PTExporter) exportTransitRoute(file *os.File, route *PTRoute) error {
	// Write route opening tag
	if _, err := file.WriteString(fmt.Sprintf("\t\t\t<transitRoute id=\"%s\">\n", escapeXML(route.ID))); err != nil {
		return err
	}

	// Export route profile (stops)
	if err := e.exportRouteProfile(file, route.ID); err != nil {
		return fmt.Errorf("failed to export route profile: %w", err)
	}

	// Export route (if any additional route data needed)
	if err := e.exportRouteElement(file, route.ID); err != nil {
		return fmt.Errorf("failed to export route element: %w", err)
	}

	// Export departures
	if err := e.exportDepartures(file, route.ID); err != nil {
		return fmt.Errorf("failed to export departures: %w", err)
	}

	// Write route closing tag
	if _, err := file.WriteString("\t\t\t</transitRoute>\n"); err != nil {
		return err
	}

	return nil
}

func (e *PTExporter) exportRouteProfile(file *os.File, routeID string) error {
	// Get route stops
	routeStops, err := e.app.GetPTRouteStops(e.processID, routeID)
	if err != nil {
		return fmt.Errorf("failed to get route stops: %w", err)
	}

	if len(routeStops) == 0 {
		return nil // No stops to export
	}

	// Write routeProfile opening tag
	if _, err := file.WriteString("\t\t\t\t<routeProfile>\n"); err != nil {
		return err
	}

	// Write each stop
	for _, rs := range routeStops {
		stopXML := fmt.Sprintf(`<stop refId="%s"`, escapeXML(rs.StopRefID))
		
		if rs.ArrivalOffset != "" {
			stopXML += fmt.Sprintf(` arrivalOffset="%s"`, rs.ArrivalOffset)
		}
		if rs.DepartureOffset != "" {
			stopXML += fmt.Sprintf(` departureOffset="%s"`, rs.DepartureOffset)
		}
		stopXML += "/>"
		
		if _, err := file.WriteString("\t\t\t\t\t" + stopXML + "\n"); err != nil {
			return err
		}
	}

	// Write routeProfile closing tag
	if _, err := file.WriteString("\t\t\t\t</routeProfile>\n"); err != nil {
		return err
	}

	return nil
}

func (e *PTExporter) exportRouteElement(file *os.File, routeID string) error {
	// Write empty route element (can be extended if needed)
	if _, err := file.WriteString("\t\t\t\t<route/>\n"); err != nil {
		return err
	}
	return nil
}

func (e *PTExporter) exportDepartures(file *os.File, routeID string) error {
	// Get departures
	departures, err := e.app.GetPTDepartures(e.processID, routeID)
	if err != nil {
		return fmt.Errorf("failed to get departures: %w", err)
	}

	if len(departures) == 0 {
		return nil // No departures to export
	}

	// Write departures opening tag
	if _, err := file.WriteString("\t\t\t\t<departures>\n"); err != nil {
		return err
	}

	// Write each departure
	for _, dep := range departures {
		depXML := fmt.Sprintf(`<departure id="%s" departureTime="%s"/>`,
			escapeXML(dep.ID), dep.DepartureTime)
		
		if _, err := file.WriteString("\t\t\t\t\t" + depXML + "\n"); err != nil {
			return err
		}
	}

	// Write departures closing tag
	if _, err := file.WriteString("\t\t\t\t</departures>\n"); err != nil {
		return err
	}

	return nil
}

// ExportPTSubset exports a subset of PT data based on a bounding box
func (a *App) ExportPTSubset(processID int, bbox BoundingBox, outputPath string) error {
	exporter := NewPTExporter(a, processID)
	return exporter.ExportSubset(bbox, outputPath)
}

// ExportSubset exports only stops and related lines within a bounding box
func (e *PTExporter) ExportSubset(bbox BoundingBox, outputPath string) error {
	// Get all stops and filter by bounding box
	allStops, err := e.app.GetPTStops(e.processID)
	if err != nil {
		return fmt.Errorf("failed to get stops: %w", err)
	}
	
	// Filter stops by bounding box
	var stops []PTStop
	for _, stop := range allStops {
		if stop.X >= bbox.West && stop.X <= bbox.East && stop.Y >= bbox.South && stop.Y <= bbox.North {
			stops = append(stops, stop)
		}
	}

	// Create a set of stop IDs for quick lookup
	stopIDSet := make(map[string]bool)
	for _, stop := range stops {
		stopIDSet[stop.ID] = true
	}

	// Find lines and routes that use these stops
	relevantLines := make(map[string]bool)
	relevantRoutes := make(map[string]bool)

	// Get all lines
	allLines, err := e.app.GetPTLines(e.processID)
	if err != nil {
		return fmt.Errorf("failed to get lines: %w", err)
	}

	// Check each line's routes
	for _, line := range allLines {
		routes, err := e.app.GetPTRoutes(e.processID, line.ID)
		if err != nil {
			continue
		}

		for _, route := range routes {
			routeStops, err := e.app.GetPTRouteStops(e.processID, route.ID)
			if err != nil {
				continue
			}

			// Check if any stop in the route is in our bbox
			for _, rs := range routeStops {
				if stopIDSet[rs.StopRefID] {
					relevantLines[line.ID] = true
					relevantRoutes[route.ID] = true
					break
				}
			}
		}
	}

	// Now export the subset
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Write XML header and root element
	if _, err := file.WriteString(`<?xml version="1.0" encoding="utf-8"?>`); err != nil {
		return err
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}
	if _, err := file.WriteString(`<!DOCTYPE transitSchedule SYSTEM "http://www.matsim.org/files/dtd/transitSchedule_v1.dtd">`); err != nil {
		return err
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}
	if _, err := file.WriteString(`<transitSchedule>`); err != nil {
		return err
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	// Export subset of stops
	if _, err := file.WriteString("\t<transitStops>\n"); err != nil {
		return err
	}
	for _, stop := range stops {
		stopXML := fmt.Sprintf(`<stopFacility id="%s" x="%.6f" y="%.6f"`,
			escapeXML(stop.ID), stop.X, stop.Y)
		if stop.Name != "" {
			stopXML += fmt.Sprintf(` name="%s"`, escapeXML(stop.Name))
		}
		stopXML += "/>"
		if _, err := file.WriteString("\t\t" + stopXML + "\n"); err != nil {
			return err
		}
	}
	if _, err := file.WriteString("\t</transitStops>\n"); err != nil {
		return err
	}

	// Export subset of lines
	if _, err := file.WriteString("\t<transitLines>\n"); err != nil {
		return err
	}
	for _, line := range allLines {
		if relevantLines[line.ID] {
			if err := e.exportTransitLineSubset(file, &line, relevantRoutes, stopIDSet); err != nil {
				return err
			}
		}
	}
	if _, err := file.WriteString("\t</transitLines>\n"); err != nil {
		return err
	}

	// Close root element
	if _, err := file.WriteString(`</transitSchedule>`); err != nil {
		return err
	}
	if _, err := file.WriteString("\n"); err != nil {
		return err
	}

	return nil
}

func (e *PTExporter) exportTransitLineSubset(file *os.File, line *PTLine, 
	relevantRoutes map[string]bool, stopIDSet map[string]bool) error {
	
	// Write line opening tag
	if _, err := file.WriteString(fmt.Sprintf("\t\t<transitLine id=\"%s\">\n", escapeXML(line.ID))); err != nil {
		return err
	}

	// Get routes for this line
	routes, err := e.app.GetPTRoutes(e.processID, line.ID)
	if err != nil {
		return fmt.Errorf("failed to get routes: %w", err)
	}

	// Export only relevant routes
	for _, route := range routes {
		if relevantRoutes[route.ID] {
			if err := e.exportTransitRoute(file, &route); err != nil {
				return fmt.Errorf("failed to export route %s: %w", route.ID, err)
			}
		}
	}

	// Write line closing tag
	if _, err := file.WriteString("\t\t</transitLine>\n"); err != nil {
		return err
	}

	return nil
}