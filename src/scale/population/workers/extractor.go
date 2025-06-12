package workers

import (
	"fmt"
	"regexp"
	"strconv"
	"sync/atomic"
)

// ExtractorWorker parses person XML and extracts coordinates
type ExtractorWorker struct {
	id          string
	inputQueue  <-chan PersonXML
	outputQueue chan<- Coordinates
	
	// XML parsing patterns
	planPattern     *regexp.Regexp
	activityPattern *regexp.Regexp
	coordPattern    *regexp.Regexp
	personIDPattern *regexp.Regexp
	
	// Metrics
	personsProcessed int64
	personsSkipped   int64
}

// NewExtractorWorker creates a new extractor worker
func NewExtractorWorker(id string, inputQueue <-chan PersonXML, outputQueue chan<- Coordinates) *ExtractorWorker {
	return &ExtractorWorker{
		id:              id,
		inputQueue:      inputQueue,
		outputQueue:     outputQueue,
		planPattern:     regexp.MustCompile(`(?s)<plan[^>]*selected="yes"[^>]*>.*?</plan>|(?s)<plan[^>]*>.*?</plan>`),
		activityPattern: regexp.MustCompile(`<activity[^>]*>`),
		coordPattern:    regexp.MustCompile(`x="([^"]*)"[^>]*y="([^"]*)"`),
		personIDPattern: regexp.MustCompile(`<person[^>]*id="([^"]*)"`),
	}
}

// Process reads from input queue and extracts coordinates
func (e *ExtractorWorker) Process() {
	// Reset counters to ensure clean start
	atomic.StoreInt64(&e.personsProcessed, 0)
	atomic.StoreInt64(&e.personsSkipped, 0)
	
	// Process until input queue is closed
	for personXML := range e.inputQueue {
		coords, err := e.extractCoordinates(personXML)
		if err != nil {
			// Skip person and continue
			atomic.AddInt64(&e.personsSkipped, 1)
			continue
		}
		
		// Send coordinates to output queue
		e.outputQueue <- *coords
		atomic.AddInt64(&e.personsProcessed, 1)
	}
	
	// fmt.Printf("ExtractorWorker %s: Completed. Processed %d, Skipped %d\n", 
	// 	e.id, e.personsProcessed, e.personsSkipped)
}

// extractCoordinates extracts coordinates from person XML
func (e *ExtractorWorker) extractCoordinates(personXML PersonXML) (*Coordinates, error) {
	// Extract person ID
	personID := e.extractPersonID(personXML.XML)
	
	// Find all plan elements
	planMatches := e.planPattern.FindAllString(personXML.XML, -1)
	if len(planMatches) == 0 {
		return nil, fmt.Errorf("no plans found")
	}
	
	// Look for selected plan first
	var selectedPlan string
	selectedPattern := regexp.MustCompile(`selected="yes"`)
	for _, plan := range planMatches {
		if selectedPattern.MatchString(plan) {
			selectedPlan = plan
			break
		}
	}
	
	// If no selected plan found, use first plan
	if selectedPlan == "" {
		selectedPlan = planMatches[0]
	}
	
	// Find first activity in selected plan
	activityMatches := e.activityPattern.FindAllString(selectedPlan, -1)
	if len(activityMatches) == 0 {
		return nil, fmt.Errorf("no activities found")
	}
	
	// Extract coordinates from first activity
	firstActivity := activityMatches[0]
	coordMatches := e.coordPattern.FindStringSubmatch(firstActivity)
	if len(coordMatches) < 3 {
		return nil, fmt.Errorf("coordinates not found")
	}
	
	// Parse X and Y coordinates
	x, err := strconv.ParseFloat(coordMatches[1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse X: %w", err)
	}
	
	y, err := strconv.ParseFloat(coordMatches[2], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Y: %w", err)
	}
	
	return &Coordinates{
		PersonID:   personID,
		X:          x,
		Y:          y,
		LineNumber: personXML.LineNumber,
	}, nil
}

// extractPersonID extracts person ID from XML
func (e *ExtractorWorker) extractPersonID(personXML string) string {
	matches := e.personIDPattern.FindStringSubmatch(personXML)
	if len(matches) >= 2 {
		return matches[1]
	}
	return "unknown"
}

// GetMetrics returns extractor metrics
func (e *ExtractorWorker) GetMetrics() (processed, skipped int64) {
	return atomic.LoadInt64(&e.personsProcessed), atomic.LoadInt64(&e.personsSkipped)
}