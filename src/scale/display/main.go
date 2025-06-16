// Package display provides a universal Terminal User Interface (TUI) solution for
// sequential logical flows across different modules in the ez-utils application.
//
// This package is designed for modules like PT (Public Transit) and future modules
// that need live status tracking and step-by-step progress visualization.
// The population module uses its own dedicated TUI system.
//
// The package is designed to be:
// - Instance-based: Multiple display instances can be created and managed independently
// - Thread-safe: All operations are protected by mutexes for concurrent access
// - Resource-efficient: Proper cleanup and monitoring optimization
// - Extensible: Configurable for different modules with varying step counts
// - Cached: Template processing and locale loading are cached for performance
// - Error-resilient: Comprehensive error handling and validation
//
// Architecture:
// - DisplayInstance: Main interface for creating and managing TUI instances
// - BusinessLogic: Separate layer for state management and validation
// - ModuleConfig: Configuration structure for different modules
//
// Usage:
//   config := &display.ModuleConfig{
//       ModuleName: "pt",           // Module name for localization
//       MaxSteps:   9,              // Number of steps in the process
//       Language:   "en",           // Language for UI text
//       Flags:      map[string]bool{"clean": true}, // Module-specific flags
//   }
//   
//   instance, err := display.NewDisplayInstance(config)
//   if err != nil {
//       return fmt.Errorf("failed to create display: %v", err)
//   }
//   defer instance.Stop() // Always cleanup
//   
//   if err := instance.Start(); err != nil {
//       return fmt.Errorf("failed to start display: %v", err)
//   }
//   
//   // Update progress during processing
//   _ = instance.SetStep(0)
//   _ = instance.SetChunkCount(100)
//   _ = instance.SetStep(1)
//   // ... continue with other steps
//   _ = instance.SetProcessComplete(true)
package display

