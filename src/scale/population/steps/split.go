package steps

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"ez-utils/src/scale/display"

	"github.com/ncw/directio"
	"golang.org/x/exp/mmap"
)

const xmlHeader = "<?xml version=\"1.0\" encoding=\"utf-8\"?>"
const docType = "<!DOCTYPE population SYSTEM \"http://www.matsim.org/files/dtd/population_v6.dtd\">"
const openTag = "<population desc=\"Switzerland Baseline\">"
const closeTag = "</population>"

// DirectIO alignment requirement
const DirectIOAlign = 4096 // 4KB alignment

// Optimized separator check - just look for the HTML comment prefix
// This is much faster than checking the entire separator string
var separatorPrefix = []byte("<!-- ")

// Define a struct to hold chunk data for worker
type chunkJob struct {
	buffer   *bytes.Buffer // Using buffer pool instead of byte slice copy
	outDir   string
	chunkNum int
}

// Create a buffer pool to reduce memory allocations and GC pressure
var bufferPool = sync.Pool{
	New: func() interface{} {
		// Start with a reasonable capacity to avoid frequent resizing
		return bytes.NewBuffer(make([]byte, 0, 1024*1024)) // 1MB initial capacity
	},
}

// Get a buffer from the pool
func getBuffer() *bytes.Buffer {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset() // Ensure it's empty
	return buf
}

// Return a buffer to the pool
func putBuffer(buf *bytes.Buffer) {
	if buf != nil {
		bufferPool.Put(buf)
	}
}

func SplitPopulation(inputFile string, chunkSize int) error {

	// Create output directory
	outDir := filepath.Join("temp", "population", "01_split_population_chunks", "chunks")
	os.MkdirAll(outDir, os.ModePerm)

	// Create a buffered channel for chunk jobs
	// Increased buffer to 50 for better throughput
	jobChan := make(chan chunkJob, 50)

	// WaitGroup to track when all chunks have been saved
	var wg sync.WaitGroup

	// Start worker goroutines (increased to 16 for better I/O parallelism)
	const numWorkers = 16
	for i := 0; i < numWorkers; i++ {
		go func() {
			for job := range jobChan {
				// Pass the complete job structure to avoid any naming confusion
				saveChunkWorker(job)
				wg.Done() // Mark this job as complete
			}
		}()
	}

	// Using memory-mapped file for faster reading
	mmapReader, err := mmap.Open(inputFile)
	if err != nil {
		// Fallback to regular file reading if mmap fails
		return fallbackSplitPopulation(inputFile, chunkSize, outDir, jobChan, &wg)
	}
	defer mmapReader.Close()

	fileSize := int64(mmapReader.Len())

	// Prepare for processing
	var (
		chunkCount     int32 = 0
		personsFound   int32 = 0
		bytesProcessed int64 = 0
		separatorCount int   = 0
	)

	// Get the first buffer from the pool instead of creating a new one
	chunkBuffer := getBuffer()

	// Add XML headers at the beginning
	chunkBuffer.WriteString(xmlHeader + "\n")
	chunkBuffer.WriteString(docType + "\n")
	chunkBuffer.WriteString(openTag + "\n")

	// Process the file in large blocks (4MB at a time)
	const blockSize = 4 * 1024 * 1024 // 4MB block size

	// For line detection
	newline := byte('\n')
	data := make([]byte, blockSize)

	// Start processing file
	var position int64 = 0

	// Create a timer for UI updates
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Launch a goroutine to update the display
	go func() {
		for range ticker.C {
			display.SetPersonsFound(int(atomic.LoadInt32(&personsFound)))
			display.SetChunkCount(int(atomic.LoadInt32(&chunkCount)))
			display.SetBytesReadMB(int(bytesProcessed / 1048576))
		}
	}()

	for position < fileSize {
		// Calculate how many bytes to read in this block
		readSize := blockSize
		if position+int64(readSize) > fileSize {
			readSize = int(fileSize - position)
		}

		// Read a block of data
		n, _ := mmapReader.ReadAt(data[:readSize], position)
		if n <= 0 {
			break
		}

		// Process the block
		blockData := data[:n]

		// Update bytes processed
		bytesProcessed += int64(n)

		// Process line by line within the block
		var i int = 0
		for i < n {
			// Find next newline
			nlPos := bytes.IndexByte(blockData[i:], newline)

			if nlPos < 0 {
				// No more newlines in this block, add to buffer and continue
				line := blockData[i:]
				chunkBuffer.Write(line)
				break
			}

			// We found a newline
			lineEnd := i + nlPos
			line := blockData[i:lineEnd]

			// Write line to buffer
			chunkBuffer.Write(line)
			chunkBuffer.WriteByte(newline)

			// Check if this is a separator line
			trimmedLine := bytes.TrimSpace(line)
			if len(trimmedLine) > 0 && bytes.HasPrefix(trimmedLine, separatorPrefix) {
				separatorCount++
				atomic.AddInt32(&personsFound, 1)

				// If we've reached the chunk size, save this chunk
				if separatorCount >= chunkSize {
					// Finalize chunk
					chunkBuffer.WriteString(closeTag + "\n")

					// Increment chunk counter
					currentChunk := atomic.AddInt32(&chunkCount, 1)

					// Add to wait group
					wg.Add(1)

					// Send the buffer directly to worker pool
					jobChan <- chunkJob{
						buffer:   chunkBuffer, // Send the entire buffer, no copying needed
						outDir:   outDir,
						chunkNum: int(currentChunk),
					}

					// Get a fresh buffer from the pool for the next chunk
					chunkBuffer = getBuffer()
					chunkBuffer.WriteString(xmlHeader + "\n")
					chunkBuffer.WriteString(docType + "\n")
					chunkBuffer.WriteString(openTag + "\n")
					separatorCount = 0
				}
			}

			// Move past this line
			i = lineEnd + 1
		}

		// Move position for next block
		position += int64(n)
	}

	// Handle the last chunk if there's any content
	if chunkBuffer.Len() > len(xmlHeader)+len(docType)+len(openTag)+10 {
		// Finalize last chunk
		chunkBuffer.WriteString(closeTag + "\n")

		// Increment chunk counter
		currentChunk := atomic.AddInt32(&chunkCount, 1)

		// Add to wait group
		wg.Add(1)

		// Send job to worker pool
		jobChan <- chunkJob{
			buffer:   chunkBuffer, // Send the buffer without copying
			outDir:   outDir,
			chunkNum: int(currentChunk),
		}
	} else {
		// If we didn't use this buffer, return it to the pool
		putBuffer(chunkBuffer)
	}

	// Stop the update ticker
	ticker.Stop()

	// Make sure the final stats are displayed
	display.SetPersonsFound(int(atomic.LoadInt32(&personsFound)))
	display.SetChunkCount(int(atomic.LoadInt32(&chunkCount)))
	display.SetBytesReadMB(int(bytesProcessed / 1048576))

	// Close the job channel (no more jobs will be sent)
	close(jobChan)

	// Wait for all chunks to be saved before returning
	wg.Wait()

	// Return success
	return nil
}

