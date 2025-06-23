// Package pt provides GTFS to XML conversion functionality.
// This file contains the main orchestrator for the Public Transit (PT) processing pipeline.
// It defines the sequence of steps, manages configuration, logging, and TUI integration.
package pt

import (
	"fmt"
	"log"
	"os"
	"time"

	displayCore "ez-utils/src/display/core"
	displayConfig "ez-utils/src/display/config"
	globalConfig "ez-utils/config"
)

// Constants for GTFS file names and directories
const (
	CalendarFile   = "calendar.txt"
	TripsFile     = "trips.txt"
	StopTimesFile = "stop_times.txt"
	StopsFile     = "stops.txt"
	RoutesFile    = "routes.txt"
	
	ServicesDir     = "02_services"
	TripsDir        = "03_trips"
	StopSequencesDir = "03_trips/stop_sequences"
	ScheduleBuildDir = "04_schedule_building"
	
	ServicePatternsFile  = "service_patterns.json"
	ServiceMappingFile   = "service_mapping.json"
	SelectedServicesFile = "selected_services.json"
	TripMappingsFile     = "trip_mappings.json"
	StopSequencesFile    = "stop_sequences.json"
	StopDetailsFile      = "stop_details.json"
	TransitScheduleFile  = "transit_schedule.xml"
	
	LogFilePrefix = "pt_processing_"
	LogTimeFormat = "20060102_150405"
	FilePermissions = 0644
	DirPermissions  = 0755
)

// FileExists checks if a file exists and is not a directory
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// SetupLogger creates a logger instance for the orchestrator
func SetupLogger(logFile *os.File) *log.Logger {
	return log.New(logFile, "", log.LstdFlags)
}

// CreateDirectoryIfNotExists creates a directory if it doesn't exist
func CreateDirectoryIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, DirPermissions)
	}
	return nil
}

// ServiceDay represents different service day types
type ServiceDay int

const (
	ServiceDayWeekday ServiceDay = iota
	ServiceDayWeekend
	ServiceDayMonday
	ServiceDayTuesday
	ServiceDayWednesday
	ServiceDayThursday
	ServiceDayFriday
	ServiceDaySaturday
	ServiceDaySunday
)

// PTConfig holds configuration for PT processing
type PTConfig struct {
	GTFSDirectory string
	ServiceDay    ServiceDay
	CleanFlag     bool
	TempDir       string
	OutputDir     string
}

// ProcessingStats tracks processing statistics
type ProcessingStats struct {
	ValidationFile      string
	ServiceCounter      int
	RouteCounter        int
	TripCounter         int
	SequenceCounter     int
	StopCounter         int
	TransitStopCounter  int
	CleanupFiles        int
	CleanupDirs         int
	CleanupBytes        int64
}

// ServicePattern represents a service pattern from calendar.txt
type ServicePattern struct {
	ServiceID string `json:"service_id"`
	Monday    bool   `json:"monday"`
	Tuesday   bool   `json:"tuesday"`
	Wednesday bool   `json:"wednesday"`
	Thursday  bool   `json:"thursday"`
	Friday    bool   `json:"friday"`
	Saturday  bool   `json:"saturday"`
	Sunday    bool   `json:"sunday"`
}

// Route represents a route from routes.txt
type Route struct {
	RouteID   string `json:"route_id"`
	RouteType string `json:"route_type"`
}

// ServiceMapping represents the mapping between services and routes
type ServiceMapping struct {
	ServiceID string  `json:"service_id"`
	Routes    []Route `json:"routes"`
}

// TripMapping represents the mapping of trips to routes and services
type TripMapping struct {
	TripID  string `json:"trip_id"`
	RouteID string `json:"route_id"`
}

// SelectedService represents a service selected for processing (legacy format)
type SelectedService struct {
	ServiceID string  `json:"service_id"`
	Routes    []Route `json:"routes"`
	Selected  bool    `json:"selected"`
	TripCount int     `json:"trip_count"`
}

