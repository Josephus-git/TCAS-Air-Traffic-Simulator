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
const CollisionThreshold = 50

// tcas detects potential mid-air collisions between a given plane (the one about to take off)
// and other planes currently in flight. If a collision is predicted under specific conditions,
// it triggers an emergency stop.
//
// Parameters:
//
//	plane: The plane attempting to take off.
//	tcasLog: The file pointer for logging planes condition before going on the flight.
func (plane *Plane) tcas(simState *SimulationState, tcasLog *os.File) []TCASEngagement {
	planeFlight := plane.FlightLog[len(plane.FlightLog)-1]
	simState.Mu.Lock()         // Lock the simulation state to safely access PlanesInFlight
	defer simState.Mu.Unlock() // Release the lock getting tcas

	fmt.Fprintf(tcasLog, "%s TCAS: Plane %s (%v) is checking for conflicts before takeoff.\n\n",
		simState.CurrentSimTime.Format("2006-01-02 15:04:05"), plane.Serial, plane.TCASCapability)

	tcasEngagementSlice := []TCASEngagement{}
	for _, otherPlane := range simState.PlanesInFlight {
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
		distanceAtCA, closestTimeForPlane1, otherPlaneStatusAtCATime := planeFlight.GetClosestApproachDetails(simState, otherPlaneFlight)

		// Condition 1: If otherPlane has landed, is about to land or at different flight altitudes, no collision concern from altitude difference
		if otherPlaneStatusAtCATime == "landed or still landing" || otherPlaneStatusAtCATime == "about to land" || otherPlaneFlight.CruisingAltitude != planeFlight.CruisingAltitude {
			fmt.Fprintf(tcasLog, "%s TCAS: Plane %s's flight path %s and Plane %s's flight path %s have closest approach (%.2f units at %v), but no worries: Other plane status is '%s' or different altitude.\n\n",
				simState.CurrentSimTime.Format("15:04:05"), plane.Serial, planeFlight.FlightID, otherPlane.Serial, otherPlaneFlight.FlightID, distanceAtCA, closestTimeForPlane1.Format("15:04:05"), otherPlaneStatusAtCATime)
			continue
		}

		// Condition 2: Check if collision distance threshold is met
		if distanceAtCA < CollisionThreshold {
			fmt.Fprintf(tcasLog, "%s TCAS ALERT: Potential collision detected between Plane %s (TCAS: %v) and Plane %s (TCAS: %v). Closest approach: %.2f units at %v.\n\n",
				simState.CurrentSimTime.Format("15:04:05"), plane.Serial, plane.TCASCapability, otherPlane.Serial, otherPlane.TCASCapability, distanceAtCA, closestTimeForPlane1.Format("15:04:05"))

			// Collision Resolution based on TCAS capabilities
			shouldCrash := false

			if plane.TCASCapability == TCASPerfect && otherPlane.TCASCapability == TCASPerfect {
				// Both perfect, no crash
				fmt.Fprintf(tcasLog, "%s TCAS: Both planes have perfect TCAS. Collision averted between %s and %s.\n\n",
					simState.CurrentSimTime.Format("2006-01-02 15:04:05"), plane.Serial, otherPlane.Serial)
				shouldCrash = false
			} else if (plane.TCASCapability == TCASPerfect && otherPlane.TCASCapability == TCASFaulty) ||
				(plane.TCASCapability == TCASFaulty && otherPlane.TCASCapability == TCASPerfect) {
				// One perfect, one faulty: 25% chance of crash
				if rand.Float64() < 0.25 {
					shouldCrash = true
				} else {
					fmt.Fprintf(tcasLog, "%s TCAS: One perfect, one faulty TCAS. Collision narrowly averted between %s and %s.\n\n",
						simState.CurrentSimTime.Format("15:04:05"), plane.Serial, otherPlane.Serial)
				}
			} else if plane.TCASCapability == TCASFaulty && otherPlane.TCASCapability == TCASFaulty {
				if rand.Float64() < 0.5 {
					shouldCrash = true
				} else {
					fmt.Fprintf(tcasLog, "%s TCAS: Two faulty TCAS. Collision narrowly averted between %s and %s.\n\n",
						simState.CurrentSimTime.Format("15:04:05"), plane.Serial, otherPlane.Serial)
				}
			}

			if shouldCrash {
				newTcasEngagement := TCASEngagement{
					EngagementID:     fmt.Sprintf("%s-%d", plane.Serial, time.Now().UnixNano()),
					FlightID:         planeFlight.FlightID,
					PlaneSerial:      plane.Serial,
					OtherPlaneSerial: otherPlane.Serial,
					TimeOfEngagement: closestTimeForPlane1,
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
				TimeOfEngagement: closestTimeForPlane1,
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
