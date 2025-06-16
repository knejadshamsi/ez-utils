package display

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

type Counter struct {
	Current int
	Total   int
}

// Model represents the state of the TUI
type Model struct {
	// Core configuration
	moduleName   string
	processTitle string      // Configured process title/description
	stepNumber   int
	maxSteps     int
	flags        map[string]bool
	steps        []StepConfig // Configured step titles and descriptions

	// Timing information
	startTime        time.Time
	elapsedTime      time.Duration
	totalStartTime   time.Time
	totalElapsedTime time.Duration
	stepDurations    map[int]time.Duration

	// UI components
	spinner     spinner.Model
	timer       timer.Model
	statusBlink bool

	// System monitoring
	cpuUsage  float64
	ramUsage  float64
	diskUsage float64

	// Status
	processComplete bool
	err             error
	
	// Theme support
	themeColors map[string]string

	// Step counters by logical group
	// XML Processing (Steps 0-1)
	chunkCount      int
	personsFound    int
	bytesReadMB     int
	firstChunkFixed bool
	lastChunkFixed  bool

	// Location Extraction (Steps 2-3)
	agentCounter int
	chunkCounter Counter
	dbCounter    Counter

	// Boundary and Binning (Steps 6-11)
	coordinateCount int
	binCounter      Counter
	agentBinCount   int

	// Scaling and Output (Steps 12-18)
	scaleCounter  Counter
	outputScale   int
	outputCounter Counter
	cleanupFiles  int
	cleanupDirs   int
	cleanupBytes  int
}

// Init initializes the model and starts animations
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		tickCmd(),
		blinkCmd(),
	)
}

// Sends a tick message every second
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Sends a blink message every 800ms
func blinkCmd() tea.Cmd {
	return tea.Tick(800*time.Millisecond, func(time.Time) tea.Msg {
		return blinkMsg{}
	})
}

// getFlag safely retrieves a flag value from the flags map
func (m Model) getFlag(flag string) bool {
	if m.flags == nil {
		return false
	}
	value, exists := m.flags[flag]
	return exists && value
}
