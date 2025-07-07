package aviation

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/config"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/util"
)

// SimulationState holds the collection of live domain objects and their current state
type SimulationState struct {
	Airports           []*Airport
	PlanesInFlight     []Plane
	Mu                 sync.Mutex
	SimStatusChannel   chan struct{}
	DifferentAltitudes bool
	SimIsRunning       bool
	SimEndedTime       time.Time
}

// GetNumberOfPlanes prompts the user to input the desired number of planes for the simulation.
// It validates the input to ensure it's an integer greater than 1 and updates the configuration.
func GetNumberOfPlanes(conf *config.Config) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to TCAS-simulator")
	notValidInput := true

	for i := 0; notValidInput; i++ {

		fmt.Print("Input the number of planes for the simulation > ")
		scanner.Scan()
		input := util.CleanInput(scanner.Text())
		if len(input) == 0 {
			fmt.Println("")
			continue
		}
		num, err := strconv.Atoi(input[0])
		if err != nil {
			fmt.Println("Please input a valid integer")
			continue
		}
		if num < 2 {
			fmt.Println("Please input a valid integer greater than 1")
			continue
		}

		conf.NoOfAirplanes = num
		notValidInput = false
	}
	notValidInput = true
	for i := 0; notValidInput; i++ {
		fmt.Print("Varying Cruise Altitudes (y/n) > ")
		scanner.Scan()
		input := util.CleanInput(scanner.Text())
		if len(input) == 0 {
			fmt.Println("")
			continue
		}

		if input[0] != "y" && input[0] != "n" {
			fmt.Println("Please input a 'y' or 'n'")
			continue
		}

		if input[0] == "y" {
			conf.DifferentAltitudes = true
		}
		notValidInput = false
	}

}

// InitializeAirports creates appropriate amount of airports and airplanes
func InitializeAirports(conf *config.Config, simState *SimulationState) {
	simState.DifferentAltitudes = conf.DifferentAltitudes

	planesCreated := 0
	airportsCreated := 0

	for i := 0; planesCreated < conf.NoOfAirplanes; i++ {
		newAirport := createAirport(airportsCreated, planesCreated, conf.NoOfAirplanes)
		planesGenerated := planesCreated
		for range newAirport.InitialPlaneAmount {
			newPlane := createPlane(planesGenerated)
			newAirport.Planes = append(newAirport.Planes, newPlane)
			planesGenerated += 1
		}
		simState.Airports = append(simState.Airports, &newAirport)
		planesCreated += newAirport.InitialPlaneAmount
		airportsCreated = i + 1
	}

	listOfAirportCoordinates := generateCoordinates(len(simState.Airports))

	for i := range simState.Airports {
		newLocation := Coordinate{listOfAirportCoordinates[i].X, listOfAirportCoordinates[i].Y, 0.0}
		simState.Airports[i].Location = newLocation
	}

	fmt.Printf("\nInitialized: %d airports, %d planes distributed among airports.\n\n",
		len(simState.Airports), conf.NoOfAirplanes)
}

// Point represents a 2D coordinate with X and Y components.
type Point struct {
	X float64
	Y float64
}

