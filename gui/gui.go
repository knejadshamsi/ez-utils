package gui

import (
	"context"
	"embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

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
	if err := InitDB("ez_utils_gui.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
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