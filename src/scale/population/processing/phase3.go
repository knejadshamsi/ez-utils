package processing

import (
	"fmt"
	"time"

	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/functions"
	"ez-utils/src/scale/population/workers"
	"ez-utils/src/scale/population/tui"
)

// Phase3Processor handles inter-phase preparation
type Phase3Processor struct {
	tuiEnabled bool
}

// NewPhase3Processor creates a new Phase 3 processor
func NewPhase3Processor() *Phase3Processor {
	return &Phase3Processor{
		tuiEnabled: true,
	}
}

// DisableTUI disables the terminal UI
func (p3p *Phase3Processor) DisableTUI() {
	p3p.tuiEnabled = false
}

// Process runs Phase 3: Inter-Phase Preparation
func (p3p *Phase3Processor) Process(mappers []*workers.MapperWorker) error {
	if p3p.tuiEnabled {
		tui.UpdatePhase(3, "Combining Local Grids")
		time.Sleep(1 * time.Second)
	}
	
	if err := p3p.postProcess(mappers); err != nil {
		return fmt.Errorf("post-processing failed: %w", err)
	}
	
	if p3p.tuiEnabled {
		tui.UpdatePhase(3, "Retention Calculation Complete")
		time.Sleep(2 * time.Second)
	}
	
	return nil
}

// postProcess performs grid combination and retention calculation
func (p3p *Phase3Processor) postProcess(mappers []*workers.MapperWorker) error {
	// Combine mapper grids
	finalGrid := make(population.DensityMap)
	for _, mapper := range mappers {
		for binID, count := range mapper.GetLocalGrid() {
			if count > 0 {
				finalGrid[binID] += count
			}
		}
	}
	population.SetDensityMap(finalGrid)
	
	if p3p.tuiEnabled {
		tui.UpdatePhase(3, "Removing Empty Bins")
		time.Sleep(500 * time.Millisecond)
	}
	
	// Calculate retention probabilities
	if p3p.tuiEnabled {
		tui.UpdatePhase(3, "Calculating Retention Ratio")
	}
	
	retentionMaps, err := functions.CalculateRetentionProbabilities(finalGrid)
	if err != nil {
		return err
	}
	
	for scale, retentionMap := range retentionMaps {
		population.SetRetentionMap(scale, retentionMap)
	}
	
	return nil
}