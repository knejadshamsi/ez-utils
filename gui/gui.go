package gui

import (
	"context"
	"embed"
	"ez-utils/gui/database"
	"ez-utils/gui/process/edit/population"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// App struct
type App struct {
	ctx         context.Context
	StartupFile string
	EditMode    string // "population", "network", "public transportation"
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// Set up logging to file
	logFile, err := os.OpenFile("ez-utils-gui.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
	}
	
	log.Printf("App.startup() called - StartupFile: '%s', EditMode: '%s'", a.StartupFile, a.EditMode)
	
	// Initialize database
	if err := database.InitDB("ez_utils_gui.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
}

// ProcessPopulationFile starts the processing of a given file path.
func (a *App) ProcessPopulationFile(filePath string) (map[string]interface{}, error) {
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
// Returns nil if there's an error, empty array if no processes found, or array of processes.
func (a *App) GetProcessesByFile(filePath string) []database.Process {
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

// GetPerson retrieves a single person by ID
func (a *App) GetPerson(tableName string, personId string) (*database.Person, error) {
	return database.GetPerson(tableName, personId)
}

// AddPerson adds a new person to a population table.
func (a *App) AddPerson(tableName string, person map[string]interface{}) (map[string]interface{}, error) {
	if err := database.AddPerson(tableName, person); err != nil {
		return nil, fmt.Errorf("failed to add person: %w", err)
	}
	return person, nil
}

// UpdatePersonPlan handles updating a person's plan XML and returns the updated person.
func (a *App) UpdatePersonPlan(tableName string, personId string, planXML string) (*database.Person, error) {
	log.Printf("UpdatePersonPlan called for person %s in table %s", personId, tableName)
	log.Printf("XML content (first 500 chars): %s", planXML[:min(500, len(planXML))])
	
	err := database.UpdatePersonXML(tableName, personId, planXML)
	if err != nil {
		return nil, fmt.Errorf("failed to update person plan: %w", err)
	}
	
	// Extract and update coordinates from the updated XML
	coords := extractCoordsFromXML(planXML)
	log.Printf("Extracted coordinates for person %s: %s", personId, coords)
	
	if coords != "" {
		err = database.UpdatePersonCoords(tableName, personId, coords)
		if err != nil {
			return nil, fmt.Errorf("failed to update person coordinates: %w", err)
		}
		log.Printf("Successfully updated coordinates for person %s to %s", personId, coords)
	} else {
		log.Printf("No coordinates found in XML for person %s", personId)
	}
	
	// Fetch and return the updated person
	updatedPerson, err := database.GetPerson(tableName, personId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated person: %w", err)
	}
	
	return updatedPerson, nil
}

// BatchUpdatePersons handles batch updating multiple persons in a single transaction
func (a *App) BatchUpdatePersons(tableName string, updates []map[string]interface{}) error {
	log.Printf("BatchUpdatePersons called for %d persons in table %s", len(updates), tableName)
	
	// Convert map updates to PersonUpdate structs
	var personUpdates []database.PersonUpdate
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
		
		personUpdates = append(personUpdates, database.PersonUpdate{
			ID:     id,
			Coords: coords,
			RawXML: rawXML,
		})
	}
	
	// Execute batch update
	err := database.BatchUpdatePersons(tableName, personUpdates)
	if err != nil {
		return fmt.Errorf("failed to batch update persons: %w", err)
	}
	
	log.Printf("Successfully batch updated %d persons", len(personUpdates))
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DeletePerson handles deleting a person's record.
func (a *App) DeletePerson(tableName string, personId string) (map[string]string, error) {
	err := database.DeletePerson(tableName, personId)
	if err != nil {
		return nil, fmt.Errorf("failed to delete person: %w", err)
	}
	return map[string]string{"message": "Person deleted successfully"}, nil
}

// extractCoordsFromXML extracts coordinates from the first activity in person XML
func extractCoordsFromXML(xmlStr string) string {
	// Updated regex to handle both self-closing and non-self-closing tags
	// Looking for pattern like: <activity ... x="123.456" ... y="789.012" ... /> or <activity ... x="123.456" ... y="789.012" ... >
	activityPattern := regexp.MustCompile(`<activity[^>]*\sx="([^"]+)"[^>]*\sy="([^"]+)"[^>]*(?:/>|>)`)
	matches := activityPattern.FindStringSubmatch(xmlStr)
	
	if len(matches) >= 3 {
		return fmt.Sprintf("%s,%s", matches[1], matches[2])
	}
	
	// Try reverse order (y before x)
	activityPatternReverse := regexp.MustCompile(`<activity[^>]*\sy="([^"]+)"[^>]*\sx="([^"]+)"[^>]*(?:/>|>)`)
	matches = activityPatternReverse.FindStringSubmatch(xmlStr)
	
	if len(matches) >= 3 {
		return fmt.Sprintf("%s,%s", matches[2], matches[1])
	}
	
	return ""
}

// processPopulationFile processes a population XML file using the new processor module
func (a *App) processPopulationFile(filePath string, processID int) {
	updateStatus := func(status string) {
		if err := database.UpdateProcessStatus(processID, status); err != nil {
			log.Printf("Failed to update process status for processID %d: %v", processID, err)
		}
	}

	updateStatus("Initializing processor...")
	processor, err := population.NewProcessor(processID)
	if err != nil {
		errStr := fmt.Sprintf("failed to create processor: %v", err)
		log.Printf("Error creating processor: %v", err)
		updateStatus(errStr)
		return
	}

	updateStatus("Processing file...")
	if err := processor.ProcessPopulationFile(filePath); err != nil {
		errStr := fmt.Sprintf("failed to process file: %v", err)
		log.Printf("Error processing file %s: %v", filePath, err)
		updateStatus(errStr)
		return
	}

	updateStatus("Completed")
	log.Printf("Successfully processed file %s", filePath)
}

// GetStartupFile returns the file path passed on startup, if any.
func (a *App) GetStartupFile() string {
	return a.StartupFile
}

// GetStartupConfig returns the startup configuration including file path and edit mode
func (a *App) GetStartupConfig() map[string]interface{} {
	log.Printf("GetStartupConfig called - StartupFile: %s, EditMode: %s", a.StartupFile, a.EditMode)
	
	config := map[string]interface{}{
		"filePath": a.StartupFile,
		"editMode": a.EditMode,
	}
	
	// Check if file exists
	if a.StartupFile != "" {
		if _, err := os.Stat(a.StartupFile); os.IsNotExist(err) {
			config["error"] = fmt.Sprintf("File not found: %s", a.StartupFile)
			log.Printf("GetStartupConfig: Startup file not found: %s", a.StartupFile)
		} else if err != nil {
			config["error"] = fmt.Sprintf("Error accessing file: %v", err)
			log.Printf("GetStartupConfig: Error accessing startup file %s: %v", a.StartupFile, err)
		} else {
			log.Printf("GetStartupConfig: Startup file exists and is accessible: %s", a.StartupFile)
		}
	} else {
		log.Printf("GetStartupConfig: No startup file provided")
	}
	
	log.Printf("GetStartupConfig returning config: %+v", config)
	return config
}

// GetProcessTelemetry retrieves telemetry data for a specific process
func (a *App) GetProcessTelemetry(processID int) (*database.ProcessTelemetry, error) {
	return database.GetTelemetry(processID)
}

// ExitApplication exits the application
func (a *App) ExitApplication() {
	os.Exit(0)
}

// GetDebugLogs returns the contents of both log files for debugging
func (a *App) GetDebugLogs() map[string]string {
	logs := make(map[string]string)
	
	// Read CLI log file
	if content, err := ioutil.ReadFile("ez-utils.log"); err == nil {
		logs["cliLog"] = string(content)
	} else {
		logs["cliLog"] = fmt.Sprintf("Error reading CLI log: %v", err)
	}
	
	// Read GUI log file
	if content, err := ioutil.ReadFile("ez-utils-gui.log"); err == nil {
		logs["guiLog"] = string(content)
	} else {
		logs["guiLog"] = fmt.Sprintf("Error reading GUI log: %v", err)
	}
	
	return logs
}

// SelectFile opens a file dialog and returns the selected file path
func (a *App) SelectFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application context not initialized")
	}
	
	filters := []runtime.FileFilter{
		{
			DisplayName: "XML Files",
			Pattern:     "*.xml",
		},
		{
			DisplayName: "All Files",
			Pattern:     "*.*",
		},
	}
	
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Select Population File",
		Filters:          filters,
		ShowHiddenFiles:  false,
		CanCreateDirectories: false,
		ResolvesAliases:  true,
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to open file dialog: %w", err)
	}
	
	return filePath, nil
}

// GetPopulationByBbox retrieves population data within a bounding box
func (a *App) GetPopulationByBbox(tableName string, minLat, minLng, maxLat, maxLng float64) ([]database.Person, error) {
	return database.GetPopulationByBbox(tableName, minLat, minLng, maxLat, maxLng)
}

// Run creates and runs the Wails application.
func Run(filePath string, editMode string) {
	// Set up logging immediately
	logFile, err := os.OpenFile("ez-utils-gui.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}
	
	log.Printf("GUI Run called with filePath: '%s', editMode: '%s'", filePath, editMode)
	
	// Create an instance of the app structure
	app := NewApp()
	app.StartupFile = filePath
	app.EditMode = editMode
	
	log.Printf("App instance created. StartupFile set to: '%s', EditMode set to: '%s'", app.StartupFile, app.EditMode)

	// Create application with options
	err = wails.Run(&options.App{
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