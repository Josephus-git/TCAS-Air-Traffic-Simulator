package aviation

import (
	"math"
	"math/rand"
	"time"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/util"
)

// TCASCapability defines the type of TCAS system installed on a plane
type TCASCapability int

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
		CruiseSpeed:    10,
		FlightLog:      []Flight{},
		TCASCapability: capability,
	}
}

// Distance calculates the Euclidean Distance between two 3D coordinates.
func Distance(p1, p2 Coordinate) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2) + math.Pow(p1.Z-p2.Z, 2))
}

// GetClosestApproachDetails calculates the time and minimum Distance at which two planes will be closest during their respective flights.
func (f1 Flight) GetClosestApproachDetails(f2 Flight) (closestTime time.Time, distanceBetweenPlanesatCA float64) {
	flight1ClosestCoord, flight2ClosestCoord := FindClosestApproachDuringTransit(f1.FlightSchedule, f2.FlightSchedule)

	flight1Distance := Distance(f1.FlightSchedule.Depature, f1.FlightSchedule.Destination)
	distBtwDepatureAndClosestApproachForFlight1 := Distance(f1.FlightSchedule.Depature, flight1ClosestCoord)

	f1fractionofCA := distBtwDepatureAndClosestApproachForFlight1 / flight1Distance

	totalFlightDuration1 := f1.DestinationArrivalTime.Sub(f1.TakeoffTime)
	closestTime = f1.TakeoffTime.Add(time.Duration(float64(totalFlightDuration1) * f1fractionofCA))

	distanceBetweenPlanesatCA = Distance(flight1ClosestCoord, flight2ClosestCoord)

	return closestTime, distanceBetweenPlanesatCA
}
