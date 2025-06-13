package tui

import (
	"time"
)

// Message types for the TUI system

// UnifiedWorkerUpdateMsg updates workers of a specific type
type UnifiedWorkerUpdateMsg struct {
	WorkerType WorkerType
	Workers    []WorkerData
	Timestamp  time.Time
}

// UnifiedPhaseUpdateMsg updates the current phase and action
type UnifiedPhaseUpdateMsg struct {
	Phase     int
	Action    string
	Timestamp time.Time
}

// UnifiedSystemUpdateMsg updates system metrics
type UnifiedSystemUpdateMsg struct {
	CPU       float64
	RAM       float64
	Disk      float64
	Timestamp time.Time
}

// UnifiedShutdownMsg signals the TUI to shutdown
type UnifiedShutdownMsg struct{}


// Helper functions to create messages

// NewWorkerUpdateMsg creates a new worker update message
func NewWorkerUpdateMsg(workerType WorkerType, workers []WorkerData) UnifiedWorkerUpdateMsg {
	return UnifiedWorkerUpdateMsg{
		WorkerType: workerType,
		Workers:    workers,
		Timestamp:  time.Now(),
	}
}

// NewPhaseUpdateMsg creates a new phase update message
func NewPhaseUpdateMsg(phase int, action string) UnifiedPhaseUpdateMsg {
	return UnifiedPhaseUpdateMsg{
		Phase:     phase,
		Action:    action,
		Timestamp: time.Now(),
	}
}

// NewSystemUpdateMsg creates a new system update message
func NewSystemUpdateMsg(cpu, ram, disk float64) UnifiedSystemUpdateMsg {
	return UnifiedSystemUpdateMsg{
		CPU:       cpu,
		RAM:       ram,
		Disk:      disk,
		Timestamp: time.Now(),
	}
}

// NewShutdownMsg creates a new shutdown message
func NewShutdownMsg() UnifiedShutdownMsg {
	return UnifiedShutdownMsg{}
}

// UnifiedWorkerMetricUpdateMsg updates only metrics for a specific worker
type UnifiedWorkerMetricUpdateMsg struct {
	WorkerID        string
	WorkerType      WorkerType
	PrimaryMetric   string
	SecondaryMetric string
	Status          WorkerStatus
	Timestamp       time.Time
}

// NewWorkerMetricUpdateMsg creates a new worker metric update message
func NewWorkerMetricUpdateMsg(workerID, primary, secondary string, workerType WorkerType, status WorkerStatus) UnifiedWorkerMetricUpdateMsg {
	return UnifiedWorkerMetricUpdateMsg{
		WorkerID:        workerID,
		WorkerType:      workerType,
		PrimaryMetric:   primary,
		SecondaryMetric: secondary,
		Status:          status,
		Timestamp:       time.Now(),
	}
}