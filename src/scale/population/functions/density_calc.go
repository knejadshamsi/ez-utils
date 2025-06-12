package functions

import (
	"fmt"
	"sync"

	"ez-utils/src/scale/population"
)

// CalculateBinID calculates the bin ID for given coordinates
func CalculateBinID(x, y float64, bounds population.GridBounds) string {
	// Use the cellSize from bounds or a default value
	cellSize := bounds.CellSize
	if cellSize == 0 {
		cellSize = 1000 // Default to 1km if not set
	}
	return population.CalculateBinID(x, y, cellSize)
}

// IncrementLocalBinCount increments the count for a bin in a local map (thread-safe)
// Returns true if the bin should be flushed to global map
func IncrementLocalBinCount(localMap map[string]int, binID string, mutex *sync.RWMutex, flushThreshold int) bool {
	mutex.Lock()
	localMap[binID]++
	count := localMap[binID]
	mutex.Unlock()
	
	// Return true if we should flush (every flushThreshold increments)
	return count%flushThreshold == 0
}

// FlushBinToGlobal flushes a specific bin's count from local map to the global density map
func FlushBinToGlobal(localMap map[string]int, binID string, mutex *sync.RWMutex) {
	mutex.Lock()
	count := localMap[binID]
	localMap[binID] = 0
	mutex.Unlock()
	
	if count > 0 {
		population.AddToDensityMap(binID, count)
	}
}

// FlushAllLocalDensityMap flushes all local counts to global density map
func FlushAllLocalDensityMap(localMap map[string]int, mutex *sync.RWMutex) {
	mutex.Lock()
	localCopy := make(map[string]int)
	for binID, count := range localMap {
		localCopy[binID] = count
		localMap[binID] = 0
	}
	mutex.Unlock()
	
	// Add all counts to global map
	for binID, count := range localCopy {
		if count > 0 {
			population.AddToDensityMap(binID, count)
		}
	}
}

// ValidateDensityMap validates the density map for consistency
func ValidateDensityMap(densityMap population.DensityMap) error {
	if densityMap == nil {
		return fmt.Errorf("density map is nil")
	}
	
	totalCount := 0
	for binID, count := range densityMap {
		if count < 0 {
			return fmt.Errorf("negative count %d for bin %s", count, binID)
		}
		totalCount += count
	}
	
	if totalCount == 0 {
		return fmt.Errorf("density map is empty - no agents found")
	}
	
	return nil
}

// GetDensityStats calculates statistics for the density map
func GetDensityStats(densityMap population.DensityMap) (int, int, float64, int, int) {
	if len(densityMap) == 0 {
		return 0, 0, 0, 0, 0
	}
	
	totalAgents := 0
	minCount := int(^uint(0) >> 1) // Max int
	maxCount := 0
	
	for _, count := range densityMap {
		totalAgents += count
		if count < minCount {
			minCount = count
		}
		if count > maxCount {
			maxCount = count
		}
	}
	
	avgCount := float64(totalAgents) / float64(len(densityMap))
	
	return totalAgents, len(densityMap), avgCount, minCount, maxCount
}