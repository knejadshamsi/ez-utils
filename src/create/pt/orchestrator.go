// Package pt provides GTFS to XML conversion functionality.
// This file contains the main orchestrator for the Public Transit (PT) processing pipeline.
// It defines the sequence of steps, manages configuration, logging, and TUI integration.
package pt

import (
	"fmt"
	"log"
	"os"
	"time"

	globalConfig "ez-utils/config"
	"ez-utils/src/create/pt/tui"
)

// Constants for GTFS file names and directories
const (
	CalendarFile  = "calendar.txt"
	TripsFile     = "trips.txt"
	StopTimesFile = "stop_times.txt"
	StopsFile     = "stops.txt"
	RoutesFile    = "routes.txt"

	ServicesDir      = "02_services"
	TripsDir         = "03_trips"
	StopSequencesDir = "03_trips/stop_sequences"
	ScheduleBuildDir = "04_schedule_building"

	ServicePatternsFile  = "service_patterns.json"
	ServiceMappingFile   = "service_mapping.json"
	SelectedServicesFile = "selected_services.json"
	TripDatabaseFile     = "trips.db"
	TripMappingsFile     = "trip_mappings.json"
	StopSequencesFile    = "stop_sequences.json"
	StopDetailsFile      = "stop_details.json"
	TransitScheduleFile  = "transit_schedule.xml"

	LogFilePrefix   = "pt_processing_"
	LogTimeFormat   = "20060102_150405"
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
	TargetDate    string // YYYY-MM-DD format, empty for auto-selection
	CleanFlag     bool
	TempDir       string
	OutputDir     string
}

// ProcessingStats tracks processing statistics
type ProcessingStats struct {
	ValidationFile     string
	ServiceCounter     int
	RouteCounter       int
	TripCounter        int
	SequenceCounter    int
	StopCounter        int
	TransitStopCounter int
	CleanupFiles       int
	CleanupDirs        int
	CleanupBytes       int64
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
	StartDate string `json:"start_date"` // YYYYMMDD from calendar.txt
	EndDate   string `json:"end_date"`   // YYYYMMDD from calendar.txt
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

// PTOrchestrator coordinates all 10 steps of PT processing
type PTOrchestrator struct {
	config       *PTConfig
	stats        *ProcessingStats
	tuiEnabled   bool
	logFile      *os.File
	logger       *log.Logger
	tui          *tui.TUI
	globalConfig *globalConfig.Config // Global config for language settings
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

// SetTargetDate sets the target date for processing
func (pto *PTOrchestrator) SetTargetDate(date string) {
	pto.config.TargetDate = date
}

// Process executes all 10 steps of the PT processing pipeline
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
				if pto.tui != nil {
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

	// Step 5: Trip Database Creation (Display Step 4)
	pto.logMessage("Starting Step 5: Trip Database Creation")
	pto.setDisplayStep(4)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.buildTripDatabase(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 5 failed: %v", err))
		return err
	}

	// Step 6: Trip Collection (Display Step 5)
	pto.logMessage("Starting Step 6: Trip Collection")
	pto.setDisplayStep(5)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.collectTrips(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 6 failed: %v", err))
		return err
	}

	// Step 7: Stop Sequence Collection (Display Step 6)
	pto.logMessage("Starting Step 7: Stop Sequence Collection")
	pto.setDisplayStep(6)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.collectStopSequences(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 7 failed: %v", err))
		return err
	}

	// Step 8: Stop Details Collection (Display Step 7)
	pto.logMessage("Starting Step 8: Stop Details Collection")
	pto.setDisplayStep(7)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.collectStopDetails(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 8 failed: %v", err))
		return err
	}

	// Step 9: Generate MATSim Transit Schedule XML (Display Step 8)
	pto.logMessage("Starting Step 9: Generate MATSim Transit Schedule XML")
	pto.setDisplayStep(8)
	if pto.checkQuitRequested() {
		pto.logMessage("Process aborted by user")
		return nil
	}
	if err := pto.generateXML(); err != nil {
		pto.logMessage(fmt.Sprintf("Step 9 failed: %v", err))
		return err
	}

	// Step 10: Clean Temporary Files (Display Step 9) - if requested
	if pto.config.CleanFlag {
		pto.logMessage("Starting Step 10: Clean Temporary Files")
		pto.setDisplayStep(9)
		if pto.checkQuitRequested() {
			pto.logMessage("Process aborted by user")
			return nil
		}
		if err := pto.cleanupFiles(); err != nil {
			pto.logMessage(fmt.Sprintf("Step 10 failed: %v", err))
			return err
		}
	}

	pto.logMessage("PT processing pipeline completed successfully")

	// Signal completion to display system for proper cleanup
	if pto.tuiEnabled && pto.tui != nil {
		pto.tui.SetProcessComplete(true)
		pto.logMessage("Set process complete in TUI")

		// Give the TUI time to handle completion
		time.Sleep(300 * time.Millisecond)

		// Now stop the TUI properly
		pto.cleanupDisplay()
	}

	return nil
}

// initializeDisplay sets up the TUI display for PT processing
func (pto *PTOrchestrator) initializeDisplay() error {
	pto.tui = tui.NewTUI()

	// Set up error callback to capture TUI errors in logs
	pto.tui.SetErrorCallback(func(err error) {
		pto.logMessage(fmt.Sprintf("TUI Runtime Error: %v", err))
	})

	if err := pto.tui.Start(); err != nil {
		return fmt.Errorf("failed to start TUI: %v", err)
	}

	pto.logMessage("TUI started successfully")
	return nil
}

// checkQuitRequested checks if user requested to quit
func (pto *PTOrchestrator) checkQuitRequested() bool {
	if pto.tuiEnabled && pto.tui != nil {
		return pto.tui.IsQuitRequested()
	}
	return false
}

// IsQuitRequested public method for external access
func (pto *PTOrchestrator) IsQuitRequested() bool {
	return pto.checkQuitRequested()
}

// cleanupDisplay shuts down the TUI
func (pto *PTOrchestrator) cleanupDisplay() {
	if pto.tui != nil {
		pto.logMessage("Stopping TUI...")

		if err := pto.tui.Stop(); err != nil {
			pto.logMessage(fmt.Sprintf("TUI stop error: %v", err))
		} else {
			pto.logMessage("TUI stopped successfully")
		}

		pto.tui = nil
		pto.logMessage("TUI cleanup completed")
	}
}

// setDisplayStep updates the display step if TUI is enabled
func (pto *PTOrchestrator) setDisplayStep(step int) {
	if pto.tuiEnabled && pto.tui != nil {
		pto.tui.SetStep(step)
		pto.logMessage(fmt.Sprintf("Set TUI step to %d", step))
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
