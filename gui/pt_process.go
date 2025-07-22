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
	X    float64 `xml:"x,attr" json:"x"`
	Y    float64 `xml:"y,attr" json:"y"`
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

type Departure struct {
	ID            string `xml:"id,attr" json:"id"`
	DepartureTime string `xml:"departureTime,attr" json:"departureTime"`
}




// ProcessPTFile is the main entry point for PT file processing
func (a *App) ProcessPTFile(filePath string) (*ProcessResult, error) {
	// Create process record
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

	// Create PT tables
	if err := a.db.CreatePTTables(int(processID)); err != nil {
		a.db.execQuery(
			"UPDATE processes SET status = ? WHERE process_id = ?",
			"", "PROCESSING_FAILED", processID,
		)
		return nil, err
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create telemetry record
	_, err = a.db.execQuery(
		"INSERT INTO process_telemetry (process_id, total_file_size) VALUES (?, ?)",
		"failed to create telemetry record",
		processID, fileInfo.Size(),
	)
	if err != nil {
		return nil, err
	}

	// Update status to INITIALIZING
	a.db.execQuery(
		"UPDATE processes SET status = ? WHERE process_id = ?",
		"failed to update process status",
		"PROCESSING", processID,
	)
	
	// Start async processing
	go a.processPTFileAsync(int(processID), filePath, fileInfo.Size())

	return &ProcessResult{
		ProcessID: int(processID),
		Message:   "PT file processing started",
	}, nil
}

func (a *App) processPTFileAsync(processID int, filePath string, fileSize int64) {
	// Update to PROCESSING
	a.db.execQuery(
		"UPDATE processes SET status = ? WHERE process_id = ?",
		"failed to update process status",
		"PROCESSING", processID,
	)
	
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
		wailsruntime.EventsEmit(a.ctx, "pt-processing-error", map[string]interface{}{
			"processID": processID,
			"error":     err.Error(),
		})
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
		wailsruntime.EventsEmit(a.ctx, "pt-processing-error", map[string]interface{}{
			"processID": processID,
			"error":     err.Error(),
		})
	}

	if _, updateErr := a.db.execQuery(
		"UPDATE processes SET status = ? WHERE process_id = ?",
		"failed to update process status",
		status, processID,
	); updateErr != nil {
	}

	// Emit completion event
	wailsruntime.EventsEmit(a.ctx, "pt-processing-complete", map[string]interface{}{
		"processID": processID,
		"status":    status,
	})
}

