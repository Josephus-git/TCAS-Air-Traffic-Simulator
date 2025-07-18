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

// TCASCapability defines the operational state of a plane's TCAS system.
const (
	TCASPerfect TCASCapability = iota // 0
	TCASFaulty
)

// TCASEngagement represents a recorded interaction between two planes, tracking its ID, involved aircraft,
// time, and the nature of the engagement (e.g., warning, crash prediction).
type TCASEngagement struct {
	EngagementID     string
	FlightID         string
	PlaneSerial      string
	OtherPlaneSerial string
	TimeOfEngagement time.Time
	WillCrash        bool
	WarningTriggered bool // Added to track if the orange warning has been shown
	Engaged          bool // Added to track if the green/red engagement has occurred
}

// Plane represents an aircraft with its key operational details and flight history.
type Plane struct {
	Serial                string
	PlaneInFlight         bool
	CruiseSpeed           float64
	FlightLog             []Flight
	TCASCapability        TCASCapability
	TCASEngagementRecords []TCASEngagement
	CurrentTCASEngagement *TCASEngagement
}

// createPlane initializes and returns a new Plane struct with a generated serial number.
func createPlane(planeCount int) *Plane {
	// Randomly assign TCAS capability
	capability := TCASPerfect
	if rand.Float64() < 0.25 { // 25% chance of faulty TCAS
		capability = TCASFaulty
	}

	return &Plane{
		Serial:         util.GenerateSerialNumber(planeCount, "p"),
		PlaneInFlight:  false,
		CruiseSpeed:    CruiseSpeed,
		FlightLog:      []Flight{},
		TCASCapability: capability,
	}
}

// Distance calculates the Euclidean Distance between two 3D coordinates.
func Distance(p1, p2 Coordinate) float64 {
	return math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2) + math.Pow(p1.Z-p2.Z, 2))
}
