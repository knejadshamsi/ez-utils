package processing

import (
	"fmt"
	"sync"
	"time"

	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/tui"
	"ez-utils/src/scale/population/workers"
)

// Buffer size constants
const (
	PersonChannelBuffer         = 2000  // Buffer for reader->reducer communication
	SelectedPersonChannelBuffer = 10000 // Buffer for reducer->writer communication  
	DoneChannelBuffer          = 100   // Buffer for completion signals
)

// Phase4Processor handles population scaling and output
type Phase4Processor struct {
	// Configuration
	config    *population.PhaseTwoConfig
	inputFile string
	
	// Workers
	readers  []*workers.ReaderWorker
	reducers []*workers.ReducerWorker
	writers  map[int][]*workers.WriterWorker // per scale
	
	// Channels
	personChannel          chan workers.PersonXML
	selectedPersonChannels map[int]chan workers.SelectedPerson // per scale
	doneChannel           chan workers.CompletionSignal
	
	// Progress tracking
	startTime      time.Time
	lastUpdateTime time.Time
	tuiEnabled     bool
	
	// State
	isRunning bool
	wg        sync.WaitGroup
}

// NewPhase4Processor creates a new Phase 4 processor
func NewPhase4Processor(config *population.PhaseTwoConfig, inputFile string) *Phase4Processor {
	selectedPersonChannels := make(map[int]chan workers.SelectedPerson)
	writers := make(map[int][]*workers.WriterWorker)
	
	// Create channels for each scale
	for _, scale := range config.Scales {
		selectedPersonChannels[scale] = make(chan workers.SelectedPerson, SelectedPersonChannelBuffer)
		writers[scale] = make([]*workers.WriterWorker, 0)
	}
	
	return &Phase4Processor{
		config:                 config,
		inputFile:              inputFile,
		personChannel:          make(chan workers.PersonXML, PersonChannelBuffer),
		writers:                writers,
		selectedPersonChannels: selectedPersonChannels,
		doneChannel:           make(chan workers.CompletionSignal, DoneChannelBuffer),
		tuiEnabled:            true,
	}
}

// DisableTUI disables the terminal UI
func (p4p *Phase4Processor) DisableTUI() {
	p4p.tuiEnabled = false
}

// Process runs Phase 4: Population Scaling and Output
func (p4p *Phase4Processor) Process() error {
	p4p.startTime = time.Now()
	p4p.isRunning = true
	
	// Step 1: Get safe points from Phase 1
	safePoints := population.GetSafeCutPoints()
	if len(safePoints) == 0 {
		return fmt.Errorf("no safe points available from Phase 1")
	}
	
	// Step 2: Create work ranges for parallel processing
	workRanges := p4p.createWorkRanges(safePoints)
	if len(workRanges) == 0 {
		return fmt.Errorf("failed to create work ranges")
	}
	
	// Step 3: Create readers and workers
	if err := p4p.createReaders(workRanges); err != nil {
		return fmt.Errorf("failed to create readers: %w", err)
	}
	
	if err := p4p.createWorkers(workRanges); err != nil {
		return fmt.Errorf("failed to create workers: %w", err)
	}
	
	// Update TUI phase
	if p4p.tuiEnabled {
		tui.UpdatePhase(4, "Starting Reader Workers")
	}
	
	// Step 4: Start readers and workers
	p4p.startWorkers()
	
	// Step 5: Monitor progress
	go p4p.monitorProgress()
	
	// Step 6: Wait for completion
	p4p.wg.Wait()
	
	// Final progress update before cleanup
	p4p.updateProgress()
	time.Sleep(100 * time.Millisecond) // Let TUI process final update
	
	p4p.isRunning = false
	
	if p4p.tuiEnabled {
		tui.UpdatePhase(4, "Population Scaling Complete")
		time.Sleep(2 * time.Second)
	}
	
	return nil
}

// createWorkRanges creates work ranges from safe points
func (p4p *Phase4Processor) createWorkRanges(safePoints []int64) []population.WorkRange {
	workRanges := make([]population.WorkRange, 0)
	
	// Create ranges between consecutive safe points
	for i := 0; i < len(safePoints)-1; i++ {
		workRanges = append(workRanges, population.WorkRange{
			StartLine: safePoints[i],
			EndLine:   safePoints[i+1] - 1,
			ChunkID:   fmt.Sprintf("chunk_%d", i+1),
		})
	}
	
	// Add final range from last safe point to end of file
	if len(safePoints) > 0 {
		workRanges = append(workRanges, population.WorkRange{
			StartLine: safePoints[len(safePoints)-1],
			EndLine:   -1, // EOF
			ChunkID:   fmt.Sprintf("chunk_%d", len(safePoints)),
		})
	}
	
	return workRanges
}

