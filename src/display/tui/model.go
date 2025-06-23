package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"

	"ez-utils/src/display/config"
)

type Counter struct {
	Current int
	Total   int
}

// Model represents the state of the TUI
type Model struct {
	// Core configuration
	ModuleName   string
	ProcessTitle string      // Configured process title/description
	StepNumber   int
	MaxSteps     int
	Flags        map[string]bool
	Steps        []config.StepConfig // Changed to config.StepConfig

	// Timing information
	StartTime        time.Time
	ElapsedTime      time.Duration
	TotalStartTime   time.Time
	TotalElapsedTime time.Duration
	StepDurations    map[int]time.Duration

	// UI components
	Spinner     spinner.Model
	Timer       timer.Model
	StatusBlink bool

	// System monitoring
	CPUUsage  float64
	RAMUsage  float64
	DiskUsage float64

	// Status
	ProcessComplete bool
	Err             error
	
	// Theme support
	ThemeColors map[string]string

	// Step counters by logical group
	// XML Processing (Steps 0-1)
	ChunkCount      int
	PersonsFound    int
	BytesReadMB     int
	FirstChunkFixed bool
	LastChunkFixed  bool

	// Location Extraction (Steps 2-3)
	AgentCounter int
	ChunkCounter Counter
	DbCounter    Counter

	// Boundary and Binning (Steps 6-11)
	CoordinateCount int
	BinCounter      Counter
	AgentBinCount   int

	// Scaling and Output (Steps 12-18)
	ScaleCounter  Counter
	OutputScale   int
	OutputCounter Counter
	CleanupFiles  int
	CleanupDirs   int
	CleanupBytes  int
}

// Init initializes the model and starts animations
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		TickCmd(), // Changed to TickCmd()
		BlinkCmd(), // Changed to BlinkCmd()
	)
}

// Sends a tick message every second
func TickCmd() tea.Cmd { // Changed to TickCmd()
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return TickMsg{} // Changed to TickMsg{}
	})
}

// Sends a blink message every 800ms
func BlinkCmd() tea.Cmd { // Changed to BlinkCmd()
	return tea.Tick(800*time.Millisecond, func(time.Time) tea.Msg {
		return BlinkMsg{} // Changed to BlinkMsg{}
	})
}

// getFlag safely retrieves a flag value from the flags map
func (m Model) GetFlag(flag string) bool { // Renamed to GetFlag and made public
	if m.Flags == nil {
		return false
	}
	value, exists := m.Flags[flag]
	return exists && value
}
