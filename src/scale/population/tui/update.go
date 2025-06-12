package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all incoming messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	
	// Handle window size changes
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	
	// Handle keyboard input
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "p":
			m.paused = !m.paused
			return m, nil
		case "r":
			// Reset/refresh display
			return m, nil
		}
	
	// Handle agent updates
	case ReaderAgentMsg:
		m.UpdateReaderAgent(msg)
		return m, nil
	
	case ExtractorAgentMsg:
		m.UpdateExtractorAgent(msg)
		return m, nil
	
	case HashMapAgentMsg:
		m.UpdateHashMapAgent(msg)
		return m, nil
	
	case GridHandlerMsg:
		m.UpdateGridHandler(msg)
		return m, nil
	
	// Handle system updates
	case ResourceUpdateMsg:
		m.UpdateSystemMetrics(msg)
		return m, nil
	
	case PhaseUpdateMsg:
		m.UpdatePhaseInfo(msg)
		return m, nil
	
	// Handle timer ticks
	case TickMsg:
		return m, m.tick()
	
	// Handle shutdown
	case ShutdownMsg:
		return m, tea.Quit
	
	// Handle pause/resume commands
	case PauseResumeMsg:
		// This would send commands back to the manager
		return m, nil
	}
	
	return m, nil
}

// tick returns a command that sends a TickMsg after a short delay
func (m Model) tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Init initializes the model and starts the tick timer
func (m Model) Init() tea.Cmd {
	return m.tick()
}

// Batch updates for efficiency
func BatchUpdates(updates ...tea.Cmd) tea.Cmd {
	return tea.Batch(updates...)
}

// Helper commands for sending messages
func SendReaderUpdate(agentID string, state interface{}, details map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		return ReaderAgentMsg{
			AgentID:           agentID,
			TotalLinesRead:    getInt64FromDetails(details, "lines_read"),
			TotalPersonsFound: getInt64FromDetails(details, "persons_found"),
			DataProcessedMB:   getFloat64FromDetails(details, "data_mb"),
			SafePointsWritten: getInt64FromDetails(details, "safe_points"),
			Timestamp:         time.Now(),
		}
	}
}

func SendExtractorUpdate(agentID string, details map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		return ExtractorAgentMsg{
			AgentID:          agentID,
			PersonsProcessed: getInt64FromDetails(details, "persons_processed"),
			PersonsSkipped:   getInt64FromDetails(details, "persons_skipped"),
			Timestamp:        time.Now(),
		}
	}
}

func SendHashMapUpdate(agentID string, details map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		return HashMapAgentMsg{
			AgentID:               agentID,
			QueueSize:             getIntFromDetails(details, "queue_size"),
			QueueFillPercentage:   getFloat64FromDetails(details, "queue_fill"),
			CoordinatesProcessed:  getInt64FromDetails(details, "coords_processed"),
			ExpansionRequests:     getInt64FromDetails(details, "expansions"),
			Timestamp:             time.Now(),
		}
	}
}

func SendResourceUpdate(cpu, ram, disk float64) tea.Cmd {
	return func() tea.Msg {
		return ResourceUpdateMsg{
			CPU:       cpu,
			RAM:       ram,
			Disk:      disk,
			Timestamp: time.Now(),
		}
	}
}

func SendPhaseUpdate(phase, inputFile string, fileSizeMB float64) tea.Cmd {
	return func() tea.Msg {
		return PhaseUpdateMsg{
			Phase:           phase,
			InputFile:       inputFile,
			InputFileSizeMB: fileSizeMB,
			Timestamp:       time.Now(),
		}
	}
}

// Helper functions to extract values from details map
func getInt64FromDetails(details map[string]interface{}, key string) int64 {
	if val, ok := details[key]; ok {
		switch v := val.(type) {
		case int64:
			return v
		case int:
			return int64(v)
		case float64:
			return int64(v)
		}
	}
	return 0
}

func getIntFromDetails(details map[string]interface{}, key string) int {
	if val, ok := details[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

func getFloat64FromDetails(details map[string]interface{}, key string) float64 {
	if val, ok := details[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0.0
}