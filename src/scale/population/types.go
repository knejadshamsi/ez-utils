package population

import (
	"time"
)

// AgentState represents the current state of an agent
type AgentState int

const (
	Available AgentState = iota
	Busy
	Paused
)

// Agent states (string constants for compatibility)
const (
	StateAvailable = "available"
	StateBusy      = "busy"
	StatePaused    = "paused"
)

// Agent types
const (
	AgentTypeReader    = "reader"
	AgentTypeExtractor = "extractor"
	AgentTypeHashMap   = "hashmap"
	AgentTypeReducer   = "reducer"
	AgentTypeDatabase  = "database"
	AgentTypeFileWriter = "file_writer"
)

// Manager commands
const (
	CommandStart  = "start"
	CommandStop   = "stop"
	CommandPause  = "pause"
	CommandResume = "resume"
)

// Error severity levels
const (
	SeverityWarning  = "warning"
	SeverityError    = "error"
	SeverityCritical = "critical"
)

// Coordinates represents a point in MTM8 coordinate system
type Coordinates struct {
	X float64
	Y float64
}

// GridBounds defines the boundaries of the spatial grid
type GridBounds struct {
	MinX     float64
	MaxX     float64
	MinY     float64
	MaxY     float64
	CellSize float64
}

// DensityMap maps bin IDs to agent counts
type DensityMap map[string]int

// RetentionMap maps bin IDs to retention probabilities
type RetentionMap map[string]float64

// AgentInfo contains information about an agent instance
type AgentInfo struct {
	ID        string
	Type      string
	State     string
	StartTime time.Time
}

// ManagerState tracks the overall state of a phase manager
type ManagerState struct {
	Phase          string                    // "phase_one" or "phase_two"
	ActiveAgents   map[string][]AgentInfo    // by agent type
	DataFlow       map[string]QueueStatus    // queue fill levels
	Completion     map[string]bool           // agent type completion
	PauseRequested bool
	ErrorState     bool
}

// QueueStatus represents the status of a data queue
type QueueStatus struct {
	Size     int
	Capacity int
	FillRate float64 // percentage filled
}

// WorkerConfig contains configuration for agent workers
type WorkerConfig struct {
	ExtractorCount  int
	HashMapCount    int
	ReducerCount    int
	DatabaseCount   int
	FileWriterCount int
	QueueSizes      QueueSizeConfig
	BatchSizes      BatchSizeConfig
}

// QueueSizeConfig defines queue sizes for different channels
type QueueSizeConfig struct {
	PersonXML      int
	Coordinate     int
	SelectedAgent  int
	StateUpdate    int
	Progress       int
	Error          int
	Command        int
	WorkAssignment int
	GridExpansion  int
	ChunkComplete  int
	BatchComplete  int
}

// BatchSizeConfig defines batch sizes for processing
type BatchSizeConfig struct {
	ChunkSize        int // For safe cut points
	GridExpansion    int // Coordinates before expansion
	DatabaseBatch    int // Records per database batch
	FileWriteBuffer  int // Buffer size for file operations
}

// GridConfig contains grid system configuration
type GridConfig struct {
	InitialBounds    GridBounds
	BinSize          float64
	ExpansionMargin  float64
}

// PhaseOneConfig contains configuration for Phase One processing
type PhaseOneConfig struct {
	ReaderCount         int    // Number of parallel readers
	SafePointInterval   int64  // Lines between safe points
	ExtractorCount      int
	HashMapCount        int
	GridConfig          GridConfig
}

// PhaseTwoConfig contains configuration for Phase Two processing
type PhaseTwoConfig struct {
	ReducerCount    int
	DatabaseCount   int
	FileWriterCount int
	OutputDir       string
	Scales          []int
}

// WorkRange represents a range of lines for parallel processing
type WorkRange struct {
	StartLine int64
	EndLine   int64
	ChunkID   string
}

// PhaseOneProgress tracks Phase One completion progress
type PhaseOneProgress struct {
	LinesRead         int64
	PersonsExtracted  int64
	CoordinatesMapped int64
	ElapsedTime       time.Duration
	IsComplete        bool
}