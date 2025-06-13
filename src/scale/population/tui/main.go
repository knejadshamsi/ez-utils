package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Global variables for the TUI
var globalProgram *tea.Program
var globalModel *UnifiedModel

// InitTUI initializes the TUI system
func InitTUI() {
	model := NewModel()
	globalModel = &model
	globalProgram = tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	
	// Start global system monitoring for real-time metrics
	StartGlobalSystemMonitor()
	
	// Start global worker tracking for real-time worker updates
	StartGlobalWorkerTracker()
	
	// Run in goroutine to prevent blocking the main process
	go func() {
		_ = globalProgram.Start()
		// Stop monitoring when TUI exits
		StopGlobalSystemMonitor()
		StopGlobalWorkerTracker()
	}()
}

// StopTUI stops the TUI
func StopTUI() {
	if globalProgram != nil {
		globalProgram.Send(NewShutdownMsg())
		globalProgram.Quit()
		
		// Clean up global references to prevent memory leaks
		globalProgram = nil
		globalModel = nil
	}
}

// UpdateWorkers updates workers for a specific type
func UpdateWorkers(workerType WorkerType, workers []WorkerData) {
	if globalProgram != nil {
		globalProgram.Send(NewWorkerUpdateMsg(workerType, workers))
	}
}

// UpdatePhase updates the current phase and action
func UpdatePhase(phase int, action string) {
	if globalProgram != nil {
		globalProgram.Send(NewPhaseUpdateMsg(phase, action))
	}
}

// UpdateSystemMetrics updates system resource metrics
func UpdateSystemMetrics(cpu, ram, disk float64) {
	if globalProgram != nil {
		globalProgram.Send(NewSystemUpdateMsg(cpu, ram, disk))
	}
}

// StartTUI creates and starts the TUI
func StartTUI() (*tea.Program, error) {
	model := NewModel()
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	return p, nil
}

