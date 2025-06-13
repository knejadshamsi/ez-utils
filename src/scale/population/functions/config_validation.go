package functions

import (
	"fmt"
	"os"
	"time"

	"ez-utils/src/scale/population"
)

// logWarningToFile logs warning messages to a unique log file
func logWarningToFile(message string) {
	timestamp := time.Now().Format("20060102-150405-000000")
	filename := fmt.Sprintf("config-validation-warning-%s.log", timestamp)
	
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return // Silently fail to avoid stdout pollution
	}
	defer file.Close()
	
	file.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), message))
}

// ConfigValidator validates population processing configuration
type ConfigValidator interface {
	// ValidatePhaseOneConfig validates Phase One configuration parameters
	ValidatePhaseOneConfig(config *population.PhaseOneConfig) error
	
	// ValidatePhaseTwoConfig validates Phase Two configuration parameters  
	ValidatePhaseTwoConfig(config *population.PhaseTwoConfig) error
	
	// ValidateSafePointSettings validates safe point discovery settings
	ValidateSafePointSettings(interval int64, numReaders int) error
}

// configValidator is the concrete implementation
type configValidator struct{}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() ConfigValidator {
	return &configValidator{}
}

// ValidatePhaseOneConfig validates Phase One configuration parameters
func (v *configValidator) ValidatePhaseOneConfig(config *population.PhaseOneConfig) error {
	if config == nil {
		return fmt.Errorf("phase one config cannot be nil")
	}

	// Validate reader count
	if config.ReaderCount <= 0 {
		return fmt.Errorf("reader count must be positive, got: %d", config.ReaderCount)
	}
	
	if config.ReaderCount > 16 {
		return fmt.Errorf("reader count too high: %d (maximum: 16)", config.ReaderCount)
	}

	// Validate safe point interval
	if err := v.ValidateSafePointSettings(config.SafePointInterval, config.ReaderCount); err != nil {
		return fmt.Errorf("safe point validation failed: %w", err)
	}

	// Validate extractor count
	if config.ExtractorCount <= 0 {
		return fmt.Errorf("extractor count must be positive, got: %d", config.ExtractorCount)
	}
	
	if config.ExtractorCount > 32 {
		return fmt.Errorf("extractor count too high: %d (maximum: 32)", config.ExtractorCount)
	}

	// Validate hashmap count
	if config.HashMapCount <= 0 {
		return fmt.Errorf("hashmap count must be positive, got: %d", config.HashMapCount)
	}
	
	if config.HashMapCount > 32 {
		return fmt.Errorf("hashmap count too high: %d (maximum: 32)", config.HashMapCount)
	}

	// Validate grid configuration
	if err := v.validateGridConfig(config.GridConfig); err != nil {
		return fmt.Errorf("grid config validation failed: %w", err)
	}

	return nil
}

// ValidatePhaseTwoConfig validates Phase Two configuration parameters
func (v *configValidator) ValidatePhaseTwoConfig(config *population.PhaseTwoConfig) error {
	if config == nil {
		return fmt.Errorf("phase two config cannot be nil")
	}

	// Validate reducer count
	if config.ReducerCount <= 0 {
		return fmt.Errorf("reducer count must be positive, got: %d", config.ReducerCount)
	}
	
	if config.ReducerCount > 16 {
		return fmt.Errorf("reducer count too high: %d (maximum: 16)", config.ReducerCount)
	}

	// Validate database count
	if config.DatabaseCount <= 0 {
		return fmt.Errorf("database count must be positive, got: %d", config.DatabaseCount)
	}
	
	if config.DatabaseCount > 8 {
		return fmt.Errorf("database count too high: %d (maximum: 8)", config.DatabaseCount)
	}

	// Validate file writer count
	if config.FileWriterCount <= 0 {
		return fmt.Errorf("file writer count must be positive, got: %d", config.FileWriterCount)
	}
	
	if config.FileWriterCount > 8 {
		return fmt.Errorf("file writer count too high: %d (maximum: 8)", config.FileWriterCount)
	}

	// Validate output directory
	if config.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Validate scales
	if len(config.Scales) == 0 {
		return fmt.Errorf("at least one scale must be specified")
	}
	
	for _, scale := range config.Scales {
		if scale < 1 || scale > 10 {
			return fmt.Errorf("scale must be between 1 and 10, got: %d", scale)
		}
	}

	return nil
}

