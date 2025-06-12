package workers

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// WriterWorker handles writing selected persons to output files
type WriterWorker struct {
	// Identity
	id          string
	workerIndex int
	scale       int
	
	// Input
	inputChannel  <-chan SelectedPerson
	doneChannel   chan<- CompletionSignal
	
	// Output
	outputDir     string
	outputFile    string
	
	// File handling
	file          *os.File
	writer        *bufio.Writer
	fileOpen      bool
	mu            sync.Mutex
	
	// Progress tracking
	personsWritten int64
	
	// State
	isRunning     bool
	headerWritten bool
}

// NewWriterWorker creates a new writer worker for a specific scale
func NewWriterWorker(id string, index int, scale int, outputDir string, inputChannel <-chan SelectedPerson) *WriterWorker {
	filename := fmt.Sprintf("scale_%02d_population_part_%d.xml", scale, index+1)
	
	return &WriterWorker{
		id:           id,
		workerIndex:  index,
		scale:        scale,
		inputChannel: inputChannel,
		outputDir:    outputDir,
		outputFile:   filename,
		fileOpen:     false,
		headerWritten: false,
	}
}

// SetDoneChannel sets the completion signal channel
func (ww *WriterWorker) SetDoneChannel(done chan<- CompletionSignal) {
	ww.doneChannel = done
}

// Process writes selected persons to the output file
func (ww *WriterWorker) Process() {
	defer func() {
		// Close file if open
		ww.closeFile()
		
		// Signal completion
		if ww.doneChannel != nil {
			ww.doneChannel <- CompletionSignal{
				WorkerID:  ww.id,
				Timestamp: time.Now(),
			}
		}
	}()
	
	ww.isRunning = true
	
	// Process all selected persons from input channel
	for selectedPerson := range ww.inputChannel {
		if err := ww.writePerson(selectedPerson); err != nil {
			fmt.Printf("Writer %s: Error writing person %s: %v\n", ww.id, selectedPerson.PersonID, err)
			continue
		}
		
		atomic.AddInt64(&ww.personsWritten, 1)
	}
	
	// Close file to ensure all data is flushed and footer is written
	ww.closeFile()
	
	ww.isRunning = false
}

// writePerson writes a single person to the output file
func (ww *WriterWorker) writePerson(selectedPerson SelectedPerson) error {
	ww.mu.Lock()
	defer ww.mu.Unlock()
	
	// Open file if not already open
	if !ww.fileOpen {
		if err := ww.openFile(); err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
	}
	
	// Write header if not already written
	if !ww.headerWritten {
		if err := ww.writeHeader(); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
		ww.headerWritten = true
	}
	
	// Write person XML
	if _, err := ww.writer.WriteString(selectedPerson.PersonXML); err != nil {
		return fmt.Errorf("failed to write person XML: %w", err)
	}
	
	// Flush periodically for better performance
	if ww.personsWritten%100 == 0 {
		if err := ww.writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer: %w", err)
		}
	}
	
	return nil
}

// openFile opens the output file for writing
func (ww *WriterWorker) openFile() error {
	// Ensure output directory exists
	if err := os.MkdirAll(ww.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	filepath := fmt.Sprintf("%s/%s", ww.outputDir, ww.outputFile)
	
	var err error
	ww.file, err = os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filepath, err)
	}
	
	ww.writer = bufio.NewWriter(ww.file)
	ww.fileOpen = true
	
	return nil
}

// writeHeader writes the XML header
func (ww *WriterWorker) writeHeader() error {
	header := `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE population SYSTEM "http://www.matsim.org/files/dtd/population_v6.dtd">

<population>
`
	_, err := ww.writer.WriteString(header)
	return err
}

// closeFile closes the output file with proper XML footer
func (ww *WriterWorker) closeFile() {
	ww.mu.Lock()
	defer ww.mu.Unlock()
	
	if !ww.fileOpen {
		return
	}
	
	// Write XML footer
	footer := `</population>
`
	if ww.writer != nil {
		ww.writer.WriteString(footer)
		ww.writer.Flush()
		ww.writer = nil
	}
	
	// Close file
	if ww.file != nil {
		ww.file.Close()
		ww.file = nil
	}
	
	ww.fileOpen = false
}

// GetMetrics returns current writing metrics
func (ww *WriterWorker) GetMetrics() int64 {
	return atomic.LoadInt64(&ww.personsWritten)
}

// GetID returns the worker ID
func (ww *WriterWorker) GetID() string {
	return ww.id
}

// GetScale returns the scale this writer handles
func (ww *WriterWorker) GetScale() int {
	return ww.scale
}

// GetOutputFile returns the output filename
func (ww *WriterWorker) GetOutputFile() string {
	return ww.outputFile
}

// IsRunning returns whether the writer is currently active
func (ww *WriterWorker) IsRunning() bool {
	return ww.isRunning
}

// GetStatus returns the current status
func (ww *WriterWorker) GetStatus() string {
	if ww.isRunning {
		return "Writing"
	}
	if ww.personsWritten > 0 {
		return "Completed"
	}
	return "Waiting"
}