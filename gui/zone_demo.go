package gui

import (
	"fmt"
)

// CreateDemoZones creates sample zones for testing
// This can be called from the frontend to set up demo zones
func (a *App) CreateDemoZones() error {
	demoZones := []*Zone{
		{
			ID:   "downtown",
			Name: "Downtown",
			Polygon: []Point{
				{X: 16.35, Y: 48.20},
				{X: 16.38, Y: 48.20},
				{X: 16.38, Y: 48.22},
				{X: 16.35, Y: 48.22},
			},
		},
		{
			ID:   "suburbs_north",
			Name: "Northern Suburbs",
			Polygon: []Point{
				{X: 16.35, Y: 48.22},
				{X: 16.38, Y: 48.22},
				{X: 16.38, Y: 48.24},
				{X: 16.35, Y: 48.24},
			},
		},
		{
			ID:   "suburbs_south",
			Name: "Southern Suburbs",
			Polygon: []Point{
				{X: 16.35, Y: 48.18},
				{X: 16.38, Y: 48.18},
				{X: 16.38, Y: 48.20},
				{X: 16.35, Y: 48.20},
			},
		},
		{
			ID:   "industrial",
			Name: "Industrial District",
			Polygon: []Point{
				{X: 16.38, Y: 48.20},
				{X: 16.41, Y: 48.20},
				{X: 16.41, Y: 48.22},
				{X: 16.38, Y: 48.22},
			},
		},
		{
			ID:   "residential_east",
			Name: "Eastern Residential",
			Polygon: []Point{
				{X: 16.41, Y: 48.18},
				{X: 16.44, Y: 48.18},
				{X: 16.44, Y: 48.22},
				{X: 16.41, Y: 48.22},
			},
		},
	}
	
	for _, zone := range demoZones {
		zone.CalculateBoundingBox()
		if err := a.db.SaveZone(zone); err != nil {
			return fmt.Errorf("failed to create demo zone %s: %w", zone.ID, err)
		}
	}
	
	return nil
}

// ClearAllZones removes all zones (for testing)
func (a *App) ClearAllZones() error {
	zones, err := a.db.GetAllZones()
	if err != nil {
		return err
	}
	
	for _, zone := range zones {
		if err := a.db.DeleteZone(zone.ID); err != nil {
			return fmt.Errorf("failed to delete zone %s: %w", zone.ID, err)
		}
	}
	
	return nil
}