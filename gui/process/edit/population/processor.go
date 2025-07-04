package population

import (
	"bytes"
	"database/sql"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"ez-utils/gui/database"
)

// ProcessTelemetry is now defined in the database package to avoid circular dependencies

// CountingReader wraps an io.Reader and counts bytes read
type CountingReader struct {
	reader    io.Reader
	bytesRead int64
	mu        sync.Mutex
}

func (cr *CountingReader) Read(p []byte) (n int, err error) {
	n, err = cr.reader.Read(p)
	cr.mu.Lock()
	cr.bytesRead += int64(n)
	cr.mu.Unlock()
	return n, err
}

func (cr *CountingReader) BytesRead() int64 {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	return cr.bytesRead
}

// Processor handles population file processing with telemetry
type Processor struct {
	db                      *sql.DB
	processID               int
	telemetry               *database.ProcessTelemetry
	telemetryMutex          sync.Mutex
	lastUpdate              time.Time
	personCounter           int64
	errorCounter            int64
	telemetryFailureCounter int64
}

// NewProcessor creates a new population processor
func NewProcessor(processID int) (*Processor, error) {
	db, err := database.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	return &Processor{
		db:        db,
		processID: processID,
		telemetry: &database.ProcessTelemetry{
			ProcessID:   processID,
			ErrorCount:  0,
			LastUpdated: time.Now(),
		},
	}, nil
}

// ProcessPopulationFile processes a population XML file with telemetry tracking
func (p *Processor) ProcessPopulationFile(filePath string) error {
	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	p.telemetry.TotalFileSize = fileInfo.Size()

	// Initialize telemetry in database
	if err := database.InitializeTelemetry(p.processID, p.telemetry.TotalFileSize); err != nil {
		return fmt.Errorf("failed to initialize telemetry: %w", err)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create counting reader
	countingReader := &CountingReader{reader: file}

	// Create table
	tableName := fmt.Sprintf("population_data_%d", p.processID)
	if err := database.CreateTable(tableName); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Start telemetry ticker
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Channel to signal processing completion
	done := make(chan error, 1)

	// Start telemetry updater in goroutine
	go func() {
		for {
			select {
			case <-ticker.C:
				p.updateTelemetry(countingReader)
			case <-done:
				// Final telemetry update
				p.updateTelemetry(countingReader)
				return
			}
		}
	}()

	// Process XML
	err = p.processXML(countingReader, tableName)
	done <- err

	if err != nil {
		return fmt.Errorf("failed to process XML: %w", err)
	}

	// Log telemetry failure statistics if any
	p.telemetryMutex.Lock()
	failCount := p.telemetryFailureCounter
	p.telemetryMutex.Unlock()
	
	if failCount > 0 {
		log.Printf("Process %d completed with %d telemetry update failures", p.processID, failCount)
	}

	return nil
}

func (p *Processor) processXML(reader io.Reader, tableName string) error {
	decoder := xml.NewDecoder(reader)
	batch := make([]database.PersonData, 0, 100)
	var personBuffer bytes.Buffer
	personEncoder := xml.NewEncoder(&personBuffer)
	personDepth := 0

	var homeX, homeY float64
	var foundCoords bool

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode XML token: %w", err)
		}

		if personDepth > 0 {
			personEncoder.EncodeToken(token)

			if !foundCoords {
				if se, ok := token.(xml.StartElement); ok && se.Name.Local == "activity" {
					for _, attr := range se.Attr {
						switch attr.Name.Local {
						case "x":
							homeX, _ = strconv.ParseFloat(attr.Value, 64)
						case "y":
							homeY, _ = strconv.ParseFloat(attr.Value, 64)
						}
					}
					if homeX != 0 || homeY != 0 {
						foundCoords = true
					}
				}
			}
		}

		switch se := token.(type) {
		case xml.StartElement:
			if se.Name.Local == "person" {
				if personDepth == 0 {
					personBuffer.Reset()
					personEncoder.EncodeToken(se)
					homeX, homeY, foundCoords = 0, 0, false
				}
				personDepth++
			}
		case xml.EndElement:
			if se.Name.Local == "person" {
				personDepth--
				if personDepth == 0 {
					personEncoder.Flush()

					id, err := extractAttribute(personBuffer.Bytes(), "person", "id")
					if err != nil {
						log.Printf("could not extract id for person, skipping: %v", err)
						p.telemetryMutex.Lock()
						p.errorCounter++
						p.telemetryMutex.Unlock()
						continue
					}

					coords := fmt.Sprintf("%f,%f", homeX, homeY)

					batch = append(batch, database.PersonData{
						ID:     id,
						Coords: coords,
						RawXML: personBuffer.String(),
					})

					p.telemetryMutex.Lock()
					p.personCounter++
					shouldUpdate := p.personCounter%100 == 0
					p.telemetryMutex.Unlock()

					if len(batch) >= 100 {
						if err := p.insertPersonBatch(tableName, batch); err != nil {
							return fmt.Errorf("failed to insert batch: %w", err)
						}
						batch = make([]database.PersonData, 0, 100)
					}

					if shouldUpdate {
						if cr, ok := reader.(*CountingReader); ok {
							p.updateTelemetry(cr)
						}
					}
				}
			}
		}
	}

	// Insert final batch
	if len(batch) > 0 {
		if err := p.insertPersonBatch(tableName, batch); err != nil {
			return fmt.Errorf("failed to insert final batch: %w", err)
		}
	}

	return nil
}

func (p *Processor) insertPersonBatch(tableName string, persons []database.PersonData) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT OR REPLACE INTO %s (id, coords, raw_xml) VALUES (?, ?, ?)", tableName))
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, person := range persons {
		if _, err := stmt.Exec(person.ID, person.Coords, person.RawXML); err != nil {
			return fmt.Errorf("failed to execute statement for person %s: %w", person.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (p *Processor) updateTelemetry(reader *CountingReader) {
	p.telemetryMutex.Lock()
	p.telemetry.BytesRead = reader.BytesRead()
	p.telemetry.PersonsExtracted = p.personCounter
	p.telemetry.ErrorCount = p.errorCounter
	p.telemetry.LastUpdated = time.Now()
	telemetry := *p.telemetry
	p.telemetryMutex.Unlock()

	// Try to update telemetry with one retry
	err := database.UpdateTelemetry(telemetry.ProcessID, telemetry.BytesRead, telemetry.PersonsExtracted, telemetry.ErrorCount)
	if err != nil {
		log.Printf("Failed to update telemetry for process %d (attempt 1): %v", p.processID, err)
		
		// Wait briefly and retry once
		time.Sleep(100 * time.Millisecond)
		err = database.UpdateTelemetry(telemetry.ProcessID, telemetry.BytesRead, telemetry.PersonsExtracted, telemetry.ErrorCount)
		if err != nil {
			p.telemetryMutex.Lock()
			p.telemetryFailureCounter++
			failCount := p.telemetryFailureCounter
			p.telemetryMutex.Unlock()
			
			log.Printf("Failed to update telemetry for process %d (attempt 2): %v. Total failures: %d", p.processID, err, failCount)
		} else {
			log.Printf("Telemetry update succeeded on retry for process %d", p.processID)
		}
	}
}

func extractAttribute(xmlData []byte, elementName, attrName string) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", err
		}
		if se, ok := token.(xml.StartElement); ok && se.Name.Local == elementName {
			for _, attr := range se.Attr {
				if attr.Name.Local == attrName {
					return attr.Value, nil
				}
			}
		}
	}
}