// createReaders creates multiple reader workers (reusing Phase 1 architecture)
func (p4p *Phase4Processor) createReaders(workRanges []population.WorkRange) error {
	// Use same reader count as configured
	numReaders := p4p.config.ReducerCount // Reuse reducer count for reader count
	if len(workRanges) < numReaders {
		numReaders = len(workRanges)
	}
	
	p4p.readers = make([]*workers.ReaderWorker, numReaders)
	
	// Distribute work ranges across readers (like Phase 1)
	for i := 0; i < numReaders; i++ {
		startIdx := i * len(workRanges) / numReaders
		endIdx := (i + 1) * len(workRanges) / numReaders
		
		// Calculate line range for this reader from assigned work ranges
		var startLine, endLine int64 = -1, -1
		
		if startIdx < len(workRanges) {
			startLine = workRanges[startIdx].StartLine
			
			if endIdx > startIdx {
				if endIdx >= len(workRanges) {
					// Last reader goes to end of file
					endLine = -1
				} else {
					endLine = workRanges[endIdx-1].EndLine
				}
			}
		}
		
		if startLine != -1 {
			p4p.readers[i] = workers.NewReaderWorkerWithRange(
				fmt.Sprintf("reader-%d", i+1),
				p4p.inputFile,
				p4p.personChannel,
				startLine,
				endLine,
			)
		}
	}
	
	return nil
}

// createWorkers creates all reducer and writer workers
func (p4p *Phase4Processor) createWorkers(workRanges []population.WorkRange) error {
	// Create reducers based on configuration
	numReducers := p4p.config.ReducerCount
	
	// Adjust number of reducers based on available work ranges
	if len(workRanges) < numReducers {
		numReducers = len(workRanges)
	}
	
	p4p.reducers = make([]*workers.ReducerWorker, numReducers)
	
	// Create reducers that receive persons from reader workers
	for i := 0; i < numReducers; i++ {
		// Calculate which work ranges this reducer should handle (for tracking)
		startIdx := i * len(workRanges) / numReducers
		endIdx := (i + 1) * len(workRanges) / numReducers
		assignedRanges := workRanges[startIdx:endIdx]
		
		reducer := workers.NewReducerWorker(
			fmt.Sprintf("reducer-%d", i+1),
			i,
			p4p.personChannel, // All reducers read from same workers.PersonXML channel
			assignedRanges,
			p4p.config.Scales,
		)
		
		// Set output channels
		reducer.SetOutputChannels(p4p.selectedPersonChannels, p4p.doneChannel)
		p4p.reducers[i] = reducer
	}
	
	// Create writers for each scale
	for _, scale := range p4p.config.Scales {
		numWriters := p4p.config.FileWriterCount
		p4p.writers[scale] = make([]*workers.WriterWorker, numWriters)
		
		for i := 0; i < numWriters; i++ {
			writer := workers.NewWriterWorker(
				fmt.Sprintf("writer-scale-%d-%d", scale, i+1),
				i,
				scale,
				p4p.config.OutputDir,
				p4p.selectedPersonChannels[scale],
			)
			writer.SetDoneChannel(p4p.doneChannel)
			p4p.writers[scale][i] = writer
		}
	}
	
	return nil
}

// startWorkers starts all reader, reducer and writer workers
func (p4p *Phase4Processor) startWorkers() {
	// Start readers
	readerWg := sync.WaitGroup{}
	for _, reader := range p4p.readers {
		if reader != nil {
			readerWg.Add(1)
			p4p.wg.Add(1)
			go func(r *workers.ReaderWorker) {
				defer p4p.wg.Done()
				defer readerWg.Done()
				r.Process()
			}(reader)
		}
	}
	
	// Close personChannel when all readers are done
	go func() {
		readerWg.Wait()
		close(p4p.personChannel)
	}()
	
	// Start reducers
	reducerWg := sync.WaitGroup{}
	for _, reducer := range p4p.reducers {
		if reducer != nil {
			reducerWg.Add(1)
			p4p.wg.Add(1)
			go func(r *workers.ReducerWorker) {
				defer p4p.wg.Done()
				defer reducerWg.Done()
				r.Process()
			}(reducer)
		}
	}
	
	// Close selectedPersonChannels when all reducers are done
	go func() {
		reducerWg.Wait()
		for _, ch := range p4p.selectedPersonChannels {
			close(ch)
		}
	}()
	
	// Start writers for each scale
	for scale, writers := range p4p.writers {
		for _, writer := range writers {
			if writer != nil {
				p4p.wg.Add(1)
				go func(w *workers.WriterWorker, s int) {
					defer p4p.wg.Done()
					w.Process()
				}(writer, scale)
			}
		}
	}
}

