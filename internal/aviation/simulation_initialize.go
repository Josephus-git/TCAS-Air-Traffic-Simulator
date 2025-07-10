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
	SimWindowOpened    bool
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

// generateCoordinates generates a slice of 2D points based on a specific pattern.
// The first coordinate is always (0,0). Subsequent coordinates are generated in stages,
// forming concentric rings.
//
// Parameters:
//
//	numCoordinates: The total number of coordinates to generate.
//
// Returns:
//
//	A slice of Point structs containing the generated coordinates.
func generateCoordinates(numCoordinates int) []Point {
	// Initialize a new random number generator.
	// Using time.Now().UnixNano() as the seed ensures that the generated
	// coordinates will be different each time the function is called.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Initialize an empty slice to store the generated points.
	points := []Point{}

	// Handle the edge case where 0 or fewer coordinates are requested.
	if numCoordinates <= 0 {
		return points
	}

	// The first coordinate is always (0,0)
	points = append(points, Point{X: 0, Y: 0})

	// If only 1 coordinate is requested, we've already added it, so return.
	if numCoordinates == 1 {
		return points
	}

	// Initialize parameters for the first stage (ring) of coordinates.
	// This stage will have 3 points.
	currentNumPointsInStage := 3
	minRadius := 150.0 // Minimum radius for the current stage.
	maxRadius := 250.0 // Maximum radius for the current stage.

	// Loop to generate points for successive stages until the desired
	// number of coordinates (numCoordinates) is reached.
	for i := 0; len(points) < numCoordinates; i++ {
		// Calculate the angular increment for points in the current stage.
		// Points are evenly distributed around 360 degrees.
		angleIncrement := 360.0 / float64(currentNumPointsInStage)

		// NEW: Generate a random offset angle for the current stage.
		// This ensures that each ring starts at a different random rotation.
		randomOffsetAngle := r.Float64() * 360.0 // Random angle between 0 and 360 degrees.

		// Generate points for the current stage.
		for j := 0; j < currentNumPointsInStage; j++ {
			// Check if we have already generated enough coordinates.
			// This is important to stop precisely at numCoordinates,
			// even if it's in the middle of a stage.
			if len(points) >= numCoordinates {
				break // Exit the inner loop.
			}

			// Generate a random radius within the current stage's defined range.
			// r.Float64() returns a pseudo-random float64 in [0.0, 1.0).
			radius := minRadius + r.Float64()*(maxRadius-minRadius)

			// Calculate the angle for the current point, adding the random offset.
			// Convert degrees to radians for trigonometric functions: radians = degrees * (pi / 180).
			angle := (float64(j)*angleIncrement + randomOffsetAngle)
			angleRad := angle * (math.Pi / 180.0)

			// Calculate the X and Y coordinates using polar to Cartesian conversion.
			// X = radius * cos(angle)
			// Y = radius * sin(angle)
			x := radius * math.Cos(angleRad)
			y := radius * math.Sin(angleRad)

			// Add the newly calculated point to our slice of points.
			points = append(points, Point{X: x, Y: y})
		}

		// Prepare for the next stage:
		// 1. The number of points in the next stage is 3 more airports than the current stage.
		currentNumPointsInStage += 3
		// 2. Update the minimum and maximum radii for the next stage.
		minRadius += 250.0
		maxRadius += 300.0
	}

	return points
}
