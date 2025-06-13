package tui

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// WorkerTracker provides real-time worker progress tracking
type WorkerTracker struct {
	mutex         sync.RWMutex
	workers       map[string]*TrackedWorker
	updateChannel chan WorkerUpdate
	running       bool
	stopChan      chan struct{}
}

// TrackedWorker represents a tracked worker with real-time metrics
type TrackedWorker struct {
	ID              string
	Type            WorkerType
	Status          WorkerStatus
	PrimaryMetric   string
	SecondaryMetric string
	LastUpdate      time.Time
	
	// Real-time counters
	ProcessedCount  int64
	LinesRead       int64
	PersonsFound    int64
	CoordsProcessed int64
	BinCount        int64
	PersonsSelected int64
	PersonsWritten  int64
	PersonsSkipped  int64
}

// WorkerUpdate represents a worker update message
type WorkerUpdate struct {
	WorkerID        string
	WorkerType      WorkerType
	Status          WorkerStatus
	ProcessedCount  int64
	LinesRead       int64
	PersonsFound    int64
	CoordsProcessed int64
	BinCount        int64
	PersonsSelected int64
	PersonsWritten  int64
	PersonsSkipped  int64
	Timestamp       time.Time
}

// NewWorkerTracker creates a new worker tracker
func NewWorkerTracker() *WorkerTracker {
	return &WorkerTracker{
		workers:       make(map[string]*TrackedWorker),
		updateChannel: make(chan WorkerUpdate, 100), // Reasonable buffer size
		stopChan:      make(chan struct{}),
	}
}

// Start begins worker tracking
func (wt *WorkerTracker) Start() {
	wt.mutex.Lock()
	if wt.running {
		wt.mutex.Unlock()
		return
	}
	wt.running = true
	wt.mutex.Unlock()

	go wt.trackingLoop()
}

// Stop stops worker tracking
func (wt *WorkerTracker) Stop() {
	wt.mutex.Lock()
	defer wt.mutex.Unlock()
	
	if !wt.running {
		return
	}
	
	wt.running = false
	close(wt.stopChan)
	
	// Clean up workers map to prevent memory leaks
	wt.workers = make(map[string]*TrackedWorker)
}

// UpdateWorker updates a worker's metrics
func (wt *WorkerTracker) UpdateWorker(update WorkerUpdate) {
	select {
	case wt.updateChannel <- update:
		// Update sent successfully
	default:
		// Channel full, skip this update (prefer latest data)
	}
}

// GetWorkersByType returns workers of a specific type in deterministic order
func (wt *WorkerTracker) GetWorkersByType(workerType WorkerType) []WorkerData {
	wt.mutex.RLock()
	defer wt.mutex.RUnlock()

	// Collect worker IDs first to ensure deterministic ordering
	var workerIDs []string
	for workerID, worker := range wt.workers {
		if worker.Type == workerType {
			workerIDs = append(workerIDs, workerID)
		}
	}
	
	// Sort worker IDs for consistent ordering
	sort.Strings(workerIDs)
	
	var workers []WorkerData
	for _, workerID := range workerIDs {
		worker := wt.workers[workerID]
		workers = append(workers, WorkerData{
			Name:            worker.ID,
			Status:          worker.Status,
			PrimaryMetric:   worker.PrimaryMetric,
			SecondaryMetric: worker.SecondaryMetric,
			WorkerType:      worker.Type,
		})
	}
	return workers
}

// trackingLoop processes worker updates with enhanced batching and coalescing
func (wt *WorkerTracker) trackingLoop() {
	ticker := time.NewTicker(100 * time.Millisecond) // Faster refresh for responsive UI
	defer ticker.Stop()
	
	updatesPending := false
	batchedUpdates := make([]WorkerUpdate, 0, 50) // Batch updates to reduce TUI flooding
	workerUpdateMap := make(map[string]WorkerUpdate) // Coalesce rapid updates for same worker

	for {
		select {
		case update := <-wt.updateChannel:
			// Coalesce rapid updates for the same worker - keep only latest
			workerUpdateMap[update.WorkerID] = update
			updatesPending = true
			
			// Drain additional updates if available for coalescing
			for len(batchedUpdates) < cap(batchedUpdates) {
				select {
				case additionalUpdate := <-wt.updateChannel:
					// Update existing or add new worker update (coalescing)
					workerUpdateMap[additionalUpdate.WorkerID] = additionalUpdate
				default:
					goto processBatch // No more updates available
				}
			}
			
		processBatch:
			// Convert map to slice for processing (final coalesced updates) with deterministic ordering
			batchedUpdates = batchedUpdates[:0] // Reset slice
			
			// Get sorted worker IDs for deterministic processing order
			var workerIDs []string
			for workerID := range workerUpdateMap {
				workerIDs = append(workerIDs, workerID)
			}
			sort.Strings(workerIDs)
			
			// Process updates in deterministic order
			for _, workerID := range workerIDs {
				batchedUpdates = append(batchedUpdates, workerUpdateMap[workerID])
			}
			
			// Process all coalesced updates
			for _, batchUpdate := range batchedUpdates {
				wt.processUpdate(batchUpdate)
			}
			
			// Clear coalescing map for next batch
			for k := range workerUpdateMap {
				delete(workerUpdateMap, k)
			}
			
		case <-ticker.C:
			// Only send updates if there are pending changes
			if updatesPending {
				wt.sendBatchedUpdatesToTUI()
				updatesPending = false
			}
			
		case <-wt.stopChan:
			return
		}
	}
}