// SelectedServicesByType represents the correct Step 4 output format per documentation
type SelectedServicesByType struct {
	BusService   string `json:"bus_service"`
	MetroService string `json:"metro_service"`
	Date         string `json:"date"`
}

// Stop represents a stop in a trip sequence
type Stop struct {
	StopID        string `json:"stop_id"`
	ArrivalTime   string `json:"arrival_time"`
	DepartureTime string `json:"departure_time"`
	StopSequence  int    `json:"stop_sequence"`
}

// StopSequence represents a stop sequence for a trip
type StopSequence struct {
	TripID string `json:"trip_id"`
	Stops  []Stop `json:"stops"`
}

// StopDetails represents detailed stop information
type StopDetails struct {
	StopID   string  `json:"stop_id"`
	StopName string  `json:"stop_name"`
	StopLat  float64 `json:"stop_lat"`
	StopLon  float64 `json:"stop_lon"`
}

// Error types
type PtError struct {
	Type    string
	Message string
	Details string
}

func (e PtError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func NewMissingFilesError(filename string) PtError {
	return PtError{
		Type:    "MissingFiles",
		Message: fmt.Sprintf("Required file not found: %s", filename),
		Details: "Ensure all required GTFS files are present in the directory",
	}
}

func NewFileParsingError(filename, details string) PtError {
	return PtError{
		Type:    "FileParsing",
		Message: fmt.Sprintf("Failed to parse file: %s", filename),
		Details: details,
	}
}

func NewInvalidServiceError(daySpecification string) PtError {
	return PtError{
		Type:    "InvalidService",
		Message: fmt.Sprintf("No valid services found for day specification: %s", daySpecification),
		Details: "Check if the specified service day exists in calendar.txt",
	}
}

func NewDirectoryError(directory, operation string) PtError {
	return PtError{
		Type:    "DirectoryError",
		Message: fmt.Sprintf("Failed to %s directory: %s", operation, directory),
		Details: "Check file permissions and available disk space",
	}
}

// PTOrchestrator coordinates all 9 steps of PT processing
type PTOrchestrator struct {
	config        *PTConfig
	stats         *ProcessingStats
	tuiEnabled    bool
	logFile       *os.File
	logger        *log.Logger
	displayInstance *displayCore.DisplayInstance
	globalConfig  *globalConfig.Config // Global config for language settings
}

// NewPTOrchestrator creates a new PTOrchestrator instance
func NewPTOrchestrator(gtfsDir string, serviceDay ServiceDay, cleanFlag bool) *PTOrchestrator {
	config := &PTConfig{
		GTFSDirectory: gtfsDir,
		ServiceDay:    serviceDay,
		CleanFlag:     cleanFlag,
		TempDir:       "temp/pt",
		OutputDir:     "output/pt",
	}
	
	return &PTOrchestrator{
		config:     config,
		stats:      &ProcessingStats{},
		tuiEnabled: true,
	}
}

// NewPTOrchestratorWithGlobalConfig creates orchestrator with global config
func NewPTOrchestratorWithGlobalConfig(gtfsDir string, serviceDay ServiceDay, cleanFlag bool, globalCfg *globalConfig.Config) *PTOrchestrator {
	orchestrator := NewPTOrchestrator(gtfsDir, serviceDay, cleanFlag)
	orchestrator.globalConfig = globalCfg
	return orchestrator
}

// DisableTUI disables TUI integration for testing
func (pto *PTOrchestrator) DisableTUI() {
	pto.tuiEnabled = false
}

// Process executes all 9 steps of the PT processing pipeline
func (pto *PTOrchestrator) Process() error {
	// Step 0: Setup logging
	if err := pto.setupLogging(); err != nil {
		return fmt.Errorf("failed to setup logging: %v", err)
	}
	defer pto.logFile.Close()

	pto.logMessage("Starting PT processing pipeline")

	// Initialize display if TUI is enabled
	if pto.tuiEnabled {
		if err := pto.initializeDisplay(); err != nil {
			pto.logMessage(fmt.Sprintf("Failed to initialize display: %v", err))
			// Continue without TUI
			pto.tuiEnabled = false
		} else {
			// Defer cleanup as fallback in case of early exit
			defer func() {
				if pto.displayInstance != nil {
					pto.cleanupDisplay()
				}
			}()
		}
	}

	// Step 0: Create necessary directories
	if err := pto.createDirectories(); err != nil {
		return err
	}

	// Step 1: File Validation (Display Step 0)
	pto.logMessage("Starting Step 1: File Validation")
	pto.setDisplayStep(0)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.validateFiles(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 1 failed: %v", err))
		return err
	}

	// Step 2: Service Pattern Identification (Display Step 1)
	pto.logMessage("Starting Step 2: Service Pattern Identification")
	pto.setDisplayStep(1)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.identifyServicePatterns(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 2 failed: %v", err))
		return err
	}

	// Step 3: Service-Route Mapping (Display Step 2)
	pto.logMessage("Starting Step 3: Service-Route Mapping")
	pto.setDisplayStep(2)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.mapServiceRoutes(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 3 failed: %v", err))
		return err
	}

	// Step 4: Service Selection (Display Step 3)
	pto.logMessage("Starting Step 4: Service Selection")
	pto.setDisplayStep(3)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.selectServices(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 4 failed: %v", err))
		return err
	}

	// Step 5: Trip Collection (Display Step 4)
	pto.logMessage("Starting Step 5: Trip Collection")
	pto.setDisplayStep(4)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.collectTrips(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 5 failed: %v", err))
		return err
	}

	// Step 6: Stop Sequence Collection (Display Step 5)
	pto.logMessage("Starting Step 6: Stop Sequence Collection")
	pto.setDisplayStep(5)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.collectStopSequences(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 6 failed: %v", err))
		return err
	}

	// Step 7: Stop Details Collection (Display Step 6)
	pto.logMessage("Starting Step 7: Stop Details Collection")
	pto.setDisplayStep(6)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.collectStopDetails(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 7 failed: %v", err))
		return err
	}

	// Step 8: Generate MATSim Transit Schedule XML (Display Step 7)
	pto.logMessage("Starting Step 8: Generate MATSim Transit Schedule XML")
	pto.setDisplayStep(7)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.generateXML(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 8 failed: %v", err))
		return err
	}

	// Step 9: Clean Temporary Files (Display Step 8) - if requested
	if pto.config.CleanFlag {
		pto.logMessage("Starting Step 9: Clean Temporary Files")
		pto.setDisplayStep(8)
		if pto.checkQuitRequested() {
			pto.logMessage("Process aborted by user")
		return nil
		}
		if err := pto.cleanupFiles(); err != nil {
			pto.logMessage(fmt.Sprintf("Step 9 failed: %v", err))
			return err
		}
	}

	pto.logMessage("PT processing pipeline completed successfully")
	
	// Signal completion to display system for proper cleanup
	if pto.tuiEnabled && pto.displayInstance != nil {
		if err := pto.displayInstance.SetProcessComplete(true); err != nil {
			pto.logMessage(fmt.Sprintf("Failed to signal completion: %v", err))
		}
		
		// Give the display system time to handle completion and cleanup
		time.Sleep(300 * time.Millisecond)
		
		// Now stop the display properly
		pto.cleanupDisplay()
	}
	
	return nil
}

// initializeDisplay sets up the TUI display for PT processing
func (pto *PTOrchestrator) initializeDisplay() error {
	var cfg *displayConfig.Config
	var err error
	
	// Use global config if available for language setting
	if pto.globalConfig != nil {
		cfg, err = displayConfig.NewPTDisplayConfig(pto.globalConfig, pto.config.CleanFlag)
	} else {
		// Fallback to default English
		cfg, err = displayConfig.NewConfig(
			"GTFS Transit Processing",
			"Converting GTFS data to MATSim transit schedule format",
			"en", // Default to English
			9,    // PT has 9 steps (Steps 1-9, display indices 0-8)
		)
		if err == nil {
			// Add PT-specific step configurations (Steps 1-9 mapped to display indices 0-8)
			cfg.AddStep("Validate GTFS Files", "Checking required GTFS files", "file", "status", "size").
				AddStep("Identify Service Patterns", "Analyzing calendar.txt for service patterns", "services", "progress").
				AddStep("Map Service Routes", "Linking services to routes (bus=3, metro=1)", "routes", "mappings").
				AddStep("Select Services", "Randomly selecting bus and metro services", "selected", "total").
				AddStep("Collect Trips", "Gathering trip data for selected services", "trips", "progress").
				AddStep("Extract Stop Sequences", "Creating individual trip files", "sequences", "stops").
				AddStep("Gather Stop Details", "Collecting stop coordinates and names", "stops", "details").
				AddStep("Generate MATSim XML", "Creating transit schedule XML with offsets", "progress", "size").
				AddStep("Cleanup Files", "Removing temporary files", "files", "dirs", "bytes").
				SetTheme("pt-blue").
				SetFlag("clean", pto.config.CleanFlag)
		}
	}
	
	if err != nil {
		return fmt.Errorf("failed to create display config: %v", err)
	}

	// Configuration is already set up by NewPTDisplayConfig or default setup above
	
	pto.displayInstance, err = displayCore.NewDisplayInstance(cfg)
	if err != nil {
		return fmt.Errorf("failed to create display instance: %v", err)
	}

	// Add completion callback for proper cleanup
	pto.displayInstance.AddCleanupFunc(func() {
		pto.logMessage("Display cleanup callback executed")
		// Force terminal restoration
		fmt.Print("\033[?25h")  // Show cursor
		fmt.Print("\033[2K")    // Clear line
		fmt.Print("\033[0m")    // Reset colors
		fmt.Print("\r")         // Return to start of line
	})

	if err := pto.displayInstance.Start(); err != nil {
		return fmt.Errorf("failed to start display: %v", err)
	}

	pto.logMessage("Display instance started successfully")
	return nil
}

// checkQuitRequested checks if user requested to quit
func (pto *PTOrchestrator) checkQuitRequested() bool {
	if pto.tuiEnabled && pto.displayInstance != nil {
		return pto.displayInstance.IsQuitRequested()
	}
	return false
}

// IsQuitRequested public method for external access
func (pto *PTOrchestrator) IsQuitRequested() bool {
	return pto.checkQuitRequested()
}

// cleanupDisplay shuts down the display using callback-based cleanup
func (pto *PTOrchestrator) cleanupDisplay() {
	if pto.displayInstance != nil {
		pto.logMessage("Stopping display...")
		
		// The display system will handle cleanup via the registered callback
		if err := pto.displayInstance.Stop(); err != nil {
			pto.logMessage(fmt.Sprintf("Display stop error: %v", err))
		} else {
			pto.logMessage("Display stopped successfully")
		}
		
		pto.displayInstance = nil
		pto.logMessage("Display cleanup completed")
	}
}

// setDisplayStep updates the display step if TUI is enabled
func (pto *PTOrchestrator) setDisplayStep(step int) {
	if pto.tuiEnabled && pto.displayInstance != nil {
		if err := pto.displayInstance.SetStep(step); err != nil {
			pto.logMessage(fmt.Sprintf("Failed to set display step: %v", err))
		}
	}
}


func (pto *PTOrchestrator) setupLogging() error {
	logFileName := LogFilePrefix + time.Now().Format(LogTimeFormat) + ".log"
	var err error
	pto.logFile, err = os.Create(logFileName)
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}
	
	pto.logger = SetupLogger(pto.logFile)
	return nil
}

func (pto *PTOrchestrator) logMessage(message string) {
	if pto.logger != nil {
		pto.logger.Printf("[PT] %s", message)
	}
}

func (pto *PTOrchestrator) createDirectories() error {
	directories := []string{
		pto.config.TempDir,
		pto.config.OutputDir,
	}
	
	for _, dir := range directories {
		if err := CreateDirectoryIfNotExists(dir); err != nil {
			pto.logMessage(fmt.Sprintf("Failed to create directory %s: %v", dir, err))
			return NewDirectoryError(dir, "create")
		}
	}
	
	pto.logMessage("Created necessary directories")
	return nil
}