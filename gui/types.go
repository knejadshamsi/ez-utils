package gui

import (
	"context"
	"io"
	"sync"
	"time"
)

// App struct - Main application structure for Wails
type App struct {
	ctx         context.Context
	db          *Database
	StartupFile string
	EditMode    string // "population", "network", "public transportation"
}

// Person - Structure for population data
type Person struct {
	ID     string  `json:"id"`
	Lng    float64 `json:"lng"`
	Lat    float64 `json:"lat"`
	RawXML string  `json:"raw_xml"`
}

// PersonData - Holds extracted information for a single person (used by processor)
type PersonData struct {
	ID     string  `json:"id"`
	Lng    float64 `json:"lng"`
	Lat    float64 `json:"lat"`
	RawXML string  `json:"raw_xml"`
}

// PersonUpdate - Used for batch updates
type PersonUpdate struct {
	ID     string
	Lng    float64
	Lat    float64
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
	NodesRead        int64     `json:"nodes_read"`
	LinksRead        int64     `json:"links_read"`
	ErrorCount       int64     `json:"error_count"`
	LastUpdated      time.Time `json:"last_updated"`
}

// Processor - Handles population file processing with telemetry
type Processor struct {
	app                     *App
	db                      *Database
	processID               int
	telemetry               *ProcessTelemetry
	telemetryMutex          sync.Mutex
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

// Network-specific types

// NodeData represents a network node for database storage.
type NodeData struct {
	ID     string  `json:"id"`
	Lng    float64 `json:"lng"`
	Lat    float64 `json:"lat"`
	RawXML string  `json:"raw_xml"`
}

// LinkData represents a network link for database storage.
type LinkData struct {
	ID       string `json:"id"`
	FromNode string `json:"from_node"`
	ToNode   string `json:"to_node"`
	RawXML   string `json:"raw_xml"`
}

// BoundingBox represents a geographic bounding box.
type BoundingBox struct {
	North float64 `json:"north"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	West  float64 `json:"west"`
}

// NodeResult represents a node query result with parsed coordinates.
type NodeResult struct {
	ID     string  `json:"id"`
	Lng    float64 `json:"lng"`
	Lat    float64 `json:"lat"`
	RawXML string  `json:"raw_xml"`
}

// LinkResult represents a link query result.
type LinkResult struct {
	ID       string `json:"id"`
	FromNode string `json:"from_node"`
	ToNode   string `json:"to_node"`
	RawXML   string `json:"raw_xml"`
}

// ProcessResult represents the result of starting a process
type ProcessResult struct {
	ProcessID int    `json:"process_id"`
	Message   string `json:"message"`
}

// PT-specific types

// Route represents a route within a line
type Route struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Stops int    `json:"stops"`
}

// Line represents a transit line with embedded routes
type Line struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Type   string  `json:"type"` // BUS, METRO, TRAM
	Routes []Route `json:"routes"`
}

// Stop represents a stop with route-specific timing
type Stop struct {
	RouteID         string                 `json:"routeId"`
	StopID          string                 `json:"stopId"`
	ArrivalOffset   string                 `json:"arrivalOffset"`
	DepartureOffset string                 `json:"departureOffset"`
	StopName        string                 `json:"stopName"`
	Lat             float64                `json:"lat"`
	Lng             float64                `json:"lng"`
	Sequence        int                    `json:"sequence"`
	Attributes      map[string]interface{} `json:"attributes,omitempty"`
}

// Departure represents a scheduled departure
type Departure struct {
	ID           string `json:"id"`
	RouteID      string `json:"routeId"`
	DepartureTime string `json:"departureTime"`
	VehicleRefID string `json:"vehicleRefId,omitempty"`
}

// PTStopUpdate represents an update to a stop
type PTStopUpdate struct {
	StopName        *string                `json:"stopName,omitempty"`
	Lat             *float64               `json:"lat,omitempty"`
	Lng             *float64               `json:"lng,omitempty"`
	ArrivalOffset   *string                `json:"arrivalOffset,omitempty"`
	DepartureOffset *string                `json:"departureOffset,omitempty"`
	Sequence        *int                   `json:"sequence,omitempty"`
	Attributes      map[string]interface{} `json:"attributes,omitempty"`
}

// Spatial query types

// ViewportBounds represents a geographic viewport for spatial queries
type ViewportBounds struct {
	MinLat float64 `json:"minLat"`
	MaxLat float64 `json:"maxLat"`
	MinLng float64 `json:"minLng"`
	MaxLng float64 `json:"maxLng"`
}

// LoadingParams represents parameters for the LoadProcessData function
type LoadingParams struct {
	ProcessId    int            `json:"processId"`
	FileEditMode string         `json:"fileEditMode"`
	Viewport     ViewportBounds `json:"viewport"`
	RandomFactor float64        `json:"randomFactor"`
	MaxElements  int            `json:"maxElements"`
	MinThreshold int            `json:"minThreshold"`
	Mode         string         `json:"mode"` // PT transport mode (BUS, METRO, TRAM)
}


// PTTelemetry represents PT processing telemetry data
type PTTelemetry struct {
	ProcessID       int    `json:"process_id"`
	TotalFileSize   int64  `json:"total_file_size"`
	BytesRead       int64  `json:"bytes_read"`
	StopsExtracted  int    `json:"stops_extracted"`
	LinesExtracted  int    `json:"lines_extracted"`
	RoutesExtracted int    `json:"routes_extracted"`
	ErrorCount      int    `json:"error_count"`
	LastUpdated     string `json:"last_updated"`
}

// PTProcessor handles PT file processing with telemetry
type PTProcessor struct {
	processID      int
	db             *Database
	app            *App
	telemetry      *PTTelemetry
	telemetryMutex sync.RWMutex
	stopCount      int64
	lineCount      int64
	routeCount     int64
	errorCount     int64
}

// PTExporter handles PT data export operations
type PTExporter struct {
	processID int
	db        *Database
	app       *App
}
