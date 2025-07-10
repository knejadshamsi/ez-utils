package gui

import (
	"encoding/xml"
)

// No shared constants currently - each workflow defines its own constants
// to avoid conflicts and maintain independence

// Read implements io.Reader interface for CountingReader
// Used by: population workflow (population_process.go), network workflow (network_process.go)
func (cr *CountingReader) Read(p []byte) (n int, err error) {
	n, err = cr.reader.Read(p)
	cr.mu.Lock()
	cr.bytesRead += int64(n)
	cr.mu.Unlock()
	return n, err
}

// BytesRead returns the total bytes read
// Used by: population workflow (population_process.go), network workflow (network_process.go)
func (cr *CountingReader) BytesRead() int64 {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	return cr.bytesRead
}

// extractAttribute extracts an attribute value from XML attributes
// Used by: network workflow (network_process.go)
// Note: Currently only used by network workflow, but kept here for potential future use by other workflows
func extractAttribute(attrs []xml.Attr, name string) string {
	for _, attr := range attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}