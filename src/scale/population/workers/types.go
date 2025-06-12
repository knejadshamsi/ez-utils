package workers

import (
	"time"
)

// PersonXML represents a complete person element from the population file
type PersonXML struct {
	XML        string
	LineNumber int64
}

// Coordinates represents extracted x,y coordinates
type Coordinates struct {
	PersonID   string
	X          float64
	Y          float64
	LineNumber int64
}

// CompletionSignal indicates a worker has finished processing
type CompletionSignal struct {
	WorkerID  string
	Timestamp time.Time
}

// QueueConfig defines the queue sizes
type QueueConfig struct {
	PersonQueueSize int
	CoordQueueSize  int
	DoneQueueSize   int
}

// DefaultQueueConfig returns default queue sizes
func DefaultQueueConfig() QueueConfig {
	return QueueConfig{
		PersonQueueSize: 10000,
		CoordQueueSize:  10000,
		DoneQueueSize:   1000,
	}
}