package aviation

import (
	"fmt"
	"log"
	"time"

	"fyne.io/fyne/v2"
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
func (ap *Airport) Land(plane *Plane, simState *SimulationState) error {
	f := simState.ConsoleLog
	log.Printf("Plane %s is attempting to land at Airport %s (%s).\n\n",
		plane.Serial, ap.Serial, ap.Location.String())
	fmt.Fprintf(f, "%s Plane %s is attempting to land at Airport %s (%s).\n\n",
		simState.CurrentSimTime.Format("2006-01-02 15:04:05"), plane.Serial, ap.Serial, ap.Location.String())

	// first we run a loop to make sure a plane is not trying to land in an airport where
	// another airplane is trying to take off
	for i := 0; ap.Runway.noOfRunwayinUse > 0 && simState.SimIsRunning; i++ {
		log.Printf("\nairport %s has %d runway(s) currently in use; plane %s cannot land until all runways are free\n\n",
			ap.Serial, ap.Runway.noOfRunwayinUse, plane.Serial)
		fmt.Fprintf(f, "%s\nairport %s has %d runway(s) currently in use; plane %s cannot land until all runways are free\n\n",
			simState.CurrentSimTime.Format("2006-01-02 15:04:05"), ap.Serial, ap.Runway.noOfRunwayinUse, plane.Serial)
		time.Sleep(TakeoffDuration)
	}
	log.Printf("Plane %s is now landing at Airport %s (%s).\n\n",
		plane.Serial, ap.Serial, ap.Location.String())
	fmt.Fprintf(f, "%sPlane %s is now landing at Airport %s (%s).\n\n",
		simState.CurrentSimTime.Format("2006-01-02 15:04:05"), plane.Serial, ap.Serial, ap.Location.String())

	// Mark a runway as in use for the landing.
	// This lock the runway so no plane can take off for the landing duration

	ap.Mu.Lock()
	ap.Runway.noOfRunwayinUse++
	ap.ReceivingPlane = true
	ap.Mu.Unlock()
	defer func() {
		ap.Mu.Lock()
		ap.ReceivingPlane = false
		ap.Mu.Unlock()
	}()
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

	// Release the runway after the landing is complete.
	ap.Runway.noOfRunwayinUse--

	// Remove the plane from the global `simState.PlanesInFlight` list.
	simState.Mu.Lock()
	planeInFlightIndex := -1
	for i, p := range simState.PlanesInFlight {
		if p.Serial == plane.Serial {
			planeInFlightIndex = i
			break
		}
	}
	if planeInFlightIndex == -1 {
		simState.Mu.Unlock() // Manual unlock before return
		return fmt.Errorf("plane %s not found in the global PlanesInFlight list", plane.Serial)
	}
	simState.PlanesInFlight = append(simState.PlanesInFlight[:planeInFlightIndex], simState.PlanesInFlight[planeInFlightIndex+1:]...)
	simState.Mu.Unlock()

	// Update the plane's status to reflect it's no longer in flight.
	plane.PlaneInFlight = false

	plane.FlightLog[len(plane.FlightLog)-1].FlightStatus = "landed"
	plane.FlightLog[len(plane.FlightLog)-1].ActualLandingTime = simState.CurrentSimTime

	// Add the now-landed plane to the destination airport's list of parked planes.
	ap.Planes = append(ap.Planes, plane) // Append the updated copy of the plane

	log.Printf("Plane %s successfully landed at Airport %s (%s). It is now parked.\n\n",
		plane.Serial, ap.Serial, ap.Location.String())
	fmt.Fprintf(f, "%sPlane %s successfully landed at Airport %s (%s). It is now parked.\n\n",
		simState.CurrentSimTime.Format("2006-01-02 15:04:05"), plane.Serial, ap.Serial, ap.Location.String())

	// Call the UI callback if registered
	if simState.OnPlaneLandCallback != nil {
		fyne.Do(func() { // Ensure UI updates are on main goroutine
			simState.OnPlaneLandCallback(plane.Serial)
		})
	}

	return nil
}
