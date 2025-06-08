package display

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var p *tea.Program
var model Model

// InitDisplay sets up the TUI with system monitoring and non-blocking operation
func InitDisplay(moduleName string, flagClean bool, flagDB bool, language string) {
	// Initialize text manager with specified language
	if err := InitializeTextManager(language); err != nil {
		// Log error but continue with fallback behavior
		// fmt.Printf("Warning: Failed to initialize text manager: %v\n", err)
	}
	spn := spinner.New()
	spn.Spinner = spinner.Line
	spn.Style = lipgloss.NewStyle().Foreground(colorMagenta)

	tmr := timer.NewWithInterval(0, time.Second)

	model = Model{
		// Basic configuration (module, flags)
		moduleName: moduleName,
		stepNumber: 0,
		flagClean:  flagClean,
		flagDB:     flagDB,

		// Time tracking
		startTime:        time.Now(),
		elapsedTime:      0,
		totalStartTime:   time.Now(),
		totalElapsedTime: 0,
		stepDurations:    make(map[int]time.Duration),

		// UI state
		spinner:     spn,
		timer:       tmr,
		statusBlink: false,

		// Resource monitoring
		cpuUsage: 0,
		ramUsage: 0,

		// Process state
		processComplete: false,
		err:             nil,

		// Progress tracking
		chunkCount:      0,
		personsFound:    0,
		bytesReadMB:     0,
		firstChunkFixed: false,
		lastChunkFixed:  false,
		agentCounter:    0,
		chunkCounter:    Counter{0, 0},
		dbCounter:       Counter{0, 0},
		coordinateCount: 0,
		binCounter:      Counter{0, 0},
		agentBinCount:   0,
		scaleCounter:    Counter{0, 0},
		outputScale:     0,
		outputCounter:   Counter{0, 0},
		cleanupFiles:    0,
		cleanupDirs:     0,
		cleanupBytes:    0,
	}

	p = tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Run in goroutine to prevent blocking the main process while UI is active
	go func() {
		_ = p.Start()
	}()

	// System monitoring in a separate goroutine
	go monitorSystem()
}

// Setters for updating the model states

func SetStep(stepNumber int) {
	if p != nil {
		p.Send(stepMsg(stepNumber))
	}
}

func SetProcessComplete(complete bool) {
	if p != nil {
		p.Send(processCompleteMsg(complete))
	}
}

// XML Processing Functions (Steps 0-1) - Chunk parsing and validation
func SetChunkCount(count int) {
	if p != nil {
		p.Send(chunkCountMsg(count))
	}
}

func SetPersonsFound(count int) {
	if p != nil {
		p.Send(personsFoundMsg(count))
	}
}

func SetBytesReadMB(mb int) {
	if p != nil {
		p.Send(bytesReadMsg(mb))
	}
}

func SetFirstChunkStatus(fixed bool) {
	if p != nil {
		p.Send(firstChunkFixedMsg(fixed))
	}
}

func SetLastChunkStatus(fixed bool) {
	if p != nil {
		p.Send(lastChunkFixedMsg(fixed))
	}
}

// Location Extraction Functions (Steps 2-3) - Geographic data processing
func SetAgentCounter(count int) {
	if p != nil {
		p.Send(agentCounterMsg(count))
	}
}

func SetChunkCounter(current, total int) {
	if p != nil {
		p.Send(chunkCounterMsg{current, total})
	}
}

func SetDbCounter(current, total int) {
	if p != nil {
		p.Send(dbCounterMsg{current, total})
	}
}

// Processing Functions (Steps 4-18) - Data transformation and cleanup
func SetCoordinateCounter(count int) {
	if p != nil {
		p.Send(coordinateCounterMsg(count))
	}
}

func SetBinCounter(current, total int) {
	if p != nil {
		p.Send(binCounterMsg{current, total})
	}
}

func SetAgentBinCounter(count int) {
	if p != nil {
		p.Send(agentBinCounterMsg(count))
	}
}

func SetScaleCounter(current, total int) {
	if p != nil {
		p.Send(scaleCounterMsg{current, total})
	}
}

func SetOutputScale(scale int) {
	if p != nil {
		p.Send(outputScaleMsg(scale))
	}
}

func SetOutputCounter(current, total int) {
	if p != nil {
		p.Send(outputCounterMsg{current, total})
	}
}

func SetCleanupCounter(files, dirs, bytes int) {
	if p != nil {
		p.Send(cleanupCounterMsg{files, dirs, bytes})
	}
}

// Message types for state updates
type stepMsg int
type processCompleteMsg bool
type systemStatsMsg struct {
	cpu float64
	ram float64
}
type tickMsg struct{}
type blinkMsg struct{}

// Message types for XML processing  (steps 0-1)
type chunkCountMsg int
type personsFoundMsg int
type bytesReadMsg int
type firstChunkFixedMsg bool
type lastChunkFixedMsg bool

// Message types for location processing phase (steps 2-3)
type agentCounterMsg int
type chunkCounterMsg struct {
	current int
	total   int
}
type dbCounterMsg struct {
	current int
	total   int
}

// Message types for data transformation (steps 4-18)
type coordinateCounterMsg int
type binCounterMsg struct {
	current int
	total   int
}
type agentBinCounterMsg int
type scaleCounterMsg struct {
	current int
	total   int
}
type outputScaleMsg int
type outputCounterMsg struct {
	current int
	total   int
}
type cleanupCounterMsg struct {
	files int
	dirs  int
	bytes int
}
