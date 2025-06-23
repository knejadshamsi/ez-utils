package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ez-utils/src/display/config"
	"ez-utils/src/display/tui"
	"ez-utils/src/display"
)

// DisplayInstance represents a single display TUI instance
type DisplayInstance struct {
	config       *config.Config
	program      *tea.Program
	model        tui.Model
	businessLogic *BusinessLogic
	asyncManager *AsyncManager
	mutex        sync.RWMutex
	running      bool
	cleanup      []func()
	quitChan     chan bool // Channel to signal quit from TUI
	
	// Live updates storage
	liveUpdates  map[int]map[string]interface{} // stepNum -> updates
	stepMessages map[int]string                  // stepNum -> custom message
}


// ModuleConfig defines configuration for different modules
type ModuleConfig struct { // This struct is deprecated and should be removed after refactoring
	ModuleName string
	MaxSteps   int
	Language   string
	Flags      map[string]bool
}

// NewDisplayInstance creates a new display instance with modern config
func NewDisplayInstance(config *config.Config) (*DisplayInstance, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	return createDisplayInstance(config)
}

// NewDisplayInstanceFromModuleConfig creates a display instance from legacy config
// DEPRECATED: Use NewDisplayInstance with Config instead
func NewDisplayInstanceFromModuleConfig(cfg *ModuleConfig) (*DisplayInstance, error) { // Renamed 'config' to 'cfg' to avoid name collision with the imported package
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	// Convert legacy config to new config format
	newConfig := &config.Config{
		Process: &config.ProcessConfig{
			Title:    cfg.ModuleName + " Processing",
			MaxSteps: cfg.MaxSteps,
			Flags:    cfg.Flags,
		},
		Theme:    config.DefaultConfig().Theme,
		Language: cfg.Language,
		Async:    config.DefaultConfig().Async,
	}
	
	return createDisplayInstance(newConfig)
}

// createDisplayInstance is the internal instance creation function
func createDisplayInstance(cfg *config.Config) (*DisplayInstance, error) { // Renamed 'config' to 'cfg'

	// Initialize text manager with specified language
	if err := display.InitializeTextManager(cfg.Language); err != nil {
		// Continue silently with fallback behavior
	}

	spn := spinner.New()
	spn.Spinner = spinner.Line
	spn.Style = lipgloss.NewStyle().Foreground(tui.ColorMagenta)

	tmr := timer.NewWithInterval(0, time.Second)

	model := tui.Model{ // Changed to tui.Model
		// Basic configuration (process, flags)
		ModuleName:   cfg.Process.Title,
		ProcessTitle: cfg.Process.Description,
		StepNumber:   0,
		MaxSteps:     cfg.Process.MaxSteps,
		Flags:        cfg.Process.Flags,
		Steps:        cfg.Process.Steps,

		// Time tracking
		StartTime:        time.Now(),
		ElapsedTime:      0,
		TotalStartTime:   time.Now(),
		TotalElapsedTime: 0,
		StepDurations:    make(map[int]time.Duration),

		// UI state
		Spinner:     spn,
		Timer:       tmr,
		StatusBlink: false,

		// Resource monitoring
		CPUUsage: 0,
		RAMUsage: 0,

		// Process state
		ProcessComplete: false,
		Err:             nil,
		
		// Theme support
		ThemeColors:     make(map[string]string),

		// Progress tracking
		ChunkCount:      0,
		PersonsFound:    0,
		BytesReadMB:     0,
		FirstChunkFixed: false,
		LastChunkFixed:  false,
		AgentCounter:    0,
		ChunkCounter:    tui.Counter{0, 0},
		DbCounter:       tui.Counter{0, 0},
		CoordinateCount: 0,
		BinCounter:      tui.Counter{0, 0},
		AgentBinCount:   0,
		ScaleCounter:    tui.Counter{0, 0},
		OutputScale:     0,
		OutputCounter:   tui.Counter{0, 0},
		CleanupFiles:    0,
		CleanupDirs:     0,
		CleanupBytes:    0,
	}

	// Create business logic instance (convert config for compatibility)
	legacyConfig := &ModuleConfig{
		ModuleName: cfg.Process.Title,
		MaxSteps:   cfg.Process.MaxSteps,
		Language:   cfg.Language,
		Flags:      cfg.Process.Flags,
	}
	businessLogic := NewBusinessLogic(legacyConfig)
	
	// Create async manager
	asyncManager := NewAsyncManager(cfg.Async)

	instance := &DisplayInstance{
		config:        cfg, // Renamed 'config' to 'cfg'
		model:         model,
		businessLogic: businessLogic,
		asyncManager:  asyncManager,
		running:       false,
		cleanup:       make([]func(), 0),
		quitChan:      make(chan bool, 1),
		liveUpdates:   make(map[int]map[string]interface{}),
		stepMessages:  make(map[int]string),
	}

	return instance, nil
}