import (
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DisplayInstance represents a single display TUI instance
type DisplayInstance struct {
	config       *Config
	program      *tea.Program
	model        Model
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
type ModuleConfig struct {
	ModuleName string
	MaxSteps   int
	Language   string
	Flags      map[string]bool
}

// NewDisplayInstance creates a new display instance with modern config
func NewDisplayInstance(config *Config) (*DisplayInstance, error) {
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
func NewDisplayInstanceFromModuleConfig(config *ModuleConfig) (*DisplayInstance, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	
	// Convert legacy config to new config format
	newConfig := &Config{
		Process: &ProcessConfig{
			Title:    config.ModuleName + " Processing",
			MaxSteps: config.MaxSteps,
			Flags:    config.Flags,
		},
		Theme:    DefaultConfig().Theme,
		Language: config.Language,
		Async:    DefaultConfig().Async,
	}
	
	return createDisplayInstance(newConfig)
}

// createDisplayInstance is the internal instance creation function
func createDisplayInstance(config *Config) (*DisplayInstance, error) {

	// Initialize text manager with specified language
	if err := InitializeTextManager(config.Language); err != nil {
		// Continue silently with fallback behavior
	}

	spn := spinner.New()
	spn.Spinner = spinner.Line
	spn.Style = lipgloss.NewStyle().Foreground(colorMagenta)

	tmr := timer.NewWithInterval(0, time.Second)

	model := Model{
		// Basic configuration (process, flags)
		moduleName:   config.Process.Title,
		processTitle: config.Process.Description,
		stepNumber:   0,
		maxSteps:     config.Process.MaxSteps,
		flags:        config.Process.Flags,
		steps:        config.Process.Steps,

		// Time tracking
		startTime:        time.Now(),
		elapsedTime:      0,
		totalStartTime:   time.Now(),
		totalElapsedTime: 0,
		stepDurations:    make(map[int]time.Duration),

		// UI state
		spinner:     spn,
		timer:       tmr,
		statusBlink: false,

		// Resource monitoring
		cpuUsage: 0,
		ramUsage: 0,

		// Process state
		processComplete: false,
		err:             nil,
		
		// Theme support
		themeColors:     make(map[string]string),

		// Progress tracking
		chunkCount:      0,
		personsFound:    0,
		bytesReadMB:     0,
		firstChunkFixed: false,
		lastChunkFixed:  false,
		agentCounter:    0,
		chunkCounter:    Counter{0, 0},
		dbCounter:       Counter{0, 0},
		coordinateCount: 0,
		binCounter:      Counter{0, 0},
		agentBinCount:   0,
		scaleCounter:    Counter{0, 0},
		outputScale:     0,
		outputCounter:   Counter{0, 0},
		cleanupFiles:    0,
		cleanupDirs:     0,
		cleanupBytes:    0,
	}

	// Create business logic instance (convert config for compatibility)
	legacyConfig := &ModuleConfig{
		ModuleName: config.Process.Title,
		MaxSteps:   config.Process.MaxSteps,
		Language:   config.Language,
		Flags:      config.Process.Flags,
	}
	businessLogic := NewBusinessLogic(legacyConfig)
	
	// Create async manager
	asyncManager := NewAsyncManager(config.Async)

	instance := &DisplayInstance{
		config:        config,
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

	// Run in goroutine to prevent blocking the main process while UI is active
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

	// System monitoring in a separate goroutine
	go d.monitorSystem()

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
	d.model.timer.Timeout = 0

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

// SetLiveUpdate sets a live update value for a specific step
func (d *DisplayInstance) SetLiveUpdate(stepNum int, key string, value interface{}) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running {
		return fmt.Errorf("display instance is not running")
	}

	// Check if this step supports this live update
	if !d.config.HasLiveUpdate(stepNum, key) {
		return fmt.Errorf("step %d does not support live update '%s'", stepNum, key)
	}

	// Store the update
	if d.liveUpdates[stepNum] == nil {
		d.liveUpdates[stepNum] = make(map[string]interface{})
	}
	d.liveUpdates[stepNum][key] = value

	// Send async update
	return d.asyncManager.SendLiveUpdate(stepNum, key, value)
}

// SetStepMessage sets a custom message for a specific step
func (d *DisplayInstance) SetStepMessage(stepNum int, template string, data map[string]interface{}) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running {
		return fmt.Errorf("display instance is not running")
	}

	// Store the message
	d.stepMessages[stepNum] = template

	// Send async update
	return d.asyncManager.SendStepMessage(stepNum, template, data)
}

// GetConfig returns the display configuration
func (d *DisplayInstance) GetConfig() *Config {
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
	themeColors := GetThemeColors(d.config.Theme.Name)
	
	// Apply custom color overrides
	for element, color := range d.config.Theme.Colors {
		themeColors[element] = color
	}

	// Apply colors to the model (this would need to be implemented in styles.go)
	// For now, just store the theme colors for later use
	d.model.themeColors = themeColors
}


// Display Instance Methods
// All methods return errors for proper error handling and are thread-safe

// SetStep updates the current step number (async)
func (d *DisplayInstance) SetStep(stepNumber int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running {
		return fmt.Errorf("display instance is not running")
	}

	// Validate step transition
	if err := d.businessLogic.ValidateStepTransition(stepNumber); err != nil {
		// Log warning but continue
		fmt.Printf("Step transition warning: %v\n", err)
	}

	// Update business logic
	d.businessLogic.UpdateStep(stepNumber)

	// Send async update
	return d.asyncManager.SendStepUpdate(stepNumber)
}


// SetProcessComplete marks the process as complete
func (d *DisplayInstance) SetProcessComplete(complete bool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	// Update business logic
	d.businessLogic.SetProcessComplete(complete)

	// Send UI update
	d.program.Send(processCompleteMsg(complete))
	return nil
}


// XML Processing Functions (Steps 0-1) - Chunk parsing and validation

// SetChunkCount updates the chunk count
func (d *DisplayInstance) SetChunkCount(count int) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	// Update business logic
	d.businessLogic.SetChunkCount(count)

	// Send UI update
	d.program.Send(chunkCountMsg(count))
	return nil
}


// SetPersonsFound updates the persons found count
func (d *DisplayInstance) SetPersonsFound(count int) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	// Update business logic
	d.businessLogic.SetPersonsFound(count)

	// Send UI update
	d.program.Send(personsFoundMsg(count))
	return nil
}


// SetBytesReadMB updates the bytes read in MB
func (d *DisplayInstance) SetBytesReadMB(mb int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(bytesReadMsg(mb))
	return nil
}


// SetFirstChunkStatus updates the first chunk fixed status
func (d *DisplayInstance) SetFirstChunkStatus(fixed bool) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(firstChunkFixedMsg(fixed))
	return nil
}


