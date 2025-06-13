package processing

import (
	"fmt"
	"time"

	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/functions"
	"ez-utils/src/scale/population/tui"
)

// Phase1Processor handles initialization and setup
type Phase1Processor struct {
	config           *population.PhaseOneConfig
	safePointService functions.SafePointService
	tuiEnabled       bool
	startTime        time.Time
}

// NewPhase1Processor creates a new Phase 1 processor
func NewPhase1Processor(config *population.PhaseOneConfig) *Phase1Processor {
	return &Phase1Processor{
		config:           config,
		safePointService: functions.NewSafePointService(),
		tuiEnabled:       true,
	}
}

// DisableTUI disables the terminal UI
func (p1p *Phase1Processor) DisableTUI() {
	p1p.tuiEnabled = false
}

// Process runs Phase 1: Initialization and Setup
func (p1p *Phase1Processor) Process(inputFile string) ([]int64, error) {
	p1p.startTime = time.Now()

	// Initialize data structures
	population.InitializeGlobalData()
	p1p.initializeGrid()

	// Update TUI phase if enabled
	if p1p.tuiEnabled {
		tui.UpdatePhase(1, "Validating Configuration and User Input")
		time.Sleep(100 * time.Millisecond) // Let TUI initialize
	}

	// Step 1: Discover safe points for parallel processing
	safePoints, err := p1p.discoverSafePoints(inputFile)
	if err != nil {
		return nil, fmt.Errorf("safe point discovery failed: %w", err)
	}

	return safePoints, nil
}

// discoverSafePoints finds safe cut points in the XML file
func (p1p *Phase1Processor) discoverSafePoints(inputFile string) ([]int64, error) {
	var progressCallback functions.ProgressCallback

	if p1p.tuiEnabled {
		tui.UpdatePhase(1, "Studying the File")
		progressCallback = func(lines int64, points int) {
			// Update system metrics during safe point discovery
			tui.UpdateSystemMetrics(25.0, 600.0, 15.0)
		}
	}

	result, err := p1p.safePointService.DiscoverSafePoints(
		inputFile,
		p1p.config.SafePointInterval,
		progressCallback,
	)
	if err != nil {
		return nil, err
	}

	// Store for Phase 2 access
	population.SetSafeCutPoints(result.SafePoints)

	return result.SafePoints, nil
}


// initializeGrid sets up the grid bounds
func (p1p *Phase1Processor) initializeGrid() {
	bounds := p1p.config.GridConfig.InitialBounds
	if bounds.CellSize == 0 {
		bounds = population.GridBounds{
			MinX:     -1000000,
			MaxX:     1000000,
			MinY:     -1000000,
			MaxY:     1000000,
			CellSize: 1000,
		}
	}
	population.SetGlobalGrid(bounds)
}