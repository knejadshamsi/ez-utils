package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Global variables for Phase Two TUI
var phaseTwoProgram *tea.Program
var phaseTwoModel *PhaseTwoTUI

// PhaseTwoTUI is a TUI for Phase Two processing (reducers and writers)
type PhaseTwoTUI struct {
	// Progress data
	startTime           time.Time
	personsProcessed    int64
	personsSelected     map[int]int64 // selected count per scale
	personsWritten      map[int]int64 // written count per scale
	workRangesQueued    int
	workRangesCompleted int
	
	// Phase Two specific data
	scales              []int          // requested scales (1%, 5%, 10%)
	reducerCount        int
	writerCount         int
	
	// Worker states
	reducers           []ReducerWorkerState
	writers            map[int][]WriterWorkerState // writers per scale
	lastAction         string
	
	// Phase status
	showStatusOnly     bool
	statusMessage      string
	
	// Display
	width      int
	height     int
	isComplete bool
	
	// Styles - all white as specified
	tableStyle lipgloss.Style
	cellStyle  lipgloss.Style
}

// ReducerWorkerState represents the state of a reducer worker
type ReducerWorkerState struct {
	Name              string
	Status            string
	PersonsProcessed  int64
	PersonsSelected   map[int]int64 // selected per scale
	WorkRangesHandled int
}

// WriterWorkerState represents the state of a writer worker
type WriterWorkerState struct {
	Name           string
	Status         string
	PersonsWritten int64
	Scale          int
	OutputFile     string
}

// NewPhaseTwoTUI creates a new Phase Two TUI
func NewPhaseTwoTUI(scales []int, reducerCount int) *PhaseTwoTUI {
	// All styles are white as per specification
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // White
	
	tui := &PhaseTwoTUI{
		startTime:       time.Now(),
		lastAction:      "Starting Phase Two",
		scales:          scales,
		reducerCount:    reducerCount,
		writerCount:     reducerCount * len(scales), // writers per scale
		personsSelected: make(map[int]int64),
		personsWritten:  make(map[int]int64),
		writers:         make(map[int][]WriterWorkerState),
		
		// All white styles as specified
		tableStyle: whiteStyle.Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("15")),
		cellStyle:  whiteStyle,
	}
	
	// Initialize counters for each scale
	for _, scale := range scales {
		tui.personsSelected[scale] = 0
		tui.personsWritten[scale] = 0
		tui.writers[scale] = make([]WriterWorkerState, 0)
	}
	
	return tui
}

// Init initializes the TUI
func (t *PhaseTwoTUI) Init() tea.Cmd {
	return t.tick()
}

// Update handles messages
func (t *PhaseTwoTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		return t, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return t, tea.Quit
		}
		
	case PhaseTwoProgressMsg:
		// Update progress data
		t.personsProcessed = msg.PersonsProcessed
		t.personsSelected = msg.PersonsSelected
		t.personsWritten = msg.PersonsWritten
		t.workRangesQueued = msg.WorkRangesQueued
		t.workRangesCompleted = msg.WorkRangesCompleted
		
		// Update workers if provided
		if len(msg.Reducers) > 0 {
			t.reducers = msg.Reducers
		}
		if len(msg.Writers) > 0 {
			for scale, writers := range msg.Writers {
				t.writers[scale] = writers
			}
		}
		
		if msg.LastAction != "" {
			t.lastAction = msg.LastAction
		}
		
		return t, nil
		
	case PhaseTwoStatusMsg:
		t.statusMessage = msg.Message
		t.showStatusOnly = msg.ClearWorkers
		return t, nil
		
	case PhaseTwoCompleteMsg:
		t.isComplete = true
		return t, tea.Quit
		
	case tickMsg:
		return t, t.tick()
	}
	
	return t, nil
}