// SetLastChunkStatus updates the last chunk fixed status
func (d *DisplayInstance) SetLastChunkStatus(fixed bool) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(lastChunkFixedMsg(fixed))
	return nil
}


// Location Extraction Functions (Steps 2-3) - Geographic data processing

// SetAgentCounter updates the agent counter
func (d *DisplayInstance) SetAgentCounter(count int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(agentCounterMsg(count))
	return nil
}


// SetChunkCounter updates the chunk processing counter
func (d *DisplayInstance) SetChunkCounter(current, total int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(chunkCounterMsg{current, total})
	return nil
}


// SetDbCounter updates the database processing counter
func (d *DisplayInstance) SetDbCounter(current, total int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(dbCounterMsg{current, total})
	return nil
}


// Processing Functions (Steps 4-18) - Data transformation and cleanup

// SetCoordinateCounter updates the coordinate counter
func (d *DisplayInstance) SetCoordinateCounter(count int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(coordinateCounterMsg(count))
	return nil
}


// SetBinCounter updates the bin processing counter
func (d *DisplayInstance) SetBinCounter(current, total int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(binCounterMsg{current, total})
	return nil
}


// SetAgentBinCounter updates the agent bin counter
func (d *DisplayInstance) SetAgentBinCounter(count int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(agentBinCounterMsg(count))
	return nil
}


// SetScaleCounter updates the scale processing counter
func (d *DisplayInstance) SetScaleCounter(current, total int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(scaleCounterMsg{current, total})
	return nil
}


// SetOutputScale updates the output scale value
func (d *DisplayInstance) SetOutputScale(scale int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(outputScaleMsg(scale))
	return nil
}


// SetOutputCounter updates the output processing counter
func (d *DisplayInstance) SetOutputCounter(current, total int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(outputCounterMsg{current, total})
	return nil
}


// SetCleanupCounter updates the cleanup counter
func (d *DisplayInstance) SetCleanupCounter(files, dirs, bytes int) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if !d.running || d.program == nil {
		return fmt.Errorf("display instance is not running")
	}

	d.program.Send(cleanupCounterMsg{files, dirs, bytes})
	return nil
}


// Message types for state updates
type stepMsg int
type processCompleteMsg bool
type systemStatsMsg struct {
	cpu  float64
	ram  float64
	disk float64
}
type tickMsg struct{}
type blinkMsg struct{}

// Message types for XML processing  (steps 0-1)
type chunkCountMsg int
type personsFoundMsg int
type bytesReadMsg int
type firstChunkFixedMsg bool
type lastChunkFixedMsg bool

// Message types for location processing phase (steps 2-3)
type agentCounterMsg int
type chunkCounterMsg struct {
	current int
	total   int
}
type dbCounterMsg struct {
	current int
	total   int
}

// Message types for data transformation (steps 4-18)
type coordinateCounterMsg int
type binCounterMsg struct {
	current int
	total   int
}
type agentBinCounterMsg int
type scaleCounterMsg struct {
	current int
	total   int
}
type outputScaleMsg int
type outputCounterMsg struct {
	current int
	total   int
}
type cleanupCounterMsg struct {
	files int
	dirs  int
	bytes int
}
