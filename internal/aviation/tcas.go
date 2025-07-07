package aviation

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/util"
)

type TCASEngagement struct {
	EngagementID     string
	FlightID         string
	PlaneSerial      string
	OtherPlaneSerial string
	TimeOfEngagement time.Time
	WillCrash        bool
	WarningTriggered bool
}

// CollisionThreshold defines the maximum distance (in units) at which two planes are considered to be in a collision course.
const CollisionThreshold = 5

// CheckPlaneStatusAtTime checks the status of a plane at a specific time based on its flight log.
func checkPlaneStatusAtTime(p Plane, checkTime time.Time) string {
	flight := p.FlightLog[len(p.FlightLog)-1]

	if checkTime.After(flight.DestinationArrivalTime) {
		// If checkTime is after arrival, plane has landed for this flight
		return "landed or still landing"
	}
	if checkTime.After(flight.TakeoffTime) && checkTime.Before(flight.DestinationArrivalTime) {
		// Plane is in transit during this flight
		// We need a more granular check for "about to land"
		// This simplified logic assumes "in transit" or "landed" for past times.
		// For future "about to land" it would be based on calculation.
		// Given the 3-second prior check, if it's already "about to land" it should be reflected in FlightStatus.
		if flight.FlightStatus == "about to land" {
			return "about to land"
		}
		return "in transit"
	}
	if checkTime.Equal(flight.TakeoffTime) {
		return "taking off"
	}
	if checkTime.Equal(flight.DestinationArrivalTime) {
		return "arriving"
	}
	return "parked/unknown" // If no flight matches the time, assume parked or not in a known flight.
}

// tcas detects potential mid-air collisions between a given plane (the one about to take off)
// and other planes currently in flight. If a collision is predicted under specific conditions,
// it triggers an emergency stop.
//
// Parameters:
//
//	plane: The plane attempting to take off.
//	tcasLog: The file pointer for logging planes condition before going on the flight.
func (plane Plane) tcas(simState *SimulationState, tcasLog *os.File) []TCASEngagement {
	planeFlight := plane.FlightLog[len(plane.FlightLog)-1]
	simState.Mu.Lock() // Lock the simulation state to safely access PlanesInFlight
	planesInFlight := simState.PlanesInFlight
	simState.Mu.Unlock() // Release the lock after copying the slice

	fmt.Fprintf(tcasLog, "%s TCAS: Plane %s (%v) is checking for conflicts before takeoff.\n\n",
		time.Now().Format("2006-01-02 15:04:05"), plane.Serial, plane.TCASCapability)

	tcasEngagementSlice := []TCASEngagement{}
	for _, otherPlane := range planesInFlight {
		// Skip checking against itself
		if plane.Serial == otherPlane.Serial {
			continue
		}

		// Ensure the other plane is indeed in flight (should be true for planesInFlight list, but good check)
		if !otherPlane.PlaneInFlight {
			continue
		}

		// Find the current active flight for the otherPlane
		otherPlaneFlight := otherPlane.FlightLog[len(otherPlane.FlightLog)-1]

		// Calculate Closest Approach Details between the potential flight paths
		closestTime, distanceAtCA := planeFlight.GetClosestApproachDetails(otherPlaneFlight)

		// Check the other plane's status at the closest approach
		otherPlaneStatusAtCheckTime := checkPlaneStatusAtTime(otherPlane, closestTime)

		// Condition 1: If otherPlane has landed, is about to land or at different flight altitudes, no collision concern from altitude difference
		if otherPlaneStatusAtCheckTime == "landed or still landing" || otherPlaneStatusAtCheckTime == "about to land" || otherPlaneFlight.CruisingAltitude != planeFlight.CruisingAltitude {
			fmt.Fprintf(tcasLog, "%s TCAS: Plane %s's flight path %s and Plane %s's flight path %s have closest approach (%.2f units at %v), but no worries: Other plane status is '%s' or different altitude.\n\n",
				time.Now().Format("15:04:05"), plane.Serial, planeFlight.FlightID, otherPlane.Serial, otherPlaneFlight.FlightID, distanceAtCA, closestTime.Format("15:04:05"), otherPlaneStatusAtCheckTime)
			continue
		}

		// Condition 2: Check if collision distance threshold is met
		if distanceAtCA < CollisionThreshold {
			fmt.Fprintf(tcasLog, "%s TCAS ALERT: Potential collision detected between Plane %s (TCAS: %v) and Plane %s (TCAS: %v). Closest approach: %.2f units at %v.\n\n",
				time.Now().Format("15:04:05"), plane.Serial, plane.TCASCapability, otherPlane.Serial, otherPlane.TCASCapability, distanceAtCA, closestTime.Format("15:04:05"))

			// Collision Resolution based on TCAS capabilities
			shouldCrash := false

			if plane.TCASCapability == TCASPerfect && otherPlane.TCASCapability == TCASPerfect {
				// Both perfect, no crash
				fmt.Fprintf(tcasLog, "%s TCAS: Both planes have perfect TCAS. Collision averted between %s and %s.\n\n",
					time.Now().Format("2006-01-02 15:04:05"), plane.Serial, otherPlane.Serial)
				shouldCrash = false
			} else if (plane.TCASCapability == TCASPerfect && otherPlane.TCASCapability == TCASFaulty) ||
				(plane.TCASCapability == TCASFaulty && otherPlane.TCASCapability == TCASPerfect) {
				// One perfect, one faulty: 50% chance of crash
				if rand.Float64() < 0.25 {
					shouldCrash = true
				} else {
					fmt.Fprintf(tcasLog, "%s TCAS: One perfect, one faulty TCAS. Collision narrowly averted between %s and %s.\n\n",
						time.Now().Format("15:04:05"), plane.Serial, otherPlane.Serial)
				}
			} else if plane.TCASCapability == TCASFaulty && otherPlane.TCASCapability == TCASFaulty {
				if rand.Float64() < 0.5 {
					shouldCrash = true
				} else {
					fmt.Fprintf(tcasLog, "%s TCAS: Two faulty TCAS. Collision narrowly averted between %s and %s.\n\n",
						time.Now().Format("15:04:05"), plane.Serial, otherPlane.Serial)
				}
			}

			if shouldCrash {
				newTcasEngagement := TCASEngagement{
					EngagementID:     plane.Serial + util.GenerateSerialNumber(len(plane.TCASEngagementRecords), "e"),
					FlightID:         planeFlight.FlightID,
					PlaneSerial:      plane.Serial,
					OtherPlaneSerial: otherPlane.Serial,
					TimeOfEngagement: closestTime,
					WillCrash:        true,
				}
				tcasEngagementSlice = append(tcasEngagementSlice, newTcasEngagement)
				continue
			}
			newTcasEngagement := TCASEngagement{
				EngagementID:     plane.Serial + util.GenerateSerialNumber(len(plane.TCASEngagementRecords), "e"),
				FlightID:         planeFlight.FlightID,
				PlaneSerial:      plane.Serial,
				OtherPlaneSerial: otherPlane.Serial,
				TimeOfEngagement: closestTime,
				WillCrash:        false,
			}
			tcasEngagementSlice = append(tcasEngagementSlice, newTcasEngagement)
		}
	}
	sort.Slice(tcasEngagementSlice, func(i, j int) bool {
		return tcasEngagementSlice[i].TimeOfEngagement.Before(tcasEngagementSlice[j].TimeOfEngagement)
	})
	return tcasEngagementSlice
}
