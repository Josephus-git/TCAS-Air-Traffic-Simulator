package aviation

import (
	"math/rand"
	"sync"

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
