package aviation

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/util"
)

// CruisingAltitudes defines 3 cruising altitude for planes in meters.
var CruisingAltitudes = [3]float64{10000.0, 11000.0, 12000.0}

// TakeoffDuration defines how long a takeoff operation physically lasts.
const TakeoffDuration = 5 * time.Second

// TakeOff prepares a plane for flight, simulates its takeoff, and updates the simulation state.
// It handles runway allocation, flight path generation, and state transitions for the plane and airport.
//
// Parameters:
//
//	plane: The Plane struct that is taking off. Note that this is passed by value
//	       the modifications to this copy are then reflected when its added to simState.PlanesInFlight.
//	simState: A pointer to the global SimulationState, allowing updates to the list of planes in flight.
//
// Returns:
//
//	*Flight: A pointer to the newly created Flight struct representing this takeoff.
//	error: An error if the takeoff cannot be initiated (e.g., no available runways, plane not found).
func (airport *Airport) TakeOff(plane *Plane, simState *SimulationState) (*Flight, error) {
	f := simState.ConsoleLog
	log.Printf("Plane %s (Cruise Speed: %.2fm/s) is attempting to takeoff from Airport %s %s\n\n",
		plane.Serial, plane.CruiseSpeed, airport.Serial, airport.Location.String())
	fmt.Fprintf(f, "%s Plane %s (Cruise Speed: %.2fm/s) is attempting to takeoff from Airport %s %s\n\n",
		simState.CurrentSimTime.Format("2006-01-02 15:04:05"), plane.Serial, plane.CruiseSpeed, airport.Serial, airport.Location.String())

	for {
		airport.Mu.Lock()
		receiving := airport.ReceivingPlane
		airport.Mu.Unlock()

		if !receiving || !simState.SimIsRunning {
			break
		}
		log.Printf("\nairport %s is currently receiving a landing plane; plane %s cannot takeoff until all landing operations are over\n\n",
			airport.Serial, plane.Serial)
		fmt.Fprintf(f, "%s \nairport %s is currently receiving a landing plane; plane %s cannot takeoff until all landing operations are over\n\n",
			simState.CurrentSimTime.Format("2006-01-02 15:04:05"), airport.Serial, plane.Serial)
		time.Sleep(LandingDuration)
	}

	for {
		// Check if there's an available runway.
		airport.Mu.Lock()
		if airport.Runway.noOfRunwayinUse >= airport.Runway.numberOfRunway {
			airport.Mu.Unlock() // Release lock immediately if no runway available
			log.Printf("\nairport %s has no available runways for takeoff (all %d of %d runway(s) in use)\n\n",
				airport.Serial, airport.Runway.noOfRunwayinUse, airport.Runway.numberOfRunway)
			fmt.Fprintf(f, "%s \nairport %s has no available runways for takeoff (all %d of %d runway(s) in use)\n\n",
				simState.CurrentSimTime.Format("2006-01-02 15:04:05"), airport.Serial, airport.Runway.noOfRunwayinUse, airport.Runway.numberOfRunway)
			time.Sleep(TakeoffDuration)
		} else {
			airport.Mu.Unlock()
			break
		}
	}
	airport.Mu.Lock()
	// Mark a runway as in use.
	airport.Runway.noOfRunwayinUse++
	airport.Mu.Unlock() // <<< IMPORTANT: Release the lock BEFORE the takeoff duration sleep

	// Simulate the physical takeoff duration. This does NOT hold the lock.
	// This allows other planes to acquire the lock and potentially start taking off
	// on another available runway immediately.
	log.Printf("Plane %s (Cruise Speed: %.2fm/s) is taking off from Airport %s %s\n\n",
		plane.Serial, plane.CruiseSpeed, airport.Serial, airport.Location.String())
	fmt.Fprintf(f, "%s Plane %s (Cruise Speed: %.2fm/s) is taking off from Airport %s %s\n\n",
		simState.CurrentSimTime.Format("2006-01-02 15:04:05"), plane.Serial, plane.CruiseSpeed, airport.Serial, airport.Location.String())

	time.Sleep(TakeoffDuration)

	// After the takeoff duration, re-acquire the lock to safely decrement the counter.
	airport.Mu.Lock()
	airport.Runway.noOfRunwayinUse--
	airport.Mu.Unlock()

	// Find and remove the plane from this airport's list of parked planes.
	planeIndex := -1
	for i, p := range airport.Planes {
		if p.Serial == plane.Serial {
			planeIndex = i
			break
		}
	}

	if planeIndex == -1 {
		return nil, fmt.Errorf("plane %s not found at airport %s to initiate takeoff", plane.Serial, airport.Serial)
	}

	// Remove the plane from the airport's Planes slice.
	airport.Planes = append(airport.Planes[:planeIndex], airport.Planes[planeIndex+1:]...)

	// Select a random destination airport for the plane.
	destinationAirport, err := airport.getRandomDestinationAirport(simState.Airports)
	if err != nil {
		return nil, fmt.Errorf("failed to select destination airport for plane %s: %w", plane.Serial, err)
	}

	// Define the flight path from the current airport to the destination.
	flightPath := FlightPath{
		Depature:    airport.Location,
		Destination: destinationAirport.Location,
	}

	// Calculate the total distance and estimated flight duration.
	flightDistance := Distance(flightPath.Depature, flightPath.Destination)
	if plane.CruiseSpeed <= 0 {
		return nil, fmt.Errorf("plane %s has an invalid cruise speed (%.2f), cannot calculate flight duration", plane.Serial, plane.CruiseSpeed)
	}
	// Assuming CruiseSpeed is in units per second, and distance is in those same units.
	flightDuration := time.Duration(flightDistance/plane.CruiseSpeed) * time.Second

	takeoffTime := simState.CurrentSimTime
	landingTime := takeoffTime.Add(flightDuration)
	var cruisingAltitude float64
	if simState.DifferentAltitudes {
		chance := rand.Float64()
		if chance < 0.33 {
			cruisingAltitude = CruisingAltitudes[0]
		} else if chance < 0.66 {
			cruisingAltitude = CruisingAltitudes[1]
		} else {
			cruisingAltitude = CruisingAltitudes[2]
		}
	} else {
		cruisingAltitude = CruisingAltitudes[0]
	}

	// Create a new Flight record with all its details.
	newFlight := Flight{
		FlightID:               plane.Serial + util.GenerateSerialNumber(len(plane.FlightLog), "f"), // Generate unique ID for this specific flight
		FlightSchedule:         flightPath,
		TakeoffTime:            takeoffTime,
		DestinationArrivalTime: landingTime,
		CruisingAltitude:       cruisingAltitude,
		DepatureAirPort:        airport.Serial,
		ArrivalAirPort:         destinationAirport.Serial,
		FlightStatus:           "in transit",
		FlightPath: FlightPath{
			Depature:    airport.Location,
			Destination: destinationAirport.Location,
		},
	}

	// Update the plane's internal state to reflect it's now in flight.
	plane.PlaneInFlight = true
	plane.FlightLog = append(plane.FlightLog, newFlight)

	// Add the updated plane to the global list of planes currently in flight.
	simState.Mu.Lock()
	simState.PlanesInFlight = append(simState.PlanesInFlight, plane)
	simState.Mu.Unlock()

	log.Printf("Plane %s (Cruise Speed: %.2fm/s) took off from Airport %s %s, heading to Airport %s %s. Estimated landing at %s.\n\n",
		plane.Serial, plane.CruiseSpeed, airport.Serial, airport.Location.String(), destinationAirport.Serial, destinationAirport.Location.String(), landingTime.Format("15:04:05"))
	fmt.Fprintf(f, "%s Plane %s (Cruise Speed: %.2fm/s) took off from Airport %s %s, heading to Airport %s %s. Estimated landing at %s.\n\n",
		simState.CurrentSimTime.Format("2006-01-02 15:04:05"), plane.Serial, plane.CruiseSpeed, airport.Serial, airport.Location.String(), destinationAirport.Serial, destinationAirport.Location.String(), landingTime.Format("15:04:05"))

	// Call the UI callback if registered
	if simState.OnPlaneTakeOffCallback != nil {
		fyne.Do(func() { // Ensure UI updates are on main goroutine
			simState.OnPlaneTakeOffCallback(plane)
		})
	}

	return &newFlight, nil
}

// getRandomDestinationAirport selects a random airport from the list of all airports
// that is not the current airport (airport). This helps in simulating inter-airport travel.
func (airport *Airport) getRandomDestinationAirport(allAirports []*Airport) (*Airport, error) {
	eligibleAirports := []*Airport{}
	for _, otherAp := range allAirports {
		if otherAp.Serial != airport.Serial { // A plane cannot fly to the airport it just took off from
			eligibleAirports = append(eligibleAirports, otherAp)
		}
	}
	if len(eligibleAirports) == 0 {
		return nil, fmt.Errorf("no other airports available to serve as a destination")
	}

	// Use a new random source to ensure sufficient randomness, especially if called frequently.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(eligibleAirports))
	return eligibleAirports[randomIndex], nil
}
