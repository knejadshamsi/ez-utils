package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all incoming messages and updates the model
func (m UnifiedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	
	// Handle window size changes
	case tea.WindowSizeMsg:
		// Ensure minimum safe dimensions
		if msg.Width < 10 {
			m.width = 80
		} else {
			m.width = msg.Width
		}
		if msg.Height < 5 {
			m.height = 24
		} else {
			m.height = msg.Height
		}
		return m, nil
	
	// Handle keyboard input
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	
	// Handle worker updates
	case UnifiedWorkerUpdateMsg:
		m.updateWorkers(msg)
		return m, nil
	
	// Handle incremental worker metric updates
	case UnifiedWorkerMetricUpdateMsg:
		(&m).updateWorkerMetrics(msg)
		return m, nil
	
	// Handle phase updates
	case UnifiedPhaseUpdateMsg:
		m.UpdatePhase(msg.Phase, msg.Action)
		return m, nil
	
	// Handle system metrics updates
	case UnifiedSystemUpdateMsg:
		m.UpdateSystemMetrics(msg.CPU, msg.RAM, msg.Disk)
		return m, nil
	
	// Handle timer ticks
	case TickMessage:
		return m, m.tick()
	
	// Handle shutdown
	case UnifiedShutdownMsg:
		m.quitting = true
		return m, tea.Quit
	}
	
	return m, nil
}

// updateWorkers updates workers based on message type (maintains full replacement for initial setup)
func (m *UnifiedModel) updateWorkers(msg UnifiedWorkerUpdateMsg) {
	switch msg.WorkerType {
	case TypeReader:
		m.UpdateReaderWorkers(msg.Workers)
	case TypeExtractor:
		m.UpdateExtractorWorkers(msg.Workers)
	case TypeMapper:
		m.UpdateMapperWorkers(msg.Workers)
	case TypeReducer:
		m.UpdateReducerWorkers(msg.Workers)
	case TypeWriter:
		m.UpdateWriterWorkers(msg.Workers)
	}
}

// updateWorkerMetrics updates only the metrics for a specific worker (in-place)
func (m *UnifiedModel) updateWorkerMetrics(msg UnifiedWorkerMetricUpdateMsg) {
	// Try to find and update existing worker
	workerArrays := [](*[]WorkerData){
		&m.readerWorkers, &m.extractorWorkers, &m.mapperWorkers, &m.reducerWorkers, &m.writerWorkers,
	}
	
	for _, workers := range workerArrays {
		for i := range *workers {
			if (*workers)[i].Name == msg.WorkerID {
				(*workers)[i].PrimaryMetric = msg.PrimaryMetric
				(*workers)[i].SecondaryMetric = msg.SecondaryMetric
				(*workers)[i].Status = msg.Status
				return
			}
		}
	}
	
	// Worker not found - create new worker in appropriate array
	newWorker := WorkerData{
		Name:            msg.WorkerID,
		Status:          msg.Status,
		PrimaryMetric:   msg.PrimaryMetric,
		SecondaryMetric: msg.SecondaryMetric,
		WorkerType:      msg.WorkerType,
	}
	
	switch msg.WorkerType {
	case TypeReader:
		m.readerWorkers = append(m.readerWorkers, newWorker)
	case TypeExtractor:
		m.extractorWorkers = append(m.extractorWorkers, newWorker)
	case TypeMapper:
		m.mapperWorkers = append(m.mapperWorkers, newWorker)
	case TypeReducer:
		m.reducerWorkers = append(m.reducerWorkers, newWorker)
	case TypeWriter:
		m.writerWorkers = append(m.writerWorkers, newWorker)
	}
}

// nextAction advances to the next action in the current phase
func (m UnifiedModel) nextAction() tea.Cmd {
	actions := getPhaseActions(m.currentPhase)
	if len(actions) == 0 {
		return nil
	}
	
	// Find current action and advance
	for i, action := range actions {
		if action == m.currentAction {
			if i < len(actions)-1 {
				return func() tea.Msg {
					return UnifiedPhaseUpdateMsg{
						Phase:  m.currentPhase,
						Action: actions[i+1],
					}
				}
			} else {
				// Last action, advance to next phase
				return m.nextPhase()
			}
		}
	}
	
	return nil
}

// nextPhase advances to the next phase
func (m UnifiedModel) nextPhase() tea.Cmd {
	if m.currentPhase < 5 {
		nextPhase := m.currentPhase + 1
		actions := getPhaseActions(nextPhase)
		firstAction := "Initializing..."
		if len(actions) > 0 {
			firstAction = actions[0]
		}
		
		return func() tea.Msg {
			return UnifiedPhaseUpdateMsg{
				Phase:  nextPhase,
				Action: firstAction,
			}
		}
	}
	
	return nil
}

// getPhaseActions returns the actions for a given phase
func getPhaseActions(phase int) []string {
	switch phase {
	case 1:
		return []string{
			"Validating Configuration and User Input",
			"Loading Population File",
			"Studying the File",
		}
	case 2:
		return []string{
			"Starting Reader Workers",
			"Starting Extractor Workers",
			"Starting Mapper Workers",
			"Processing Population Data",
			"Mapping Complete",
		}
	case 3:
		return []string{
			"Combining Local Grids",
			"Removing Empty Bins",
			"Calculating Retention Ratio",
			"Retention Calculation Complete",
		}
	case 4:
		return []string{
			"Starting Reader Workers",
			"Starting Reducer Workers",
			"Starting Writer Workers",
			"Processing Scale Selection",
			"Population Scaling Complete",
		}
	case 5:
		return []string{
			"Combining Part Files",
			"Validating Output Files",
			"Cleaning Temporary Files",
			"Process Complete",
		}
	default:
		return []string{}
	}
}