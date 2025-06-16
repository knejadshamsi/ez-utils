// Package display/business provides the business logic layer for the display package.
// This file separates UI concerns from business logic, making the code more maintainable
// and testable.
//
// The BusinessLogic struct manages:
// - State tracking and validation for module processes
// - Step timing and progression monitoring
// - Progress counters organized by logical processing phases
// - Flag management for conditional features (like cleanup, database operations)
//
// Design principles:
// - Single Responsibility: Each method has a clear, single purpose
// - Immutable reads: Getter methods don't modify state
// - Validation: Step transitions are validated for consistency
// - Organization: Counters are grouped by logical processing phases
// - Error handling: All operations can fail gracefully
//
// This layer is completely independent of the UI framework (Bubble Tea)
// and can be tested in isolation.
package display

import (
	"fmt"
	"time"
)

// BusinessLogic handles all non-UI business logic for the display package
type BusinessLogic struct {
	moduleName string
	maxSteps   int
	flags      map[string]bool
	
	// State tracking
	currentStep      int
	processComplete  bool
	stepDurations    map[int]time.Duration
	stepStartTime    time.Time
	totalStartTime   time.Time
	
	// Progress counters - organized by processing phase
	xmlProcessing     XMLProcessingData
	locationProcessing LocationProcessingData
	dataProcessing    DataProcessingData
	systemStats       SystemStatsData
}

// XMLProcessingData holds counters for XML processing phase (steps 0-1)
type XMLProcessingData struct {
	ChunkCount      int
	PersonsFound    int
	BytesReadMB     int
	FirstChunkFixed bool
	LastChunkFixed  bool
}

// LocationProcessingData holds counters for location processing phase (steps 2-3)
type LocationProcessingData struct {
	AgentCounter int
	ChunkCounter Counter
	DbCounter    Counter
}

// DataProcessingData holds counters for data processing phase (steps 4-18)
type DataProcessingData struct {
	CoordinateCount int
	BinCounter      Counter
	AgentBinCount   int
	ScaleCounter    Counter
	OutputScale     int
	OutputCounter   Counter
	CleanupFiles    int
	CleanupDirs     int
	CleanupBytes    int
}

// SystemStatsData holds system monitoring data
type SystemStatsData struct {
	CPUUsage float64
	RAMUsage float64
}

// NewBusinessLogic creates a new business logic instance
func NewBusinessLogic(config *ModuleConfig) *BusinessLogic {
	return &BusinessLogic{
		moduleName:     config.ModuleName,
		maxSteps:       config.MaxSteps,
		flags:          config.Flags,
		currentStep:    0,
		processComplete: false,
		stepDurations:  make(map[int]time.Duration),
		totalStartTime: time.Now(),
		stepStartTime:  time.Now(),
	}
}

// UpdateStep updates the current step and tracks timing
func (bl *BusinessLogic) UpdateStep(stepNumber int) {
	if bl.currentStep >= 0 && stepNumber != bl.currentStep {
		// Store duration for completed step
		bl.stepDurations[bl.currentStep] = time.Since(bl.stepStartTime)
		bl.stepStartTime = time.Now()
	}
	bl.currentStep = stepNumber
}

// SetProcessComplete marks the process as complete
func (bl *BusinessLogic) SetProcessComplete(complete bool) {
	bl.processComplete = complete
}

// GetFlag retrieves a flag value safely
func (bl *BusinessLogic) GetFlag(flag string) bool {
	if bl.flags == nil {
		return false
	}
	value, exists := bl.flags[flag]
	return exists && value
}

// GetCurrentStep returns the current step number
func (bl *BusinessLogic) GetCurrentStep() int {
	return bl.currentStep
}

// GetMaxSteps returns the maximum number of steps
func (bl *BusinessLogic) GetMaxSteps() int {
	return bl.maxSteps
}

// GetModuleName returns the module name
func (bl *BusinessLogic) GetModuleName() string {
	return bl.moduleName
}

// IsProcessComplete returns whether the process is complete
func (bl *BusinessLogic) IsProcessComplete() bool {
	return bl.processComplete
}

// GetElapsedTime returns the elapsed time for current step and total
func (bl *BusinessLogic) GetElapsedTime() (stepElapsed, totalElapsed time.Duration) {
	stepElapsed = time.Since(bl.stepStartTime)
	totalElapsed = time.Since(bl.totalStartTime)
	return
}

