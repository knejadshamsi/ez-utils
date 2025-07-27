package gui

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

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
	db, err := NewDatabase("ez_utils_gui.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	a.db = db
	
	// Create zone tables
	if err := db.CreateZoneTables(); err != nil {
		log.Printf("Error creating zone tables: %v", err)
	}
}

// SetDatabase sets the database for testing purposes
func (a *App) SetDatabase(db *Database) {
	a.db = db
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

// ValidationResult represents the result of XML validation
type ValidationResult struct {
	IsValid bool   `json:"isValid"`
	Error   string `json:"error"`
}

// ValidatePopulationXML validates that the XML file has proper structure for population data
func (a *App) ValidatePopulationXML(filePath string) (*ValidationResult, error) {
	log.Printf("ValidatePopulationXML called with filePath: %s", filePath)
	
	// Check if file exists and is readable
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &ValidationResult{
			IsValid: false,
			Error:   "File not found",
		}, nil
	}
	
	// Check file extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".xml") {
		return &ValidationResult{
			IsValid: false,
			Error:   "File must have .xml extension",
		}, nil
	}
	
	// Open and read the file
	file, err := os.Open(filePath)
	if err != nil {
		return &ValidationResult{
			IsValid: false,
			Error:   fmt.Sprintf("Cannot read file: %v", err),
		}, nil
	}
	defer file.Close()
	
	// Parse XML and check structure
	decoder := xml.NewDecoder(file)
	
	var foundPopulation bool
	var foundPerson bool
	var depth int
	
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return &ValidationResult{
				IsValid: false,
				Error:   fmt.Sprintf("Invalid XML format: %v", err),
			}, nil
		}
		
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			
			// Check for population root element
			if depth == 1 && t.Name.Local == "population" {
				foundPopulation = true
			}
			
			// Check for person element
			if t.Name.Local == "person" {
				foundPerson = true
			}
			
			// Early exit if we found both required elements
			if foundPopulation && foundPerson {
				return &ValidationResult{
					IsValid: true,
					Error:   "",
				}, nil
			}
			
		case xml.EndElement:
			depth--
		}
	}
	
	// Check validation results
	if !foundPopulation {
		return &ValidationResult{
			IsValid: false,
			Error:   "Missing <population> root element",
		}, nil
	}
	
	if !foundPerson {
		return &ValidationResult{
			IsValid: false,
			Error:   "No <person> elements found",
		}, nil
	}
	
	return &ValidationResult{
		IsValid: true,
		Error:   "",
	}, nil
}

// ValidateNetworkXML validates that the XML file has proper structure for network data
func (a *App) ValidateNetworkXML(filePath string) (*ValidationResult, error) {
	log.Printf("ValidateNetworkXML called with filePath: %s", filePath)
	
	// Check if file exists and is readable
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &ValidationResult{
			IsValid: false,
			Error:   "File not found",
		}, nil
	}
	
	// Check file extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".xml") {
		return &ValidationResult{
			IsValid: false,
			Error:   "File must have .xml extension",
		}, nil
	}
	
	// Open and read the file
	file, err := os.Open(filePath)
	if err != nil {
		return &ValidationResult{
			IsValid: false,
			Error:   fmt.Sprintf("Cannot read file: %v", err),
		}, nil
	}
	defer file.Close()
	
	// Parse XML and check structure
	decoder := xml.NewDecoder(file)
	
	var foundNetwork bool
	var foundNodes bool
	var foundLinks bool
	var depth int
	
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return &ValidationResult{
				IsValid: false,
				Error:   fmt.Sprintf("Invalid XML format: %v", err),
			}, nil
		}
		
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			
			// Check for network root element
			if depth == 1 && t.Name.Local == "network" {
				foundNetwork = true
			}
			
			// Check for nodes and links elements
			if t.Name.Local == "nodes" {
				foundNodes = true
			}
			if t.Name.Local == "links" {
				foundLinks = true
			}
			
			// Early exit if we found all required elements
			if foundNetwork && foundNodes && foundLinks {
				return &ValidationResult{
					IsValid: true,
					Error:   "",
				}, nil
			}
			
		case xml.EndElement:
			depth--
		}
	}
	
	// Check validation results
	if !foundNetwork {
		return &ValidationResult{
			IsValid: false,
			Error:   "Missing <network> root element",
		}, nil
	}
	
	if !foundNodes {
		return &ValidationResult{
			IsValid: false,
			Error:   "Missing <nodes> element",
		}, nil
	}
	
	if !foundLinks {
		return &ValidationResult{
			IsValid: false,
			Error:   "Missing <links> element",
		}, nil
	}
	
	return &ValidationResult{
		IsValid: true,
		Error:   "",
	}, nil
}

