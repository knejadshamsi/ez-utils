package processing

import (
	"fmt"
	"sync"
	"time"

	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/functions"
	"ez-utils/src/scale/population/workers"
	"ez-utils/src/scale/population/tui"

	tea "github.com/charmbracelet/bubbletea"
)

// PopulationProcessor handles the complete population processing pipeline
// This replaces all the over-engineered managers and orchestrators
type PopulationProcessor struct {
	// Configuration
	config *population.PhaseOneConfig
	
	// Services
	safePointService functions.SafePointService
	
	// Channels
	personQueue chan workers.PersonXML
	coordQueue  chan workers.Coordinates
	doneQueue   chan workers.CompletionSignal
	
	// Workers
	readers    []*workers.ReaderWorker
	extractors []*workers.ExtractorWorker
	mappers    []*workers.MapperWorker
	
	// Progress tracking
	startTime    time.Time
	tuiProgram   *tea.Program
	tuiEnabled   bool
	
	// State
	isRunning bool
	wg        sync.WaitGroup
}

// NewPopulationProcessor creates a new population processor
func NewPopulationProcessor(config *population.PhaseOneConfig) *PopulationProcessor {
	queueConfig := workers.DefaultQueueConfig()
	
	return &PopulationProcessor{
		config:           config,
		safePointService: functions.NewSafePointService(),
		personQueue:      make(chan workers.PersonXML, queueConfig.PersonQueueSize),
		coordQueue:       make(chan workers.Coordinates, queueConfig.CoordQueueSize),
		doneQueue:        make(chan workers.CompletionSignal, queueConfig.DoneQueueSize),
		tuiEnabled:       true,
	}
}

// DisableTUI disables the terminal UI
func (pp *PopulationProcessor) DisableTUI() {
	pp.tuiEnabled = false
}

// Process runs the complete population processing pipeline
func (pp *PopulationProcessor) Process(inputFile string) error {
	pp.startTime = time.Now()
	pp.isRunning = true
	
	// Initialize data structures
	population.InitializeGlobalData()
	pp.initializeGrid()
	
	// Start TUI if enabled
	if pp.tuiEnabled {
		if err := pp.startTUI(); err != nil {
			return fmt.Errorf("failed to start TUI: %w", err)
		}
	}
	
	// Step 1: Discover safe points for parallel processing
	safePoints, err := pp.discoverSafePoints(inputFile)
	if err != nil {
		return fmt.Errorf("safe point discovery failed: %w", err)
	}
	
	// Step 2: Create workers based on configuration
	if err := pp.createWorkers(inputFile, safePoints); err != nil {
		return fmt.Errorf("failed to create workers: %w", err)
	}
	
	// Step 3: Start all workers
	pp.startWorkers()
	
	// Step 4: Wait for completion
	pp.wg.Wait()
	
	// Step 5: Post-processing
	if pp.tuiEnabled && pp.tuiProgram != nil {
		tui.UpdateStatus(pp.tuiProgram, "Phase 1 completed!\n\nCombining grids...", true)
		time.Sleep(1 * time.Second)
	}
	
	if err := pp.postProcess(); err != nil {
		return fmt.Errorf("post-processing failed: %w", err)
	}
	
	pp.isRunning = false
	
	if pp.tuiEnabled && pp.tuiProgram != nil {
		tui.UpdateStatus(pp.tuiProgram, "Processing completed successfully!", true)
		time.Sleep(2 * time.Second)
		pp.tuiProgram.Send(tea.Quit())
		time.Sleep(100 * time.Millisecond) // Give TUI time to exit cleanly
	}
	
	elapsed := time.Since(pp.startTime)
	fmt.Printf("\nPhase One completed in %s\n", elapsed.Round(time.Second))
	
	return nil
}

