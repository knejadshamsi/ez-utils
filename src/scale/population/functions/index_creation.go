package functions

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// CreateIndexFile creates a new index file for safe cut points
func CreateIndexFile(basePath, inputFileName string) (*os.File, string, error) {
	// Generate index file name based on input file
	indexFileName := generateIndexFileName(inputFileName)
	indexPath := filepath.Join(basePath, indexFileName)
	
	// Create the index file
	indexFile, err := os.Create(indexPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create index file %s: %w", indexPath, err)
	}
	
	// Write header comment
	header := fmt.Sprintf("# Safe cut points for %s\n# Format: line_number\n", inputFileName)
	if _, err := indexFile.WriteString(header); err != nil {
		indexFile.Close()
		return nil, "", fmt.Errorf("failed to write header to index file: %w", err)
	}
	
	return indexFile, indexPath, nil
}

// WriteSafeCutPoint writes a safe cut point to the index file
func WriteSafeCutPoint(indexFile *os.File, lineNumber int64) error {
	if indexFile == nil {
		return fmt.Errorf("index file is nil")
	}
	
	_, err := fmt.Fprintf(indexFile, "%d\n", lineNumber)
	if err != nil {
		return fmt.Errorf("failed to write safe cut point %d: %w", lineNumber, err)
	}
	
	return nil
}

// LoadSafeCutPoints loads safe cut points from an index file
func LoadSafeCutPoints(indexFilePath string) ([]int64, error) {
	if !fileExists(indexFilePath) {
		return nil, fmt.Errorf("index file does not exist: %s", indexFilePath)
	}
	
	content, err := os.ReadFile(indexFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read index file: %w", err)
	}
	
	lines := strings.Split(string(content), "\n")
	var cutPoints []int64
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse line number
		lineNum, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			// Skip invalid lines but continue processing
			continue
		}
		
		cutPoints = append(cutPoints, lineNum)
	}
	
	return cutPoints, nil
}

// ValidateSafeCutPoints validates that safe cut points are in ascending order
func ValidateSafeCutPoints(cutPoints []int64) error {
	if len(cutPoints) == 0 {
		return fmt.Errorf("no safe cut points found")
	}
	
	for i := 1; i < len(cutPoints); i++ {
		if cutPoints[i] <= cutPoints[i-1] {
			return fmt.Errorf("safe cut points not in ascending order at index %d: %d <= %d", 
				i, cutPoints[i], cutPoints[i-1])
		}
	}
	
	return nil
}

// CalculateWorkRanges calculates work ranges based on safe cut points
func CalculateWorkRanges(cutPoints []int64, maxWorkers int) []WorkRange {
	if len(cutPoints) == 0 {
		return nil
	}
	
	var ranges []WorkRange
	
	// Create ranges between consecutive cut points
	for i := 0; i < len(cutPoints)-1; i++ {
		ranges = append(ranges, WorkRange{
			StartLine: cutPoints[i],
			EndLine:   cutPoints[i+1] - 1, // Exclusive end
			ChunkID:   fmt.Sprintf("chunk_%d", i+1),
		})
	}
	
	// Add final range from last cut point to end of file (will be determined later)
	if len(cutPoints) > 0 {
		ranges = append(ranges, WorkRange{
			StartLine: cutPoints[len(cutPoints)-1],
			EndLine:   -1, // Will be set to EOF during processing
			ChunkID:   fmt.Sprintf("chunk_%d", len(cutPoints)),
		})
	}
	
	return ranges
}

// WorkRange represents a range of lines for parallel processing
type WorkRange struct {
	StartLine int64
	EndLine   int64  // -1 means EOF
	ChunkID   string
}

// generateIndexFileName generates an index file name based on the input file
func generateIndexFileName(inputFileName string) string {
	// Remove extension and add .index
	name := strings.TrimSuffix(inputFileName, filepath.Ext(inputFileName))
	return fmt.Sprintf("%s.index", name)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetIndexFileStats returns statistics about an index file
func GetIndexFileStats(indexFilePath string) (int, int64, int64, error) {
	cutPoints, err := LoadSafeCutPoints(indexFilePath)
	if err != nil {
		return 0, 0, 0, err
	}
	
	if len(cutPoints) == 0 {
		return 0, 0, 0, nil
	}
	
	count := len(cutPoints)
	firstLine := cutPoints[0]
	lastLine := cutPoints[len(cutPoints)-1]
	
	return count, firstLine, lastLine, nil
}