func (p *PTProcessor) process(countingReader *CountingReader) error {
	decoder := xml.NewDecoder(countingReader)
	
	// Batch processing
	const batchSize = 100
	var stopBatch []PTStopData
	var lineBatch []PTLineData
	var routeBatch []PTRouteData
	var routeStopBatch []PTRouteStopData
	var departureBatch []PTDepartureData

	var currentLine *TransitLine
	var currentRoute *TransitRoute
	var currentLineRawXML string
	var currentRouteRawXML string
	var routeStopOrder int
	var currentTransportMode string
	var lineTransportModes = make(map[string]string) // Track transport modes by line ID

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
				var stopXML strings.Builder
				stopXML.WriteString(p.captureElementStart(&se))
				
				// Extract attributes
				for _, attr := range se.Attr {
					switch attr.Name.Local {
					case "id":
						stop.ID = attr.Value
					case "x":
						stop.X, _ = strconv.ParseFloat(attr.Value, 64)
					case "y":
						stop.Y, _ = strconv.ParseFloat(attr.Value, 64)
					case "name":
						stop.Name = attr.Value
					}
				}
				
				// Capture inner content until matching end element
				depth := 1
				for depth > 0 {
					token, err := decoder.Token()
					if err != nil {
						break
					}
					
					switch t := token.(type) {
					case xml.StartElement:
						stopXML.WriteString(p.captureElementStart(&t))
						depth++
					case xml.EndElement:
						stopXML.WriteString(fmt.Sprintf("</%s>", t.Name.Local))
						depth--
					case xml.CharData:
						stopXML.Write(t)
					}
				}
				
				if stop.ID != "" {
					stopBatch = append(stopBatch, PTStopData{
						ID:     stop.ID,
						X:      stop.X,
						Y:      stop.Y,
						Name:   stop.Name,
						RawXML: stopXML.String(),
					})
					atomic.AddInt64(&p.stopCount, 1)

					if len(stopBatch) >= batchSize {
						if err := p.insertStopBatch(stopBatch); err != nil {
							atomic.AddInt64(&p.errorCount, 1)
							return fmt.Errorf("failed to insert stop batch: %w", err)
						}
						stopBatch = stopBatch[:0]
					}
				}

			case "transitLine":
				currentLine = &TransitLine{}
				currentLineRawXML = p.captureElementStart(&se)
				for _, attr := range se.Attr {
					if attr.Name.Local == "id" {
						currentLine.ID = attr.Value
						break
					}
				}

			case "transitRoute":
				if currentLine != nil {
					currentRoute = &TransitRoute{}
					currentRouteRawXML = p.captureElementStart(&se)
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
						routeStopBatch = append(routeStopBatch, PTRouteStopData{
							RouteID:         currentRoute.ID,
							StopRefID:       routeStop.RefID,
							StopOrder:       routeStopOrder,
							ArrivalOffset:   routeStop.ArrivalOffset,
							DepartureOffset: routeStop.DepartureOffset,
						})
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
						departureBatch = append(departureBatch, PTDepartureData{
							ID:            departure.ID,
							RouteID:       currentRoute.ID,
							DepartureTime: departure.DepartureTime,
						})
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
							currentTransportMode = strings.TrimSpace(string(t))
							if currentLine != nil {
								// Store the transport mode for this line
								lineTransportModes[currentLine.ID] = strings.ToUpper(currentTransportMode)
							}
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
					// Save route
					routeBatch = append(routeBatch, PTRouteData{
						ID:     currentRoute.ID,
						LineID: currentLine.ID,
						RawXML: currentRouteRawXML + "</transitRoute>",
					})
					atomic.AddInt64(&p.routeCount, 1)
					
					if len(routeBatch) >= batchSize {
						if err := p.insertRouteBatch(routeBatch); err != nil {
							atomic.AddInt64(&p.errorCount, 1)
							return fmt.Errorf("failed to insert route batch: %w", err)
						}
						routeBatch = routeBatch[:0]
					}

					// Insert route stops
					if len(routeStopBatch) > 0 {
						if err := p.insertRouteStopBatch(routeStopBatch); err != nil {
							atomic.AddInt64(&p.errorCount, 1)
							return fmt.Errorf("failed to insert route stop batch: %w", err)
						}
						routeStopBatch = routeStopBatch[:0]
					}

					// Insert departures
					if len(departureBatch) > 0 {
						if err := p.insertDepartureBatch(departureBatch); err != nil {
							atomic.AddInt64(&p.errorCount, 1)
							return fmt.Errorf("failed to insert departure batch: %w", err)
						}
						departureBatch = departureBatch[:0]
					}
					
					currentRoute = nil
				}

			case "transitLine":
				if currentLine != nil {
					// Save line with mode from routes (default to BUS if not found)
					mode := lineTransportModes[currentLine.ID]
					if mode == "" {
						mode = "BUS"
					}
					lineBatch = append(lineBatch, PTLineData{
						ID:     currentLine.ID,
						Mode:   mode,
						RawXML: currentLineRawXML + "</transitLine>",
					})
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
	if len(stopBatch) > 0 {
		if err := p.insertStopBatch(stopBatch); err != nil {
			atomic.AddInt64(&p.errorCount, 1)
			return fmt.Errorf("failed to insert final stop batch: %w", err)
		}
	}
	if len(lineBatch) > 0 {
		if err := p.insertLineBatch(lineBatch); err != nil {
			atomic.AddInt64(&p.errorCount, 1)
			return fmt.Errorf("failed to insert final line batch: %w", err)
		}
	}
	if len(routeBatch) > 0 {
		if err := p.insertRouteBatch(routeBatch); err != nil {
			atomic.AddInt64(&p.errorCount, 1)
			return fmt.Errorf("failed to insert final route batch: %w", err)
		}
	}
	if len(routeStopBatch) > 0 {
		if err := p.insertRouteStopBatch(routeStopBatch); err != nil {
			atomic.AddInt64(&p.errorCount, 1)
			return fmt.Errorf("failed to insert final route stop batch: %w", err)
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

	// Emit event
	wailsruntime.EventsEmit(p.app.ctx, "pt-telemetry-update", telemetry)
}

// Batch insert methods
func (p *PTProcessor) insertStopBatch(stops []PTStopData) error {
	return p.db.InsertPTStopBatch(p.processID, stops)
}

func (p *PTProcessor) insertLineBatch(lines []PTLineData) error {
	return p.db.InsertPTLineBatch(p.processID, lines)
}

func (p *PTProcessor) insertRouteBatch(routes []PTRouteData) error {
	return p.db.InsertPTRouteBatch(p.processID, routes)
}

func (p *PTProcessor) insertRouteStopBatch(routeStops []PTRouteStopData) error {
	return p.db.InsertPTRouteStopBatch(p.processID, routeStops)
}

func (p *PTProcessor) insertDepartureBatch(departures []PTDepartureData) error {
	return p.db.InsertPTDepartureBatch(p.processID, departures)
}

// GetPTTelemetry retrieves telemetry data for a PT process
func (a *App) GetPTTelemetry(processID int) (*PTTelemetry, error) {
	return a.db.GetPTTelemetry(processID)
}