// View renders the Phase Two TUI - simplified like Phase One
func (t *PhaseTwoTUI) View() string {
	if t.width == 0 {
		return "Initializing Phase Two..."
	}
	
	// If showing status only, display just the status message
	if t.showStatusOnly {
		content := t.statusMessage
		return t.tableStyle.Render(content)
	}
	
	// Main progress display like Phase One
	content := fmt.Sprintf("Persons Processed: %d\nPersons Selected: %d", t.personsProcessed, t.getTotalSelected())
	
	// Add reducer table
	if len(t.reducers) > 0 {
		content += "\n\nReducer Workers:\n"
		content += "┌─────────────┬────────────┬───────────┐\n"
		content += "│ Name        │ Status     │ Processed │\n"
		content += "├─────────────┼────────────┼───────────┤\n"
		
		for _, reducer := range t.reducers {
			name := fmt.Sprintf("%-11s", reducer.Name)
			status := fmt.Sprintf("%-10s", reducer.Status)
			processed := fmt.Sprintf("%9d", reducer.PersonsProcessed)
			content += fmt.Sprintf("│ %s │ %s │ %s │\n", name, status, processed)
		}
		
		content += "└─────────────┴────────────┴───────────┘"
	}
	
	// Add writer table
	if t.hasWriters() {
		content += "\n\nWriter Workers:\n"
		content += "┌─────────────┬────────────┬───────────┬───────────┐\n"
		content += "│ Name        │ Status     │ Written   │ Scale     │\n"
		content += "├─────────────┼────────────┼───────────┼───────────┤\n"
		
		for _, scale := range t.scales {
			writers := t.writers[scale]
			for _, writer := range writers {
				name := fmt.Sprintf("%-11s", writer.Name)
				status := fmt.Sprintf("%-10s", writer.Status)
				written := fmt.Sprintf("%9d", writer.PersonsWritten)
				scaleStr := fmt.Sprintf("%8d%%", scale)
				content += fmt.Sprintf("│ %s │ %s │ %s │ %s │\n", name, status, written, scaleStr)
			}
		}
		
		content += "└─────────────┴────────────┴───────────┴───────────┘"
	}
	
	// Wrap in table with white borders
	return t.tableStyle.Render(content)
}

// Helper methods
func (t *PhaseTwoTUI) getTotalSelected() int64 {
	total := int64(0)
	for _, count := range t.personsSelected {
		total += count
	}
	return total
}

func (t *PhaseTwoTUI) hasWriters() bool {
	for _, writers := range t.writers {
		if len(writers) > 0 {
			return true
		}
	}
	return false
}

// tick returns a command that sends a tick message after a delay
func (t *PhaseTwoTUI) tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// Message types for Phase Two
type PhaseTwoProgressMsg struct {
	PersonsProcessed    int64
	PersonsSelected     map[int]int64
	PersonsWritten      map[int]int64
	WorkRangesQueued    int
	WorkRangesCompleted int
	Reducers            []ReducerWorkerState
	Writers             map[int][]WriterWorkerState
	LastAction          string
}

type PhaseTwoCompleteMsg struct{}

type PhaseTwoStatusMsg struct {
	Message      string
	ClearWorkers bool
}

// Global functions for Phase Two TUI management
func InitPhaseTwoTUI(scales []int, reducerCount int) {
	phaseTwoModel = NewPhaseTwoTUI(scales, reducerCount)
	phaseTwoProgram = tea.NewProgram(phaseTwoModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
	
	// Run in goroutine to prevent blocking
	go func() {
		_ = phaseTwoProgram.Start()
	}()
}

// UpdatePhaseTwoProgress sends a progress update to the Phase Two TUI
func UpdatePhaseTwoProgress(msg PhaseTwoProgressMsg) {
	if phaseTwoProgram != nil {
		phaseTwoProgram.Send(msg)
	}
}

// UpdatePhaseTwoStatus sends a status message to the Phase Two TUI
func UpdatePhaseTwoStatus(message string, clearWorkers bool) {
	if phaseTwoProgram != nil {
		phaseTwoProgram.Send(PhaseTwoStatusMsg{
			Message:      message,
			ClearWorkers: clearWorkers,
		})
	}
}

// SignalPhaseTwoComplete signals completion to the Phase Two TUI
func SignalPhaseTwoComplete() {
	if phaseTwoProgram != nil {
		phaseTwoProgram.Send(PhaseTwoCompleteMsg{})
	}
}

// StopPhaseTwoTUI stops the Phase Two TUI
func StopPhaseTwoTUI() {
	if phaseTwoProgram != nil {
		phaseTwoProgram.Send(PhaseTwoCompleteMsg{})
		phaseTwoProgram.Quit()
	}
}

// StartPhaseTwoTUI creates and starts the Phase Two TUI
func StartPhaseTwoTUI(scales []int, reducerCount int) (*tea.Program, error) {
	model := NewPhaseTwoTUI(scales, reducerCount)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	return p, nil
}