package functions

import (
	"fmt"

	"ez-utils/src/scale/population"
)

// CalculateRetentionProbabilities calculates retention probabilities for all scales (1-10%)
func CalculateRetentionProbabilities(densityMap population.DensityMap) (map[int]population.RetentionMap, error) {
	if densityMap == nil || len(densityMap) == 0 {
		return nil, fmt.Errorf("density map is empty")
	}
	
	retentionMaps := make(map[int]population.RetentionMap)
	
	// Calculate total population
	totalPopulation := 0
	for _, count := range densityMap {
		totalPopulation += count
	}
	
	if totalPopulation == 0 {
		return nil, fmt.Errorf("no persons found in density map")
	}
	
	// For each scale (1-10%), calculate retention probabilities
	for scale := 1; scale <= 10; scale++ {
		retentionMap := make(population.RetentionMap)
		targetPercentage := float64(scale) / 100.0
		
		// Simple uniform retention strategy for Phase One
		// This ensures all bins have the same retention probability
		// Phase Two will implement more sophisticated density-based strategies
		for binID, count := range densityMap {
			if count > 0 {
				retentionMap[binID] = targetPercentage
			}
		}
		
		retentionMaps[scale] = retentionMap
	}
	
	return retentionMaps, nil
}

// ValidateRetentionMaps validates that retention maps are consistent
func ValidateRetentionMaps(retentionMaps map[int]population.RetentionMap) error {
	if retentionMaps == nil {
		return fmt.Errorf("retention maps is nil")
	}
	
	// Check that we have all scales 1-10
	for scale := 1; scale <= 10; scale++ {
		if _, exists := retentionMaps[scale]; !exists {
			return fmt.Errorf("missing retention map for scale %d%%", scale)
		}
	}
	
	// Validate individual retention maps
	for scale, retentionMap := range retentionMaps {
		if retentionMap == nil {
			return fmt.Errorf("retention map for scale %d%% is nil", scale)
		}
		
		if len(retentionMap) == 0 {
			return fmt.Errorf("retention map for scale %d%% is empty", scale)
		}
		
		// Check probability values are valid
		for binID, prob := range retentionMap {
			if prob < 0.0 || prob > 1.0 {
				return fmt.Errorf("invalid probability %.3f for bin %s in scale %d%%", prob, binID, scale)
			}
		}
	}
	
	return nil
}

// GetRetentionStats calculates statistics for retention maps
func GetRetentionStats(retentionMaps map[int]population.RetentionMap) map[int]RetentionStats {
	stats := make(map[int]RetentionStats)
	
	for scale, retentionMap := range retentionMaps {
		var totalProb float64
		minProb := 1.0
		maxProb := 0.0
		binCount := 0
		
		for _, prob := range retentionMap {
			totalProb += prob
			if prob < minProb {
				minProb = prob
			}
			if prob > maxProb {
				maxProb = prob
			}
			binCount++
		}
		
		avgProb := 0.0
		if binCount > 0 {
			avgProb = totalProb / float64(binCount)
		}
		
		stats[scale] = RetentionStats{
			Scale:             scale,
			BinCount:          binCount,
			AverageProb:       avgProb,
			MinProb:           minProb,
			MaxProb:           maxProb,
			TotalExpectedKept: totalProb,
		}
	}
	
	return stats
}

// RetentionStats contains statistics for a retention map
type RetentionStats struct {
	Scale             int
	BinCount          int
	AverageProb       float64
	MinProb           float64
	MaxProb           float64
	TotalExpectedKept float64
}