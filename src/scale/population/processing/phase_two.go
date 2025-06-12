package processing

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/tui"
	"ez-utils/src/scale/population/workers"

	tea "github.com/charmbracelet/bubbletea"
)

// PhaseTwoProcessor handles the complete Phase 2 parallel processing pipeline
type PhaseTwoProcessor struct {
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
	startTime  time.Time
	tuiProgram *tea.Program
	tuiEnabled bool
	
	// State
	isRunning bool
	wg        sync.WaitGroup
}

// NewPhaseTwoProcessor creates a new Phase 2 processor
func NewPhaseTwoProcessor(config *population.PhaseTwoConfig, inputFile string) *PhaseTwoProcessor {
	selectedPersonChannels := make(map[int]chan workers.SelectedPerson)
	writers := make(map[int][]*workers.WriterWorker)
	
	// Create channels for each scale
	for _, scale := range config.Scales {
		selectedPersonChannels[scale] = make(chan workers.SelectedPerson, 10000) // Larger buffer to prevent blocking
		writers[scale] = make([]*workers.WriterWorker, 0)
	}
	
	return &PhaseTwoProcessor{
		config:                 config,
		inputFile:              inputFile,
		personChannel:          make(chan workers.PersonXML, 2000), // Buffered for reader->reducer communication
		writers:                writers,
		selectedPersonChannels: selectedPersonChannels,
		doneChannel:           make(chan workers.CompletionSignal, 100),
		tuiEnabled:            true,
	}
}

// DisableTUI disables the terminal UI
func (ptp *PhaseTwoProcessor) DisableTUI() {
	ptp.tuiEnabled = false
}

// Process runs the complete Phase 2 processing pipeline
func (ptp *PhaseTwoProcessor) Process() error {
	ptp.startTime = time.Now()
	ptp.isRunning = true
	
	// Step 1: Get safe points from Phase 1
	safePoints := population.GetSafeCutPoints()
	if len(safePoints) == 0 {
		return fmt.Errorf("no safe points available from Phase 1")
	}
	
	// Step 2: Create work ranges for parallel processing
	workRanges := ptp.createWorkRanges(safePoints)
	if len(workRanges) == 0 {
		return fmt.Errorf("failed to create work ranges")
	}
	
	// Step 3: Create readers and workers
	if err := ptp.createReaders(workRanges); err != nil {
		return fmt.Errorf("failed to create readers: %w", err)
	}
	
	if err := ptp.createWorkers(workRanges); err != nil {
		return fmt.Errorf("failed to create workers: %w", err)
	}
	
	// Start TUI after workers are created
	if ptp.tuiEnabled {
		if err := ptp.startTUI(); err != nil {
			return fmt.Errorf("failed to start TUI: %w", err)
		}
	}
	
	// Step 4: Start readers and workers
	ptp.startWorkers()
	
	// Step 5: Monitor progress
	go ptp.monitorProgress()
	
	// Step 6: Wait for completion
	ptp.wg.Wait()
	
	// Final progress update before cleanup
	ptp.updateProgress()
	time.Sleep(100 * time.Millisecond) // Let TUI process final update
	
	// Step 7: Combine part files into single files per scale
	if err := ptp.combinePartFiles(); err != nil {
		fmt.Printf("Warning: Failed to combine part files: %v\n", err)
	}
	
	// Step 8: Cleanup
	ptp.cleanup()
	
	ptp.isRunning = false
	
	elapsed := time.Since(ptp.startTime)
	fmt.Printf("\nPhase Two completed in %s\n", elapsed.Round(time.Second))
	ptp.printSummary()
	
	if ptp.tuiEnabled && ptp.tuiProgram != nil {
		tui.UpdatePhaseTwoStatus("Phase Two completed successfully!", true)
		time.Sleep(2 * time.Second)
		tui.SignalPhaseTwoComplete()
	}
	
	return nil
}

