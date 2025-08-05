package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// TUI represents the main TUI controller
type TUI struct {
	state         *State
	program       *tea.Program
	ctx           context.Context
	cancel        context.CancelFunc
	errorCallback func(error) // Callback for TUI errors
}

// Model wraps the state for Bubble Tea
type Model struct {
	state *State
}

// Messages for Bubble Tea
type tickMsg struct{}
type blinkMsg struct{}
type systemStatsMsg struct {
	cpu float64
	ram float64
}

// NewTUI creates a new TUI instance
func NewTUI() *TUI {
	state := NewState()

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Line
	s.Style = Styles.Spinner
	state.Spinner = s

	ctx, cancel := context.WithCancel(context.Background())

	tui := &TUI{
		state:  state,
		ctx:    ctx,
		cancel: cancel,
	}

	return tui
}

// SetErrorCallback sets the error callback function
func (t *TUI) SetErrorCallback(callback func(error)) {
	t.errorCallback = callback
}

// Start initializes and starts the TUI program
func (t *TUI) Start() error {
	// Log TUI initialization attempt
	if t.errorCallback != nil {
		t.errorCallback(fmt.Errorf("TUI Debug: Initializing Bubble Tea program"))
	}

	model := Model{state: t.state}

	// Try to create the program - this might fail with terminal setup issues
	t.program = tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if t.errorCallback != nil {
		t.errorCallback(fmt.Errorf("TUI Debug: Bubble Tea program created successfully"))
	}

	// Start background tasks
	go t.startTicker()
	go t.startSystemMonitor()

	if t.errorCallback != nil {
		t.errorCallback(fmt.Errorf("TUI Debug: Background tasks started"))
	}

	// Run the program in a goroutine
	go func() {
		if t.errorCallback != nil {
			t.errorCallback(fmt.Errorf("TUI Debug: Starting Bubble Tea program.Run()"))
		}

		_, err := t.program.Run()
		if err != nil {
			// Call error callback if set, otherwise log to stderr as fallback
			if t.errorCallback != nil {
				t.errorCallback(fmt.Errorf("TUI program error: %v", err))
			} else {
				fmt.Fprintf(os.Stderr, "TUI Error (no callback set): %v\n", err)
			}
		} else {
			// Log successful completion
			if t.errorCallback != nil {
				t.errorCallback(fmt.Errorf("TUI Debug: Bubble Tea program completed normally"))
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down the TUI
func (t *TUI) Stop() error {
	if t.errorCallback != nil {
		t.errorCallback(fmt.Errorf("TUI Debug: Stop() called"))
	}

	if t.cancel != nil {
		t.cancel()
		if t.errorCallback != nil {
			t.errorCallback(fmt.Errorf("TUI Debug: Context cancelled"))
		}
	}

	if t.program != nil {
		t.program.Quit()
		if t.errorCallback != nil {
			t.errorCallback(fmt.Errorf("TUI Debug: Program quit signal sent"))
		}
	}

	return nil
}

// SetStep updates the current step
func (t *TUI) SetStep(step int) {
	t.state.SetStep(step)
	if t.program != nil {
		t.program.Send(tickMsg{}) // Trigger refresh
	}
}

// UpdateCounter updates live data for current step
func (t *TUI) UpdateCounter(key string, value interface{}) {
	t.state.UpdateCounter(key, value)
	if t.program != nil {
		t.program.Send(tickMsg{}) // Trigger refresh
	}
}

// SetSystemStats updates CPU and RAM usage
func (t *TUI) SetSystemStats(cpu, ram float64) {
	if t.program != nil {
		t.program.Send(systemStatsMsg{cpu: cpu, ram: ram})
	}
}

// SetProcessComplete marks the process as complete
func (t *TUI) SetProcessComplete(complete bool) {
	t.state.SetProcessComplete(complete)
	if t.program != nil {
		t.program.Send(tickMsg{}) // Trigger refresh
	}
}

// IsQuitRequested checks if the user requested to quit
func (t *TUI) IsQuitRequested() bool {
	return t.state.QuitRequested
}

// Bubble Tea Model interface implementation

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		blinkCmd(),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.state.RequestQuit()
			return m, tea.Quit
		}

	case tickMsg:
		m.state.UpdateElapsedTime()

	case blinkMsg:
		m.state.ToggleStatusBlink()

	case systemStatsMsg:
		m.state.SetSystemStats(msg.cpu, msg.ram)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.state.Spinner, cmd = m.state.Spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the TUI
func (m Model) View() string {
	if m.state.ProcessComplete {
		return ""
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		RenderHeader(m.state),
		RenderPrevious(m.state),
		RenderOngoing(m.state),
		RenderSystem(m.state),
		RenderUpcoming(m.state),
	)
}

// Background task functions

func (t *TUI) startTicker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if t.program != nil {
				t.program.Send(tickMsg{})
			}
		case <-t.ctx.Done():
			return
		}
	}
}

func (t *TUI) startSystemMonitor() {
	// Get initial system stats immediately
	cpuPercent, err := cpu.Percent(500*time.Millisecond, false)
	cpuUsage := 0.0
	if err == nil && len(cpuPercent) > 0 {
		cpuUsage = cpuPercent[0]
	}

	memInfo, err := mem.VirtualMemory()
	ramUsage := 0.0
	if err == nil {
		ramUsage = float64(memInfo.Used) / 1024 / 1024 // Convert to MB
	}

	if t.program != nil {
		t.program.Send(systemStatsMsg{cpu: cpuUsage, ram: ramUsage})
	}

	// Now start the ticker for regular updates
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get real CPU usage with 500ms sampling period
			cpuPercent, err := cpu.Percent(500*time.Millisecond, false)
			cpuUsage := 0.0
			if err == nil && len(cpuPercent) > 0 {
				cpuUsage = cpuPercent[0]
			}

			// Get real RAM usage
			memInfo, err := mem.VirtualMemory()
			ramUsage := 0.0
			if err == nil {
				ramUsage = float64(memInfo.Used) / 1024 / 1024 // Convert to MB
			}

			if t.program != nil {
				t.program.Send(systemStatsMsg{cpu: cpuUsage, ram: ramUsage})
			}
		case <-t.ctx.Done():
			return
		}
	}
}

// Helper functions for Bubble Tea commands

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func blinkCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return blinkMsg{}
	})
}
