package tui

import (
	"time"

	"ez-utils/src/scale/population"
)

// TUI message types for Bubble Tea
type AgentUpdateMsg struct {
	AgentType  string
	AgentID    string
	State      population.AgentState
	Timestamp  time.Time
	Details    map[string]interface{}
}

type ResourceUpdateMsg struct {
	CPU        float64
	RAM        float64
	Disk       float64
	Timestamp  time.Time
}

type PhaseUpdateMsg struct {
	Phase           string
	InputFile       string
	InputFileSizeMB float64
	GridDimensions  population.GridBounds
	BinCount        int
	DensityMapSize  int
	SafePointsCount int
	Timestamp       time.Time
}

type ReaderAgentMsg struct {
	AgentID           string
	State             population.AgentState
	TotalLinesRead    int64
	TotalPersonsFound int64
	DataProcessedMB   float64
	SafePointsWritten int64
	Timestamp         time.Time
}

type ExtractorAgentMsg struct {
	AgentID          string
	State            population.AgentState
	PersonsProcessed int64
	PersonsSkipped   int64
	Timestamp        time.Time
}

type HashMapAgentMsg struct {
	AgentID               string
	State                 population.AgentState
	QueueSize             int
	QueueFillPercentage   float64
	CoordinatesProcessed  int64
	ExpansionRequests     int64
	Timestamp             time.Time
}

type GridHandlerMsg struct {
	CurrentBounds     population.GridBounds
	BinCount          int
	ExpansionRequests int64
	LastExpansion     time.Time
	Timestamp         time.Time
}

// Command messages
type PauseResumeMsg struct {
	AgentType string
	AgentID   string
	Command   string // "pause" or "resume"
}

type ShutdownMsg struct{}

type TickMsg time.Time