// Start initializes and starts the TUI program with async support
func (d *DisplayInstance) Start() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.running {
		return fmt.Errorf("display instance is already running")
	}

	// Apply theme to model
	d.applyTheme()

	d.program = tea.NewProgram(d.model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	d.running = true

	// Start async manager
	if err := d.asyncManager.Start(d.program); err != nil {
		return fmt.Errorf("failed to start async manager: %w", err)
	}

	// Start system monitoring in a goroutine
	go d.monitorSystem()

	// Run TUI program in a goroutine to prevent blocking the main process while UI is active
	go func() {
		_, err := d.program.Run()
		if err == nil {
			// TUI exited normally (user pressed 'q'), signal quit
			select {
			case d.quitChan <- true:
			default:
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down the display instance
func (d *DisplayInstance) Stop() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running {
		return nil
	}

	// Stop async manager first
	if d.asyncManager != nil {
		if err := d.asyncManager.Stop(); err != nil {
			// Stop silently
		}
	}

	// Run cleanup functions in reverse order
	for i := len(d.cleanup) - 1; i >= 0; i-- {
		if d.cleanup[i] != nil {
			d.cleanup[i]()
		}
	}
	d.cleanup = d.cleanup[:0]

	// Stop the program
	if d.program != nil {
		d.program.Quit()
		// Give the program time to shut down
		time.Sleep(100 * time.Millisecond)
		d.program = nil
	}

	// Clear the model's timers and spinners
	d.model.Timer.Timeout = 0 // Changed to d.model.Timer.Timeout

	d.running = false
	return nil
}

// IsRunning returns whether the display instance is currently running
func (d *DisplayInstance) IsRunning() bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.running
}

// IsQuitRequested checks if the user requested to quit (non-blocking)
func (d *DisplayInstance) IsQuitRequested() bool {
	select {
	case <-d.quitChan:
		return true
	default:
		return false
	}
}

// AddCleanupFunc adds a cleanup function to be called when stopping
func (d *DisplayInstance) AddCleanupFunc(cleanupFunc func()) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.cleanup = append(d.cleanup, cleanupFunc)
}

// GetConfig returns the display configuration
func (d *DisplayInstance) GetConfig() *config.Config {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.config
}

// GetLiveUpdates returns all live updates for a step
func (d *DisplayInstance) GetLiveUpdates(stepNum int) map[string]interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if updates, exists := d.liveUpdates[stepNum]; exists {
		// Return a copy to prevent external modification
		copy := make(map[string]interface{})
		for k, v := range updates {
			copy[k] = v
		}
		return copy
	}
	return make(map[string]interface{})
}

// GetAsyncStats returns statistics about the async message system
func (d *DisplayInstance) GetAsyncStats() map[string]interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if d.asyncManager == nil {
		return make(map[string]interface{})
	}
	return d.asyncManager.GetQueueStats()
}

// applyTheme applies the configured theme to the display
func (d *DisplayInstance) applyTheme() {
	if d.config.Theme == nil {
		return
	}

	// Get theme colors
	themeColors := config.GetThemeColors(d.config.Theme.Name)
	
	// Apply custom color overrides
	for element, color := range d.config.Theme.Colors {
		themeColors[element] = color
	}

	// Apply colors to the model (this would need to be implemented in styles.go)
	// For now, just store the theme colors for later use
	d.model.ThemeColors = themeColors
}
