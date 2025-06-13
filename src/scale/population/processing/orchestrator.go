package processing

import (
	"fmt"
	"time"

	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/tui"
)

// PopulationScalingOrchestrator coordinates all 5 phases of population scaling
type PopulationScalingOrchestrator struct {
	phaseOneConfig *population.PhaseOneConfig
	phaseTwoConfig *population.PhaseTwoConfig
	inputFile      string
	tuiEnabled     bool
}

// NewPopulationScalingOrchestrator creates a new orchestrator
func NewPopulationScalingOrchestrator(phaseOneConfig *population.PhaseOneConfig, phaseTwoConfig *population.PhaseTwoConfig, inputFile string) *PopulationScalingOrchestrator {
	return &PopulationScalingOrchestrator{
		phaseOneConfig: phaseOneConfig,
		phaseTwoConfig: phaseTwoConfig,
		inputFile:      inputFile,
		tuiEnabled:     true,
	}
}

// DisableTUI disables the terminal UI for all phases
func (pso *PopulationScalingOrchestrator) DisableTUI() {
	pso.tuiEnabled = false
}

// Process runs the complete 5-phase population scaling pipeline
func (pso *PopulationScalingOrchestrator) Process() error {
	
	// Initialize TUI system for real-time tracking
	if pso.tuiEnabled {
		tui.InitTUI()
		defer tui.StopTUI()
		// Give TUI time to initialize
		time.Sleep(200 * time.Millisecond)
	}
	
	// Phase 1: Initialization and Setup
	phase1 := NewPhase1Processor(pso.phaseOneConfig)
	if !pso.tuiEnabled {
		phase1.DisableTUI()
	}
	
	safePoints, err := phase1.Process(pso.inputFile)
	if err != nil {
		return fmt.Errorf("Phase 1 failed: %w", err)
	}
	
	// Phase 2: Density Mapping
	phase2 := NewPhase2Processor(pso.phaseOneConfig)
	if !pso.tuiEnabled {
		phase2.DisableTUI()
	}
	
	if err := phase2.Process(pso.inputFile, safePoints); err != nil {
		return fmt.Errorf("Phase 2 failed: %w", err)
	}
	
	// Phase 3: Inter-Phase Preparation
	phase3 := NewPhase3Processor()
	if !pso.tuiEnabled {
		phase3.DisableTUI()
	}
	
	if err := phase3.Process(phase2.GetMappers()); err != nil {
		return fmt.Errorf("Phase 3 failed: %w", err)
	}
	
	// Phase 4: Population Scaling and Output
	phase4 := NewPhase4Processor(pso.phaseTwoConfig, pso.inputFile)
	if !pso.tuiEnabled {
		phase4.DisableTUI()
	}
	
	if err := phase4.Process(); err != nil {
		return fmt.Errorf("Phase 4 failed: %w", err)
	}
	
	// Phase 5: File Finalization and Cleanup
	phase5 := NewPhase5Processor(pso.phaseTwoConfig)
	if !pso.tuiEnabled {
		phase5.DisableTUI()
	}
	
	if err := phase5.Process(); err != nil {
		return fmt.Errorf("Phase 5 failed: %w", err)
	}
	
	// Clear terminal after completion
	if pso.tuiEnabled {
		fmt.Print("\033[H\033[2J") // Clear screen and move cursor to top
	}
	
	return nil
}