// ValidatePTXML validates that the XML file has proper structure for public transport data
func (a *App) ValidatePTXML(filePath string) (*ValidationResult, error) {
	log.Printf("ValidatePTXML called with filePath: %s", filePath)
	
	// Check if file exists and is readable
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &ValidationResult{
			IsValid: false,
			Error:   "File not found",
		}, nil
	}
	
	// Check file extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".xml") {
		return &ValidationResult{
			IsValid: false,
			Error:   "File must have .xml extension",
		}, nil
	}
	
	// Open and read the file
	file, err := os.Open(filePath)
	if err != nil {
		return &ValidationResult{
			IsValid: false,
			Error:   fmt.Sprintf("Cannot read file: %v", err),
		}, nil
	}
	defer file.Close()
	
	// Parse XML and check structure
	decoder := xml.NewDecoder(file)
	
	var foundTransitSchedule bool
	var foundTransitStops bool
	var foundTransitLines bool
	var depth int
	
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return &ValidationResult{
				IsValid: false,
				Error:   fmt.Sprintf("Invalid XML format: %v", err),
			}, nil
		}
		
		switch t := token.(type) {
		case xml.StartElement:
			depth++
			
			// Check for transitSchedule root element
			if depth == 1 && t.Name.Local == "transitSchedule" {
				foundTransitSchedule = true
			}
			
			// Check for transitStops and transitLines elements
			if t.Name.Local == "transitStops" {
				foundTransitStops = true
			}
			if t.Name.Local == "transitLines" {
				foundTransitLines = true
			}
			
			// Early exit if we found all required elements
			if foundTransitSchedule && foundTransitStops && foundTransitLines {
				return &ValidationResult{
					IsValid: true,
					Error:   "",
				}, nil
			}
			
		case xml.EndElement:
			depth--
		}
	}
	
	// Check validation results
	if !foundTransitSchedule {
		return &ValidationResult{
			IsValid: false,
			Error:   "Missing <transitSchedule> root element",
		}, nil
	}
	
	if !foundTransitStops {
		return &ValidationResult{
			IsValid: false,
			Error:   "Missing <transitStops> element",
		}, nil
	}
	
	if !foundTransitLines {
		return &ValidationResult{
			IsValid: false,
			Error:   "Missing <transitLines> element",
		}, nil
	}
	
	return &ValidationResult{
		IsValid: true,
		Error:   "",
	}, nil
}

// StartProcessing routes to the appropriate processor based on edit mode
func (a *App) StartProcessing(filePath string, fileEditMode string) (map[string]any, error) {
	switch fileEditMode {
	case "POPULATION":
		return a.ProcessPopulationFile(filePath)
	case "NETWORK":
		return a.ProcessNetworkFile(filePath)
	case "PT":
		result, err := a.ProcessPTFile(filePath)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"message":    result.Message,
			"process_id": result.ProcessID,
		}, nil
	}
	return nil, fmt.Errorf("unknown file edit mode: %s", fileEditMode)
}

// LoadProcessData loads the processed data for editing with spatial filtering based on LoadingParams
func (a *App) LoadProcessData(params LoadingParams) (map[string]any, error) {
	log.Printf("LoadProcessData called with params: %+v", params)
	
	// Calculate 60% of viewport bounds
	filteredViewport := calculate60PercentViewport(params.Viewport)
	
	switch params.FileEditMode {
	case "POPULATION":
		// Count total population first
		count, err := a.db.CountPopulationInTable(params.ProcessId)
		if err != nil {
			return nil, fmt.Errorf("failed to count population: %v", err)
		}
		
		var persons []PersonData
		
		// If count is below min threshold, load all persons
		if params.MinThreshold > 0 && count <= params.MinThreshold {
			persons, err = a.db.GetAllPopulation(params.ProcessId)
			if err != nil {
				return nil, fmt.Errorf("failed to load all population data: %v", err)
			}
		} else {
			// Otherwise use viewport filtering with randomization
			persons, err = a.db.GetPopulationInBounds(
				params.ProcessId, 
				filteredViewport,
				params.RandomFactor,
				params.MaxElements,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to load population data: %v", err)
			}
		}
		
		return map[string]any{
			"data": persons,
			"type": "population",
		}, nil
		
	case "NETWORK":
		nodes, err := a.db.GetNodesInBounds(
			params.ProcessId,
			filteredViewport,
			params.RandomFactor,
			params.MaxElements,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load network nodes: %v", err)
		}
		
		links, err := a.db.GetLinksForNodesInBounds(
			params.ProcessId,
			filteredViewport,
			params.RandomFactor,
			params.MaxElements,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load network links: %v", err)
		}
		
		return map[string]any{
			"data": map[string]any{
				"nodes": nodes,
				"links": links,
			},
			"type": "network",
		}, nil
		
	case "PT":
		// PT loads all line summaries for the selected mode - no viewport filtering needed
		// Cities rarely have more than a few hundred lines total
		mode := params.Mode
		if mode == "" {
			mode = "BUS" // Default mode
		}
		
		summaries, err := a.db.GetPTLineSummaries(params.ProcessId, mode)
		if err != nil {
			return nil, fmt.Errorf("failed to load PT line summaries: %v", err)
		}
		
		return map[string]any{
			"data": map[string]any{
				"summaries": summaries,
			},
			"type": "pt",
		}, nil
		
	default:
		return nil, fmt.Errorf("unknown file edit mode: %s", params.FileEditMode)
	}
}

// Helper function to calculate 60% viewport
func calculate60PercentViewport(viewport ViewportBounds) ViewportBounds {
	latMargin := (viewport.MaxLat - viewport.MinLat) * 0.2  // 20% margin = 60% viewport
	lngMargin := (viewport.MaxLng - viewport.MinLng) * 0.2
	
	return ViewportBounds{
		MinLat: viewport.MinLat + latMargin,
		MaxLat: viewport.MaxLat - latMargin,
		MinLng: viewport.MinLng + lngMargin,
		MaxLng: viewport.MaxLng + lngMargin,
	}
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
			Assets: FrontendAssets,
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