// createWorkRanges creates work ranges from safe points
func (ptp *PhaseTwoProcessor) createWorkRanges(safePoints []int64) []population.WorkRange {
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
func (ptp *PhaseTwoProcessor) createReaders(workRanges []population.WorkRange) error {
	// Use same reader count as configured
	numReaders := ptp.config.ReducerCount // Reuse reducer count for reader count
	if len(workRanges) < numReaders {
		numReaders = len(workRanges)
	}
	
	ptp.readers = make([]*workers.ReaderWorker, numReaders)
	
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
			ptp.readers[i] = workers.NewReaderWorkerWithRange(
				fmt.Sprintf("reader-%d", i+1),
				ptp.inputFile,
				ptp.personChannel,
				startLine,
				endLine,
			)
		}
	}
	
	return nil
}

// createWorkers creates all reducer and writer workers
func (ptp *PhaseTwoProcessor) createWorkers(workRanges []population.WorkRange) error {
	// Create reducers based on configuration
	numReducers := ptp.config.ReducerCount
	
	// Adjust number of reducers based on available work ranges
	if len(workRanges) < numReducers {
		numReducers = len(workRanges)
	}
	
	ptp.reducers = make([]*workers.ReducerWorker, numReducers)
	
	// Create reducers that receive persons from reader workers
	for i := 0; i < numReducers; i++ {
		// Calculate which work ranges this reducer should handle (for tracking)
		startIdx := i * len(workRanges) / numReducers
		endIdx := (i + 1) * len(workRanges) / numReducers
		assignedRanges := workRanges[startIdx:endIdx]
		
		reducer := workers.NewReducerWorker(
			fmt.Sprintf("reducer-%d", i+1),
			i,
			ptp.personChannel, // All reducers read from same workers.PersonXML channel
			assignedRanges,
			ptp.config.Scales,
		)
		
		// Set output channels
		reducer.SetOutputChannels(ptp.selectedPersonChannels, ptp.doneChannel)
		ptp.reducers[i] = reducer
	}
	
	// Create writers for each scale
	for _, scale := range ptp.config.Scales {
		numWriters := ptp.config.FileWriterCount
		ptp.writers[scale] = make([]*workers.WriterWorker, numWriters)
		
		for i := 0; i < numWriters; i++ {
			writer := workers.NewWriterWorker(
				fmt.Sprintf("writer-scale-%d-%d", scale, i+1),
				i,
				scale,
				ptp.config.OutputDir,
				ptp.selectedPersonChannels[scale],
			)
			writer.SetDoneChannel(ptp.doneChannel)
			ptp.writers[scale][i] = writer
		}
	}
	
	return nil
}

// startWorkers starts all reader, reducer and writer workers
func (ptp *PhaseTwoProcessor) startWorkers() {
	// Start readers
	readerWg := sync.WaitGroup{}
	for _, reader := range ptp.readers {
		if reader != nil {
			readerWg.Add(1)
			ptp.wg.Add(1)
			go func(r *workers.ReaderWorker) {
				defer ptp.wg.Done()
				defer readerWg.Done()
				r.Process()
			}(reader)
		}
	}
	
	// Close personChannel when all readers are done
	go func() {
		readerWg.Wait()
		close(ptp.personChannel)
	}()
	
	// Start reducers
	reducerWg := sync.WaitGroup{}
	for _, reducer := range ptp.reducers {
		if reducer != nil {
			reducerWg.Add(1)
			ptp.wg.Add(1)
			go func(r *workers.ReducerWorker) {
				defer ptp.wg.Done()
				defer reducerWg.Done()
				r.Process()
			}(reducer)
		}
	}
	
	// Close selectedPersonChannels when all reducers are done
	go func() {
		reducerWg.Wait()
		for _, ch := range ptp.selectedPersonChannels {
			close(ch)
		}
	}()
	
	// Start writers for each scale
	for scale, writers := range ptp.writers {
		for _, writer := range writers {
			if writer != nil {
				ptp.wg.Add(1)
				go func(w *workers.WriterWorker, s int) {
					defer ptp.wg.Done()
					w.Process()
				}(writer, scale)
			}
		}
	}
}

