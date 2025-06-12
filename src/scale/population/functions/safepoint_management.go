package functions

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"ez-utils/src/scale/population"
)

// SafePointService handles safe point discovery for parallel file processing
type SafePointService interface {
	// DiscoverSafePoints scans the file and returns safe cut points
	DiscoverSafePoints(inputFile string, interval int64, progressCallback ProgressCallback) (*SafePointResult, error)
	
	// ValidateConfiguration validates safe point discovery parameters
	ValidateConfiguration(interval int64, numReaders int) error
}

// ProgressCallback reports discovery progress (lines scanned, points found)
type ProgressCallback func(linesScanned int64, pointsFound int)

// SafePointResult contains the results of safe point discovery
type SafePointResult struct {
	SafePoints  []int64
	TotalLines  int64
	WorkRanges  []population.WorkRange
}

// safePointService is the concrete implementation
type safePointService struct{
	validator ConfigValidator
}

// NewSafePointService creates a new safe point service
func NewSafePointService() SafePointService {
	return &safePointService{
		validator: NewConfigValidator(),
	}
}

// DiscoverSafePoints implements the SafePointService interface
func (s *safePointService) DiscoverSafePoints(inputFile string, interval int64, progressCallback ProgressCallback) (*SafePointResult, error) {
	// Validate inputs using centralized validator
	if err := s.validator.ValidateSafePointSettings(interval, 1); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	file, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	
	var safePoints []int64
	var currentLine int64
	var lastPersonLine int64
	nextTarget := interval

	// Always add line 1 as the first safe point
	safePoints = append(safePoints, 1)

	for scanner.Scan() {
		line := scanner.Text()
		currentLine++

		// Detect person start tags more precisely
		trimmedLine := strings.TrimSpace(line)
		if s.isPersonStartTag(trimmedLine) {
			lastPersonLine = currentLine
		}

		// Check if we've reached the next target interval
		if currentLine >= nextTarget {
			safePoint := s.findNearestSafePoint(lastPersonLine)
			if safePoint > 0 && (len(safePoints) == 0 || safePoint > safePoints[len(safePoints)-1]) {
				safePoints = append(safePoints, safePoint)
			}

			nextTarget += interval

			// Report progress
			if progressCallback != nil {
				progressCallback(currentLine, len(safePoints))
			}
		}

		// Frequent progress updates for responsiveness
		if currentLine%1000 == 0 && progressCallback != nil {
			progressCallback(currentLine, len(safePoints))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	// Final progress update
	if progressCallback != nil {
		progressCallback(currentLine, len(safePoints))
	}

	// Validate results
	if err := s.validateSafePoints(safePoints, interval); err != nil {
		return nil, fmt.Errorf("safe point validation failed: %w", err)
	}

	// Calculate work ranges
	workRanges := s.calculateWorkRanges(safePoints, currentLine)

	return &SafePointResult{
		SafePoints: safePoints,
		TotalLines: currentLine,
		WorkRanges: workRanges,
	}, nil
}

// ValidateConfiguration validates safe point discovery parameters
func (s *safePointService) ValidateConfiguration(interval int64, numReaders int) error {
	return s.validator.ValidateSafePointSettings(interval, numReaders)
}

// isPersonStartTag detects if a line contains a person start tag
func (s *safePointService) isPersonStartTag(trimmedLine string) bool {
	return strings.HasPrefix(trimmedLine, "<person") && !strings.Contains(trimmedLine, "</person>")
}

// findNearestSafePoint finds the safe point (line before <person>)
func (s *safePointService) findNearestSafePoint(personLine int64) int64 {
	if personLine <= 1 {
		return 1
	}
	return personLine - 1 // Line before <person> is safe
}

// validateSafePoints validates that safe points are in ascending order and sufficient
func (s *safePointService) validateSafePoints(safePoints []int64, interval int64) error {
	if len(safePoints) == 0 {
		return fmt.Errorf("no safe points discovered")
	}

	// For small files, a single safe point (start of file) is acceptable
	// This allows single reader/reducer processing for files that don't need parallelization
	if len(safePoints) == 1 {
		// Ensure the single safe point is valid (should be line 1)
		if safePoints[0] != 1 {
			return fmt.Errorf("invalid single safe point: %d (should be 1)", safePoints[0])
		}
		// Single safe point is valid for small files
		return nil
	}

	// For multiple safe points, ensure they are in ascending order
	for i := 1; i < len(safePoints); i++ {
		if safePoints[i] <= safePoints[i-1] {
			return fmt.Errorf("safe points not in ascending order at index %d: %d <= %d",
				i, safePoints[i], safePoints[i-1])
		}
	}

	return nil
}

// calculateWorkRanges calculates work ranges between consecutive safe points
func (s *safePointService) calculateWorkRanges(safePoints []int64, totalLines int64) []population.WorkRange {
	if len(safePoints) == 0 {
		return nil
	}

	var ranges []population.WorkRange

	// Create ranges between consecutive safe points
	for i := 0; i < len(safePoints)-1; i++ {
		ranges = append(ranges, population.WorkRange{
			StartLine: safePoints[i],
			EndLine:   safePoints[i+1] - 1, // Exclusive end
			ChunkID:   fmt.Sprintf("chunk_%d", i+1),
		})
	}

	// Add final range from last safe point to end of file
	if len(safePoints) > 0 {
		ranges = append(ranges, population.WorkRange{
			StartLine: safePoints[len(safePoints)-1],
			EndLine:   totalLines,
			ChunkID:   fmt.Sprintf("chunk_%d", len(safePoints)),
		})
	}

	return ranges
}