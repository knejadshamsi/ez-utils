package gui

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/xml"
	"ez-utils/gui/database"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct
type App struct {
	ctx         context.Context
	StartupFile string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Initialize database
	if err := database.InitDB("ez_utils_gui.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
}

// ProcessFile starts the processing of a given file path.
func (a *App) ProcessFile(filePath string) (map[string]interface{}, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	processID, err := database.CreateProcess(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create process record: %w", err)
	}

	go a.processPopulationFile(filePath, processID)

	response := map[string]interface{}{
		"message":   "Processing started",
		"processId": processID,
	}
	return response, nil
}

// GetProcesses retrieves all process records from the database.
func (a *App) GetProcesses() ([]database.Process, error) {
	return database.GetProcesses()
}

// CheckProcessingStatus retrieves the status of a specific process.
func (a *App) CheckProcessingStatus(processID int) (string, error) {
	return database.GetProcessStatus(processID)
}

// GetProcessesByFile retrieves all process records for a given file path.
func (a *App) GetProcessesByFile(filePath string) ([]database.Process, error) {
	return database.GetProcessesByFile(filePath)
}

// DeleteProcess deletes a process and its associated data.
func (a *App) DeleteProcess(processID int) error {
	tableName := fmt.Sprintf("population_data_%d", processID)
	// First, drop the associated table if it exists
	if err := database.DropTable(tableName); err != nil {
		// We can potentially ignore "no such table" errors, but for now, we'll return it.
		return fmt.Errorf("failed to drop table '%s': %w", tableName, err)
	}
	// Then, delete the process record
	if err := database.DeleteProcess(processID); err != nil {
		return fmt.Errorf("failed to delete process record with ID %d: %w", processID, err)
	}
	return nil
}

// GetPopulation retrieves all processed population data for a given table.
func (a *App) GetPopulation(tableName string) ([]database.Person, error) {
	return database.GetPopulationData(tableName)
}

// AddPerson adds a new person to a population table.
func (a *App) AddPerson(tableName string, person map[string]interface{}) (map[string]interface{}, error) {
	if err := database.AddPerson(tableName, person); err != nil {
		return nil, fmt.Errorf("failed to add person: %w", err)
	}
	return person, nil
}

// UpdatePersonPlan handles updating a person's plan XML.
func (a *App) UpdatePersonPlan(tableName string, personId string, planXML string) (map[string]string, error) {
	err := database.UpdatePersonXML(tableName, personId, planXML)
	if err != nil {
		return nil, fmt.Errorf("failed to update person plan: %w", err)
	}
	return map[string]string{"message": "Plan updated successfully"}, nil
}

// DeletePerson handles deleting a person's record.
func (a *App) DeletePerson(tableName string, personId string) (map[string]string, error) {
	err := database.DeletePerson(tableName, personId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete person: %w", err)
	}
	return map[string]string{"message": "Person deleted successfully"}, nil
}

const personBatchSize = 100

// PersonData holds the extracted information for a single person.
type PersonData struct {
	ID     string
	Coords string // "x,y"
	RawXML string
}

// processPopulationFile reads a MATSim population XML file, extracts person data,
// and stores it in a new database table using a memory-efficient streaming parser
// and batched database inserts.
func (a *App) processPopulationFile(filePath string, processID int) {
	tableName := fmt.Sprintf("population_data_%d", processID)

	updateStatus := func(status string) {
		if err := database.UpdateProcessStatus(processID, status); err != nil {
			log.Printf("Failed to update process status for processID %d: %v", processID, err)
		}
	}

	updateStatus("Reading XML file...")
	file, err := os.Open(filePath)
	if err != nil {
		errStr := fmt.Sprintf("failed to open file: %v", err)
		log.Printf("Error opening file %s: %v", filePath, err)
		updateStatus(errStr)
		return
	}
	defer file.Close()

	updateStatus("Parsing persons and plans...")
	db, err := database.GetDB()
	if err != nil {
		errStr := fmt.Sprintf("failed to get DB connection: %v", err)
		log.Printf("Error processing population file: %s", errStr)
		updateStatus(errStr)
		return
	}

	updateStatus("Creating database and tables...")
	if err := database.CreateTable(tableName); err != nil {
		errStr := fmt.Sprintf("failed to create table: %v", err)
		log.Printf("Error creating table %s: %v", tableName, err)
		updateStatus(errStr)
		return
	}

	decoder := xml.NewDecoder(file)
	batch := make([]PersonData, 0, personBatchSize)
	var personBuffer bytes.Buffer
	var personEncoder = xml.NewEncoder(&personBuffer)
	personDepth := 0

	var homeX, homeY float64
	var foundCoords bool

	updateStatus("Inserting person records...")
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			errStr := fmt.Sprintf("failed to decode XML token: %v", err)
			log.Printf("XML decoding error: %v", err)
			updateStatus(errStr)
			return
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
						continue
					}

					coords := fmt.Sprintf("%f,%f", homeX, homeY)

					batch = append(batch, PersonData{
						ID:     id,
						Coords: coords,
						RawXML: personBuffer.String(),
					})

					if len(batch) >= personBatchSize {
						if err := insertPersonBatch(db, tableName, batch); err != nil {
							errStr := fmt.Sprintf("failed to insert batch: %v", err)
							log.Printf(errStr)
							updateStatus(errStr)
							return
						}
						batch = make([]PersonData, 0, personBatchSize)
					}
				}
			}
		}
	}

	if len(batch) > 0 {
		if err := insertPersonBatch(db, tableName, batch); err != nil {
			errStr := fmt.Sprintf("failed to insert final batch: %v", err)
			log.Printf(errStr)
			updateStatus(errStr)
			return
		}
	}

	updateStatus("Finalizing and cleaning up...")
	time.Sleep(2 * time.Second)

	updateStatus("Completed")
	log.Printf("Successfully processed file %s into table %s", filePath, tableName)
}

// insertPersonBatch inserts a slice of PersonData into the database in a single transaction.
func insertPersonBatch(db *sql.DB, tableName string, persons []PersonData) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback is a no-op if the transaction is committed.

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT OR REPLACE INTO %s (id, coords, raw_xml) VALUES (?, ?, ?)", tableName))
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, p := range persons {
		if _, err := stmt.Exec(p.ID, p.Coords, p.RawXML); err != nil {
			// The transaction will be rolled back by defer.
			return fmt.Errorf("failed to execute statement for person %s: %w", p.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// extractAttribute is a helper to find an attribute value from a raw XML element.
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

// GetStartupFile returns the file path passed on startup, if any.
func (a *App) GetStartupFile() string {
	return a.StartupFile
}

// Run creates and runs the Wails application.
func Run(filePath string) {
	// Create an instance of the app structure
	app := NewApp()
	app.StartupFile = filePath

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "EZ-Utils GUI",
		Width:         10000,
		Height:        10000,
		DisableResize: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Error running Wails app: %v", err)
	}
}