// monitorProgress monitors worker progress and updates TUI
func (p4p *Phase4Processor) monitorProgress() {
	ticker := time.NewTicker(100 * time.Millisecond) // Faster refresh for responsive UI
	defer ticker.Stop()
	
	for p4p.isRunning {
		select {
		case <-ticker.C:
			p4p.updateProgress()
		case <-p4p.doneChannel:
			// Worker completed, but we'll let the timer handle updates to avoid flooding
			// Removed throttledUpdate() call to prevent multiple update paths
		}
	}
}

// updateProgress updates the TUI with current progress
func (p4p *Phase4Processor) updateProgress() {
	// Always track workers for real-time monitoring (independent of TUI display)
	
	totalProcessed := int64(0)
	totalSelected := make(map[int]int64)
	totalWritten := make(map[int]int64)
	
	// Collect metrics from reducers
	for _, reducer := range p4p.reducers {
		if reducer != nil {
			processed, selected := reducer.GetMetrics()
			totalProcessed += processed
			
			for scale, count := range selected {
				totalSelected[scale] += count
			}
		}
	}
	
	// Collect metrics from writers
	for scale, scaleWriters := range p4p.writers {
		for _, writer := range scaleWriters {
			if writer != nil {
				written := writer.GetMetrics()
				totalWritten[scale] += written
			}
		}
	}
	
	// Send update to TUI
	p4p.updateUnifiedWorkers(totalProcessed, totalSelected, totalWritten)
}


// updateUnifiedWorkers updates workers in the TUI for Phase 4 with real-time tracking
func (p4p *Phase4Processor) updateUnifiedWorkers(totalProcessed int64, totalSelected, totalWritten map[int]int64) {
	// Send real-time updates for reader workers
	for _, reader := range p4p.readers {
		if reader != nil {
			lines, persons := reader.GetMetrics()
			status := tui.StatusActive
			if !p4p.isRunning {
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
	}
	
	// Send real-time updates for reducer workers
	for i, reducer := range p4p.reducers {
		if reducer != nil {
			processed, selectedMap := reducer.GetMetrics()
			status := tui.StatusActive
			if !p4p.isRunning {
				status = tui.StatusCompleted
			}
			
			// Sum up selected from all scales for display
			totalSelectedForWorker := int64(0)
			for _, count := range selectedMap {
				totalSelectedForWorker += count
			}
			
			tui.TrackWorkerUpdate(tui.WorkerUpdate{
				WorkerID:        fmt.Sprintf("Reducer-%d", i+1),
				WorkerType:      tui.TypeReducer,
				Status:          status,
				ProcessedCount:  processed,
				PersonsSelected: totalSelectedForWorker,
				Timestamp:       time.Now(),
			})
		}
	}
	
	// Send real-time updates for writer workers
	for scale, writers := range p4p.writers {
		for i, writer := range writers {
			if writer != nil {
				written := writer.GetMetrics() // Get actual written count from writer
				status := tui.StatusActive
				if !p4p.isRunning {
					status = tui.StatusCompleted
				}
				
				tui.TrackWorkerUpdate(tui.WorkerUpdate{
					WorkerID:       fmt.Sprintf("Writer-Scale-%d-%d", scale, i+1),
					WorkerType:     tui.TypeWriter,
					Status:         status,
					PersonsWritten: written,
					Timestamp:      time.Now(),
				})
			}
		}
	}
}

// startTUI initializes and starts the TUI

// GetProgress returns processing summary for display
func (p4p *Phase4Processor) GetProgress() (int64, map[int]int64, map[int]int64) {
	totalProcessed := int64(0)
	totalSelected := make(map[int]int64)
	totalWritten := make(map[int]int64)
	
	// Collect metrics from reducers
	for _, reducer := range p4p.reducers {
		if reducer != nil {
			processed, selected := reducer.GetMetrics()
			totalProcessed += processed
			
			for scale, count := range selected {
				totalSelected[scale] += count
			}
		}
	}
	
	// Collect metrics from writers
	for scale, scaleWriters := range p4p.writers {
		for _, writer := range scaleWriters {
			if writer != nil {
				written := writer.GetMetrics()
				totalWritten[scale] += written
			}
		}
	}
	
	return totalProcessed, totalSelected, totalWritten
}