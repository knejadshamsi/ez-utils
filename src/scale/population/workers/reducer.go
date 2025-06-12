package workers

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"

	"ez-utils/src/scale/population"
)

// ScalingCoordinates represents extracted coordinates for scaling
type ScalingCoordinates struct {
	X float64
	Y float64
}


// ReducerWorker processes XML chunks and applies retention probability
type ReducerWorker struct {
	// Identity
	id         string
	workerIndex int
	
	// Input
	personChannel   <-chan PersonXML
	workRanges      []population.WorkRange // Used for assignment tracking
	scales          []int
	
	// Output channels
	selectedPersons map[int]chan SelectedPerson // per scale
	doneChannel     chan<- CompletionSignal
	
	// Progress tracking
	personsProcessed int64
	personsSelected  map[int]*int64  // Use pointers for atomic operations
	
	// XML parsing patterns (compiled once)
	planPattern     *regexp.Regexp
	activityPattern *regexp.Regexp
	coordPattern    *regexp.Regexp
	selectedPattern *regexp.Regexp
	
	// Random generator for retention decisions
	rng *rand.Rand
}

// SelectedPerson represents a person selected for a specific scale
type SelectedPerson struct {
	PersonXML string
	PersonID  string
	Scale     int
	BinID     string
	ChunkID   string
}

// NewReducerWorker creates a new reducer worker
func NewReducerWorker(id string, index int, personChannel <-chan PersonXML, workRanges []population.WorkRange, scales []int) *ReducerWorker {
	// Initialize personsSelected with atomic int64 pointers
	personsSelected := make(map[int]*int64)
	for _, scale := range scales {
		personsSelected[scale] = new(int64)
	}
	
	return &ReducerWorker{
		id:              id,
		workerIndex:     index,
		personChannel:   personChannel,
		workRanges:      workRanges,
		scales:          scales,
		personsSelected: personsSelected,
		
		// Pre-compile regex patterns for performance
		planPattern:     regexp.MustCompile(`(?s)<plan[^>]*selected="yes"[^>]*>.*?</plan>|(?s)<plan[^>]*>.*?</plan>`),
		activityPattern: regexp.MustCompile(`<activity[^>]*>`),
		coordPattern:    regexp.MustCompile(`x="([^"]*)"[^>]*y="([^"]*)"`),
		selectedPattern: regexp.MustCompile(`selected="yes"`),
		
		// Individual random source to avoid contention
		rng: rand.New(rand.NewSource(time.Now().UnixNano() + int64(index))),
	}
}

// SetOutputChannels sets the output channels for selected persons
func (rw *ReducerWorker) SetOutputChannels(channels map[int]chan SelectedPerson, done chan<- CompletionSignal) {
	rw.selectedPersons = channels
	rw.doneChannel = done
}

// Process processes persons received from the streaming reader
func (rw *ReducerWorker) Process() {
	defer func() {
		// Signal completion (don't close channels - let processor manage them)
		if rw.doneChannel != nil {
			rw.doneChannel <- CompletionSignal{
				WorkerID:  rw.id,
				Timestamp: time.Now(),
			}
		}
	}()
	
	// Process persons from input channel until closed
	for personXML := range rw.personChannel {
		chunkID := fmt.Sprintf("line_%d", personXML.LineNumber) // Use line number as chunk ID
		if err := rw.processPerson(personXML.XML, chunkID); err != nil {
			// Skip person on error (continue processing)
		}
		atomic.AddInt64(&rw.personsProcessed, 1)
	}
}


// processPerson processes a single person and applies retention logic
func (rw *ReducerWorker) processPerson(personXML string, chunkID string) error {
	// Extract coordinates from person XML
	coords, err := rw.extractCoordinates(personXML)
	if err != nil {
		return err // Skip person
	}
	
	// Calculate bin ID for retention lookup
	gridBounds := population.GetGlobalGrid()
	binID := population.CalculateBinID(coords.X, coords.Y, gridBounds.CellSize)
	
	// Extract person ID for tracking
	personID := rw.extractPersonID(personXML)
	
	// Apply retention probability for each requested scale
	for _, scale := range rw.scales {
		retentionMap := population.GetRetentionMap(scale)
		if retentionMap == nil {
			continue
		}
		
		// Get retention probability for this bin
		probability, exists := retentionMap[binID]
		if !exists {
			probability = 0.0 // Default to 0% retention if bin not found
		}
		
		// Coin toss decision
		if rw.rng.Float64() < probability {
			// Person selected for this scale - send to writer
			selectedPerson := SelectedPerson{
				PersonXML: personXML,
				PersonID:  personID,
				Scale:     scale,
				BinID:     binID,
				ChunkID:   chunkID,
			}
			
			// Send to appropriate scale channel (blocking to ensure no data loss)
			rw.selectedPersons[scale] <- selectedPerson
			atomic.AddInt64(rw.personsSelected[scale], 1)
		}
	}
	
	return nil
}

// extractCoordinates extracts coordinates from person XML
func (rw *ReducerWorker) extractCoordinates(personXML string) (*ScalingCoordinates, error) {
	// Find all plan elements
	planMatches := rw.planPattern.FindAllString(personXML, -1)
	if len(planMatches) == 0 {
		return nil, fmt.Errorf("no plans found")
	}
	
	// Look for selected plan first
	var selectedPlan string
	for _, plan := range planMatches {
		if rw.selectedPattern.MatchString(plan) {
			selectedPlan = plan
			break
		}
	}
	
	// If no selected plan found, use first plan
	if selectedPlan == "" {
		selectedPlan = planMatches[0]
	}
	
	// Find first activity in selected plan
	activityMatches := rw.activityPattern.FindAllString(selectedPlan, -1)
	if len(activityMatches) == 0 {
		return nil, fmt.Errorf("no activities found")
	}
	
	// Extract coordinates from first activity
	firstActivity := activityMatches[0]
	coordMatches := rw.coordPattern.FindStringSubmatch(firstActivity)
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
	
	return &ScalingCoordinates{
		X: x,
		Y: y,
	}, nil
}

// extractPersonID extracts person ID from person XML
func (rw *ReducerWorker) extractPersonID(personXML string) string {
	personIDPattern := regexp.MustCompile(`<person[^>]*id="([^"]*)`)
	matches := personIDPattern.FindStringSubmatch(personXML)
	if len(matches) >= 2 {
		return matches[1]
	}
	return "unknown"
}

// GetMetrics returns current processing metrics
func (rw *ReducerWorker) GetMetrics() (int64, map[int]int64) {
	processed := atomic.LoadInt64(&rw.personsProcessed)
	selected := make(map[int]int64)
	for scale, countPtr := range rw.personsSelected {
		selected[scale] = atomic.LoadInt64(countPtr)
	}
	return processed, selected
}

// GetID returns the worker ID
func (rw *ReducerWorker) GetID() string {
	return rw.id
}

// GetChunkIDs returns all chunk IDs being processed
func (rw *ReducerWorker) GetChunkIDs() []string {
	chunkIDs := make([]string, len(rw.workRanges))
	for i, workRange := range rw.workRanges {
		chunkIDs[i] = workRange.ChunkID
	}
	return chunkIDs
}