// discoverSafePoints finds safe cut points in the XML file
func (pp *PopulationProcessor) discoverSafePoints(inputFile string) ([]int64, error) {
	var progressCallback functions.ProgressCallback
	
	if pp.tuiEnabled && pp.tuiProgram != nil {
		tui.UpdatePhase(pp.tuiProgram, "Phase 0", 0, 0)
		progressCallback = func(lines int64, points int) {
			tui.UpdatePhase(pp.tuiProgram, "Phase 0", points, lines)
		}
	}
	
	result, err := pp.safePointService.DiscoverSafePoints(
		inputFile, 
		pp.config.SafePointInterval, 
		progressCallback,
	)
	if err != nil {
		return nil, err
	}
	
	// Store for Phase 2 access
	population.SetSafeCutPoints(result.SafePoints)
	
	if pp.tuiEnabled && pp.tuiProgram != nil {
		tui.UpdatePhase(pp.tuiProgram, "Phase 1", len(result.SafePoints), result.TotalLines)
	}
	
	return result.SafePoints, nil
}

// createWorkers creates all worker instances
func (pp *PopulationProcessor) createWorkers(inputFile string, safePoints []int64) error {
	// Create multiple readers based on safe points
	pp.readers = pp.createMultipleReaders(inputFile, safePoints)
	
	// Set progress callbacks
	for _, reader := range pp.readers {
		reader.SetProgressCallback(pp.updateProgress)
	}
	
	// Create extractors
	pp.extractors = make([]*workers.ExtractorWorker, pp.config.ExtractorCount)
	for i := 0; i < pp.config.ExtractorCount; i++ {
		pp.extractors[i] = workers.NewExtractorWorker(
			fmt.Sprintf("extractor-%d", i+1),
			pp.personQueue,
			pp.coordQueue,
		)
	}
	
	// Create mappers
	pp.mappers = make([]*workers.MapperWorker, pp.config.HashMapCount)
	for i := 0; i < pp.config.HashMapCount; i++ {
		pp.mappers[i] = workers.NewMapperWorker(
			fmt.Sprintf("mapper-%d", i+1),
			pp.coordQueue,
			pp.doneQueue,
		)
	}
	
	return nil
}

// createMultipleReaders creates multiple parallel readers
func (pp *PopulationProcessor) createMultipleReaders(inputFile string, safePoints []int64) []*workers.ReaderWorker {
	readers := make([]*workers.ReaderWorker, 0)
	
	// Use configured number of readers
	numReaders := pp.config.ReaderCount
	
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
			pp.personQueue,
			startLine,
			endLine,
		))
	}
	
	return readers
}

// startWorkers starts all workers
func (pp *PopulationProcessor) startWorkers() {
	// Start readers
	readersWg := sync.WaitGroup{}
	for _, reader := range pp.readers {
		readersWg.Add(1)
		pp.wg.Add(1)
		go func(r *workers.ReaderWorker) {
			defer pp.wg.Done()
			defer readersWg.Done()
			r.Process()
		}(reader)
	}
	
	// Close personQueue when all readers finish
	go func() {
		readersWg.Wait()
		close(pp.personQueue)
	}()
	
	// Start extractors
	extractorsWg := sync.WaitGroup{}
	for _, extractor := range pp.extractors {
		extractorsWg.Add(1)
		pp.wg.Add(1)
		go func(e *workers.ExtractorWorker) {
			defer pp.wg.Done()
			defer extractorsWg.Done()
			e.Process()
		}(extractor)
	}
	
	// Close coordQueue when all extractors finish
	go func() {
		extractorsWg.Wait()
		close(pp.coordQueue)
	}()
	
	// Start mappers
	for _, mapper := range pp.mappers {
		pp.wg.Add(1)
		go func(m *workers.MapperWorker) {
			defer pp.wg.Done()
			m.Process()
		}(mapper)
	}
	
	// Monitor completion
	pp.wg.Add(1)
	go pp.monitorCompletion()
}

// monitorCompletion monitors when to close channels
func (pp *PopulationProcessor) monitorCompletion() {
	defer pp.wg.Done()
	
	// Wait for all mappers to complete
	completedMappers := 0
	for completedMappers < len(pp.mappers) {
		<-pp.doneQueue
		completedMappers++
	}
	close(pp.doneQueue)
}

