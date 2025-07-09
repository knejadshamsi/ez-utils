package gui

import (
	"context"
	"database/sql"
	"io"
	"sync"
	"time"
)

// App struct - Main application structure for Wails
type App struct {
	ctx         context.Context
	StartupFile string
	EditMode    string // "population", "network", "public transportation"
}

// Person - Structure for population data
type Person struct {
	ID     string `json:"id"`
	Coords string `json:"coords"`
	RawXML string `json:"raw_xml"`
}

// PersonData - Holds extracted information for a single person (used by processor)
type PersonData struct {
	ID     string
	Coords string
	RawXML string
}

// PersonUpdate - Used for batch updates
type PersonUpdate struct {
	ID     string
	Coords string
	RawXML string
}

// Process - Structure for a record from the 'processes' table
type Process struct {
	ProcessID int       `json:"process_id"`
	FilePath  string    `json:"file_path"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// ProcessTelemetry - Represents telemetry data for a process
type ProcessTelemetry struct {
	ProcessID        int       `json:"process_id"`
	TotalFileSize    int64     `json:"total_file_size"`
	BytesRead        int64     `json:"bytes_read"`
	PersonsExtracted int64     `json:"persons_extracted"`
	ErrorCount       int64     `json:"error_count"`
	LastUpdated      time.Time `json:"last_updated"`
}

// Processor - Handles population file processing with telemetry
type Processor struct {
	db                      *sql.DB
	processID               int
	telemetry               *ProcessTelemetry
	telemetryMutex          sync.Mutex
	lastUpdate              time.Time
	personCounter           int64
	errorCounter            int64
	telemetryFailureCounter int64
}

// CountingReader - Wraps an io.Reader and counts bytes read
type CountingReader struct {
	reader    io.Reader
	bytesRead int64
	mu        sync.Mutex
}
