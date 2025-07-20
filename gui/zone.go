package gui

import (
	"fmt"
	"math"
)

// Point represents a 2D coordinate
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Zone represents a geographic zone with a polygon boundary
type Zone struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Polygon  []Point `json:"polygon"`
	BoundingBox struct {
		MinX float64 `json:"minX"`
		MinY float64 `json:"minY"`
		MaxX float64 `json:"maxX"`
		MaxY float64 `json:"maxY"`
	} `json:"boundingBox"`
}

// CalculateBoundingBox computes the bounding box for a zone
func (z *Zone) CalculateBoundingBox() {
	if len(z.Polygon) == 0 {
		return
	}
	
	z.BoundingBox.MinX = z.Polygon[0].X
	z.BoundingBox.MaxX = z.Polygon[0].X
	z.BoundingBox.MinY = z.Polygon[0].Y
	z.BoundingBox.MaxY = z.Polygon[0].Y
	
	for _, p := range z.Polygon[1:] {
		z.BoundingBox.MinX = math.Min(z.BoundingBox.MinX, p.X)
		z.BoundingBox.MaxX = math.Max(z.BoundingBox.MaxX, p.X)
		z.BoundingBox.MinY = math.Min(z.BoundingBox.MinY, p.Y)
		z.BoundingBox.MaxY = math.Max(z.BoundingBox.MaxY, p.Y)
	}
}

// ContainsPoint checks if a point is inside the zone using ray casting algorithm
// This implementation handles edge cases properly:
// - Points on vertices are considered inside
// - Points on edges are considered inside  
// - Horizontal edges are handled correctly
func (z *Zone) ContainsPoint(x, y float64) bool {
	// Quick bounding box check first
	if x < z.BoundingBox.MinX || x > z.BoundingBox.MaxX ||
	   y < z.BoundingBox.MinY || y > z.BoundingBox.MaxY {
		return false
	}
	
	// Ray casting algorithm
	n := len(z.Polygon)
	if n < 3 {
		return false
	}
	
	inside := false
	p1 := z.Polygon[0]
	
	for i := 1; i <= n; i++ {
		p2 := z.Polygon[i%n]
		
		// Check if point is on the edge
		if isPointOnSegment(x, y, p1.X, p1.Y, p2.X, p2.Y) {
			return true // Points on edges are inside
		}
		
		// Ray casting check
		if y > math.Min(p1.Y, p2.Y) && y <= math.Max(p1.Y, p2.Y) {
			if x <= math.Max(p1.X, p2.X) {
				if p1.Y != p2.Y {
					xinters := (y-p1.Y)*(p2.X-p1.X)/(p2.Y-p1.Y) + p1.X
					if p1.X == p2.X || x <= xinters {
						inside = !inside
					}
				}
			}
		}
		p1 = p2
	}
	
	return inside
}

// isPointOnSegment checks if a point lies on a line segment
func isPointOnSegment(px, py, x1, y1, x2, y2 float64) bool {
	const epsilon = 1e-10
	
	// Check if point is within bounding box of segment
	if px < math.Min(x1, x2)-epsilon || px > math.Max(x1, x2)+epsilon ||
	   py < math.Min(y1, y2)-epsilon || py > math.Max(y1, y2)+epsilon {
		return false
	}
	
	// Check if point is collinear with segment
	// Using cross product: (p-p1) × (p2-p1) = 0
	cross := (px-x1)*(y2-y1) - (py-y1)*(x2-x1)
	return math.Abs(cross) < epsilon
}

// ParseCoordinates extracts x,y coordinates from a comma-separated string
func ParseCoordinates(coords string) (float64, float64, error) {
	var x, y float64
	_, err := fmt.Sscanf(coords, "%f,%f", &x, &y)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid coordinates format: %s", coords)
	}
	return x, y, nil
}

// ToGeoJSON converts the zone to a GeoJSON Feature
func (z *Zone) ToGeoJSON() map[string]interface{} {
	coordinates := make([][]float64, len(z.Polygon))
	for i, p := range z.Polygon {
		coordinates[i] = []float64{p.X, p.Y}
	}
	
	// Close the polygon by adding first point at the end if needed
	if len(coordinates) > 0 {
		first := coordinates[0]
		last := coordinates[len(coordinates)-1]
		if first[0] != last[0] || first[1] != last[1] {
			coordinates = append(coordinates, first)
		}
	}
	
	return map[string]interface{}{
		"type": "Feature",
		"properties": map[string]interface{}{
			"id":   z.ID,
			"name": z.Name,
		},
		"geometry": map[string]interface{}{
			"type":        "Polygon",
			"coordinates": []interface{}{coordinates},
		},
	}
}

// ZoneFromGeoJSON creates a Zone from a GeoJSON Feature
func ZoneFromGeoJSON(data map[string]interface{}) (*Zone, error) {
	zone := &Zone{}
	
	// Extract properties
	if props, ok := data["properties"].(map[string]interface{}); ok {
		if id, ok := props["id"].(string); ok {
			zone.ID = id
		}
		if name, ok := props["name"].(string); ok {
			zone.Name = name
		}
	}
	
	// Extract geometry
	if geom, ok := data["geometry"].(map[string]interface{}); ok {
		if geomType, ok := geom["type"].(string); ok && geomType == "Polygon" {
			if coords, ok := geom["coordinates"].([]interface{}); ok && len(coords) > 0 {
				if ring, ok := coords[0].([]interface{}); ok {
					zone.Polygon = make([]Point, 0, len(ring))
					for _, coord := range ring {
						if c, ok := coord.([]interface{}); ok && len(c) >= 2 {
							if x, ok := c[0].(float64); ok {
								if y, ok := c[1].(float64); ok {
									zone.Polygon = append(zone.Polygon, Point{X: x, Y: y})
								}
							}
						}
					}
				}
			}
		}
	}
	
	// Remove duplicate last point if it's the same as first
	if len(zone.Polygon) > 1 {
		first := zone.Polygon[0]
		last := zone.Polygon[len(zone.Polygon)-1]
		if first.X == last.X && first.Y == last.Y {
			zone.Polygon = zone.Polygon[:len(zone.Polygon)-1]
		}
	}
	
	zone.CalculateBoundingBox()
	return zone, nil
}

