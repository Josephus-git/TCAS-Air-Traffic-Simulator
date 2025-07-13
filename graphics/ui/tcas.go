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

// tcasCore handles the collision resolution logic and returns a TCASEngagement record.
// It also updates the plane's TCASEngagementRecords.
func tcasCore(simState *aviation.SimulationState, plane1, plane2 *aviation.Plane) aviation.TCASEngagement {
	tcasLog := simState.TCASLog
	shouldCrash := false

	// Assume engagement is happening at current simulation time for logging purposes
	engagementTime := simState.CurrentSimTime
	plane1Flight := plane1.FlightLog[len(plane1.FlightLog)-1] // Assuming current flight is the last one

	if plane1.TCASCapability == aviation.TCASPerfect && plane2.TCASCapability == aviation.TCASPerfect {
		// Both perfect, no crash
		fmt.Fprintf(tcasLog, "%s TCAS: Both planes have perfect TCAS. Collision averted between %s and %s.\n\n",
			engagementTime.Format("2006-01-02 15:04:05"), plane1.Serial, plane2.Serial)
		shouldCrash = false
	} else if (plane1.TCASCapability == aviation.TCASPerfect && plane2.TCASCapability == aviation.TCASFaulty) ||
		(plane1.TCASCapability == aviation.TCASFaulty && plane2.TCASCapability == aviation.TCASPerfect) {
		// One perfect, one faulty: 25% chance of crash
		if rand.Float64() < 0.25 {
			shouldCrash = true
			fmt.Fprintf(tcasLog, "%s TCAS: One perfect, one faulty TCAS. Collision occurred between %s and %s.\n\n",
				engagementTime.Format("15:04:05"), plane1.Serial, plane2.Serial)
		} else {
			fmt.Fprintf(tcasLog, "%s TCAS: One perfect, one faulty TCAS. Collision narrowly averted between %s and %s.\n\n",
				engagementTime.Format("15:04:05"), plane1.Serial, plane2.Serial)
		}
	} else if plane1.TCASCapability == aviation.TCASFaulty && plane2.TCASCapability == aviation.TCASFaulty {
		// Both faulty: 50% chance of crash
		if rand.Float64() < 0.5 {
			shouldCrash = true
			fmt.Fprintf(tcasLog, "%s TCAS: Two faulty TCAS. Collision occurred between %s and %s.\n\n",
				engagementTime.Format("15:04:05"), plane1.Serial, plane2.Serial)
		} else {
			fmt.Fprintf(tcasLog, "%s TCAS: Two faulty TCAS. Collision narrowly averted between %s and %s.\n\n",
				engagementTime.Format("15:04:05"), plane1.Serial, plane2.Serial)
		}
	}

	// Create and return the TCAS Engagement record
	newTcasEngagement := aviation.TCASEngagement{
		EngagementID:     fmt.Sprintf("E-%s-%s-%d", plane1.Serial, plane2.Serial, time.Now().UnixNano()), // Unique ID for engagement
		FlightID:         plane1Flight.FlightID,
		PlaneSerial:      plane1.Serial,
		OtherPlaneSerial: plane2.Serial,
		TimeOfEngagement: engagementTime,
		WillCrash:        shouldCrash,
		WarningTriggered: false, // This will be set true if the orange warning was shown, but here it's about engagement itself
		Engaged:          true,  // Mark as engaged (green/red state)
	}

	// Store the engagement record in both planes' history
	plane1.TCASEngagementRecords = append(plane1.TCASEngagementRecords, newTcasEngagement)
	plane2.TCASEngagementRecords = append(plane2.TCASEngagementRecords, newTcasEngagement)

	// Set current engagement for both planes
	plane1.CurrentTCASEngagement = &newTcasEngagement
	plane2.CurrentTCASEngagement = &newTcasEngagement

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
		// The UI should handle removing the plane if it's considered fully landed
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
