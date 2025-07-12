package aviation

import (
	"math"
	"math/rand"
	"time"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/util"
)

// TCASCapability defines the type of TCAS system installed on a plane
type TCASCapability int

// CruiseSpeed defines the speed of all planes
const CruiseSpeed = 10.0

// Plane represents an aircraft with its key operational details and flight history.
type Plane struct {
	Serial                 string
	PlaneInFlight          bool
	CruiseSpeed            float64
	FlightLog              []Flight
	TCASCapability         TCASCapability
	TCASEngagementRecords  []TCASEngagement
	CurrentTCASEngagements []TCASEngagement
}

const (
	TCASPerfect TCASCapability = iota // 0
	TCASFaulty
)

// createPlane initializes and returns a new Plane struct with a generated serial number.
func createPlane(planeCount int) Plane {
	// Randomly assign TCAS capability
	capability := TCASPerfect
	if rand.Float64() < 0.25 { // 25% chance of faulty TCAS
		capability = TCASFaulty
	}

	return Plane{
		Serial:         util.GenerateSerialNumber(planeCount, "p"),
		PlaneInFlight:  false,
		CruiseSpeed:    CruiseSpeed,
		FlightLog:      []Flight{},
		TCASCapability: capability,
	}
}

// getPlanePosition calculates the plane's position at a given checkTime.
// It interpolates the position along the current flight path based on the elapsed time.
func (flight Flight) getPlanePosition(checkTime time.Time) Coordinate {
	var pct float64

	// Check if the checkTime is within the flight's active period
	if checkTime.After(flight.TakeoffTime) && checkTime.Before(flight.DestinationArrivalTime) {
		totalDuration := flight.DestinationArrivalTime.Sub(flight.TakeoffTime)
		elapsedDuration := checkTime.Sub(flight.TakeoffTime) // Use checkTime, not simState.CurrentSimTime
		if totalDuration > 0 {
			pct = float64(elapsedDuration) / float64(totalDuration)
			// Clamp pct between 0 and 1 to ensure it's within the path bounds
			pct = clamp(pct, 0.0, 1.0)
		} else {
			pct = 0.0 // If total duration is 0, the plane hasn't moved or arrived instantly
		}
	} else if checkTime.After(flight.DestinationArrivalTime) {
		// If checkTime is after arrival, the plane is at its destination
		return flight.FlightPath.Destination
	} else {
		// If checkTime is before takeoff, the plane is at its departure point
		return flight.FlightPath.Depature
	}

	// Calculate the vector from departure to destination
	pathVector := flight.FlightPath.Destination.subtract(flight.FlightPath.Depature)

	// Calculate the current position by adding a fraction of the pathVector to the departure point
	currentPosition := flight.FlightPath.Depature.add(pathVector.mulScalar(pct))

	return currentPosition
}

// Distance calculates the Euclidean Distance between two 3D coordinates.
func Distance(p1, p2 Coordinate) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2) + math.Pow(p1.Z-p2.Z, 2))
}
