package config

import "ez-utils/src/scale/population"

// ToPhaseOneConfig converts config to population.PhaseOneConfig
func (c *Config) ToPhaseOneConfig() *population.PhaseOneConfig {
	return &population.PhaseOneConfig{
		ReaderCount:       c.Workers.Readers,
		SafePointInterval: c.Workers.SafePointInterval,
		ExtractorCount:    c.Workers.Extractors,
		HashMapCount:      c.Workers.Hashmap,
		GridConfig: population.GridConfig{
			InitialBounds: population.GridBounds{
				MinX:     -1000000,
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
	scales := c.Population.Scales
	if len(scales) == 0 {
		scales = []int{1, 5, 10}
	}
	
	return &population.PhaseTwoConfig{
		ReducerCount:    c.Workers.Reducers,
		DatabaseCount:   c.Workers.Database,
		FileWriterCount: c.Workers.FileWriters,
		OutputDir:       c.Population.OutputDir,
		Scales:          scales,
	}
}