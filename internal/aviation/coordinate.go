package aviation

import (
	"fmt"
)

// All functions here are helpers to implement calculations on flight paths

// Coordinate represents a 3D Coordinate
// may be changed to latitude logitude altitude
type Coordinate struct {
	X, Y, Z float64
}

// Coordinate.String() helper for better print output
func (c Coordinate) String() string {
	return fmt.Sprintf("(%.0f, %.0f, %.0f)", c.X, c.Y, c.Z)
}
