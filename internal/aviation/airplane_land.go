package aviation

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LandingDuration defines how long a landing operation physically lasts.
const LandingDuration = 7 * time.Second

// Epsilon is a small value used for floating-point comparisons,
// particularly when checking if coordinates are approximately equal.
const Epsilon = 0.1 // meters, adjust as needed for precision of coordinates

// Land handles the process of a plane landing at an airport.
// It verifies the plane's intended destination, strictly manages runway availability,
// simulates the landing process, and updates the plane's and the global simulation's states.
//
// Parameters:
//
//	plane: The Plane struct that is attempting to land. This is passed by value;
//	       its modifications will be reflected when it's re-added to the airport's list.
//	simState: A pointer to the global SimulationState, necessary for removing the plane
//	          from the `PlanesInFlight` list.
//
// Returns:
//
//	error: An error if the landing cannot proceed (e.g., wrong destination,
//	       runways are currently in use, or the plane is not found in flight).
func (ap *Airport) Land(plane Plane, simState *SimulationState, f *os.File) error {
	log.Printf("Plane %s is attempting to land at Airport %s (%s).\n\n",
		plane.Serial, ap.Serial, ap.Location.String())
	fmt.Fprintf(f, "%s Plane %s is attempting to land at Airport %s (%s).\n\n",
		time.Now().Format("2006-01-02 15:04:05"), plane.Serial, ap.Serial, ap.Location.String())

	// first we run a loop to make sure a plane is not trying to land in an airport where
	// another airplane is trying to take off
	for i := 0; ap.Runway.noOfRunwayinUse > 0 && simState.SimIsRunning; i++ {
		log.Printf("\nairport %s has %d runway(s) currently in use; plane %s cannot land until all runways are free\n\n",
			ap.Serial, ap.Runway.noOfRunwayinUse, plane.Serial)
		fmt.Fprintf(f, "%s\nairport %s has %d runway(s) currently in use; plane %s cannot land until all runways are free\n\n",
			time.Now().Format("2006-01-02 15:04:05"), ap.Serial, ap.Runway.noOfRunwayinUse, plane.Serial)
		time.Sleep(TakeoffDuration)
	}
	log.Printf("Plane %s is now landing at Airport %s (%s).\n\n",
		plane.Serial, ap.Serial, ap.Location.String())
	fmt.Fprintf(f, "%sPlane %s is now landing at Airport %s (%s).\n\n",
		time.Now().Format("2006-01-02 15:04:05"), plane.Serial, ap.Serial, ap.Location.String())

	// Mark a runway as in use for the landing.
	// This lock the runway so no plane can take off for the landing duration

	ap.Mu.Lock()
	ap.Runway.noOfRunwayinUse++
	ap.ReceivingPlane = true
	ap.Mu.Unlock()
	defer func() { ap.ReceivingPlane = false }()
	time.Sleep(LandingDuration)

	// Retrieve the current flight details from the plane's log.
	if len(plane.FlightLog) == 0 {
		return fmt.Errorf("plane %s has no flight history; cannot initiate landing", plane.Serial)
	}
	// Get the most recent flight from the log.
	currentFlight := plane.FlightLog[len(plane.FlightLog)-1]

	plane.FlightLog[len(plane.FlightLog)-1].FlightStatus = "about to land"

	// Verify that this airport is the plane's intended destination.
	// We use the 'distance' function with an Epsilon to account for floating-point inaccuracies.
	if Distance(ap.Location, currentFlight.FlightSchedule.Destination) > Epsilon {
		return fmt.Errorf("plane %s attempting to land at airport %s (%s), but its destination for current flight %s is %s",
			plane.Serial, ap.Serial, ap.Location.String(), currentFlight.FlightID, currentFlight.FlightSchedule.Destination.String())
	}

	// Acquire the airport's mutex lock. This protects the runway state and other
	// airport-specific shared resources during the critical landing operation.
	ap.Mu.Lock()
	defer ap.Mu.Unlock() // Ensure the lock is released when the function exits

	// 5. Simulate the physical landing duration.
	// The lock is held during this time, preventing other takeoffs or landings
	// from this airport (due to the strict rule and lock).

	// 6. Release the runway after the landing is complete.
	ap.Runway.noOfRunwayinUse--

	// 7. Remove the plane from the global `simState.PlanesInFlight` list.
	planeInFlightIndex := -1
	for i, p := range simState.PlanesInFlight {
		if p.Serial == plane.Serial {
			planeInFlightIndex = i
			break
		}
	}

	if planeInFlightIndex == -1 {
		// This scenario should ideally not happen if the simulation logic is robust,
		// as a plane should only be landed if it's currently in flight.
		return fmt.Errorf("plane %s not found in the global PlanesInFlight list; cannot complete landing at airport %s", plane.Serial, ap.Serial)
	}

	// Remove the plane from the slice without changing its capacity.
	simState.PlanesInFlight = append(simState.PlanesInFlight[:planeInFlightIndex], simState.PlanesInFlight[planeInFlightIndex+1:]...)

	// 8. Update the plane's status to reflect it's no longer in flight.
	plane.PlaneInFlight = false // Update the local copy
	plane.CurrentTCASEngagements = []TCASEngagement{}

	plane.FlightLog[len(plane.FlightLog)-1].FlightStatus = "landed"
	plane.FlightLog[len(plane.FlightLog)-1].ActualLandingTime = time.Now()

	// 9. Add the now-landed plane to the destination airport's list of parked planes.
	ap.Planes = append(ap.Planes, plane) // Append the updated copy of the plane

	log.Printf("Plane %s successfully landed at Airport %s (%s). It is now parked.\n\n",
		plane.Serial, ap.Serial, ap.Location.String())
	fmt.Fprintf(f, "%sPlane %s successfully landed at Airport %s (%s). It is now parked.\n\n",
		time.Now().Format("2006-01-02 15:04:05"), plane.Serial, ap.Serial, ap.Location.String())

	return nil
}