// postProcess performs grid combination and retention calculation
func (pp *PopulationProcessor) postProcess() error {
	// Combine mapper grids
	finalGrid := make(population.DensityMap)
	for _, mapper := range pp.mappers {
		for binID, count := range mapper.GetLocalGrid() {
			if count > 0 {
				finalGrid[binID] += count
			}
		}
	}
	population.SetDensityMap(finalGrid)
	
	// Calculate retention probabilities
	retentionMaps, err := functions.CalculateRetentionProbabilities(finalGrid)
	if err != nil {
		return err
	}
	
	for scale, retentionMap := range retentionMaps {
		population.SetRetentionMap(scale, retentionMap)
	}
	
	return nil
}

// updateProgress updates progress in TUI
func (pp *PopulationProcessor) updateProgress(lines, persons int64) {
	if pp.tuiEnabled && pp.tuiProgram != nil {
		// Aggregate totals from ALL readers
		var totalLines, totalPersons int64
		for _, reader := range pp.readers {
			l, p := reader.GetMetrics()
			totalLines += l
			totalPersons += p
		}
		
		workers := pp.buildWorkerStates()
		tui.UpdateProgress(pp.tuiProgram, tui.ProgressMsg{
			LinesRead:    totalLines,
			PersonsFound: totalPersons,
			Workers:      workers,
		})
	}
}

// buildWorkerStates builds current worker states for TUI
func (pp *PopulationProcessor) buildWorkerStates() []tui.WorkerState {
	states := make([]tui.WorkerState, 0)
	
	// Add readers
	for _, reader := range pp.readers {
		lines, persons := reader.GetMetrics()
		status := "Reading"
		if !pp.isRunning {
			status = "Completed"
		}
		states = append(states, tui.WorkerState{
			Name:      reader.GetID(),
			Status:    status,
			Extracted: persons,
		})
		_ = lines // Use if needed
	}
	
	// Add extractors
	for i, extractor := range pp.extractors {
		processed, _ := extractor.GetMetrics()
		states = append(states, tui.WorkerState{
			Name:      fmt.Sprintf("Extractor-%d", i+1),
			Status:    "Extracting",
			Extracted: processed,
		})
	}
	
	// Add mappers
	for i, mapper := range pp.mappers {
		mapped, bins := mapper.GetMetrics()
		states = append(states, tui.WorkerState{
			Name:      fmt.Sprintf("Mapper-%d", i+1),
			Status:    "Mapping",
			Extracted: mapped,
			BinCount:  bins,
		})
	}
	
	return states
}

// startTUI initializes and starts the TUI
func (pp *PopulationProcessor) startTUI() error {
	simpleTUI := tui.NewSimpleTUI()
	pp.tuiProgram = tea.NewProgram(simpleTUI)
	
	go func() {
		if _, err := pp.tuiProgram.Run(); err != nil {
			fmt.Printf("TUI error: %v\n", err)
		}
	}()
	
	time.Sleep(100 * time.Millisecond) // Let TUI initialize
	return nil
}

// initializeGrid sets up the grid bounds
func (pp *PopulationProcessor) initializeGrid() {
	bounds := pp.config.GridConfig.InitialBounds
	if bounds.CellSize == 0 {
		bounds = population.GridBounds{
			MinX:     -1000000,
			MaxX:     1000000,
			MinY:     -1000000,
			MaxY:     1000000,
			CellSize: 1000,
		}
	}
	population.SetGlobalGrid(bounds)
}

// GetProgress returns current progress
func (pp *PopulationProcessor) GetProgress() *population.PhaseOneProgress {
	var linesRead, personsFound int64
	
	for _, reader := range pp.readers {
		lines, persons := reader.GetMetrics()
		linesRead += lines
		personsFound += persons
	}
	
	var extracted, mapped int64
	for _, e := range pp.extractors {
		p, _ := e.GetMetrics()
		extracted += p
	}
	
	for _, m := range pp.mappers {
		p, _ := m.GetMetrics()
		mapped += p
	}
	
	return &population.PhaseOneProgress{
		LinesRead:         linesRead,
		PersonsExtracted:  extracted,
		CoordinatesMapped: mapped,
		ElapsedTime:       time.Since(pp.startTime),
		IsComplete:        !pp.isRunning,
	}
}