package population

import (
	"fmt"
	"sync"
)

// Package-level variables for phase communication
// These allow Phase Two to access Phase One results without complex communication
var (
	// globalDensityMap holds the density map from Phase One
	globalDensityMap DensityMap
	densityMapMutex  sync.RWMutex

	// retentionMaps holds retention probability maps for each scale (1-10%)
	retentionMaps map[int]RetentionMap
	retentionMutex sync.RWMutex

	// safeCutPoints holds the in-memory safe points array for Phase Two reducers
	safeCutPoints []int64
	safePointsMutex sync.RWMutex

	// globalGrid holds the current grid bounds
	globalGrid GridBounds
	gridMutex  sync.RWMutex
)

// InitializeGlobalData initializes package-level data structures
func InitializeGlobalData() {
	globalDensityMap = make(DensityMap)
	retentionMaps = make(map[int]RetentionMap)
	for i := 1; i <= 10; i++ {
		retentionMaps[i] = make(RetentionMap)
	}
}

// SetDensityMap safely sets the global density map
func SetDensityMap(dm DensityMap) {
	densityMapMutex.Lock()
	defer densityMapMutex.Unlock()
	globalDensityMap = dm
}

// GetDensityMap safely retrieves the global density map
func GetDensityMap() DensityMap {
	densityMapMutex.RLock()
	defer densityMapMutex.RUnlock()
	
	// Return a copy to prevent external modification
	copy := make(DensityMap)
	for k, v := range globalDensityMap {
		copy[k] = v
	}
	return copy
}

// SetRetentionMap safely sets the retention map for a specific scale
func SetRetentionMap(scale int, rm RetentionMap) {
	if scale < 1 || scale > 10 {
		return
	}
	
	retentionMutex.Lock()
	defer retentionMutex.Unlock()
	retentionMaps[scale] = rm
}

// GetRetentionMap safely retrieves the retention map for a specific scale
func GetRetentionMap(scale int) RetentionMap {
	if scale < 1 || scale > 10 {
		return nil
	}
	
	retentionMutex.RLock()
	defer retentionMutex.RUnlock()
	
	// Return a copy to prevent external modification
	copy := make(RetentionMap)
	for k, v := range retentionMaps[scale] {
		copy[k] = v
	}
	return copy
}

// GetAllRetentionMaps safely retrieves all retention maps
func GetAllRetentionMaps() map[int]RetentionMap {
	retentionMutex.RLock()
	defer retentionMutex.RUnlock()
	
	// Return a deep copy
	allMaps := make(map[int]RetentionMap)
	for scale, rm := range retentionMaps {
		copy := make(RetentionMap)
		for k, v := range rm {
			copy[k] = v
		}
		allMaps[scale] = copy
	}
	return allMaps
}

// SetSafeCutPoints safely sets the safe cut points array
func SetSafeCutPoints(points []int64) {
	safePointsMutex.Lock()
	defer safePointsMutex.Unlock()
	safeCutPoints = make([]int64, len(points))
	copy(safeCutPoints, points)
}

// GetSafeCutPoints safely retrieves the safe cut points array
func GetSafeCutPoints() []int64 {
	safePointsMutex.RLock()
	defer safePointsMutex.RUnlock()
	
	// Return a copy to prevent external modification
	copySlice := make([]int64, len(safeCutPoints))
	copy(copySlice, safeCutPoints)
	return copySlice
}

// SetGlobalGrid safely sets the global grid bounds
func SetGlobalGrid(bounds GridBounds) {
	gridMutex.Lock()
	defer gridMutex.Unlock()
	globalGrid = bounds
}

// GetGlobalGrid safely retrieves the global grid bounds
func GetGlobalGrid() GridBounds {
	gridMutex.RLock()
	defer gridMutex.RUnlock()
	return globalGrid
}

// IncrementBinCount safely increments the count for a bin
func IncrementBinCount(binID string) {
	densityMapMutex.Lock()
	defer densityMapMutex.Unlock()
	globalDensityMap[binID]++
}

// AddToDensityMap safely adds a count to a bin
func AddToDensityMap(binID string, count int) {
	densityMapMutex.Lock()
	defer densityMapMutex.Unlock()
	globalDensityMap[binID] += count
}

// GetBinCount safely retrieves the count for a bin
func GetBinCount(binID string) int {
	densityMapMutex.RLock()
	defer densityMapMutex.RUnlock()
	return globalDensityMap[binID]
}

// CalculateBinID calculates the bin ID for given coordinates
func CalculateBinID(x, y float64, cellSize float64) string {
	row := int(y / cellSize)
	col := int(x / cellSize)
	return fmt.Sprintf("%d,%d", row, col)
}

// IsCoordinateInGrid checks if coordinates fall within current grid bounds
func IsCoordinateInGrid(x, y float64) bool {
	gridMutex.RLock()
	defer gridMutex.RUnlock()
	
	return x >= globalGrid.MinX && x <= globalGrid.MaxX &&
		   y >= globalGrid.MinY && y <= globalGrid.MaxY
}