// Fallback method if memory mapping fails
func fallbackSplitPopulation(inputFile string, chunkSize int, outDir string, jobChan chan chunkJob, wg *sync.WaitGroup) error {
	// Try to open the file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	// Create a 4MB buffer for reading
	reader := bufio.NewReaderSize(file, 4*1024*1024)

	// Get a buffer from the pool
	chunkBuffer := getBuffer()

	separatorCount := 0
	chunkCount := 0
	personsFound := 0
	bytesRead := int64(0)

	// Initial chunk setup
	chunkBuffer.WriteString(xmlHeader + "\n")
	chunkBuffer.WriteString(docType + "\n")
	chunkBuffer.WriteString(openTag + "\n")

	// Create a timer for UI updates
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Launch a goroutine to update the display
	go func() {
		for range ticker.C {
			display.SetPersonsFound(personsFound)
			display.SetChunkCount(chunkCount)
			display.SetBytesReadMB(int(bytesRead / 1048576))
		}
	}()

	// Read line by line with optimized buffer
	lineBuf := make([]byte, 512) // Most XML lines should fit in this
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil && err != io.ErrUnexpectedEOF {
			if err == io.EOF {
				break // End of file
			}
			return fmt.Errorf("error reading line: %w", err)
		}

		// Copy to our line buffer
		copied := copy(lineBuf, line)
		lineLen := copied

		// Handle lines longer than our buffer
		for isPrefix && err == nil {
			line, isPrefix, err = reader.ReadLine()
			if lineLen+len(line) > len(lineBuf) {
				// Expand the buffer if needed
				newBuf := make([]byte, len(lineBuf)*2)
				copy(newBuf, lineBuf[:lineLen])
				lineBuf = newBuf
			}
			lineLen += copy(lineBuf[lineLen:], line)
		}

		// Update bytes read counter
		bytesRead += int64(lineLen + 1) // +1 for newline

		// Write the line to our chunk buffer
		chunkBuffer.Write(lineBuf[:lineLen])
		chunkBuffer.WriteByte('\n')

		// Check if this is a separator line
		if lineLen > 4 && bytes.HasPrefix(bytes.TrimSpace(lineBuf[:lineLen]), separatorPrefix) {
			separatorCount++
			personsFound++

			// When the counter reaches CHUNK_SIZE, save the buffer as a new chunk
			if separatorCount >= chunkSize {
				chunkCount++
				// Finalize chunk
				chunkBuffer.WriteString(closeTag + "\n")

				// Add to wait group
				wg.Add(1)

				// Send job to worker pool (directly send the buffer)
				jobChan <- chunkJob{
					buffer:   chunkBuffer,
					outDir:   outDir,
					chunkNum: chunkCount,
				}

				// Get a new buffer for the next chunk
				chunkBuffer = getBuffer()
				chunkBuffer.WriteString(xmlHeader + "\n")
				chunkBuffer.WriteString(docType + "\n")
				chunkBuffer.WriteString(openTag + "\n")
				separatorCount = 0

				// Update display
				display.SetChunkCount(chunkCount)
			}
		}

		// Break if we reached EOF
		if err == io.EOF {
			break
		}
	}

	// Save last chunk if there's content
	if chunkBuffer.Len() > len(xmlHeader)+len(docType)+len(openTag)+10 {
		chunkCount++
		chunkBuffer.WriteString(closeTag + "\n")

		// Add to wait group
		wg.Add(1)

		// Send job to worker pool
		jobChan <- chunkJob{
			buffer:   chunkBuffer,
			outDir:   outDir,
			chunkNum: chunkCount,
		}

		display.SetChunkCount(chunkCount)
	} else {
		// Return the buffer to the pool if we didn't use it
		putBuffer(chunkBuffer)
	}

	// Stop the ticker
	ticker.Stop()

	// Make sure final stats are updated
	display.SetPersonsFound(personsFound)
	display.SetBytesReadMB(int(bytesRead / 1048576))
	return nil
}

