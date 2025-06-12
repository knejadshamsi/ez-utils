package functions

import (
	"ez-utils/src/scale/population"
)

// CalculateExpandedBounds calculates minimum bounds to fit all coordinates
func CalculateExpandedBounds(currentBounds population.GridBounds, coords []population.Coordinates) population.GridBounds {
	if len(coords) == 0 {
		return currentBounds
	}
	
	// Start with current bounds
	newBounds := currentBounds
	
	// Expand to accommodate all coordinates
	for _, coord := range coords {
		if coord.X < newBounds.MinX {
			newBounds.MinX = coord.X
		}
		if coord.X > newBounds.MaxX {
			newBounds.MaxX = coord.X
		}
		if coord.Y < newBounds.MinY {
			newBounds.MinY = coord.Y
		}
		if coord.Y > newBounds.MaxY {
			newBounds.MaxY = coord.Y
		}
	}
	
	return newBounds
}

// NeedsExpansion checks if the new bounds require expansion
func NeedsExpansion(currentBounds, newBounds population.GridBounds) bool {
	return newBounds.MinX < currentBounds.MinX ||
		   newBounds.MinY < currentBounds.MinY ||
		   newBounds.MaxX > currentBounds.MaxX ||
		   newBounds.MaxY > currentBounds.MaxY
}

// ApplyExpansionMargins adds safety margins to prevent frequent re-expansions
func ApplyExpansionMargins(bounds population.GridBounds, expansionMargin float64) population.GridBounds {
	return population.GridBounds{
		MinX: bounds.MinX - expansionMargin,
		MinY: bounds.MinY - expansionMargin,
		MaxX: bounds.MaxX + expansionMargin,
		MaxY: bounds.MaxY + expansionMargin,
		CellSize: bounds.CellSize, // Fix: Preserve CellSize during margin expansion
	}
}

// CalculateGridCell calculates which grid cell a coordinate belongs to
func CalculateGridCell(x, y float64, bounds population.GridBounds) (int, int) {
	col := int((x - bounds.MinX) / bounds.CellSize)
	row := int((y - bounds.MinY) / bounds.CellSize)
	return row, col
}

// GetCellBounds returns the bounds of a specific grid cell
func GetCellBounds(row, col int, bounds population.GridBounds) population.GridBounds {
	minX := bounds.MinX + float64(col)*bounds.CellSize
	minY := bounds.MinY + float64(row)*bounds.CellSize
	
	return population.GridBounds{
		MinX: minX,
		MinY: minY,
		MaxX: minX + bounds.CellSize,
		MaxY: minY + bounds.CellSize,
		CellSize: bounds.CellSize, // Fix: Preserve CellSize in cell bounds
	}
}

// IsCoordinateInBounds checks if a coordinate is within bounds
func IsCoordinateInBounds(x, y float64, bounds population.GridBounds) bool {
	return x >= bounds.MinX && x <= bounds.MaxX &&
		   y >= bounds.MinY && y <= bounds.MaxY
}