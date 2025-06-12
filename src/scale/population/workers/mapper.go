package workers

import (
	"sync"
	"sync/atomic"

	"ez-utils/src/scale/population"
)

// MapperWorker maintains an independent local spatial grid and counts persons per bin
type MapperWorker struct {
	id          string
	inputQueue  <-chan Coordinates
	doneQueue   chan<- CompletionSignal
	
	// Local independent grid - no shared state!
	localGrid     map[string]int
	localMutex    sync.RWMutex
	cellSize      float64
	
	// Metrics
	coordinatesProcessed int64
}

// NewMapperWorker creates a new mapper worker with independent local grid
func NewMapperWorker(id string, inputQueue <-chan Coordinates, doneQueue chan<- CompletionSignal) *MapperWorker {
	// Get cell size from global grid bounds
	bounds := population.GetGlobalGrid()
	cellSize := bounds.CellSize
	if cellSize == 0 {
		cellSize = 1000 // Default to 1km if not set
	}
	
	return &MapperWorker{
		id:        id,
		inputQueue: inputQueue,
		doneQueue:  doneQueue,
		localGrid:  make(map[string]int),
		cellSize:   cellSize,
	}
}

// Process reads coordinates and updates the local independent grid
func (m *MapperWorker) Process() {
	// Reset counter to ensure clean start
	atomic.StoreInt64(&m.coordinatesProcessed, 0)
	
	// Process until input queue is closed
	for coord := range m.inputQueue {
		m.processCoordinate(coord)
	}
	
	// Signal completion
	m.doneQueue <- CompletionSignal{
		WorkerID: m.id,
	}
}

// processCoordinate processes a single coordinate in the local grid
func (m *MapperWorker) processCoordinate(coord Coordinates) {
	// Calculate bin ID using local cell size
	binID := population.CalculateBinID(coord.X, coord.Y, m.cellSize)
	
	// Increment count in local grid (thread-safe)
	m.localMutex.Lock()
	m.localGrid[binID]++
	m.localMutex.Unlock()
	
	// Track that we processed a coordinate
	atomic.AddInt64(&m.coordinatesProcessed, 1)
}

// GetLocalGrid returns a copy of the local grid for combination
func (m *MapperWorker) GetLocalGrid() map[string]int {
	m.localMutex.RLock()
	defer m.localMutex.RUnlock()
	
	// Return a copy to prevent external modification
	copy := make(map[string]int)
	for binID, count := range m.localGrid {
		copy[binID] = count
	}
	return copy
}

// GetMetrics returns mapper metrics
func (m *MapperWorker) GetMetrics() (processed, gridSize int64) {
	m.localMutex.RLock()
	gridSize = int64(len(m.localGrid))
	m.localMutex.RUnlock()
	return atomic.LoadInt64(&m.coordinatesProcessed), gridSize
}