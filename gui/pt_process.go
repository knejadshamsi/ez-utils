package gui

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// PT XML structures
type TransitSchedule struct {
	TransitStops []TransitStop `xml:"transitStops>stopFacility"`
	TransitLines []TransitLine `xml:"transitLines>transitLine"`
}

type TransitStop struct {
	ID   string  `xml:"id,attr" json:"id"`
	Lng  float64 `xml:"x,attr" json:"lng"`
	Lat  float64 `xml:"y,attr" json:"lat"`
	Name string  `xml:"name,attr" json:"name"`
}

type TransitLine struct {
	ID            string         `xml:"id,attr" json:"id"`
	TransitRoutes []TransitRoute `xml:"transitRoute" json:"transitRoutes"`
}

type TransitRoute struct {
	ID            string       `xml:"id,attr" json:"id"`
	TransportMode string       `xml:"transportMode" json:"transportMode"`
	RouteProfile  RouteProfile `xml:"routeProfile" json:"routeProfile"`
	Departures    []Departure  `xml:"departures>departure" json:"departures"`
}

type RouteProfile struct {
	Stops []RouteStop `xml:"stop" json:"stops"`
}

type RouteStop struct {
	RefID           string `xml:"refId,attr" json:"refId"`
	ArrivalOffset   string `xml:"arrivalOffset,attr" json:"arrivalOffset"`
	DepartureOffset string `xml:"departureOffset,attr" json:"departureOffset"`
}


// ProcessPTFile is the main entry point for PT file processing
func (a *App) ProcessPTFile(filePath string) (*ProcessResult, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// Create process record synchronously to get real process ID
	result, err := a.db.execQuery(
		createProcessQuery,
		"failed to create process record",
		filePath,
	)
	if err != nil {
		return nil, err
	}

	processID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get process ID: %w", err)
	}

	// Ensure telemetry table has PT columns before creating any records
	if err := a.db.AlterTelemetryTableForPT(); err != nil {
		a.db.execQuery(
			"UPDATE processes SET status = ? WHERE process_id = ?",
			"", "PROCESSING_FAILED", processID,
		)
		return nil, fmt.Errorf("failed to update telemetry table: %w", err)
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		a.db.execQuery(
			"UPDATE processes SET status = ? WHERE process_id = ?",
			"", "PROCESSING_FAILED", processID,
		)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create telemetry record with PT columns available
	_, err = a.db.execQuery(
		"INSERT INTO process_telemetry (process_id, total_file_size, stops_extracted, lines_extracted, routes_extracted) VALUES (?, ?, 0, 0, 0)",
		"failed to create telemetry record",
		processID, fileInfo.Size(),
	)
	if err != nil {
		a.db.execQuery(
			"UPDATE processes SET status = ? WHERE process_id = ?",
			"", "PROCESSING_FAILED", processID,
		)
		return nil, err
	}

	// Create PT tables synchronously
	if err := a.db.CreatePTTables(int(processID)); err != nil {
		a.db.execQuery(
			"UPDATE processes SET status = ? WHERE process_id = ?",
			"", "PROCESSING_FAILED", processID,
		)
		return nil, err
	}

	// Start async processing with real process ID
	go a.processPTFileAsync(int(processID), filePath, fileInfo.Size())

	return &ProcessResult{
		ProcessID: int(processID),
		Message:   "PT file processing started",
	}, nil
}

func (a *App) processPTFileAsync(processID int, filePath string, fileSize int64) {
	processor := &PTProcessor{
		processID: processID,
		db:        a.db,
		app:       a,
		telemetry: &PTTelemetry{
			ProcessID:     processID,
			TotalFileSize: fileSize,
		},
	}

	// Open file and create counting reader BEFORE starting telemetry
	file, err := os.Open(filePath)
	if err != nil {
		a.db.execQuery(
			"UPDATE processes SET status = ? WHERE process_id = ?",
			"",
			"PROCESSING_FAILED", processID,
		)
		return
	}
	defer file.Close()

	countingReader := &CountingReader{
		reader: file,
	}

	// Start telemetry updates with counting reader
	done := make(chan bool)
	go processor.updateTelemetry(countingReader, done)

	// Process the file
	err = processor.process(countingReader)

	// Stop telemetry updates
	done <- true

	// Update final status
	status := "PROCESSING_SUCCESS"
	if err != nil {
		status = "PROCESSING_FAILED"
	}

	if _, updateErr := a.db.execQuery(
		"UPDATE processes SET status = ? WHERE process_id = ?",
		"failed to update process status",
		status, processID,
	); updateErr != nil {
	}
	
	wailsruntime.EventsEmit(a.ctx, "process:status", map[string]interface{}{
		"processId": processID,
		"status":    status,
	})
}

