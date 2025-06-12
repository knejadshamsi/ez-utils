package tui

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Global variables following the same pattern as the working display module
var p *tea.Program
var model *SimpleTUI

// SimpleTUI is a minimal TUI for the queue-based system following the simplified specification
type SimpleTUI struct {
	// Progress data
	startTime        time.Time
	linesRead        int64
	personsFound     int64
	coordsExtracted  int64
	coordsMapped     int64
	personQueueSize  int
	coordQueueSize   int
	
	// Phase 0 data
	currentPhase     string // "Phase 0", "Phase 1", "Phase 2"
	safePointsFound  int
	totalLinesScanned int64
	
	// System metrics
	cpuUsage         float64
	ramUsage         float64
	diskUsage        float64
	
	// Worker states
	workers          []WorkerState
	lastAction       string
	
	// Phase status
	showStatusOnly   bool
	statusMessage    string
	
	// Display
	width      int
	height     int
	isComplete bool
	
	// Styles - all white as specified
	tableStyle   lipgloss.Style
	cellStyle    lipgloss.Style
}

// WorkerState represents the state of a worker
type WorkerState struct {
	Name      string
	Status    string
	Extracted int64
	BinCount  int64  // Number of unique bins for mappers
}

// NewSimpleTUI creates a new simple TUI
func NewSimpleTUI() *SimpleTUI {
	// All styles are white as per specification
	whiteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15")) // White
	
	return &SimpleTUI{
		startTime:         time.Now(),
		lastAction:        "Starting Phase 0: Safe Point Discovery",
		currentPhase:      "Phase 0",
		safePointsFound:   0,
		totalLinesScanned: 0,
		workers:           make([]WorkerState, 0),
		
		// All white styles as specified
		tableStyle: whiteStyle.Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("15")),
		cellStyle:  whiteStyle,
	}
}

// Init initializes the TUI
func (t *SimpleTUI) Init() tea.Cmd {
	return t.tick()
}

// Update handles messages
func (t *SimpleTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		
	case ProgressMsg:
		// Update progress data
		t.linesRead = msg.LinesRead
		t.personsFound = msg.PersonsFound
		t.coordsExtracted = msg.CoordsExtracted
		t.coordsMapped = msg.CoordsMapped
		t.personQueueSize = msg.PersonQueueSize
		t.coordQueueSize = msg.CoordQueueSize
		
		// Update the global model pointer so setter functions can access current values
		if model != nil {
			model.linesRead = msg.LinesRead
			model.personsFound = msg.PersonsFound
		}
		
		// Update workers and last action if provided
		if len(msg.Workers) > 0 {
			t.workers = msg.Workers
		} else {
			t.updateWorkers()
		}
		
		if msg.LastAction != "" {
			t.lastAction = msg.LastAction
		}
		
		return t, nil
		
	case StatusMsg:
		t.statusMessage = msg.Message
		t.showStatusOnly = msg.ClearWorkers
		return t, nil
		
	case PhaseMsg:
		t.currentPhase = msg.Phase
		t.safePointsFound = msg.SafePointsFound
		t.totalLinesScanned = msg.LinesScanned
		return t, nil
		
	case CompleteMsg:
		t.isComplete = true
		return t, tea.Quit
		
	case tickMsg:
		return t, t.tick()
	}
	
	return t, nil
}

