package aviation

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Simulation parameters
// FlightMonitorInterval is how often the monitor checks planes for landing time
const FlightMonitorInterval = 100 * time.Millisecond

// FlightNumberCount is a global counter used to generate unique flight numbers.
var FlightNumberCount int

// simulationCancelFunc is a global variable to hold the cancel function for the simulation context,
// this allows EmergencyStop to trigger cancellation of the simulation from anywhere
var simulationCancelFunc context.CancelFunc

// stopTrigger is a pointer to time.Timer, it is stopped during emergency stop
var stopTrigger *time.Timer

// startSimulationInit initializes and starts the TCAS simulation, managing goroutines for takeoffs and landings.
// It sets up a context for graceful shutdown and waits for all simulation activities to complete.
func StartSimulation(simState *SimulationState, durationMinutes time.Duration) {
	simState.SimIsRunning = true
	f := simState.ConsoleLog
	FlightNumberCount = 0

	defer func() { simState.SimIsRunning = false }()
	defer func() { simState.SimEndedTime = time.Now() }()
	defer func() { fmt.Print("\nTCAS-simulator > ") }()

	defer func() { f.Close() }()
	log.Printf("\n--- TCAS Simulation Started for %d minute(s) ---", durationMinutes)
	fmt.Fprintf(f, "%s\n--- TCAS Simulation Started for %d minute(s) ---\n",
		time.Now().Format("2006-01-02 15:04:05"), durationMinutes)

	fmt.Printf("TCAS logs can be found in logs/tcasLogs.txt. \n\n")

	// WaitGroup to keep track of running goroutines
	var wg sync.WaitGroup

	// Create a cancellable context for the simulation.
	// This context will be passed to all goroutines.
	// The cancel function is stored globally and also called when the duration expires.
	var ctx context.Context
	simulationDuration := time.Duration(durationMinutes) * time.Minute
	ctx, simulationCancelFunc = context.WithCancel(context.Background())

	// Set a timer to automatically call cancel after the specified duration.
	// This ensures the simulation stops even if EmergencyStop is not called.
	stopTrigger = time.AfterFunc(simulationDuration, func() {
		if simState.SimIsRunning {
			log.Printf("\n--- Simulation Duration (%d minutes) Reached. Initiating shutdown... ---", durationMinutes)
			fmt.Fprintf(f, "%s\n--- Simulation Duration (%d minutes) Reached. Initiating shutdown... ---\n",
				time.Now().Format("2006-01-02 15:04:05"), durationMinutes)
		}
		if simulationCancelFunc != nil {
			simulationCancelFunc() // Trigger cancellation
		}
	})

	// Start the takeoff simulation (using your provided startSimulation function)
	// Pass ctx and wg to startSimulation so airport goroutines can respect shutdown
	startAirports(simState, ctx, &wg)

	// --- Start Flight Monitoring Goroutine (for landings) ---
	log.Printf("--- Starting Flight Landing and TCAS Monitor ---\n\n")
	fmt.Fprintf(f, "%s--- Starting Flight Landing and TCAS Monitor ---, \n\n",
		time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("--- Varying Altitudes: %v ---\n\n", simState.DifferentAltitudes)
	fmt.Fprintf(f, "%s--- Varying Altitudes: %v ---, \n\n",
		time.Now().Format("2006-01-02 15:04:05"), simState.DifferentAltitudes)
	fmt.Println("Remember type 'q' and hit Enter to immediately stop the simulation if needed")

	wg.Add(1) // Add for the monitor goroutine
	go func(globalSimState *SimulationState, ctx context.Context) {
		defer wg.Done()

		for simState.SimIsRunning {
			select {
			case <-ctx.Done(): // Check if the main simulation context is done
				log.Printf("Flight monitor stopping.")
				fmt.Fprintf(f, "%sFlight monitor stopping .\n",
					time.Now().Format("2006-01-02 15:04:05"))
				return // Exit goroutine
			default:
				// Continue monitoring
			}

			select {
			case <-time.After(FlightMonitorInterval):
				// This case executes if the FlightMonitorInterval duration passes.
			case <-ctx.Done():
				// This case executes if the context (ctx) is cancelled.
				log.Printf("Flight monitor stopping during sleep.")
				fmt.Fprintf(f, "%sFlight monitor stopping during sleep.\n",
					time.Now().Format("2006-01-02 15:04:05"))
				return // Exits the goroutine immediately.
			}

			// We need to safely access and potentially modify globalSimState.PlanesInFlight.
			// It's safer to copy the list of planes to be processed, then release the lock,
			// and then process the copy. This prevents deadlocks if Land() tries to acquire
			// other locks (like airport.Mu) while globalSimState.Mu is held.
			globalSimState.Mu.Lock()
			planesToLand := []*Plane{}

			for _, p := range globalSimState.PlanesInFlight {
				if len(p.FlightLog) > 0 {
					currentFlight := p.FlightLog[len(p.FlightLog)-1]
					// Check if current time is past or at the plane's scheduled landing time
					if simState.CurrentSimTime.After(currentFlight.DestinationArrivalTime) || simState.CurrentSimTime.Equal(currentFlight.DestinationArrivalTime) {
						planesToLand = append(planesToLand, p)
					}
				}
			}
			globalSimState.Mu.Unlock() // Release lock on global state after identifying planes

			// Process the planes that are ready to land
			for _, p := range planesToLand {
				select {
				case <-ctx.Done():
					log.Printf("Flight monitor stopping while processing planes.")
					fmt.Fprintf(f, "%sFlight monitor stopping while processing planes.\n",
						time.Now().Format("2006-01-02 15:04:05"))
					return
				default:
				}

				// Find the corresponding destination airport object
				currentFlight := p.FlightLog[len(p.FlightLog)-1]
				var destinationAirport *Airport = nil
				for i := range globalSimState.Airports {
					ap := globalSimState.Airports[i]
					// Match airport by location, using Epsilon for robust float comparison
					if Distance(ap.Location, currentFlight.FlightSchedule.Destination) <
						Epsilon {
						destinationAirport = ap
						break
					}
				}

				if destinationAirport != nil {
					// Call the Land function. It handles its own internal locking for runway use
					// and updates globalSimState.PlanesInFlight by removing the landed plane.
					// The Land function itself acquires the necessary simState.Mu lock for its modification.
					err := destinationAirport.Land(p, globalSimState)
					if err != nil {
						// This error could be due to runway busy. The plane remains in PlanesInFlight
						// and will be retried in the next monitor interval.
					}
				} else {
					log.Printf("Monitor Error: Destination airport not found for plane %s (arrival coord: %s)\n",
						p.Serial, currentFlight.FlightSchedule.Destination.String())
					fmt.Fprintf(f, "%sMonitor Error: Destination airport not found for plane %s (arrival coord: %s)\n",
						time.Now().Format("2006-01-02 15:04:05"), p.Serial, currentFlight.FlightSchedule.Destination.String())
				}
			}

		}
	}(simState, ctx)

	// This wg.Wait() will block Start() until all goroutines have gracefully exited
	wg.Wait()

	log.Printf("\n--- All simulation goroutines have stopped. ---")
	fmt.Fprintf(f, "%s\n--- All simulation goroutines have stopped. ---\n",
		time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("Final Simulation State Summary:")
	fmt.Fprintf(f, "%sFinal Simulation State Summary:\n",
		time.Now().Format("2006-01-02 15:04:05"))
	simState.Mu.Lock() // Acquire lock to safely read final count of planes in flight
	log.Printf("  Planes currently in flight: %d", len(simState.PlanesInFlight))
	fmt.Fprintf(f, "%s  Planes currently in flight: %d\n",
		time.Now().Format("2006-01-02 15:04:05"), len(simState.PlanesInFlight))
	simState.Mu.Unlock()

	for i := range simState.Airports {
		ap := simState.Airports[i]
		ap.Mu.Lock() // Acquire lock for each airport to safely read its parked planes count
		log.Printf("  Airport %s has %d planes parked.", ap.Serial, len(ap.Planes))
		fmt.Fprintf(f, "%s  Airport %s has %d planes parked.\n",
			time.Now().Format("2006-01-02 15:04:05"), ap.Serial, len(ap.Planes))
		ap.Mu.Unlock()
	}
	log.Printf("--- TCAS Simulation Ended ---")
	fmt.Fprintf(f, "%s--- TCAS Simulation Ended ---\n",
		time.Now().Format("2006-01-02 15:04:05"))

}
