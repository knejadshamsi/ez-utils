package config

import (
	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/functions"
)

// ToPhaseOneConfig converts config to population.PhaseOneConfig
func (c *Config) ToPhaseOneConfig() *population.PhaseOneConfig {
	return &population.PhaseOneConfig{
		ReaderCount:       c.Workers.Readers,
		SafePointInterval: c.Workers.SafePointInterval,
		ExtractorCount:    c.Workers.Extractors,
		HashMapCount:      c.Workers.Hashmap,
		GridConfig: population.GridConfig{
			InitialBounds: population.GridBounds{
				MinX:     -1000000, // Default bounds - could be configurable
				MaxX:     1000000,
				MinY:     -1000000,
				MaxY:     1000000,
				CellSize: 1000,
			},
			BinSize:         1000,
			ExpansionMargin: 10000,
		},
	}
}

// ToPhaseTwoConfig converts config to population.PhaseTwoConfig
func (c *Config) ToPhaseTwoConfig() *population.PhaseTwoConfig {
	return &population.PhaseTwoConfig{
		ReducerCount:    c.Workers.Reducers,
		DatabaseCount:   c.Workers.Database,
		FileWriterCount: c.Workers.FileWriters,
		OutputDir:       c.Population.OutputDir,
		Scales:          []int{1, 5, 10}, // Standard scales: 1%, 5%, 10%
	}
}

// ValidatePopulationConfig validates population-specific configuration using the centralized validator
func (c *Config) ValidatePopulationConfig() error {
	validator := functions.NewConfigValidator()
	
	// Validate Phase One config
	phaseOneConfig := c.ToPhaseOneConfig()
	if err := validator.ValidatePhaseOneConfig(phaseOneConfig); err != nil {
		return err
	}
	
	// Validate Phase Two config
	phaseTwoConfig := c.ToPhaseTwoConfig()
	if err := validator.ValidatePhaseTwoConfig(phaseTwoConfig); err != nil {
		return err
	}
	
	return nil
}