// ValidateSafePointSettings validates safe point discovery settings
func (v *configValidator) ValidateSafePointSettings(interval int64, numReaders int) error {
	if interval <= 0 {
		return fmt.Errorf("safe point interval must be positive, got: %d", interval)
	}

	// No minimum interval - small files should work with small intervals

	// Maximum interval to prevent memory issues
	const maxInterval = 1000000
	if interval > maxInterval {
		return fmt.Errorf("safe point interval too large: %d (maximum: %d lines)", interval, maxInterval)
	}

	if numReaders <= 0 {
		return fmt.Errorf("number of readers must be positive, got: %d", numReaders)
	}

	// Warn if interval is too small for effective parallelization
	const recommendedMinInterval = 10000
	if interval < recommendedMinInterval && numReaders > 2 {
		// Log warning to file instead of stdout to avoid interfering with TUI
		logWarningToFile(fmt.Sprintf("Warning: safe point interval %d may be too small for %d readers (recommended: >= %d)", 
			interval, numReaders, recommendedMinInterval))
	}

	return nil
}

// validateGridConfig validates grid configuration settings
func (v *configValidator) validateGridConfig(config population.GridConfig) error {
	bounds := config.InitialBounds

	// Validate cell size
	if bounds.CellSize <= 0 {
		return fmt.Errorf("grid cell size must be positive, got: %f", bounds.CellSize)
	}
	
	if bounds.CellSize < 1 {
		return fmt.Errorf("grid cell size too small: %f (minimum: 1.0)", bounds.CellSize)
	}
	
	if bounds.CellSize > 100000 {
		return fmt.Errorf("grid cell size too large: %f (maximum: 100000.0)", bounds.CellSize)
	}

	// Validate bounds
	if bounds.MaxX <= bounds.MinX {
		return fmt.Errorf("invalid X bounds: max (%f) must be greater than min (%f)", bounds.MaxX, bounds.MinX)
	}
	
	if bounds.MaxY <= bounds.MinY {
		return fmt.Errorf("invalid Y bounds: max (%f) must be greater than min (%f)", bounds.MaxY, bounds.MinY)
	}

	// Check for reasonable grid size
	gridWidth := (bounds.MaxX - bounds.MinX) / bounds.CellSize
	gridHeight := (bounds.MaxY - bounds.MinY) / bounds.CellSize
	
	if gridWidth > 100000 || gridHeight > 100000 {
		return fmt.Errorf("grid dimensions too large: %fx%f cells (maximum: 100000x100000)", gridWidth, gridHeight)
	}

	// Validate expansion margin
	if config.ExpansionMargin < 0 {
		return fmt.Errorf("expansion margin cannot be negative, got: %f", config.ExpansionMargin)
	}

	return nil
}

// ValidationRules contains the validation constants used throughout the system
type ValidationRules struct {
	// Safe point settings
	MinSafePointInterval int64
	MaxSafePointInterval int64
	RecommendedMinInterval int64
	
	// Worker limits
	MaxReaders     int
	MaxExtractors  int
	MaxHashMaps    int
	MaxReducers    int
	MaxDatabaseWorkers int
	MaxFileWriters int
	
	// Grid limits
	MinCellSize    float64
	MaxCellSize    float64
	MaxGridCells   float64
	
	// Scale limits
	MinScale int
	MaxScale int
}

// GetValidationRules returns the validation rules used by the system
func GetValidationRules() ValidationRules {
	return ValidationRules{
		MinSafePointInterval:   1000,
		MaxSafePointInterval:   1000000,
		RecommendedMinInterval: 10000,
		
		MaxReaders:         16,
		MaxExtractors:      32,
		MaxHashMaps:        32,
		MaxReducers:        16,
		MaxDatabaseWorkers: 8,
		MaxFileWriters:     8,
		
		MinCellSize:  1.0,
		MaxCellSize:  100000.0,
		MaxGridCells: 100000.0,
		
		MinScale: 1,
		MaxScale: 10,
	}
}