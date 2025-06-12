package tui

import (
	"fmt"
	"time"

	"ez-utils/src/scale/population"
)

// Model represents the main TUI state
type Model struct {
	// Core state
	startTime       time.Time
	currentPhase    string
	
	// System metrics
	cpuUsage        float64
	ramUsage        float64
	diskUsage       float64
	lastUpdate      time.Time
	
	// Phase information
	inputFile       string
	inputFileSizeMB float64
	gridBounds      population.GridBounds
	binCount        int
	densityMapSize  int
	safePointsCount int
	
	// Agent states
	readerAgent     ReaderAgentState
	extractorAgents map[string]ExtractorAgentState
	hashmapAgents   map[string]HashMapAgentState
	gridHandler     GridHandlerState
	
	// UI state
	styles          StyleDefinitions
	width           int
	height          int
	paused          bool
	
	// Channels for updates
	updateChan      chan interface{}
}

// Agent state structures
type ReaderAgentState struct {
	ID               string
	State            population.AgentState
	TotalLinesRead   int64
	TotalPersonsFound int64
	DataProcessedMB  float64
	SafePointsWritten int64
	LastUpdate       time.Time
}

type ExtractorAgentState struct {
	ID               string
	State            population.AgentState
	PersonsProcessed int64
	PersonsSkipped   int64
	LastUpdate       time.Time
}

type HashMapAgentState struct {
	ID                   string
	State                population.AgentState
	QueueSize            int
	QueueFillPercentage  float64
	CoordinatesProcessed int64
	ExpansionRequests    int64
	LastUpdate           time.Time
}

type GridHandlerState struct {
	CurrentBounds     population.GridBounds
	BinCount          int
	ExpansionRequests int64
	LastExpansion     time.Time
	LastUpdate        time.Time
}

// NewModel creates a new TUI model
func NewModel() Model {
	return Model{
		startTime:       time.Now(),
		currentPhase:    "Phase One",
		extractorAgents: make(map[string]ExtractorAgentState),
		hashmapAgents:   make(map[string]HashMapAgentState),
		styles:          NewStyleDefinitions(),
		updateChan:      make(chan interface{}, 100),
	}
}

// Update methods for different agent types
func (m *Model) UpdateReaderAgent(msg ReaderAgentMsg) {
	m.readerAgent = ReaderAgentState{
		ID:                msg.AgentID,
		State:             msg.State,
		TotalLinesRead:    msg.TotalLinesRead,
		TotalPersonsFound: msg.TotalPersonsFound,
		DataProcessedMB:   msg.DataProcessedMB,
		SafePointsWritten: msg.SafePointsWritten,
		LastUpdate:        msg.Timestamp,
	}
}

func (m *Model) UpdateExtractorAgent(msg ExtractorAgentMsg) {
	m.extractorAgents[msg.AgentID] = ExtractorAgentState{
		ID:               msg.AgentID,
		State:            msg.State,
		PersonsProcessed: msg.PersonsProcessed,
		PersonsSkipped:   msg.PersonsSkipped,
		LastUpdate:       msg.Timestamp,
	}
}

func (m *Model) UpdateHashMapAgent(msg HashMapAgentMsg) {
	m.hashmapAgents[msg.AgentID] = HashMapAgentState{
		ID:                   msg.AgentID,
		State:                msg.State,
		QueueSize:            msg.QueueSize,
		QueueFillPercentage:  msg.QueueFillPercentage,
		CoordinatesProcessed: msg.CoordinatesProcessed,
		ExpansionRequests:    msg.ExpansionRequests,
		LastUpdate:           msg.Timestamp,
	}
}

func (m *Model) UpdateGridHandler(msg GridHandlerMsg) {
	m.gridHandler = GridHandlerState{
		CurrentBounds:     msg.CurrentBounds,
		BinCount:          msg.BinCount,
		ExpansionRequests: msg.ExpansionRequests,
		LastExpansion:     msg.LastExpansion,
		LastUpdate:        msg.Timestamp,
	}
}

func (m *Model) UpdateSystemMetrics(msg ResourceUpdateMsg) {
	m.cpuUsage = msg.CPU
	m.ramUsage = msg.RAM
	m.diskUsage = msg.Disk
	m.lastUpdate = msg.Timestamp
}

func (m *Model) UpdatePhaseInfo(msg PhaseUpdateMsg) {
	m.currentPhase = msg.Phase
	m.inputFile = msg.InputFile
	m.inputFileSizeMB = msg.InputFileSizeMB
	m.gridBounds = msg.GridDimensions
	m.binCount = msg.BinCount
	m.densityMapSize = msg.DensityMapSize
	m.safePointsCount = msg.SafePointsCount
}

// Helper methods
func (m *Model) GetElapsedTime() time.Duration {
	return time.Since(m.startTime)
}

func (m *Model) FormatElapsedTime() string {
	elapsed := m.GetElapsedTime()
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	seconds := int(elapsed.Seconds()) % 60
	return m.styles.timeInfo.Render(fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds))
}

func (m *Model) GetActiveExtractorCount() int {
	count := 0
	for _, agent := range m.extractorAgents {
		if agent.State == population.Busy || agent.State == population.Available {
			count++
		}
	}
	return count
}

func (m *Model) GetTotalExtractorCount() int {
	return len(m.extractorAgents)
}

func (m *Model) GetActiveHashMapCount() int {
	count := 0
	for _, agent := range m.hashmapAgents {
		if agent.State == population.Busy || agent.State == population.Available {
			count++
		}
	}
	return count
}

func (m *Model) GetTotalHashMapCount() int {
	return len(m.hashmapAgents)
}

func (m *Model) GetTotalPersonsProcessed() int64 {
	total := int64(0)
	for _, agent := range m.extractorAgents {
		total += agent.PersonsProcessed
	}
	return total
}

func (m *Model) GetTotalPersonsSkipped() int64 {
	total := int64(0)
	for _, agent := range m.extractorAgents {
		total += agent.PersonsSkipped
	}
	return total
}

func (m *Model) GetTotalCoordinatesProcessed() int64 {
	total := int64(0)
	for _, agent := range m.hashmapAgents {
		total += agent.CoordinatesProcessed
	}
	return total
}

func (m *Model) GetTotalExpansionRequests() int64 {
	total := int64(0)
	for _, agent := range m.hashmapAgents {
		total += agent.ExpansionRequests
	}
	return total
}

func (m *Model) GetAverageQueueFill() float64 {
	if len(m.hashmapAgents) == 0 {
		return 0.0
	}
	
	total := 0.0
	for _, agent := range m.hashmapAgents {
		total += agent.QueueFillPercentage
	}
	return total / float64(len(m.hashmapAgents))
}

// State conversion helpers
func (m *Model) GetAgentStateString(state population.AgentState) string {
	switch state {
	case population.Available:
		return "AVAILABLE"
	case population.Busy:
		return "BUSY"
	case population.Paused:
		return "PAUSED"
	default:
		return "UNKNOWN"
	}
}