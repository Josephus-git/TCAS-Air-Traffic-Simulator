package aviation

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/util"
)

// Airport represents an Airport with its location
type Airport struct {
	Serial             string
	Location           Coordinate
	InitialPlaneAmount int
	Runway             runway
	Planes             []Plane
	Mu                 sync.Mutex
	ReceivingPlane     bool
}

// runway represents the state of an airport's runways.
type runway struct {
	numberOfRunway  int
	noOfRunwayinUse int
}

// createAirport initializes and returns a new Airport struct.
// It generates a serial number, plane capacity, and runway details for the airport.
func createAirport(airportCount, planecount, totalNumPlanes int) Airport {
	return Airport{
		Serial:             util.GenerateSerialNumber(airportCount, "ap"),
		InitialPlaneAmount: generatePlaneCapacity(totalNumPlanes, planecount),
		Runway:             generateRunway(),
	}
}

// generateRunway creates and returns a new runway configuration.
func generateRunway() runway {
	randomNumber := rand.Intn(3) + 1
	return runway{
		numberOfRunway:  randomNumber,
		noOfRunwayinUse: 0,
	}
}

// generatePlaneCapacity calculates a random number of planes to create,
// adjusting the quantity based on the total target and already generated planes.
func generatePlaneCapacity(totalPlanes, planeGenerated int) int {
	var randomNumber int
	if totalPlanes < 20 {
		planeToCreate := totalPlanes - planeGenerated
		if planeToCreate <= 3 {
			randomNumber = planeToCreate
		} else {
			randomNumber = rand.Intn(2) + 1
		}

	} else if totalPlanes < 100 {
		planeToCreate := totalPlanes - planeGenerated
		if planeToCreate <= 6 {
			randomNumber = planeToCreate
		} else {
			randomNumber = rand.Intn(5) + 1
		}

	} else {
		planeToCreate := totalPlanes - planeGenerated
		if planeToCreate <= 30 {
			randomNumber = planeToCreate
		} else {
			randomNumber = rand.Intn(20) + 10
		}

	}
	return randomNumber
}

// Simulation parameters

// AirportLaunchIntervalMin is the min random delay before an airport tries to launch a plane
const AirportLaunchIntervalMin = 1 * time.Second

// AirportLaunchIntervalMax is the max random delay before an airport tries to launch a plane
const AirportLaunchIntervalMax = 60 * time.Second

// startAirports launches goroutines for each airport to handle takeoffs.
func startAirports(simState *SimulationState, ctx context.Context, wg *sync.WaitGroup, f, tcasLog *os.File) {
	log.Printf("--- Starting Airport Launch Operations ---")
	fmt.Fprintf(f, "%s--- Starting Airport Launch Operations ---\n",
		time.Now().Format("2006-01-02 15:04:05"))
	for i := range simState.Airports {
		ap := simState.Airports[i] // Get a pointer to the airport
		wg.Add(1)                  // Add to WaitGroup for each airport goroutine
		go func(airport *Airport) {
			defer wg.Done()
			airportRand := rand.New(rand.NewSource(time.Now().UnixNano() + int64(i)*1000)) // Unique seed for each airport

			for {
				select {
				case <-ctx.Done(): // Check if the main simulation context is done
					// stopping all airport launch operations
					return // Exit goroutine
				default:
					// Continue operation
				}

				sleepDuration := time.Duration(airportRand.Intn(int(AirportLaunchIntervalMax.Seconds()-AirportLaunchIntervalMin.Seconds())+1)+int(AirportLaunchIntervalMin.Seconds())) * time.Second //wait 5 to 10 seconds
				select {
				case <-time.After(sleepDuration):
				case <-ctx.Done():
					// stoping all airport launch operation during sleep
					return
				}

				airport.Mu.Lock() // Lock airport to safely check and pick a plane
				if len(airport.Planes) > 0 {
					planeToTakeOff := airport.Planes[0] // Pick the first available plane for simplicity
					airport.Mu.Unlock()                 // Unlock airport before calling TakeOff

					// IMPORTANT: Pass the global simState here.
					_, err := airport.TakeOff(planeToTakeOff, simState, f, tcasLog) // Pass the simState from main
					if err != nil {
						// log.Printf("error taking off from %s: %v", airport.Serial, err)
					}
				} else {
					airport.Mu.Unlock() // Always ensure lock is released
					// log.Printf("Airport %s has no planes to take off.", airport.Serial)
					time.Sleep(1 * time.Second)
				}
			}
		}(ap) // Pass airport pointer
	}
}