// Worker function that saves chunks using optimized I/O
func saveChunkWorker(job chunkJob) {
	defer putBuffer(job.buffer) // Return the buffer to the pool when done

	data := job.buffer.Bytes()
	dataLen := len(data)
	filename := filepath.Join(job.outDir, fmt.Sprintf("chunk_%04d.xml", job.chunkNum))

	// Only use direct I/O for larger chunks
	if dataLen > 1*1024*1024 { // Files larger than 1MB
		// Try direct I/O first
		if err := saveWithDirectIO(filename, data); err == nil {
			return // Successfully saved with direct I/O
		}
	}

	// Fall back to standard I/O
	saveWithStandardIO(filename, data)
}

// Save file using direct I/O to bypass OS cache
func saveWithDirectIO(filename string, data []byte) error {
	// Calculate aligned size (DirectIO requires aligned buffers)
	alignedSize := (len(data) + DirectIOAlign - 1) & ^(DirectIOAlign - 1)

	// Create an aligned buffer
	alignedBuf := directio.AlignedBlock(alignedSize)
	copy(alignedBuf, data)

	// Open file with direct I/O
	f, err := directio.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write data
	_, err = f.Write(alignedBuf[:len(data)])
	return err
}

// Standard I/O fallback
func saveWithStandardIO(filename string, data []byte) error {
	// Create file
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Use a larger buffer for maximum write throughput (256KB)
	writer := bufio.NewWriterSize(f, 256*1024)

	// Write the complete chunk data in one operation
	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	// Ensure all data is flushed to disk
	return writer.Flush()
}

func SplitPopulationPart2(inputFile string, chunkSize int) error {

	// Get chunk count from the directory
	outDir := filepath.Join("temp", "population", "01_split_population_chunks", "chunks")
	files, _ := os.ReadDir(outDir)
	chunkCount := len(files)

	// Reset both status indicators to false to show we're working
	display.SetFirstChunkStatus(false)
	display.SetLastChunkStatus(false)

	// Fix the first chunk
	if err := fixFirstChunk(outDir, 1); err != nil {
		return fmt.Errorf("failed to fix first chunk: %w", err)
	}
	display.SetFirstChunkStatus(true)

	// Fix the last chunk
	if err := fixLastChunk(outDir, chunkCount); err != nil {
		return fmt.Errorf("failed to fix last chunk: %w", err)
	}
	display.SetLastChunkStatus(true)

	// Return success
	return nil
}

