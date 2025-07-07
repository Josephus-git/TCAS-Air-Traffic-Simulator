package aviation

import (
	"fmt"
	"math"
)

// Most functions here are helpers to implement the finding of closest approach between flight paths

// Coordinate represents a 3D Coordinate
// may be changed to latitude logitude altitude
type Coordinate struct {
	X, Y, Z float64
}

// Coordinate.String() helper for better print output
func (c Coordinate) String() string {
	return fmt.Sprintf("(%.0f, %.0f, %.0f)", c.X, c.Y, c.Z)
}

// add returns the sum of two Coordinates (3D vector)
func (c Coordinate) add(other Coordinate) Coordinate {
	return Coordinate{c.X + other.X, c.Y + other.Y, c.Z + other.Z}
}

// subtract returns the difference of two Coordinates (3D vector)
func (c Coordinate) subtract(other Coordinate) Coordinate {
	return Coordinate{c.X - other.X, c.Y - other.Y, c.Z - other.Z}
}

// mulScalar returns the Coordinates scaled by a scalar
func (c Coordinate) mulScalar(s float64) Coordinate {
	return Coordinate{c.X * s, c.Y * s, c.Z * s}
}

// dot returns the dot product of two Coordinates
func (c Coordinate) dot(other Coordinate) float64 {
	return (c.X * other.X) + (c.Y * other.Y) + (c.Z * other.Z)
}

// clamp limits a value to a specific range, ensuring it falls within a minimum and maximum boundary
func clamp(val, min, max float64) float64 {
	return math.Max(min, math.Min(val, max))
}
