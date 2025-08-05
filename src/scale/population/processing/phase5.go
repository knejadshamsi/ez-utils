package processing

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ez-utils/src/scale/population"
	"ez-utils/src/scale/population/tui"
)

// logPhase5ErrorToFile logs error messages to a unique log file
func logPhase5ErrorToFile(message string) {
	timestamp := time.Now().Format("20060102-150405-000000")
	filename := fmt.Sprintf("phase5-error-%s.log", timestamp)
	
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return // Silently fail to avoid stdout pollution
	}
	defer file.Close()
	
	file.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), message))
}

// Phase5Processor handles file finalization and cleanup
type Phase5Processor struct {
	config     *population.PhaseTwoConfig
	tuiEnabled bool
}

// NewPhase5Processor creates a new Phase 5 processor
func NewPhase5Processor(config *population.PhaseTwoConfig) *Phase5Processor {
	return &Phase5Processor{
		config:     config,
		tuiEnabled: true,
	}
}

// DisableTUI disables the terminal UI
func (p5p *Phase5Processor) DisableTUI() {
	p5p.tuiEnabled = false
}

// Process runs Phase 5: File Finalization and Cleanup
func (p5p *Phase5Processor) Process() error {
	if p5p.tuiEnabled {
		tui.UpdatePhase(5, "Combining Part Files")
		time.Sleep(1 * time.Second)
	}

	// Step 1: Combine part files into single files per scale
	if err := p5p.combinePartFiles(); err != nil {
		return fmt.Errorf("failed to combine part files: %w", err)
	}

	if p5p.tuiEnabled {
		tui.UpdatePhase(5, "Validating Output Files")
		time.Sleep(1 * time.Second)
	}

	// Step 2: Print summary
	p5p.printSummary()

	if p5p.tuiEnabled {
		tui.UpdatePhase(5, "Cleaning Temporary Files")
		time.Sleep(1 * time.Second)
	}

	// Step 3: Cleanup
	p5p.cleanup()

	if p5p.tuiEnabled {
		tui.UpdatePhase(5, "Process Complete")
		time.Sleep(2 * time.Second)
	}

	return nil
}

// combinePartFiles combines all part files into single files per scale
func (p5p *Phase5Processor) combinePartFiles() error {
	for _, scale := range p5p.config.Scales {
		if err := p5p.combineScaleFiles(scale); err != nil {
			return fmt.Errorf("failed to combine files for scale %d: %w", scale, err)
		}
	}

	return nil
}

// combineScaleFiles combines all part files for a specific scale
func (p5p *Phase5Processor) combineScaleFiles(scale int) error {
	// Generate part file names
	partFiles := make([]string, 0, p5p.config.FileWriterCount)
	for i := 1; i <= p5p.config.FileWriterCount; i++ {
		partFile := filepath.Join(p5p.config.OutputDir, fmt.Sprintf("scale_%02d_population_part_%d.xml", scale, i))
		if _, err := os.Stat(partFile); err == nil {
			partFiles = append(partFiles, partFile)
		}
	}

	if len(partFiles) == 0 {
		return fmt.Errorf("no part files found for scale %d", scale)
	}

	// Create combined file
	combinedFile := filepath.Join(p5p.config.OutputDir, fmt.Sprintf("scale_%02d_population.xml", scale))
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
		if err := p5p.extractPersonsFromFile(partFile, writer); err != nil {
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
			logPhase5ErrorToFile(fmt.Sprintf("Warning: Failed to remove part file %s: %v", partFile, err))
		}
	}

	return nil
}

// extractPersonsFromFile extracts person elements from a part file
func (p5p *Phase5Processor) extractPersonsFromFile(filename string, writer *bufio.Writer) error {
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

// cleanup performs post-processing cleanup
func (p5p *Phase5Processor) cleanup() {
	// Additional cleanup can be added here (temporary files, etc.)
}

// printSummary prints processing results
func (p5p *Phase5Processor) printSummary() {
	// Summary output removed per user request
}