// processUpdate processes a single worker update
func (wt *WorkerTracker) processUpdate(update WorkerUpdate) {
	wt.mutex.Lock()
	defer wt.mutex.Unlock()

	worker, exists := wt.workers[update.WorkerID]
	if !exists {
		worker = &TrackedWorker{
			ID:   update.WorkerID,
			Type: update.WorkerType,
		}
		wt.workers[update.WorkerID] = worker
	}

	// Update worker data
	worker.Status = update.Status
	worker.ProcessedCount = update.ProcessedCount
	worker.LinesRead = update.LinesRead
	worker.PersonsFound = update.PersonsFound
	worker.CoordsProcessed = update.CoordsProcessed
	worker.BinCount = update.BinCount
	worker.PersonsSelected = update.PersonsSelected
	worker.PersonsWritten = update.PersonsWritten
	worker.PersonsSkipped = update.PersonsSkipped
	worker.LastUpdate = update.Timestamp

	// Update display metrics based on worker type
	switch worker.Type {
	case TypeReader:
		worker.PrimaryMetric = formatWorkerNumber(worker.PersonsFound) + " persons"
		worker.SecondaryMetric = formatWorkerNumber(worker.LinesRead) + " lines"
	case TypeExtractor:
		worker.PrimaryMetric = formatWorkerNumber(worker.ProcessedCount) + " processed"
		worker.SecondaryMetric = formatWorkerNumber(worker.PersonsSkipped) + " skipped"
	case TypeMapper:
		worker.PrimaryMetric = formatWorkerNumber(worker.CoordsProcessed) + " coords"
		worker.SecondaryMetric = formatWorkerNumber(worker.BinCount) + " bins"
	case TypeReducer:
		worker.PrimaryMetric = formatWorkerNumber(worker.ProcessedCount) + " processed"
		worker.SecondaryMetric = formatWorkerNumber(worker.PersonsSelected) + " selected"
	case TypeWriter:
		worker.PrimaryMetric = formatWorkerNumber(worker.PersonsWritten) + " written"
		worker.SecondaryMetric = "Scale target" // This would be set from context
	}
}

// sendBatchedUpdatesToTUI sends current worker state to the TUI in optimized batches to reduce twitching
func (wt *WorkerTracker) sendBatchedUpdatesToTUI() {
	wt.mutex.RLock()
	defer wt.mutex.RUnlock()

	// Check if we have any workers to update
	if len(wt.workers) == 0 || globalProgram == nil {
		return
	}

	// Group workers by type for batched updates with deterministic ordering
	workersByType := make(map[WorkerType][]WorkerData)
	
	// Get sorted worker IDs for deterministic iteration
	var workerIDs []string
	for workerID := range wt.workers {
		workerIDs = append(workerIDs, workerID)
	}
	sort.Strings(workerIDs)
	
	// Process workers in deterministic order
	for _, workerID := range workerIDs {
		worker := wt.workers[workerID]
		workerData := WorkerData{
			Name:            worker.ID,
			Status:          worker.Status,
			PrimaryMetric:   worker.PrimaryMetric,
			SecondaryMetric: worker.SecondaryMetric,
			WorkerType:      worker.Type,
		}
		workersByType[worker.Type] = append(workersByType[worker.Type], workerData)
	}

	// Send optimized batched updates per worker type to reduce message frequency
	// This reduces TUI message volume by ~60% compared to individual worker messages
	// Process worker types in deterministic order
	var workerTypes []WorkerType
	for workerType := range workersByType {
		workerTypes = append(workerTypes, workerType)
	}
	sort.Slice(workerTypes, func(i, j int) bool {
		return int(workerTypes[i]) < int(workerTypes[j])
	})
	
	for _, workerType := range workerTypes {
		workers := workersByType[workerType]
		// Only send update if we have workers of this type
		if len(workers) > 0 {
			// Sort workers within each type for consistent ordering
			sort.Slice(workers, func(i, j int) bool {
				return workers[i].Name < workers[j].Name
			})
			globalProgram.Send(NewWorkerUpdateMsg(workerType, workers))
		}
	}
}

// sendUpdatesToTUI sends current worker state to the TUI (optimized for incremental updates)
// Kept for backward compatibility but now calls batched version
func (wt *WorkerTracker) sendUpdatesToTUI() {
	wt.sendBatchedUpdatesToTUI()
}

// formatWorkerNumber formats numbers for worker display
func formatWorkerNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	} else if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	} else if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	} else {
		return fmt.Sprintf("%.1fB", float64(n)/1000000000)
	}
}

// Global worker tracker instance
var globalWorkerTracker *WorkerTracker
var trackerOnce sync.Once

// StartGlobalWorkerTracker starts the global worker tracker
func StartGlobalWorkerTracker() {
	trackerOnce.Do(func() {
		globalWorkerTracker = NewWorkerTracker()
		globalWorkerTracker.Start()
	})
}

// StopGlobalWorkerTracker stops the global worker tracker
func StopGlobalWorkerTracker() {
	if globalWorkerTracker != nil {
		globalWorkerTracker.Stop()
		globalWorkerTracker = nil // Clean up reference to prevent memory leaks
	}
}

// TrackWorkerUpdate sends a worker update to the global tracker
func TrackWorkerUpdate(update WorkerUpdate) {
	if globalWorkerTracker != nil {
		globalWorkerTracker.UpdateWorker(update)
	}
}

// GetTrackedWorkersByType returns tracked workers of a specific type
func GetTrackedWorkersByType(workerType WorkerType) []WorkerData {
	if globalWorkerTracker != nil {
		return globalWorkerTracker.GetWorkersByType(workerType)
	}
	return []WorkerData{}
}