// calculateDistance calculates the Euclidean distance between two 2D points.
func calculateDistance(p1, p2 Point) float64 {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// generateCoordinates generates a list of (X, Y) coordinates based on specified spacing and density rules.
//
// The rules are:
//  1. Minimum Separation: Every new coordinate generated must be at least 50 units
//     away from *all* previously generated coordinates. This prevents overlap and
//     ensures a minimum spacing between all points.
//  2. Initial Clustering (First 4 Points): For the first four coordinates,
//     the generation attempts to place them within a 50 to 100 unit range from a
//     randomly selected existing point. This helps in forming a relatively compact
//     initial group, while strictly adhering to the 50-unit minimum separation
//     from all other points.
//  3. Spreading Mechanism (5th Point Onwards): To avoid "overpopulation" in one
//     area and encourage the coordinates to spread out across the "map", for the
//     fifth coordinate and all subsequent ones, the generation is guided. New
//     points are primarily generated outward from the coordinate that is currently
//     farthest from the origin (0,0) among all existing points. They will be placed
//     at least 50 units away from this "most distant" point.
func generateCoordinates(numCoordinates int) []Point {
	if numCoordinates <= 0 {
		return []Point{}
	}

	// Initialize random number generator with a unique seed based on current time.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	coordinates := make([]Point, 0, numCoordinates)
	minDist := 50.0 // The minimum required distance between any two coordinates

	// Maximum attempts to find a valid position for a single coordinate.
	// If a valid spot isn't found after this many attempts, the function will stop.
	maxAttemptsPerPoint := 5000

	for i := range numCoordinates {
		foundValidPoint := false
		attempts := 0

		for !foundValidPoint && attempts < maxAttemptsPerPoint {
			var candidateX, candidateY float64

			if len(coordinates) == 0 {
				// For the very first coordinate, start near the origin with a small random offset.
				candidateX = r.Float64()*10 - 5 // Range -5 to 5
				candidateY = r.Float64()*10 - 5 // Range -5 to 5
			} else {
				// Find the point currently farthest from the origin (0,0).
				// This point serves as a reference for expanding the map outwards later.
				mostDistantPoint := Point{X: 0.0, Y: 0.0}
				maxDistFromOriginSq := -1.0
				for _, p := range coordinates {
					distSq := p.X*p.X + p.Y*p.Y
					if distSq > maxDistFromOriginSq {
						maxDistFromOriginSq = distSq
						mostDistantPoint = p
					}
				}

				// Strategy for generating the next candidate point:
				if len(coordinates) < 4 {
					// For the first 4 points, select a random existing point as a reference.
					// Attempt to place the new point within 50 to 100 units from this reference.
					// This encourages a relatively compact initial grouping.
					referencePoint := coordinates[r.Intn(len(coordinates))] // Pick a random existing point
					angle := r.Float64() * 2 * math.Pi                      // Random angle for direction
					// Distance from the reference point, between min_dist and 100.0
					distanceFromRef := r.Float64()*(100.0-minDist) + minDist
					candidateX = referencePoint.X + distanceFromRef*math.Cos(angle)
					candidateY = referencePoint.Y + distanceFromRef*math.Sin(angle)
				} else {
					// For the 5th point and onwards, use the most distant point from origin as reference.
					// This ensures new points expand the map, preventing clumping ("overpopulation").
					// Generate the new point at least 50 units away from this most distant point.
					// A range of 50 to 150 units from the reference is used to provide some variability.
					referencePoint := mostDistantPoint
					angle := r.Float64() * 2 * math.Pi
					// Aim for 50-150 from reference
					distanceFromRef := r.Float64()*(minDist+100.0-minDist) + minDist
					candidateX = referencePoint.X + distanceFromRef*math.Cos(angle)
					candidateY = referencePoint.Y + distanceFromRef*math.Sin(angle)
				}
			}

			candidatePoint := Point{X: candidateX, Y: candidateY}

			// Validate the candidate: Ensure it is at least `minDist` away from *all* existing points.
			isValid := true
			for _, existingPoint := range coordinates {
				if calculateDistance(candidatePoint, existingPoint) < minDist {
					isValid = false
					break // If too close to any point, this candidate is invalid
				}
			}

			if isValid {
				coordinates = append(coordinates, candidatePoint)
				foundValidPoint = true
			}
			attempts++
		}

		if !foundValidPoint {
			// If after many attempts a valid spot isn't found, print a warning and stop early.
			// This can happen if the constraints are too strict for the desired number of points.
			fmt.Printf("Warning: Could not find a valid coordinate after %d attempts for point %d. Stopping generation.\n", maxAttemptsPerPoint, i+1)
			break
		}
	}
	return coordinates
}
