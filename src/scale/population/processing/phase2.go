package processing

import (
	"fmt"
	"sync"
	"time"

	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/workers"
	"ez-utils/src/scale/population/tui"
)

// Phase2Processor handles density mapping
type Phase2Processor struct {
	config      *population.PhaseOneConfig
	
	// Channels
	personQueue chan workers.PersonXML
	coordQueue  chan workers.Coordinates
	doneQueue   chan workers.CompletionSignal
	
	// Workers
	readers    []*workers.ReaderWorker
	extractors []*workers.ExtractorWorker
	mappers    []*workers.MapperWorker
	
	// Progress tracking
	startTime      time.Time
	lastUpdateTime time.Time
	tuiEnabled     bool
	
	// State
	isRunning bool
	wg        sync.WaitGroup
}

// NewPhase2Processor creates a new Phase 2 processor
func NewPhase2Processor(config *population.PhaseOneConfig) *Phase2Processor {
	queueConfig := workers.DefaultQueueConfig()
	
	return &Phase2Processor{
		config:      config,
		personQueue: make(chan workers.PersonXML, queueConfig.PersonQueueSize),
		coordQueue:  make(chan workers.Coordinates, queueConfig.CoordQueueSize),
		doneQueue:   make(chan workers.CompletionSignal, queueConfig.DoneQueueSize),
		tuiEnabled:  true,
	}
}

// DisableTUI disables the terminal UI
func (p2p *Phase2Processor) DisableTUI() {
	p2p.tuiEnabled = false
}

// Process runs Phase 2: Density Mapping
func (p2p *Phase2Processor) Process(inputFile string, safePoints []int64) error {
	p2p.startTime = time.Now()
	p2p.isRunning = true
	
	if p2p.tuiEnabled {
		tui.UpdatePhase(2, "Starting Reader Workers")
	}
	
	// Step 1: Create workers based on configuration
	if err := p2p.createWorkers(inputFile, safePoints); err != nil {
		return fmt.Errorf("failed to create workers: %w", err)
	}
	
	// Step 2: Start all workers
	p2p.startWorkers()
	
	// Step 3: Wait for completion
	p2p.wg.Wait()
	
	p2p.isRunning = false
	
	if p2p.tuiEnabled {
		tui.UpdatePhase(2, "Mapping Complete")
		time.Sleep(1 * time.Second)
	}
	
	return nil
}

// createWorkers creates all worker instances
func (p2p *Phase2Processor) createWorkers(inputFile string, safePoints []int64) error {
	// Create multiple readers based on safe points
	p2p.readers = p2p.createMultipleReaders(inputFile, safePoints)
	
	// Set progress callbacks
	for _, reader := range p2p.readers {
		reader.SetProgressCallback(p2p.updateProgress)
	}
	
	// Create extractors
	p2p.extractors = make([]*workers.ExtractorWorker, p2p.config.ExtractorCount)
	for i := 0; i < p2p.config.ExtractorCount; i++ {
		p2p.extractors[i] = workers.NewExtractorWorker(
			fmt.Sprintf("extractor-%d", i+1),
			p2p.personQueue,
			p2p.coordQueue,
		)
	}
	
	// Create mappers
	p2p.mappers = make([]*workers.MapperWorker, p2p.config.HashMapCount)
	for i := 0; i < p2p.config.HashMapCount; i++ {
		p2p.mappers[i] = workers.NewMapperWorker(
			fmt.Sprintf("mapper-%d", i+1),
			p2p.coordQueue,
			p2p.doneQueue,
		)
	}
	
	return nil
}

// createMultipleReaders creates multiple parallel readers
func (p2p *Phase2Processor) createMultipleReaders(inputFile string, safePoints []int64) []*workers.ReaderWorker {
	readers := make([]*workers.ReaderWorker, 0)
	
	// Use configured number of readers
	numReaders := p2p.config.ReaderCount
	
	// Calculate work ranges from safe points
	workRanges := make([]population.WorkRange, 0)
	for i := 0; i < len(safePoints)-1; i++ {
		workRanges = append(workRanges, population.WorkRange{
			StartLine: safePoints[i],
			EndLine:   safePoints[i+1] - 1,
			ChunkID:   fmt.Sprintf("chunk_%d", i+1),
		})
	}
	
	// Add final range
	if len(safePoints) > 0 {
		workRanges = append(workRanges, population.WorkRange{
			StartLine: safePoints[len(safePoints)-1],
			EndLine:   -1, // EOF
			ChunkID:   fmt.Sprintf("chunk_%d", len(safePoints)),
		})
	}
	
	// Adjust number of readers based on available work ranges
	// For small files with few work ranges, we don't need multiple readers
	actualReaders := numReaders
	if len(workRanges) < numReaders {
		actualReaders = len(workRanges)
	}
	
	// Create readers based on actual need
	for i := 0; i < actualReaders; i++ {
		startIdx := i * len(workRanges) / actualReaders
		endIdx := (i + 1) * len(workRanges) / actualReaders
		if i == actualReaders-1 {
			endIdx = len(workRanges)
		}
		
		// Handle single work range case
		if startIdx >= len(workRanges) {
			break
		}
		
		startLine := workRanges[startIdx].StartLine
		endLine := workRanges[endIdx-1].EndLine
		
		readers = append(readers, workers.NewReaderWorkerWithRange(
			fmt.Sprintf("reader-%d", i+1),
			inputFile,
			p2p.personQueue,
			startLine,
			endLine,
		))
	}
	
	return readers
}