// View renders the TUI according to the simplified specification
func (t *SimpleTUI) View() string {
	if t.width == 0 {
		return "Initializing..."
	}
	
	// If showing status only, display just the status message
	if t.showStatusOnly {
		content := t.statusMessage
		return t.tableStyle.Render(content)
	}
	
	// Phase and progress display
	var content string
	if t.currentPhase == "Phase 0" {
		content = fmt.Sprintf("%s: Safe Point Discovery\nSafe Points Found: %d\nLines Scanned: %d", 
			t.currentPhase, t.safePointsFound, t.totalLinesScanned)
		if t.totalLinesScanned > 0 {
			// Calculate approximate progress percentage
			estimatedTotal := t.totalLinesScanned * 2 // Rough estimate
			if estimatedTotal > 0 {
				progress := float64(t.totalLinesScanned) / float64(estimatedTotal) * 100
				if progress > 100 {
					progress = 100
				}
				content += fmt.Sprintf("\nProgress: %.1f%%", progress)
			}
		}
		
		// Add last action if in Phase 0
		if t.lastAction != "" {
			content += fmt.Sprintf("\nStatus: %s", t.lastAction)
		}
	} else {
		content = fmt.Sprintf("%s\nLines Read: %d\nPersons Found: %d", 
			t.currentPhase, t.linesRead, t.personsFound)
	}
	
	// Add worker tables
	if len(t.workers) > 0 {
		// Separate readers, extractors and mappers
		var readers []WorkerState
		var extractors []WorkerState
		var mappers []WorkerState
		
		for _, worker := range t.workers {
			if strings.HasPrefix(worker.Name, "reader-") || strings.HasPrefix(worker.Name, "Reader") {
				readers = append(readers, worker)
			} else if strings.HasPrefix(worker.Name, "Extractor") {
				extractors = append(extractors, worker)
			} else if strings.HasPrefix(worker.Name, "Mapper") {
				mappers = append(mappers, worker)
			}
		}
		
		// Show readers table (for multiple readers)
		if len(readers) > 0 {
			content += "\n\nReader Workers:\n"
			content += "┌─────────────┬────────────┬───────────┐\n"
			content += "│ Name        │ Status     │ Lines Read│\n"
			content += "├─────────────┼────────────┼───────────┤\n"
			
			for _, worker := range readers {
				name := fmt.Sprintf("%-11s", worker.Name)
				status := fmt.Sprintf("%-10s", worker.Status)
				linesRead := fmt.Sprintf("%9d", worker.Extracted) // Using Extracted field for lines read
				content += fmt.Sprintf("│ %s │ %s │ %s │\n", name, status, linesRead)
			}
			
			content += "└─────────────┴────────────┴───────────┘"
		}
		
		// Show extractors
		if len(extractors) > 0 {
			content += "\n\nExtractor Workers:\n"
			content += "┌─────────────┬────────────┬───────────┐\n"
			content += "│ Name        │ Status     │ Extracted │\n"
			content += "├─────────────┼────────────┼───────────┤\n"
			
			for _, worker := range extractors {
				name := fmt.Sprintf("%-11s", worker.Name)
				status := fmt.Sprintf("%-10s", worker.Status)
				extracted := fmt.Sprintf("%9d", worker.Extracted)
				content += fmt.Sprintf("│ %s │ %s │ %s │\n", name, status, extracted)
			}
			
			content += "└─────────────┴────────────┴───────────┘"
		}
		
		// Show mappers
		if len(mappers) > 0 {
			content += "\n\nMapper Workers:\n"
			content += "┌─────────────┬────────────┬───────────┬───────────┐\n"
			content += "│ Name        │ Status     │ Mapped    │ Bins      │\n"
			content += "├─────────────┼────────────┼───────────┼───────────┤\n"
			
			for _, worker := range mappers {
				name := fmt.Sprintf("%-11s", worker.Name)
				status := fmt.Sprintf("%-10s", worker.Status)
				mapped := fmt.Sprintf("%9d", worker.Extracted) // Using Extracted field for mapped count
				bins := fmt.Sprintf("%9d", worker.BinCount)     // Show unique bin count
				content += fmt.Sprintf("│ %s │ %s │ %s │ %s │\n", name, status, mapped, bins)
			}
			
			content += "└─────────────┴────────────┴───────────┴───────────┘"
		}
	}
	
	// Optional: Show elapsed time
	// elapsed := time.Since(t.startTime)
	// hours := int(elapsed.Hours())
	// minutes := int(elapsed.Minutes()) % 60
	// seconds := int(elapsed.Seconds()) % 60
	
	// Wrap in table with white borders
	return t.tableStyle.Render(content)
}

// updateSystemMetrics gets basic system metrics
func (t *SimpleTUI) updateSystemMetrics() {
	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Simple approximations for display
	t.cpuUsage = float64(runtime.NumGoroutine()) * 0.1 // Approximate based on goroutines
	t.ramUsage = float64(m.Alloc) / (1024 * 1024)     // MB of allocated memory
	t.diskUsage = 0.0                                  // Not implemented for simplicity
}