func (p *PTProcessor) process(countingReader *CountingReader) error {
	decoder := xml.NewDecoder(countingReader)

	// Batch processing
	const batchSize = 100
	var lineBatch []Line
	var stopBatch []Stop
	var departureBatch []Departure

	// Temporary storage for building complete structures
	var currentLine *TransitLine
	var currentRoute *TransitRoute
	var lineRoutes = make(map[string][]Route) // Map line ID to routes
	var routeStops = make(map[string][]RouteStop) // Map route ID to stops
	var stopDetails = make(map[string]TransitStop) // Map stop ID to stop details
	var routeStopOrder int
	var currentTransportMode string

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			atomic.AddInt64(&p.errorCount, 1)
			continue
		}

		switch se := token.(type) {
		case xml.StartElement:
			switch se.Name.Local {
			case "stopFacility":
				var stop TransitStop
				// Extract attributes
				for _, attr := range se.Attr {
					switch attr.Name.Local {
					case "id":
						stop.ID = attr.Value
					case "x":
						stop.Lng, _ = strconv.ParseFloat(attr.Value, 64)
					case "y":
						stop.Lat, _ = strconv.ParseFloat(attr.Value, 64)
					case "name":
						stop.Name = attr.Value
					}
				}

				// Skip element content
				depth := 1
				for depth > 0 {
					token, err := decoder.Token()
					if err != nil {
						break
					}
					switch token.(type) {
					case xml.StartElement:
						depth++
					case xml.EndElement:
						depth--
					}
				}

				// Store stop details for later use
				if stop.ID != "" {
					stopDetails[stop.ID] = stop
					atomic.AddInt64(&p.stopCount, 1)
				}

			case "transitLine":
				currentLine = &TransitLine{}
				for _, attr := range se.Attr {
					if attr.Name.Local == "id" {
						currentLine.ID = attr.Value
						break
					}
				}

			case "transitRoute":
				if currentLine != nil {
					currentRoute = &TransitRoute{}
					routeStopOrder = 0
					currentTransportMode = "" // Reset for this route
					for _, attr := range se.Attr {
						if attr.Name.Local == "id" {
							currentRoute.ID = attr.Value
							break
						}
					}
				}

			case "stop":
				if currentRoute != nil {
					var routeStop RouteStop
					for _, attr := range se.Attr {
						switch attr.Name.Local {
						case "refId":
							routeStop.RefID = attr.Value
						case "arrivalOffset":
							routeStop.ArrivalOffset = attr.Value
						case "departureOffset":
							routeStop.DepartureOffset = attr.Value
						}
					}
					if routeStop.RefID != "" {
						// Store route stops temporarily
						if routeStops[currentRoute.ID] == nil {
							routeStops[currentRoute.ID] = []RouteStop{}
						}
						routeStops[currentRoute.ID] = append(routeStops[currentRoute.ID], routeStop)
						routeStopOrder++
					}
				}

			case "departure":
				if currentRoute != nil {
					var departure Departure
					for _, attr := range se.Attr {
						switch attr.Name.Local {
						case "id":
							departure.ID = attr.Value
						case "departureTime":
							departure.DepartureTime = attr.Value
						}
					}
					if departure.ID != "" {
						departure.RouteID = currentRoute.ID
						departureBatch = append(departureBatch, departure)
					}
				}

			case "transportMode":
				if currentRoute != nil {
					// Read the transportMode text content
					for {
						token, err := decoder.Token()
						if err != nil {
							break
						}
						switch t := token.(type) {
						case xml.CharData:
							currentTransportMode = strings.ToUpper(strings.TrimSpace(string(t)))
						case xml.EndElement:
							if t.Name.Local == "transportMode" {
								goto transportModeDone
							}
						}
					}
				transportModeDone:
				}
			}

		case xml.EndElement:
			switch se.Name.Local {
			case "transitRoute":
				if currentLine != nil && currentRoute != nil {
					// Store route in temporary map
					stops := routeStops[currentRoute.ID]
					route := Route{
						ID:    currentRoute.ID,
						Name:  currentRoute.ID, // Default to ID, can be enhanced later
						Stops: len(stops),
					}
					
					if lineRoutes[currentLine.ID] == nil {
						lineRoutes[currentLine.ID] = []Route{}
					}
					lineRoutes[currentLine.ID] = append(lineRoutes[currentLine.ID], route)
					
					// Create Stop entries for this route
					for i, rs := range stops {
						if stopDetail, exists := stopDetails[rs.RefID]; exists {
							stop := Stop{
								RouteID:         currentRoute.ID,
								StopID:          rs.RefID,
								ArrivalOffset:   rs.ArrivalOffset,
								DepartureOffset: rs.DepartureOffset,
								StopName:        stopDetail.Name,
								Lat:             stopDetail.Lat,
								Lng:             stopDetail.Lng,
								Sequence:        i,
							}
							stopBatch = append(stopBatch, stop)
							
							if len(stopBatch) >= batchSize {
								if err := p.insertStopBatch(stopBatch); err != nil {
									atomic.AddInt64(&p.errorCount, 1)
									return fmt.Errorf("failed to insert stop batch: %w", err)
								}
								stopBatch = stopBatch[:0]
							}
						}
					}
					
					// Store departures temporarily
					if len(departureBatch) > 0 {
						if err := p.insertDepartureBatch(departureBatch); err != nil {
							atomic.AddInt64(&p.errorCount, 1)
							return fmt.Errorf("failed to insert departure batch: %w", err)
						}
						departureBatch = departureBatch[:0]
					}
					
					atomic.AddInt64(&p.routeCount, 1)
					currentRoute = nil
				}

			case "transitLine":
				if currentLine != nil {
					// Determine transport mode (default to BUS)
					mode := currentTransportMode
					if mode == "" {
						mode = "BUS"
					}
					
					// Create Line with embedded routes
					routes := lineRoutes[currentLine.ID]
					if routes == nil {
						routes = []Route{}
					}
					
					line := Line{
						ID:     currentLine.ID,
						Name:   currentLine.ID, // Default to ID, can be enhanced later
						Type:   mode,
						Routes: routes,
					}
					
					lineBatch = append(lineBatch, line)
					atomic.AddInt64(&p.lineCount, 1)

					if len(lineBatch) >= batchSize {
						if err := p.insertLineBatch(lineBatch); err != nil {
							atomic.AddInt64(&p.errorCount, 1)
							return fmt.Errorf("failed to insert line batch: %w", err)
						}
						lineBatch = lineBatch[:0]
					}

					currentLine = nil
				}
			}
		}
	}

	// Insert remaining batches
	if len(lineBatch) > 0 {
		if err := p.insertLineBatch(lineBatch); err != nil {
			atomic.AddInt64(&p.errorCount, 1)
			return fmt.Errorf("failed to insert final line batch: %w", err)
		}
	}
	if len(stopBatch) > 0 {
		if err := p.insertStopBatch(stopBatch); err != nil {
			atomic.AddInt64(&p.errorCount, 1)
			return fmt.Errorf("failed to insert final stop batch: %w", err)
		}
	}
	if len(departureBatch) > 0 {
		if err := p.insertDepartureBatch(departureBatch); err != nil {
			atomic.AddInt64(&p.errorCount, 1)
			return fmt.Errorf("failed to insert final departure batch: %w", err)
		}
	}

	return nil
}