// startWorkers starts all workers
func (p2p *Phase2Processor) startWorkers() {
	// Start readers
	readersWg := sync.WaitGroup{}
	for _, reader := range p2p.readers {
		readersWg.Add(1)
		p2p.wg.Add(1)
		go func(r *workers.ReaderWorker) {
			defer p2p.wg.Done()
			defer readersWg.Done()
			r.Process()
		}(reader)
	}
	
	// Close personQueue when all readers finish
	go func() {
		readersWg.Wait()
		close(p2p.personQueue)
	}()
	
	// Start extractors
	extractorsWg := sync.WaitGroup{}
	for _, extractor := range p2p.extractors {
		extractorsWg.Add(1)
		p2p.wg.Add(1)
		go func(e *workers.ExtractorWorker) {
			defer p2p.wg.Done()
			defer extractorsWg.Done()
			e.Process()
		}(extractor)
	}
	
	// Close coordQueue when all extractors finish
	go func() {
		extractorsWg.Wait()
		close(p2p.coordQueue)
	}()
	
	// Start mappers
	for _, mapper := range p2p.mappers {
		p2p.wg.Add(1)
		go func(m *workers.MapperWorker) {
			defer p2p.wg.Done()
			m.Process()
		}(mapper)
	}
	
	// Monitor completion
	p2p.wg.Add(1)
	go p2p.monitorCompletion()
}

// monitorCompletion monitors when to close channels
func (p2p *Phase2Processor) monitorCompletion() {
	defer p2p.wg.Done()
	
	// Wait for all mappers to complete
	completedMappers := 0
	for completedMappers < len(p2p.mappers) {
		<-p2p.doneQueue
		completedMappers++
	}
	close(p2p.doneQueue)
}

// updateProgress updates progress in TUI with throttling
func (p2p *Phase2Processor) updateProgress(lines, persons int64) {
	// Throttle updates to prevent flooding - standardized to 250ms to match TUI system
	now := time.Now()
	if p2p.lastUpdateTime == (time.Time{}) {
		p2p.lastUpdateTime = now
	}
	
	if now.Sub(p2p.lastUpdateTime) < 100*time.Millisecond {
		return // Skip this update to reduce twitching
	}
	p2p.lastUpdateTime = now
	
	// Always track workers for real-time monitoring (independent of TUI display)
	var totalLines, totalPersons int64
	for _, reader := range p2p.readers {
		l, p := reader.GetMetrics()
		totalLines += l
		totalPersons += p
	}
	
	// Update workers for tracking system
	p2p.updateUnifiedWorkers(totalLines, totalPersons)
}

// updateUnifiedWorkers updates workers in the TUI with real-time tracking
func (p2p *Phase2Processor) updateUnifiedWorkers(totalLines, totalPersons int64) {
	// Send real-time updates for reader workers
	for _, reader := range p2p.readers {
		lines, persons := reader.GetMetrics()
		status := tui.StatusActive
		if !p2p.isRunning {
			status = tui.StatusCompleted
		}
		
		tui.TrackWorkerUpdate(tui.WorkerUpdate{
			WorkerID:     reader.GetID(),
			WorkerType:   tui.TypeReader,
			Status:       status,
			LinesRead:    lines,
			PersonsFound: persons,
			Timestamp:    time.Now(),
		})
	}
	
	// Send real-time updates for extractor workers
	for i, extractor := range p2p.extractors {
		processed, _ := extractor.GetMetrics()
		status := tui.StatusActive
		if !p2p.isRunning {
			status = tui.StatusCompleted
		}
		
		tui.TrackWorkerUpdate(tui.WorkerUpdate{
			WorkerID:       fmt.Sprintf("Extractor-%d", i+1),
			WorkerType:     tui.TypeExtractor,
			Status:         status,
			ProcessedCount: processed,
			PersonsSkipped: 0, // No skip data available yet
			Timestamp:      time.Now(),
		})
	}
	
	// Send real-time updates for mapper workers
	for i, mapper := range p2p.mappers {
		mapped, bins := mapper.GetMetrics()
		status := tui.StatusActive
		if !p2p.isRunning {
			status = tui.StatusCompleted
		}
		
		tui.TrackWorkerUpdate(tui.WorkerUpdate{
			WorkerID:        fmt.Sprintf("Mapper-%d", i+1),
			WorkerType:      tui.TypeMapper,
			Status:          status,
			CoordsProcessed: mapped,
			BinCount:        bins,
			Timestamp:       time.Now(),
		})
	}
}

// GetMappers returns the mapper workers for grid access
func (p2p *Phase2Processor) GetMappers() []*workers.MapperWorker {
	return p2p.mappers
}