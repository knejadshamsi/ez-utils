package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// UnifiedModel represents the main TUI state for all phases
type UnifiedModel struct {
	// Core phase tracking
	currentPhase    int    // 1-5
	currentAction   string
	startTime       time.Time
	
	// System metrics
	cpuUsage        float64
	ramUsage        float64
	diskUsage       float64
	
	// Phase-specific worker data
	readerWorkers     []WorkerData
	extractorWorkers  []WorkerData
	mapperWorkers     []WorkerData
	reducerWorkers    []WorkerData
	writerWorkers     []WorkerData
	
	// UI state
	width           int
	height          int
	quitting        bool
	
	// Styles
	styles          UnifiedStyles
}

// WorkerData represents worker information across all types
type WorkerData struct {
	Name            string
	Status          WorkerStatus
	PrimaryMetric   string
	SecondaryMetric string
	WorkerType      WorkerType
}

// WorkerStatus represents the status of a worker
type WorkerStatus int

const (
	StatusActive WorkerStatus = iota
	StatusCompleted
	StatusIdle
)

// WorkerType represents the type of worker
type WorkerType int

const (
	TypeReader WorkerType = iota
	TypeExtractor
	TypeMapper
	TypeReducer
	TypeWriter
)

// NewModel creates a new TUI model
func NewModel() UnifiedModel {
	return UnifiedModel{
		startTime:        time.Now(),
		currentPhase:     1,
		currentAction:    "Validating Configuration and User Input",
		readerWorkers:    make([]WorkerData, 0),
		extractorWorkers: make([]WorkerData, 0),
		mapperWorkers:    make([]WorkerData, 0),
		reducerWorkers:   make([]WorkerData, 0),
		writerWorkers:    make([]WorkerData, 0),
		width:           80,  // Default minimum width
		height:          24,  // Default minimum height
		styles:           NewStyles(),
	}
}

// Init initializes the model
func (m UnifiedModel) Init() tea.Cmd {
	return m.tick()
}

// tick returns a command that sends a tick message after a delay
func (m UnifiedModel) tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMessage(t)
	})
}

// TickMessage represents a timer tick
type TickMessage time.Time

// Helper methods
func (m *UnifiedModel) GetElapsedTime() string {
	elapsed := time.Since(m.startTime)
	minutes := int(elapsed.Minutes())
	seconds := int(elapsed.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func (m *UnifiedModel) GetPhase() string {
	phaseNames := []string{"ONE", "TWO", "THREE", "FOUR", "FIVE"}
	if m.currentPhase > 0 && m.currentPhase <= 5 {
		return fmt.Sprintf("= PHASE %s OF FIVE =", phaseNames[m.currentPhase-1])
	}
	return "= PHASE ONE OF FIVE ="
}

// TUI sizing constants
const (
	MinTerminalWidth  = 40  // Minimum usable terminal width
	MinTerminalHeight = 10  // Minimum usable terminal height
	DefaultWidth      = 120 // Reasonable default width
	DefaultHeight     = 30  // Reasonable default height
	BorderPadding     = 2   // Space used by borders
	TablePadding      = 4   // Space used by table formatting
)

// safeWidth returns a safe width value that prevents negative calculations
func (m *UnifiedModel) safeWidth() int {
	if m.width < MinTerminalWidth {
		return DefaultWidth
	}
	return m.width
}

// safeHeight returns a safe height value
func (m *UnifiedModel) safeHeight() int {
	if m.height < MinTerminalHeight {
		return DefaultHeight
	}
	return m.height
}

// safeBorderWidth returns a safe width for border strings.Repeat()
func (m *UnifiedModel) safeBorderWidth() int {
	width := m.safeWidth() - BorderPadding
	if width < 1 {
		return 1 // Minimum border width
	}
	return width
}

// safeTableWidth returns a safe width for tables
func (m *UnifiedModel) safeTableWidth() int {
	width := m.safeWidth() - TablePadding
	if width < MinTerminalWidth/2 {
		return MinTerminalWidth / 2 // Minimum table width
	}
	return width
}

func (m *UnifiedModel) GetSystemMetrics() SystemMetrics {
	// Get real-time metrics from system monitor
	cpu, ram, disk := GetCurrentSystemMetrics()
	
	// Don't modify state in getter - return current metrics
	return SystemMetrics{
		CPU:  int(cpu),
		RAM:  int(ram),
		DISK: int(disk),
	}
}

// SystemMetrics represents system resource usage
type SystemMetrics struct {
	CPU  int
	RAM  int
	DISK int
}

// Update worker data methods
func (m *UnifiedModel) UpdateReaderWorkers(workers []WorkerData) {
	m.readerWorkers = workers
}

func (m *UnifiedModel) UpdateExtractorWorkers(workers []WorkerData) {
	m.extractorWorkers = workers
}

func (m *UnifiedModel) UpdateMapperWorkers(workers []WorkerData) {
	m.mapperWorkers = workers
}

func (m *UnifiedModel) UpdateReducerWorkers(workers []WorkerData) {
	m.reducerWorkers = workers
}

func (m *UnifiedModel) UpdateWriterWorkers(workers []WorkerData) {
	m.writerWorkers = workers
}

func (m *UnifiedModel) UpdateSystemMetrics(cpu, ram, disk float64) {
	m.cpuUsage = cpu
	m.ramUsage = ram
	m.diskUsage = disk
}

func (m *UnifiedModel) UpdatePhase(phase int, action string) {
	m.currentPhase = phase
	m.currentAction = action
}

// GetActiveWorkers returns workers for the current phase
func (m *UnifiedModel) GetActiveWorkers() []WorkerData {
	switch m.currentPhase {
	case 1:
		return []WorkerData{} // No workers in phase 1
	case 2:
		// Return readers, extractors, and mappers
		workers := make([]WorkerData, 0)
		workers = append(workers, m.readerWorkers...)
		workers = append(workers, m.extractorWorkers...)
		workers = append(workers, m.mapperWorkers...)
		return workers
	case 3:
		return []WorkerData{} // No active workers in phase 3
	case 4:
		// Return readers, reducers, and writers
		workers := make([]WorkerData, 0)
		workers = append(workers, m.readerWorkers...)
		workers = append(workers, m.reducerWorkers...)
		workers = append(workers, m.writerWorkers...)
		return workers
	case 5:
		return []WorkerData{} // No workers in phase 5
	default:
		return []WorkerData{}
	}
}