// updateWorkers updates the worker list based on current state
func (t *SimpleTUI) updateWorkers() {
	// Clear existing workers
	t.workers = t.workers[:0]
	
	// Add reader worker
	t.workers = append(t.workers, WorkerState{
		Name:      "Reader",
		Status:    "Reading",
		Extracted: 0, // Reader doesn't extract
	})
	
	// Add extractor workers (example: 4 workers)
	for i := 1; i <= 4; i++ {
		status := "Available"
		extracted := int64(i * 100) // Sample extracted count
		if t.isComplete {
			status = "Completed"
			extracted = int64(i * 500)
		} else if i%2 == 0 {
			status = "Extracting"
			extracted = int64(i * 150)
		}
		t.workers = append(t.workers, WorkerState{
			Name:      fmt.Sprintf("Extractor-%d", i),
			Status:    status,
			Extracted: extracted,
		})
	}
	
	// Add mapper workers (example: 3 workers)
	for i := 1; i <= 3; i++ {
		status := "Mapping"
		if t.isComplete {
			status = "Completed"
		}
		t.workers = append(t.workers, WorkerState{
			Name:      fmt.Sprintf("Mapper-%d", i),
			Status:    status,
			Extracted: 0, // Mappers don't extract
		})
	}
}

// Message types
type tickMsg struct{}

type ProgressMsg struct {
	LinesRead       int64
	PersonsFound    int64
	CoordsExtracted int64
	CoordsMapped    int64
	PersonQueueSize int
	CoordQueueSize  int
	Workers         []WorkerState
	LastAction      string
}

type CompleteMsg struct{}

type StatusMsg struct {
	Message     string
	ClearWorkers bool  // If true, hide worker tables and show only status
}

type PhaseMsg struct {
	Phase            string
	SafePointsFound  int
	LinesScanned     int64
}

// UpdateProgress sends a progress update to the TUI
func UpdateProgress(p *tea.Program, msg ProgressMsg) {
	p.Send(msg)
}

// SignalComplete signals completion to the TUI
func SignalComplete(p *tea.Program) {
	p.Send(CompleteMsg{})
}

// UpdateStatus sends a status message to the TUI
func UpdateStatus(p *tea.Program, message string, clearWorkers bool) {
	if p != nil {
		p.Send(StatusMsg{
			Message:     message,
			ClearWorkers: clearWorkers,
		})
	}
}

// UpdatePhase sends a phase update to the TUI
func UpdatePhase(p *tea.Program, phase string, safePoints int, linesScanned int64) {
	if p != nil {
		p.Send(PhaseMsg{
			Phase:           phase,
			SafePointsFound: safePoints,
			LinesScanned:    linesScanned,
		})
	}
}

// tick returns a command that sends a tick message after a delay
func (t *SimpleTUI) tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

// InitSimpleTUI sets up the SimpleTUI following the same pattern as the global display
func InitSimpleTUI() {
	model = NewSimpleTUI()
	p = tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	
	// Run in goroutine to prevent blocking the main process while UI is active
	go func() {
		_ = p.Start()
	}()
}

// Setter functions following the same pattern as the global display
func SetLinesRead(count int64) {
	if p != nil {
		p.Send(ProgressMsg{
			LinesRead:    count,
			PersonsFound: model.personsFound, // Keep existing value
		})
	}
}

func SetPersonsFound(count int64) {
	if p != nil {
		p.Send(ProgressMsg{
			LinesRead:    model.linesRead, // Keep existing value
			PersonsFound: count,
		})
	}
}

func SetProgress(linesRead, personsFound int64) {
	if p != nil {
		p.Send(ProgressMsg{
			LinesRead:    linesRead,
			PersonsFound: personsFound,
			Workers:      getDefaultWorkers(), // Add worker info
		})
	}
}

func SetProgressWithWorkers(linesRead, personsFound int64, workers []WorkerState) {
	if p != nil {
		p.Send(ProgressMsg{
			LinesRead:    linesRead,
			PersonsFound: personsFound,
			Workers:      workers,
		})
	}
}

// getDefaultWorkers returns sample extractor workers for demonstration
func getDefaultWorkers() []WorkerState {
	return []WorkerState{
		{Name: "Extractor-1", Status: "Extracting", Extracted: 245},
		{Name: "Extractor-2", Status: "Available", Extracted: 189},
		{Name: "Extractor-3", Status: "Extracting", Extracted: 312},
		{Name: "Extractor-4", Status: "Available", Extracted: 156},
	}
}

func StopSimpleTUI() {
	if p != nil {
		p.Send(CompleteMsg{})
		p.Quit()
	}
}

// StartSimpleTUI creates and starts the simple TUI with proper alt screen
func StartSimpleTUI() (*tea.Program, error) {
	model := NewSimpleTUI()
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	return p, nil
}