// GetStepDuration returns the duration for a specific step
func (bl *BusinessLogic) GetStepDuration(step int) (time.Duration, bool) {
	duration, exists := bl.stepDurations[step]
	return duration, exists
}

// Setter methods for XML processing data
func (bl *BusinessLogic) SetChunkCount(count int) {
	bl.xmlProcessing.ChunkCount = count
}

func (bl *BusinessLogic) SetPersonsFound(count int) {
	bl.xmlProcessing.PersonsFound = count
}

func (bl *BusinessLogic) SetBytesReadMB(mb int) {
	bl.xmlProcessing.BytesReadMB = mb
}

func (bl *BusinessLogic) SetFirstChunkStatus(fixed bool) {
	bl.xmlProcessing.FirstChunkFixed = fixed
}

func (bl *BusinessLogic) SetLastChunkStatus(fixed bool) {
	bl.xmlProcessing.LastChunkFixed = fixed
}

// Setter methods for location processing data
func (bl *BusinessLogic) SetAgentCounter(count int) {
	bl.locationProcessing.AgentCounter = count
}

func (bl *BusinessLogic) SetChunkCounter(current, total int) {
	bl.locationProcessing.ChunkCounter = Counter{current, total}
}

func (bl *BusinessLogic) SetDbCounter(current, total int) {
	bl.locationProcessing.DbCounter = Counter{current, total}
}

// Setter methods for data processing
func (bl *BusinessLogic) SetCoordinateCounter(count int) {
	bl.dataProcessing.CoordinateCount = count
}

func (bl *BusinessLogic) SetBinCounter(current, total int) {
	bl.dataProcessing.BinCounter = Counter{current, total}
}

func (bl *BusinessLogic) SetAgentBinCounter(count int) {
	bl.dataProcessing.AgentBinCount = count
}

func (bl *BusinessLogic) SetScaleCounter(current, total int) {
	bl.dataProcessing.ScaleCounter = Counter{current, total}
}

func (bl *BusinessLogic) SetOutputScale(scale int) {
	bl.dataProcessing.OutputScale = scale
}

func (bl *BusinessLogic) SetOutputCounter(current, total int) {
	bl.dataProcessing.OutputCounter = Counter{current, total}
}

func (bl *BusinessLogic) SetCleanupCounter(files, dirs, bytes int) {
	bl.dataProcessing.CleanupFiles = files
	bl.dataProcessing.CleanupDirs = dirs
	bl.dataProcessing.CleanupBytes = bytes
}

// Setter methods for system stats
func (bl *BusinessLogic) SetSystemStats(cpu, ram float64) {
	bl.systemStats.CPUUsage = cpu
	bl.systemStats.RAMUsage = ram
}

// Getter methods for data access
func (bl *BusinessLogic) GetXMLProcessingData() XMLProcessingData {
	return bl.xmlProcessing
}

func (bl *BusinessLogic) GetLocationProcessingData() LocationProcessingData {
	return bl.locationProcessing
}

func (bl *BusinessLogic) GetDataProcessingData() DataProcessingData {
	return bl.dataProcessing
}

func (bl *BusinessLogic) GetSystemStatsData() SystemStatsData {
	return bl.systemStats
}

// ValidateStepTransition validates if a step transition is valid
func (bl *BusinessLogic) ValidateStepTransition(newStep int) error {
	if newStep < 0 {
		return fmt.Errorf("step number cannot be negative")
	}
	
	if newStep > bl.maxSteps {
		return fmt.Errorf("step number %d exceeds maximum steps %d", newStep, bl.maxSteps)
	}
	
	// Allow going backwards (for error recovery) and skipping steps (for conditional logic)
	// but warn about unusual transitions
	if newStep > bl.currentStep+2 {
		return fmt.Errorf("warning: large step jump from %d to %d", bl.currentStep, newStep)
	}
	
	return nil
}

// ShouldSkipStep determines if a step should be skipped based on flags
func (bl *BusinessLogic) ShouldSkipStep(step int) bool {
	switch step {
	case 3:
		return !bl.GetFlag("db")
	case 18:
		return !bl.GetFlag("clean")
	default:
		return false
	}
}