// monitorProgress monitors worker progress and updates TUI
func (ptp *PhaseTwoProcessor) monitorProgress() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for ptp.isRunning {
		select {
		case <-ticker.C:
			ptp.updateProgress()
		case <-ptp.doneChannel:
			// Worker completed, update progress
			ptp.updateProgress()
		}
	}
}

// updateProgress updates the TUI with current progress
func (ptp *PhaseTwoProcessor) updateProgress() {
	if !ptp.tuiEnabled {
		return
	}
	
	// Collect reducer states
	reducers := make([]tui.ReducerWorkerState, 0)
	totalProcessed := int64(0)
	totalSelected := make(map[int]int64)
	
	for _, reducer := range ptp.reducers {
		if reducer != nil {
			processed, selected := reducer.GetMetrics()
			totalProcessed += processed
			
			for scale, count := range selected {
				totalSelected[scale] += count
			}
			
			reducers = append(reducers, tui.ReducerWorkerState{
				Name:             reducer.GetID(),
				Status:           "Processing",
				PersonsProcessed: processed,
				PersonsSelected:  selected,
			})
		}
	}
	
	// Collect writer states
	writers := make(map[int][]tui.WriterWorkerState)
	totalWritten := make(map[int]int64)
	
	for scale, scaleWriters := range ptp.writers {
		writers[scale] = make([]tui.WriterWorkerState, 0)
		
		for _, writer := range scaleWriters {
			if writer != nil {
				written := writer.GetMetrics()
				totalWritten[scale] += written
				
				writers[scale] = append(writers[scale], tui.WriterWorkerState{
					Name:           writer.GetID(),
					Status:         writer.GetStatus(),
					PersonsWritten: written,
					Scale:          scale,
					OutputFile:     writer.GetOutputFile(),
				})
			}
		}
	}
	
	// Send update to TUI
	tui.UpdatePhaseTwoProgress(tui.PhaseTwoProgressMsg{
		PersonsProcessed: totalProcessed,
		PersonsSelected:  totalSelected,
		PersonsWritten:   totalWritten,
		Reducers:         reducers,
		Writers:          writers,
		LastAction:       "Processing chunks in parallel",
	})
}

// startTUI initializes and starts the TUI
func (ptp *PhaseTwoProcessor) startTUI() error {
	numReducers := len(ptp.reducers)
	tui.InitPhaseTwoTUI(ptp.config.Scales, numReducers)
	
	// Allow TUI to fully initialize
	time.Sleep(100 * time.Millisecond)
	
	tui.UpdatePhaseTwoStatus("Starting Phase Two parallel processing...", true)
	time.Sleep(500 * time.Millisecond)
	tui.UpdatePhaseTwoStatus("", false)
	
	// Give TUI time to clear status and prepare for progress updates
	time.Sleep(100 * time.Millisecond)
	return nil
}

// cleanup performs post-processing cleanup
func (ptp *PhaseTwoProcessor) cleanup() {
	// Close done channel (selectedPersonChannels are closed in startWorkers)
	close(ptp.doneChannel)
	
	// Additional cleanup can be added here (temporary files, etc.)
}

// printSummary prints processing results
func (ptp *PhaseTwoProcessor) printSummary() {
	fmt.Println("\n=== Phase Two Summary ===")
	
	totalProcessed := int64(0)
	for _, reducer := range ptp.reducers {
		if reducer != nil {
			processed, _ := reducer.GetMetrics()
			totalProcessed += processed
		}
	}
	
	fmt.Printf("Total persons processed: %d\n", totalProcessed)
	
	for _, scale := range ptp.config.Scales {
		totalSelected := int64(0)
		totalWritten := int64(0)
		
		// Count selected from reducers
		for _, reducer := range ptp.reducers {
			if reducer != nil {
				_, selected := reducer.GetMetrics()
				if count, exists := selected[scale]; exists {
					totalSelected += count
				}
			}
		}
		
		// Count written from writers
		for _, writer := range ptp.writers[scale] {
			if writer != nil {
				totalWritten += writer.GetMetrics()
			}
		}
		
		fmt.Printf("Scale %d%%: %d selected, %d written\n", scale, totalSelected, totalWritten)
	}
}

