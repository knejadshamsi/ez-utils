package workers

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ReaderWorker reads the population file and outputs complete person XML elements
type ReaderWorker struct {
	id            string        // Reader ID for multiple readers
	inputFile     string
	outputQueue   chan<- PersonXML
	chunkSize     int
	shouldCloseQueue bool        // Whether this reader should close the queue when done
	
	// Line range for parallel processing
	startLine     int64         // Starting line (inclusive)
	endLine       int64         // Ending line (inclusive, -1 for EOF)
	
	// Processing state
	personBuffer  strings.Builder
	inPerson      bool
	personCount   int
	currentLine   int64         // Current line being processed
	
	// Metrics
	totalLinesRead    int64
	totalPersonsFound int64
	
	// Progress callback
	onProgress func(lines, persons int64)
}

// NewReaderWorker creates a new reader worker (for single reader mode)
func NewReaderWorker(inputFile string, outputQueue chan<- PersonXML, chunkSize int) *ReaderWorker {
	return &ReaderWorker{
		id:               "Reader", // Changed to match TUI expectations
		inputFile:        inputFile,
		outputQueue:      outputQueue,
		chunkSize:        chunkSize,
		shouldCloseQueue: true, // Single reader should close queue
		startLine:        1,
		endLine:          -1, // Read entire file
	}
}

// NewReaderWorkerWithRange creates a new reader worker with line range (for multiple readers)
func NewReaderWorkerWithRange(id, inputFile string, outputQueue chan<- PersonXML, startLine, endLine int64) *ReaderWorker {
	return &ReaderWorker{
		id:               id,
		inputFile:        inputFile,
		outputQueue:      outputQueue,
		chunkSize:        1000, // Not used in range mode
		shouldCloseQueue: false, // Multiple readers don't close queue individually
		startLine:        startLine,
		endLine:          endLine,
	}
}

// Process reads the file and sends person XML to the output queue
func (r *ReaderWorker) Process() error {
	// Open input file
	file, err := os.Open(r.inputFile)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	r.currentLine = 0
	
	// Skip to start line if needed
	for r.currentLine < r.startLine-1 && scanner.Scan() {
		r.currentLine++
	}
	
	// Process lines within the assigned range
	for scanner.Scan() {
		line := scanner.Text()
		r.currentLine++
		r.totalLinesRead++
		
		// Check if we've reached the end of our assigned range
		if r.endLine > 0 && r.currentLine > r.endLine {
			break
		}
		
		r.processLine(line)
		
		// Update progress every 1000 lines or when person found
		if r.totalLinesRead%1000 == 0 && r.onProgress != nil {
			r.onProgress(r.totalLinesRead, r.totalPersonsFound)
		}
	}
	
	// Handle any remaining person in buffer
	if r.inPerson && r.personBuffer.Len() > 0 {
		r.sendPersonXML()
	}
	
	// Final progress update
	if r.onProgress != nil {
		r.onProgress(r.totalLinesRead, r.totalPersonsFound)
	}
	
	// Close output queue if this reader is responsible for it (single reader mode)
	if r.shouldCloseQueue {
		close(r.outputQueue)
	}
	
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}
	
	// fmt.Printf("\nReaderWorker: Completed. Read %d lines, found %d persons\n", 
	// 	r.totalLinesRead, r.totalPersonsFound)
	
	return nil
}

// processLine processes a single line of the file
func (r *ReaderWorker) processLine(line string) {
	// Search for <person in the line
	if strings.Contains(line, "<person") && !r.inPerson {
		r.inPerson = true
		r.personBuffer.Reset()
		r.personBuffer.WriteString(line)
		r.personBuffer.WriteString("\n")
		return
	}
	
	// If we're in a person, add line to buffer
	if r.inPerson {
		r.personBuffer.WriteString(line)
		r.personBuffer.WriteString("\n")
		
		// Detect </person> and send as task to queue
		if strings.Contains(line, "</person>") {
			// Complete person found
			r.sendPersonXML()
			
			// Increment person count
			r.totalPersonsFound++
			r.personCount++
			
			// Update progress when person found
			if r.onProgress != nil {
				r.onProgress(r.totalLinesRead, r.totalPersonsFound)
			}
			
			// Reset state
			r.inPerson = false
			r.personBuffer.Reset()
		}
	}
}

// sendPersonXML sends complete person XML to the output queue
func (r *ReaderWorker) sendPersonXML() {
	personXML := PersonXML{
		XML:        r.personBuffer.String(),
		LineNumber: r.currentLine,
	}
	
	// Send to queue (blocking if full - natural backpressure)
	r.outputQueue <- personXML
}

// SetProgressCallback sets the progress callback function
func (r *ReaderWorker) SetProgressCallback(callback func(lines, persons int64)) {
	r.onProgress = callback
}

// GetID returns the reader ID
func (r *ReaderWorker) GetID() string {
	return r.id
}

// GetMetrics returns reader metrics
func (r *ReaderWorker) GetMetrics() (linesRead, personsFound int64) {
	return r.totalLinesRead, r.totalPersonsFound
}