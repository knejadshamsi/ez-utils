package steps

import (
	"math"
)

// ConvertMTM8ToWGS84 converts from Modified Transverse Mercator (MTM) Zone 8 to WGS84 coordinates
// MTM8 is a projection commonly used in Quebec, Canada
func ConvertMTM8ToWGS84(x, y float64) (longitude, latitude float64) {
	// MTM8 parameters (Quebec Zone 8)
	const (
		k0            = 0.9999                  // Scale factor
		centralLat    = 0.0                     // Central latitude (in radians)
		centralLon    = -73.5 * math.Pi / 180.0 // Central longitude (in radians) for MTM8
		falseEasting  = 304800.0                // False easting (meters)
		falseNorthing = 0.0                     // False northing (meters)

		// WGS84 ellipsoid parameters
		a = 6378137.0           // Semi-major axis
		f = 1.0 / 298.257223563 // Flattening
	)

	// Derived constants
	e2 := 2.0*f - f*f // Eccentricity squared
	ep2 := e2 / (1.0 - e2)

	// Adjust for false easting/northing
	x = x - falseEasting
	y = y - falseNorthing

	// Inverse transverse mercator projection
	// Step 1: Adjust for central meridian
	M := y / k0

	// Compute footpoint latitude
	mu := M / (a * (1.0 - e2/4.0 - 3.0*e2*e2/64.0 - 5.0*e2*e2*e2/256.0))

	// Compute latitude with series expansion
	e1 := (1.0 - math.Sqrt(1.0-e2)) / (1.0 + math.Sqrt(1.0-e2))
	phi1 := mu + (3.0*e1/2.0-27.0*e1*e1*e1/32.0)*math.Sin(2.0*mu) +
		(21.0*e1*e1/16.0-55.0*e1*e1*e1*e1/32.0)*math.Sin(4.0*mu) +
		(151.0*e1*e1*e1/96.0)*math.Sin(6.0*mu) +
		(1097.0*e1*e1*e1*e1/512.0)*math.Sin(8.0*mu)

	// Compute auxiliary values
	sinPhi := math.Sin(phi1)
	cosPhi := math.Cos(phi1)
	tanPhi := sinPhi / cosPhi

	n := a / math.Sqrt(1.0-e2*sinPhi*sinPhi)
	t := tanPhi * tanPhi
	c := ep2 * cosPhi * cosPhi
	r := a * (1.0 - e2) / math.Pow(1.0-e2*sinPhi*sinPhi, 1.5)
	d := x / (n * k0)

	// Compute latitude and longitude
	lat := phi1 - (n*tanPhi/r)*(d*d/2.0-(5.0+3.0*t+10.0*c-4.0*c*c-9.0*ep2)*d*d*d*d/24.0+
		(61.0+90.0*t+298.0*c+45.0*t*t-252.0*ep2-3.0*c*c)*d*d*d*d*d*d/720.0)

	lon := centralLon + (d-(1.0+2.0*t+c)*d*d*d/6.0+
		(5.0-2.0*c+28.0*t-3.0*c*c+8.0*ep2+24.0*t*t)*d*d*d*d*d/120.0)/cosPhi

	// Convert from radians to degrees
	latitude = lat * 180.0 / math.Pi
	longitude = lon * 180.0 / math.Pi

	// Round to 6 decimal places for consistency
	return RoundCoordinates(longitude, latitude)
}

// RoundCoordinates rounds the coordinates to 6 decimal places for consistency
func RoundCoordinates(x, y float64) (float64, float64) {
	factor := 1000000.0
	return math.Round(x*factor) / factor, math.Round(y*factor) / factor
}
