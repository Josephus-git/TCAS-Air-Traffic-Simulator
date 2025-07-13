package ui

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
)

// TriggerTCAS engages the early warning for planes
const TriggerTCAS = 50.0

// TriggerEngageTCAS displays the planes engaging in TCAS manauver, if successful, green else red
const TriggerEngageTCAS = 15.0

// tcasCore handles the collision resolution logic and ensures a TCASEngagement record
// is stored only once per flight for a given pair of planes.
// It returns the relevant TCASEngagement (either newly created or existing).
func tcasCore(simState *aviation.SimulationState, plane1, plane2 *aviation.Plane) aviation.TCASEngagement {

	// Determine the current flight ID for plane1 (assuming it's the active flight)
	plane1FlightID := ""
	if len(plane1.FlightLog) > 0 {
		plane1FlightID = plane1.FlightLog[len(plane1.FlightLog)-1].FlightID
	}
	if plane1FlightID == "" {
		// This plane is not on an active flight. Defensive check.
		return aviation.TCASEngagement{}
	}

	// Try to find an existing engagement record for this pair and flight ID
	var existingEngagement *aviation.TCASEngagement
	// Check plane1's records first, as it's the primary plane in this context.
	for i := range plane1.TCASEngagementRecords {
		rec := &plane1.TCASEngagementRecords[i]
		// An engagement involves two planes, so check both directions of the pair
		if ((rec.PlaneSerial == plane1.Serial && rec.OtherPlaneSerial == plane2.Serial) ||
			(rec.PlaneSerial == plane2.Serial && rec.OtherPlaneSerial == plane1.Serial)) &&
			rec.FlightID == plane1FlightID && rec.Engaged { // Ensure it's an actual engagement, not just a warning
			existingEngagement = rec
			break
		}
	}

	if existingEngagement != nil {
		// Engagement already exists for this flight and pair, reuse it.
		// Its WillCrash status is already determined when it was first created.
		return *existingEngagement
	}

	// No existing engagement found, so create a new one.
	tcasLog := simState.TCASLog
	shouldCrash := false // This will be determined for the new engagement

	engagementTime := simState.CurrentSimTime

	// Determine shouldCrash based on TCAS capabilities (your existing logic)
	if plane1.TCASCapability == aviation.TCASPerfect && plane2.TCASCapability == aviation.TCASPerfect {
		fmt.Fprintf(tcasLog, "%s TCAS: Both perfect. Averted between %s and %s.\n\n", engagementTime.Format("2006-01-02 15:04:05"), plane1.Serial, plane2.Serial)
		shouldCrash = false
	} else if (plane1.TCASCapability == aviation.TCASPerfect && plane2.TCASCapability == aviation.TCASFaulty) ||
		(plane1.TCASCapability == aviation.TCASFaulty && plane2.TCASCapability == aviation.TCASPerfect) {
		if rand.Float64() < 0.25 {
			shouldCrash = true
			fmt.Fprintf(tcasLog, "%s TCAS: One perfect, one faulty. Collision occurred between %s and %s.\n\n", engagementTime.Format("2006-01-02 15:04:05"), plane1.Serial, plane2.Serial)
		} else {
			fmt.Fprintf(tcasLog, "%s TCAS: One perfect, one faulty. Averted between %s and %s.\n\n", engagementTime.Format("2006-01-02 15:04:05"), plane1.Serial, plane2.Serial)
		}
	} else if plane1.TCASCapability == aviation.TCASFaulty && plane2.TCASCapability == aviation.TCASFaulty {
		if rand.Float64() < 0.5 {
			shouldCrash = true
			fmt.Fprintf(tcasLog, "%s TCAS: Both faulty. Collision occurred between %s and %s.\n\n", engagementTime.Format("2006-01-02 15:04:05"), plane1.Serial, plane2.Serial)
		} else {
			fmt.Fprintf(tcasLog, "%s TCAS: Both faulty. Averted between %s and %s.\n\n", engagementTime.Format("2006-01-02 15:04:05"), plane1.Serial, plane2.Serial)
		}
	}

	newTcasEngagement := aviation.TCASEngagement{
		EngagementID:     fmt.Sprintf("E-%s-%s-%d", plane1.Serial, plane2.Serial, time.Now().UnixNano()), // Unique ID
		FlightID:         plane1FlightID,                                                                 // Associate with the specific flight
		PlaneSerial:      plane1.Serial,
		OtherPlaneSerial: plane2.Serial,
		TimeOfEngagement: engagementTime,
		WillCrash:        shouldCrash, // Determined here, will be consistent for both planes.
		WarningTriggered: false,       // This is an *engagement*, not just a warning
		Engaged:          true,        // Mark as engaged (green/red state)
	}

	// Store the new engagement record in both planes' histories.
	// Since TCASEngagement is a struct (value type), a copy is appended.
	// This is fine as WillCrash is set once at creation.
	plane1.TCASEngagementRecords = append(plane1.TCASEngagementRecords, newTcasEngagement)
	plane2.TCASEngagementRecords = append(plane2.TCASEngagementRecords, newTcasEngagement)

	return newTcasEngagement
}

// planeCurrentPosition calculates the current position of a plane along its flight path.
// This is crucial for real-time animation.
func planeCurrentPosition(plane *aviation.Plane, simTime time.Time) (aviation.Coordinate, bool) {
	if len(plane.FlightLog) == 0 {
		return aviation.Coordinate{}, false
	}

	currentFlight := plane.FlightLog[len(plane.FlightLog)-1]

	if simTime.Before(currentFlight.TakeoffTime) {
		// Plane hasn't taken off yet, return its departure airport's location
		return currentFlight.FlightSchedule.Depature, false
	} else if simTime.After(currentFlight.DestinationArrivalTime) {
		// Plane has landed, return its destination airport's location
		return currentFlight.FlightSchedule.Destination, false
	} else {
		// Plane is in transit
		totalDuration := float64(currentFlight.DestinationArrivalTime.Sub(currentFlight.TakeoffTime))
		elapsedDuration := float64(simTime.Sub(currentFlight.TakeoffTime))

		if totalDuration == 0 { // Avoid division by zero
			return currentFlight.FlightSchedule.Depature, true
		}

		// Interpolation factor (0.0 at takeoff, 1.0 at arrival)
		t := elapsedDuration / totalDuration

		// Linear interpolation for X, Y, Z
		x := currentFlight.FlightSchedule.Depature.X + t*(currentFlight.FlightSchedule.Destination.X-currentFlight.FlightSchedule.Depature.X)
		y := currentFlight.FlightSchedule.Depature.Y + t*(currentFlight.FlightSchedule.Destination.Y-currentFlight.FlightSchedule.Depature.Y)
		z := currentFlight.FlightSchedule.Depature.Z + t*(currentFlight.FlightSchedule.Destination.Z-currentFlight.FlightSchedule.Depature.Z)

		return aviation.Coordinate{X: x, Y: y, Z: z}, true
	}
}