func fixFirstChunk(outDir string, chunkNum int) error {
	filename := filepath.Join(outDir, fmt.Sprintf("chunk_%04d.xml", chunkNum))

	// Read file using optimized approach
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Find occurrences of <population token (not just the exact openTag)
	// This is more flexible than checking for the exact openTag string
	populationTokenCount := bytes.Count(data, []byte("<population"))

	if populationTokenCount > 1 {
		// Find the first occurrence of <population
		firstPopPos := bytes.Index(data, []byte("<population"))
		if firstPopPos < 0 {
			fmt.Printf("Error: No opening population tag found in %s\n", filename)
			return fmt.Errorf("no opening population tag found")
		}

		// Find the end of line after first <population
		firstTagEnd := firstPopPos
		for firstTagEnd < len(data) && data[firstTagEnd] != '\n' {
			firstTagEnd++
		}
		if firstTagEnd < len(data) {
			firstTagEnd++ // Include the newline
		}

		// Find the second occurrence of <population
		secondPopPos := bytes.Index(data[firstTagEnd:], []byte("<population"))
		if secondPopPos >= 0 {
			secondPopPos += firstTagEnd // Adjust to absolute position

			// Find the start of the line containing the second <population
			lineStart := secondPopPos
			for lineStart > 0 && data[lineStart-1] != '\n' {
				lineStart--
			}

			// Find the end of the line containing the second <population
			lineEnd := secondPopPos
			for lineEnd < len(data) && data[lineEnd] != '\n' {
				lineEnd++
			}
			if lineEnd < len(data) {
				lineEnd++ // Include the newline
			}

			// Create a new buffer with fixed content
			var fixedBuffer bytes.Buffer

			// Write content before the second <population line
			fixedBuffer.Write(data[:lineStart])

			// Write content after the second <population line (skip the entire line)
			fixedBuffer.Write(data[lineEnd:])

			// Save the fixed content
			if err := saveWithStandardIO(filename, fixedBuffer.Bytes()); err != nil {
				fmt.Printf("Error saving fixed file %s: %v\n", filename, err)
				return err
			}
			return nil
		}
	}

	// If we didn't find multiple <population tokens or couldn't fix them,
	// just ensure the file has a proper structure
	var fixedBuffer bytes.Buffer

	// Check if xml header exists
	if !bytes.Contains(data, []byte(xmlHeader)) {
		fixedBuffer.WriteString(xmlHeader + "\n")
	}

	// Check if doctype exists
	if !bytes.Contains(data, []byte(docType)) {
		fixedBuffer.WriteString(docType + "\n")
	}

	// Check if opening population tag exists
	if !bytes.Contains(data, []byte("<population")) {
		fixedBuffer.WriteString(openTag + "\n")
		fixedBuffer.Write(data)
	} else {
		// File contains a population tag, just use the original content
		fixedBuffer.Write(data)
	}

	// Save the fixed content
	if err := saveWithStandardIO(filename, fixedBuffer.Bytes()); err != nil {
		fmt.Printf("Error saving fixed file %s: %v\n", filename, err)
		return err
	}
	return nil
}

func fixLastChunk(outDir string, chunkNum int) error {
	filename := filepath.Join(outDir, fmt.Sprintf("chunk_%04d.xml", chunkNum))

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Count occurrences of closing tags
	closingTagCount := bytes.Count(data, []byte("</population>"))

	// If no closing tag found, add one
	if closingTagCount == 0 {
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		writer.WriteString(closeTag + "\n")
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer: %w", err)
		}
		return nil
	}

	// If multiple closing tags found, remove the second one
	if closingTagCount > 1 {
		// Find the first occurrence
		firstClosePos := bytes.Index(data, []byte("</population>"))
		if firstClosePos < 0 {
			// This shouldn't happen as we already counted them
			return fmt.Errorf("couldn't find first closing tag")
		}

		// Find the end of the first closing tag
		firstCloseEnd := firstClosePos + len("</population>")

		// Find the second occurrence
		secondClosePos := bytes.Index(data[firstCloseEnd:], []byte("</population>"))
		if secondClosePos < 0 {
			// This shouldn't happen as we already counted them
			return fmt.Errorf("couldn't find second closing tag")
		}
		secondClosePos += firstCloseEnd // Adjust to absolute position

		// Find the start of the line containing the second closing tag
		lineStart := secondClosePos
		for lineStart > 0 && data[lineStart-1] != '\n' {
			lineStart--
		}

		// Find the end of the line containing the second closing tag
		lineEnd := secondClosePos + len("</population>")
		for lineEnd < len(data) && data[lineEnd] != '\n' {
			lineEnd++
		}
		if lineEnd < len(data) {
			lineEnd++ // Include the newline
		}

		// Create fixed buffer
		var fixedBuffer bytes.Buffer

		// Write everything before the second closing tag line
		fixedBuffer.Write(data[:lineStart])

		// Write everything after the second closing tag line
		fixedBuffer.Write(data[lineEnd:])

		// Save the fixed content, try direct I/O first, then standard I/O as fallback
		if err := saveWithDirectIO(filename, fixedBuffer.Bytes()); err != nil {
			return saveWithStandardIO(filename, fixedBuffer.Bytes())
		}
		return nil
	}
	return nil
}
