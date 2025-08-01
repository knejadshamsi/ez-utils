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


// IsPointInPolygon checks if a point is inside a polygon using ray casting algorithm
// Used by: population filtering with zones
func IsPointInPolygon(point Point, polygon []Point) bool {
	if len(polygon) < 3 {
		return false
	}
	
	inside := false
	p1 := polygon[0]
	
	for i := 1; i <= len(polygon); i++ {
		p2 := polygon[i%len(polygon)]
		
		minY, maxY := p1.Y, p2.Y
		if p1.Y > p2.Y {
			minY, maxY = p2.Y, p1.Y
		}
		
		if point.Y > minY && point.Y <= maxY {
			maxX := p1.X
			if p2.X > maxX {
				maxX = p2.X
			}
			
			if point.X <= maxX {
				if p1.Y != p2.Y {
					xinters := (point.Y-p1.Y)*(p2.X-p1.X)/(p2.Y-p1.Y) + p1.X
					if p1.X == p2.X || point.X <= xinters {
						inside = !inside
					}
				}
			}
		}
		p1 = p2
	}
	
	return inside
}