// combinePartFiles combines all part files into single files per scale
func (ptp *PhaseTwoProcessor) combinePartFiles() error {
	fmt.Println("\n=== Combining Part Files ===")
	
	for _, scale := range ptp.config.Scales {
		if err := ptp.combineScaleFiles(scale); err != nil {
			return fmt.Errorf("failed to combine files for scale %d: %w", scale, err)
		}
		fmt.Printf("Scale %d%%: Combined %d part files\n", scale, ptp.config.FileWriterCount)
	}
	
	return nil
}

// combineScaleFiles combines all part files for a specific scale
func (ptp *PhaseTwoProcessor) combineScaleFiles(scale int) error {
	// Generate part file names
	partFiles := make([]string, 0, ptp.config.FileWriterCount)
	for i := 1; i <= ptp.config.FileWriterCount; i++ {
		partFile := filepath.Join(ptp.config.OutputDir, fmt.Sprintf("scale_%02d_population_part_%d.xml", scale, i))
		if _, err := os.Stat(partFile); err == nil {
			partFiles = append(partFiles, partFile)
		}
	}
	
	if len(partFiles) == 0 {
		return fmt.Errorf("no part files found for scale %d", scale)
	}
	
	// Create combined file
	combinedFile := filepath.Join(ptp.config.OutputDir, fmt.Sprintf("scale_%02d_population.xml", scale))
	outputFile, err := os.Create(combinedFile)
	if err != nil {
		return fmt.Errorf("failed to create combined file: %w", err)
	}
	defer outputFile.Close()
	
	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()
	
	// Write XML header
	header := `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE population SYSTEM "http://www.matsim.org/files/dtd/population_v6.dtd">

<population>
`
	if _, err := writer.WriteString(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	
	// Process each part file
	for _, partFile := range partFiles {
		if err := ptp.extractPersonsFromFile(partFile, writer); err != nil {
			return fmt.Errorf("failed to extract persons from %s: %w", partFile, err)
		}
	}
	
	// Write XML footer
	footer := `</population>
`
	if _, err := writer.WriteString(footer); err != nil {
		return fmt.Errorf("failed to write footer: %w", err)
	}
	
	// Clean up part files after successful combination
	for _, partFile := range partFiles {
		if err := os.Remove(partFile); err != nil {
			fmt.Printf("Warning: Failed to remove part file %s: %v\n", partFile, err)
		}
	}
	
	return nil
}

// extractPersonsFromFile extracts person elements from a part file
func (ptp *PhaseTwoProcessor) extractPersonsFromFile(filename string, writer *bufio.Writer) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	inPersonElement := false
	personBuffer := strings.Builder{}
	
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		
		// Skip XML header, DOCTYPE, and population tags
		if strings.HasPrefix(trimmedLine, "<?xml") ||
		   strings.HasPrefix(trimmedLine, "<!DOCTYPE") ||
		   trimmedLine == "<population>" ||
		   trimmedLine == "</population>" ||
		   trimmedLine == "" {
			continue
		}
		
		// Check if this line starts a person element
		if strings.HasPrefix(trimmedLine, "<person") {
			inPersonElement = true
			personBuffer.Reset()
			personBuffer.WriteString(line)
			personBuffer.WriteString("\n")
			continue
		}
		
		// If we're inside a person element, collect lines
		if inPersonElement {
			personBuffer.WriteString(line)
			personBuffer.WriteString("\n")
			
			// Check if this line ends the person element
			if strings.HasPrefix(trimmedLine, "</person>") {
				// Write the complete person element
				if _, err := writer.WriteString(personBuffer.String()); err != nil {
					return fmt.Errorf("failed to write person element: %w", err)
				}
				inPersonElement = false
				personBuffer.Reset()
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	
	return nil
}