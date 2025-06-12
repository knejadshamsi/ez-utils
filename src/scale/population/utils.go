package population

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"strings"
)

// File utilities

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileInfo returns file information
func GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// CreateTempFile creates a temporary file in the system temp directory
func CreateTempFile(prefix string) (*os.File, error) {
	return os.CreateTemp("", prefix)
}

// XML utilities

// ExtractPersonID extracts the person ID from a person XML element
func ExtractPersonID(personXML string) (string, error) {
	type Person struct {
		XMLName xml.Name `xml:"person"`
		ID      string   `xml:"id,attr"`
	}
	
	var p Person
	if err := xml.Unmarshal([]byte(personXML), &p); err != nil {
		return "", err
	}
	
	return p.ID, nil
}

// ExtractFirstActivityCoordinates extracts coordinates from the first activity in a person XML
func ExtractFirstActivityCoordinates(personXML string) (*Coordinates, error) {
	type Activity struct {
		X float64 `xml:"x,attr"`
		Y float64 `xml:"y,attr"`
	}
	
	type Plan struct {
		Selected   string     `xml:"selected,attr"`
		Activities []Activity `xml:"activity"`
	}
	
	type Person struct {
		Plans []Plan `xml:"plan"`
	}
	
	var p Person
	if err := xml.Unmarshal([]byte(personXML), &p); err != nil {
		return nil, err
	}
	
	// Find selected plan or use first plan
	var selectedPlan *Plan
	for i := range p.Plans {
		if p.Plans[i].Selected == "yes" {
			selectedPlan = &p.Plans[i]
			break
		}
	}
	
	if selectedPlan == nil && len(p.Plans) > 0 {
		selectedPlan = &p.Plans[0]
	}
	
	if selectedPlan == nil || len(selectedPlan.Activities) == 0 {
		return nil, fmt.Errorf("no activities found in person")
	}
	
	// Return first activity coordinates
	return &Coordinates{
		X: selectedPlan.Activities[0].X,
		Y: selectedPlan.Activities[0].Y,
	}, nil
}

// IsPersonElement checks if a line contains a person element start
func IsPersonElement(line string) bool {
	return strings.Contains(line, "<person")
}

// IsPersonEndElement checks if a line contains a person element end
func IsPersonEndElement(line string) bool {
	return strings.Contains(line, "</person>")
}

// Coordinate utilities

// CalculateDistance calculates Euclidean distance between two points
func CalculateDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// CalculateGridCell calculates which grid cell a coordinate belongs to
func CalculateGridCell(x, y, cellSize, originX, originY float64) (int, int) {
	col := int((x - originX) / cellSize)
	row := int((y - originY) / cellSize)
	return row, col
}

// Performance utilities

// GetMemoryUsage returns current memory usage in MB
func GetMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc / 1024 / 1024
}

// GetCPUCount returns the number of available CPU cores
func GetCPUCount() int {
	return runtime.NumCPU()
}

// FormatBytes formats bytes into human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Compression utilities

// CompressString compresses a string using gzip
func CompressString(s string) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	
	if _, err := gz.Write([]byte(s)); err != nil {
		return nil, err
	}
	
	if err := gz.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// DecompressString decompresses a gzip compressed byte array to string
func DecompressString(data []byte) (string, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer r.Close()
	
	result, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	
	return string(result), nil
}

// Buffer pool utilities

// BufferPool manages a pool of reusable buffers
type BufferPool struct {
	pool chan *bufio.Writer
	size int
}

// NewBufferPool creates a new buffer pool
func NewBufferPool(size int) *BufferPool {
	pool := make(chan *bufio.Writer, size)
	return &BufferPool{
		pool: pool,
		size: size,
	}
}

// Get retrieves a buffer from the pool or creates a new one
func (bp *BufferPool) Get(w io.Writer) *bufio.Writer {
	select {
	case buf := <-bp.pool:
		buf.Reset(w)
		return buf
	default:
		return bufio.NewWriter(w)
	}
}

// Put returns a buffer to the pool
func (bp *BufferPool) Put(buf *bufio.Writer) {
	buf.Flush()
	select {
	case bp.pool <- buf:
		// Buffer returned to pool
	default:
		// Pool is full, let buffer be garbage collected
	}
}