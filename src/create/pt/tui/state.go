package tui

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
)

// State - Central state store (like Svelte store)
type State struct {
	// Process State
	CurrentStep     int
	TotalSteps      int
	ProcessComplete bool
	QuitRequested   bool

	// Timing
	StartTime     time.Time
	ElapsedTime   time.Duration
	StepDurations map[int]time.Duration // step -> duration

	// UI State
	Spinner     spinner.Model
	StatusBlink bool

	// System Stats
	CPUUsage float64
	RAMUsage float64

	// Step Data - Live counters for each step
	StepCounters map[int]StepCounter

	// PT Steps Definition
	Steps []PTStep
}

// PTStep - Definition of each processing step
type PTStep struct {
	ID          int
	Name        string
	Description string
}

// StepCounter - Live data for current step
type StepCounter struct {
	// Step 0: File validation
	FilesValidated int
	CurrentFile    string
	FileSize       int

	// Step 1: Service patterns
	ServicesFound int
	Progress      int // percentage

	// Step 2: Route mapping
	RoutesMapped int
	Mappings     int

	// Step 3: Service selection
	Selected int
	Total    int

	// Step 4: Database creation
	RowsImported int
	DatabaseSize int

	// Step 5: Trip collection
	TripsCollected int
	TripProgress   int

	// Step 6: Stop sequences
	Sequences int
	Stops     int

	// Step 7: Stop details
	StopDetails int

	// Step 8: XML generation
	XMLProgress int
	OutputSize  int

	// Step 9: Cleanup
	FilesDeleted int
	DirsDeleted  int
	BytesFreed   int
}

// NewState creates initial state
func NewState() *State {
	// Define the 10 PT processing steps
	steps := []PTStep{
		{0, "Validate GTFS Files", "Validating GTFS data integrity"},
		{1, "Identify Service Patterns", "Analyzing calendar.txt for service patterns"},
		{2, "Map Service Routes", "Linking services to routes"},
		{3, "Select Services", "Selecting bus/metro services"},
		{4, "Build Trip Database", "Creating SQLite database from trips.txt"},
		{5, "Collect Trips", "Querying selected trips from database"},
		{6, "Extract Stop Sequences", "Collecting stop sequences"},
		{7, "Gather Stop Details", "Collecting stop details"},
		{8, "Generate XML", "Creating MATSim transit schedule"},
		{9, "Cleanup Files", "Cleaning temporary files"},
	}

	return &State{
		CurrentStep:     0,
		TotalSteps:      10,
		ProcessComplete: false,
		QuitRequested:   false,
		StartTime:       time.Now(),
		ElapsedTime:     0,
		StepDurations:   make(map[int]time.Duration),
		StatusBlink:     false,
		CPUUsage:        0,
		RAMUsage:        0,
		StepCounters:    make(map[int]StepCounter),
		Steps:           steps,
	}
}

// SetStep updates the current step
func (s *State) SetStep(step int) {
	if s.CurrentStep >= 0 && step != s.CurrentStep {
		// Record duration for previous step
		s.StepDurations[s.CurrentStep] = s.ElapsedTime
	}
	s.CurrentStep = step
}

// UpdateCounter updates live data for current step
func (s *State) UpdateCounter(key string, value interface{}) {
	counter := s.StepCounters[s.CurrentStep]

	// Helper function to parse string or return int directly
	parseInt := func(v interface{}) int {
		switch val := v.(type) {
		case int:
			return val
		case string:
			// Parse string like "123" to int
			if parsed, err := strconv.Atoi(val); err == nil {
				return parsed
			}
			// Parse percentage strings like "85%" to int
			if strings.HasSuffix(val, "%") {
				if parsed, err := strconv.Atoi(strings.TrimSuffix(val, "%")); err == nil {
					return parsed
				}
			}
		}
		return 0
	}

	switch key {
	// Step 0: File validation
	case "file":
		if v, ok := value.(string); ok {
			counter.CurrentFile = v
		}
	case "files_validated":
		counter.FilesValidated = parseInt(value)
	case "file_size", "size":
		counter.FileSize = parseInt(value)
	case "status":
		// Status is for display only, no counter update needed

	// Step 1: Service patterns
	case "services":
		counter.ServicesFound = parseInt(value)
	case "progress":
		counter.Progress = parseInt(value)

	// Step 2: Route mapping
	case "routes":
		counter.RoutesMapped = parseInt(value)
	case "mappings":
		counter.Mappings = parseInt(value)

	// Step 3: Service selection
	case "selected":
		counter.Selected = parseInt(value)
	case "total":
		counter.Total = parseInt(value)

	// Step 4: Database creation
	case "rows":
		counter.RowsImported = parseInt(value)
	case "db_size":
		counter.DatabaseSize = parseInt(value)

	// Step 5: Trip collection
	case "trips":
		counter.TripsCollected = parseInt(value)
	case "trip_progress":
		counter.TripProgress = parseInt(value)

	// Step 6: Stop sequences
	case "sequences":
		counter.Sequences = parseInt(value)
	case "stops":
		// Handle both Step 6 sequences and Step 7 stop details
		if s.CurrentStep == 6 {
			counter.Stops = parseInt(value)
		} else if s.CurrentStep == 7 {
			counter.StopDetails = parseInt(value)
		}

	// Step 7: Stop details (also handled by "stops" above)
	case "stop_details":
		counter.StopDetails = parseInt(value)

	// Step 8: XML generation
	case "xml_progress_alt":
		// Handle Step 8 XML progress (don't conflict with other steps' progress)
		if s.CurrentStep == 8 {
			counter.XMLProgress = parseInt(value)
		}
	case "xml_progress":
		counter.XMLProgress = parseInt(value)
	case "output_size_alt":
		// Handle Step 8 output size
		if s.CurrentStep == 8 {
			counter.OutputSize = parseInt(value)
		}
	case "output_size":
		counter.OutputSize = parseInt(value)

	// Step 9: Cleanup
	case "files":
		counter.FilesDeleted = parseInt(value)
	case "dirs":
		counter.DirsDeleted = parseInt(value)
	case "bytes":
		counter.BytesFreed = parseInt(value)
	}

	s.StepCounters[s.CurrentStep] = counter
}

// SetSystemStats updates CPU and RAM usage
func (s *State) SetSystemStats(cpu, ram float64) {
	s.CPUUsage = cpu
	s.RAMUsage = ram
}

// SetProcessComplete marks the process as complete
func (s *State) SetProcessComplete(complete bool) {
	s.ProcessComplete = complete
}

// RequestQuit sets the quit flag
func (s *State) RequestQuit() {
	s.QuitRequested = true
}

// UpdateElapsedTime updates the elapsed time
func (s *State) UpdateElapsedTime() {
	s.ElapsedTime = time.Since(s.StartTime)
}

// ToggleStatusBlink toggles the status indicator
func (s *State) ToggleStatusBlink() {
	s.StatusBlink = !s.StatusBlink
}
