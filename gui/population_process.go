package gui

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"
)


// ProcessPopulationFile starts the processing of a given file path.
func (a *App) ProcessPopulationFile(filePath string) (map[string]any, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	result, err := a.db.execQuery(createProcessQuery, fmt.Sprintf("failed to create process for file %s", filePath), filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create process record: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	processID := int(id)

	go a.processPopulationFile(filePath, processID)

	response := map[string]any{
		"message":   "Processing started",
		"processId": processID,
	}
	return response, nil
}

// processPopulationFile processes a population XML file using the new processor module
func (a *App) processPopulationFile(filePath string, processID int) {
	updateStatus := func(status string) {
		if _, err := a.db.execQuery(updateProcessStatusQuery, fmt.Sprintf("failed to update process status for ID %d", processID), status, processID); err != nil {
			log.Printf("Failed to update process status for processID %d: %v", processID, err)
		}
	}

	updateStatus("INITIALIZING")
	processor, err := NewProcessor(a.db, processID)
	if err != nil {
		log.Printf("Error creating processor: %v", err)
		updateStatus("FAILED")
		return
	}

	updateStatus("PROCESSING")
	if err := processor.ProcessPopulationFile(filePath); err != nil {
		log.Printf("Error processing file %s: %v", filePath, err)
		updateStatus("FAILED")
		return
	}

	updateStatus("COMPLETED")
	log.Printf("Successfully processed file %s", filePath)
}

// GetPopulation retrieves all processed population data for a given table.
func (a *App) GetPopulation(tableName string) ([]Person, error) {
	persons, err := a.db.GetPopulationData(tableName)
	if err != nil {
		return make([]Person, 0), err
	}
	if persons == nil {
		return make([]Person, 0), nil
	}
	return persons, nil
}

// GetPopulationByBbox retrieves population data within a bounding box
func (a *App) GetPopulationByBbox(tableName string, minLat, minLng, maxLat, maxLng float64) ([]Person, error) {
	persons, err := a.db.GetPopulationByBbox(tableName, minLat, minLng, maxLat, maxLng)
	if err != nil {
		return make([]Person, 0), err
	}
	if persons == nil {
		return make([]Person, 0), nil
	}
	return persons, nil
}

// GetPerson retrieves a single person by ID
func (a *App) GetPerson(tableName string, personId string) (*Person, error) {
	return a.db.GetPerson(tableName, personId)
}

// addPerson adds a new person to a population table (internal function for interpreter)
func (a *App) addPerson(tableName string, person map[string]any) (map[string]any, error) {
	// Extract person data
	id, ok := person["id"].(string)
	if !ok || id == "" {
		return nil, fmt.Errorf("person ID is required")
	}
	coords, _ := person["coords"].(string)
	rawXML, _ := person["raw_xml"].(string)
	
	if err := a.db.AddPerson(tableName, id, coords, rawXML); err != nil {
		return nil, err
	}
	return person, nil
}

// updatePersonPlan handles updating a person's plan XML and returns the updated person (internal function for interpreter)
func (a *App) updatePersonPlan(tableName string, personId string, planXML string) (*Person, error) {
	log.Printf("UpdatePersonPlan called for person %s in table %s", personId, tableName)
	limit := 500
	if len(planXML) < limit {
		limit = len(planXML)
	}
	log.Printf("XML content (first %d chars): %s", limit, planXML[:limit])
	
	if err := a.db.UpdatePersonXML(tableName, personId, planXML); err != nil {
		return nil, fmt.Errorf("failed to update person plan: %w", err)
	}
	
	// Extract and update coordinates from the updated XML
	// Updated regex to handle both self-closing and non-self-closing tags
	// Looking for pattern like: <activity ... x="123.456" ... y="789.012" ... /> or <activity ... x="123.456" ... y="789.012" ... >
	activityPattern := regexp.MustCompile(`<activity[^>]*\sx="([^"]+)"[^>]*\sy="([^"]+)"[^>]*(?:/>|>)`)
	matches := activityPattern.FindStringSubmatch(planXML)
	
	var coords string
	if len(matches) >= 3 {
		coords = fmt.Sprintf("%s,%s", matches[1], matches[2])
	} else {
		// Try reverse order (y before x)
		activityPatternReverse := regexp.MustCompile(`<activity[^>]*\sy="([^"]+)"[^>]*\sx="([^"]+)"[^>]*(?:/>|>)`)
		matches = activityPatternReverse.FindStringSubmatch(planXML)
		
		if len(matches) >= 3 {
			coords = fmt.Sprintf("%s,%s", matches[2], matches[1])
		}
	}
	log.Printf("Extracted coordinates for person %s: %s", personId, coords)
	
	if coords != "" {
		if err := a.db.UpdatePersonCoords(tableName, personId, coords); err != nil {
			return nil, fmt.Errorf("failed to update person coordinates: %w", err)
		}
		log.Printf("Successfully updated coordinates for person %s to %s", personId, coords)
	} else {
		log.Printf("No coordinates found in XML for person %s", personId)
	}
	
	// Fetch and return the updated person
	updatedPerson, err := a.db.GetPerson(tableName, personId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated person: %w", err)
	}
	
	return updatedPerson, nil
}

// batchUpdatePersons handles batch updating multiple persons in a single transaction (internal function for interpreter)
func (a *App) batchUpdatePersons(tableName string, updates []map[string]any) error {
	log.Printf("BatchUpdatePersons called for %d persons in table %s", len(updates), tableName)
	
	// Convert map updates to PersonUpdate structs
	var personUpdates []PersonUpdate
	for _, update := range updates {
		id, ok := update["id"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid id in update")
		}
		
		coords, ok := update["coords"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid coords in update for person %s", id)
		}
		
		rawXML, ok := update["raw_xml"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid raw_xml in update for person %s", id)
		}
		
		personUpdates = append(personUpdates, PersonUpdate{
			ID:     id,
			Coords: coords,
			RawXML: rawXML,
		})
	}
	
	// Execute batch update
	err := a.db.BatchUpdatePersons(tableName, personUpdates)
	if err != nil {
		return fmt.Errorf("failed to batch update persons: %w", err)
	}
	
	log.Printf("Successfully batch updated %d persons", len(personUpdates))
	return nil
}

// deletePerson handles deleting a person's record (internal function for interpreter)
func (a *App) deletePerson(tableName string, personId string) (map[string]string, error) {
	if err := a.db.DeletePerson(tableName, personId); err != nil {
		return nil, fmt.Errorf("failed to delete person: %w", err)
	}
	return map[string]string{"message": "Person deleted successfully"}, nil
}

// NewProcessor creates a new population processor
func NewProcessor(db *Database, processID int) (*Processor, error) {
	return &Processor{
		db:        db,
		processID: processID,
		telemetry: &ProcessTelemetry{
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
	if _, err := p.db.execQuery(initTelemetryQuery, fmt.Sprintf("failed to initialize telemetry for process %d", p.processID), p.processID, p.telemetry.TotalFileSize); err != nil {
		return err
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
	if err := p.db.execTableQuery(fmt.Sprintf(createPopulationTableQuery, tableName)); err != nil {
		return err
	}

	// Start telemetry ticker
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Channel to signal processing completion
	done := make(chan error, 1)
	// Channel to signal telemetry goroutine has finished
	telemetryDone := make(chan struct{})

	// Start telemetry updater in goroutine
	go func() {
		defer close(telemetryDone)
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
	
	// Wait for telemetry goroutine to finish
	<-telemetryDone

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
	batch := make([]PersonData, 0, 100)
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

					// Extract person ID from XML
					var id string
					foundID := false
					decoder2 := xml.NewDecoder(bytes.NewReader(personBuffer.Bytes()))
					for {
						token2, err2 := decoder2.Token()
						if err2 != nil {
							break
						}
						if se, ok := token2.(xml.StartElement); ok && se.Name.Local == "person" {
							for _, attr := range se.Attr {
								if attr.Name.Local == "id" {
									id = attr.Value
									foundID = true
									break
								}
							}
							break
						}
					}
					
					if !foundID || id == "" {
						log.Printf("could not extract id for person, skipping")
						p.telemetryMutex.Lock()
						p.errorCounter++
						p.telemetryMutex.Unlock()
						continue
					}

					coords := fmt.Sprintf("%f,%f", homeX, homeY)

					batch = append(batch, PersonData{
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
						batch = make([]PersonData, 0, 100)
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

func (p *Processor) insertPersonBatch(tableName string, persons []PersonData) error {
	tx, err := p.db.GetConn().Begin()
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
	_, err := p.db.execQuery(updateTelemetryQuery, fmt.Sprintf("failed to update telemetry for process %d", telemetry.ProcessID), telemetry.BytesRead, telemetry.PersonsExtracted, telemetry.ErrorCount, telemetry.ProcessID)
	if err != nil {
		log.Printf("Failed to update telemetry for process %d (attempt 1): %v", p.processID, err)
		
		// Wait briefly and retry once
		time.Sleep(100 * time.Millisecond)
		_, err = p.db.execQuery(updateTelemetryQuery, fmt.Sprintf("failed to update telemetry for process %d", telemetry.ProcessID), telemetry.BytesRead, telemetry.PersonsExtracted, telemetry.ErrorCount, telemetry.ProcessID)
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