// captureRawXML is no longer needed - we'll capture XML manually during parsing

func (p *PTProcessor) captureElementStart(start *xml.StartElement) string {
	var rawXML strings.Builder
	rawXML.WriteString("<")
	rawXML.WriteString(start.Name.Local)

	for _, attr := range start.Attr {
		rawXML.WriteString(fmt.Sprintf(` %s="%s"`, attr.Name.Local, attr.Value))
	}
	rawXML.WriteString(">")

	return rawXML.String()
}

func (p *PTProcessor) updateTelemetry(countingReader *CountingReader, done chan bool) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			p.sendTelemetryUpdate(countingReader)
			return
		case <-ticker.C:
			p.sendTelemetryUpdate(countingReader)
		}
	}
}

func (p *PTProcessor) sendTelemetryUpdate(countingReader *CountingReader) {
	bytesRead := int64(0)
	if countingReader != nil {
		bytesRead = countingReader.BytesRead()
	}

	p.telemetryMutex.RLock()
	telemetry := &PTTelemetry{
		ProcessID:       p.telemetry.ProcessID,
		TotalFileSize:   p.telemetry.TotalFileSize,
		BytesRead:       bytesRead,
		StopsExtracted:  int(atomic.LoadInt64(&p.stopCount)),
		LinesExtracted:  int(atomic.LoadInt64(&p.lineCount)),
		RoutesExtracted: int(atomic.LoadInt64(&p.routeCount)),
		ErrorCount:      int(atomic.LoadInt64(&p.errorCount)),
		LastUpdated:     time.Now().Format(time.RFC3339),
	}
	p.telemetryMutex.RUnlock()

	// Update database using the dedicated method
	if err := p.db.UpdatePTTelemetry(
		telemetry.ProcessID,
		telemetry.BytesRead,
		telemetry.StopsExtracted,
		telemetry.LinesExtracted,
		telemetry.RoutesExtracted,
		telemetry.ErrorCount,
	); err != nil {
		return
	}

	// Emit events
	wailsruntime.EventsEmit(p.app.ctx, "pt-telemetry-update", telemetry)
	wailsruntime.EventsEmit(p.app.ctx, "process:telemetry", map[string]interface{}{
		"processId":        telemetry.ProcessID,
		"bytesRead":        telemetry.BytesRead,
		"stopsExtracted":   telemetry.StopsExtracted,
		"linesExtracted":   telemetry.LinesExtracted,
		"routesExtracted":  telemetry.RoutesExtracted,
		"errorCount":       telemetry.ErrorCount,
		"lastUpdated":      telemetry.LastUpdated,
	})
}

// Batch insert methods
func (p *PTProcessor) insertStopBatch(stops []Stop) error {
	return p.db.InsertPTStopBatch(p.processID, stops)
}

func (p *PTProcessor) insertLineBatch(lines []Line) error {
	return p.db.InsertPTLineBatch(p.processID, lines)
}

func (p *PTProcessor) insertDepartureBatch(departures []Departure) error {
	return p.db.InsertPTDepartureBatch(p.processID, departures)
}

// GetPTTelemetry retrieves telemetry data for a PT process
func (a *App) GetPTTelemetry(processID int) (*PTTelemetry, error) {
	return a.db.GetPTTelemetry(processID)
}
