package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
)

// Simulation parameters

// AirportLaunchIntervalMin is the min random delay before an airport tries to launch a plane
const AirportLaunchIntervalMin = 1 * time.Second

// AirportLaunchIntervalMax is the max random delay before an airport tries to launch a plane
const AirportLaunchIntervalMax = 60 * time.Second

// FlightMonitorInterval is how often the monitor checks planes for landing time
const FlightMonitorInterval = 500 * time.Millisecond

var FlightNumberCount int

// simulationCancelFunc is a global variable to hold the cancel function for the simulation context,
// this allows EmergencyStop to trigger cancellation of the simulation from anywhere
var simulationCancelFunc context.CancelFunc

// stopTrigger is a pointer to time.Timer, it is stopped during emergency stop
var stopTrigger *time.Timer

// startSimulationInit initializes and starts the TCAS simulation, managing goroutines for takeoffs and landings.
// It sets up a context for graceful shutdown and waits for all simulation activities to complete.
func startSimulation(simState *aviation.SimulationState, durationMinutes time.Duration, f, tcasLog *os.File) {
	FlightNumberCount = 0
	defer close(simState.SimStatusChannel) // Ensures SimStatuschannel is closed when startSimulation function exits
	defer func() { simState.SimIsRunning = false }()
	defer func() { simState.SimEndedTime = time.Now() }()
	defer func() { fmt.Print("\nTCAS-simulator > ") }()
	defer func() { f.Close() }()
	log.Printf("\n--- TCAS Simulation Started for %d minute(s) ---", durationMinutes)
	fmt.Fprintf(f, "%s\n--- TCAS Simulation Started for %d minute(s) ---\n",
		time.Now().Format("2006-01-02 15:04:05"), durationMinutes)
	fmt.Printf("To initiate an emergency stop, type 'q' and press Enter.\n\n")
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
	startAirports(simState, ctx, &wg, f, tcasLog)

	// --- Start Flight Monitoring Goroutine (for landings) ---
	log.Printf("--- Starting Flight Landing and TCAS Monitor ---\n\n")
	fmt.Fprintf(f, "%s--- Starting Flight Landing and TCAS Monitor ---, \n\n",
		time.Now().Format("2006-01-02 15:04:05"))

	wg.Add(1) // Add for the monitor goroutine
	go func(globalSimState *aviation.SimulationState, ctx context.Context) {
		defer wg.Done()

		for i := 0; simState.SimIsRunning; i++ {
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

			time.Sleep(FlightMonitorInterval) // Sleep to avoid busy-waiting and reduce CPU usage

			// We need to safely access and potentially modify globalSimState.PlanesInFlight.
			// It's safer to copy the list of planes to be processed, then release the lock,
			// and then process the copy. This prevents deadlocks if Land() tries to acquire
			// other locks (like airport.Mu) while globalSimState.Mu is held.
			globalSimState.Mu.Lock()
			planesToLand := []aviation.Plane{}
			type monitorTCASEngagement struct {
				plane      aviation.Plane
				engagement aviation.TCASEngagement
			}
			planesToEngageTCASManeuver := []monitorTCASEngagement{}
			currentTime := time.Now()

			for _, p := range globalSimState.PlanesInFlight {
				if len(p.FlightLog) > 0 {
					currentFlight := p.FlightLog[len(p.FlightLog)-1]
					// Check if current time is past or at the plane's scheduled landing time
					if currentTime.After(currentFlight.DestinationArrivalTime) || currentTime.Equal(currentFlight.DestinationArrivalTime) {
						planesToLand = append(planesToLand, p)
					}
				}
				if len(p.CurrentTCASEngagements) > 0 {
					for _, tcasE := range p.CurrentTCASEngagements {
						// in order to engage early tcas warning, we will call an alarm 3 seconds before the
						// collition time so the maneuver can take place
						if currentTime.After(tcasE.TimeOfEngagement.Add(-3*time.Second)) || currentTime.Equal(tcasE.TimeOfEngagement.Add(-3*time.Second)) {
							newEngagement := monitorTCASEngagement{
								plane:      p,
								engagement: tcasE,
							}
							planesToEngageTCASManeuver = append(planesToEngageTCASManeuver, newEngagement)
						}
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
				var destinationAirport *aviation.Airport = nil
				for i := range globalSimState.Airports {
					ap := globalSimState.Airports[i]
					// Match airport by location, using Epsilon for robust float comparison
					if aviation.Distance(ap.Location, currentFlight.FlightSchedule.Destination) < aviation.Epsilon {
						destinationAirport = ap
						break
					}
				}

				if destinationAirport != nil {
					// Call the Land function. It handles its own internal locking for runway use
					// and updates globalSimState.PlanesInFlight by removing the landed plane.
					// The Land function itself acquires the necessary simState.Mu lock for its modification.
					err := destinationAirport.Land(p, globalSimState, f)
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

			// Process the planes that are ready to engage Tcas
			for _, tcasEngagement := range planesToEngageTCASManeuver {
				select {
				case <-ctx.Done():
					log.Printf("Flight monitor stopping while processing planes.")
					fmt.Fprintf(f, "%sFlight monitor stopping while processing planes.\n",
						time.Now().Format("2006-01-02 15:04:05"))
					return
				default:
				}

				// Find the corresponding plane to engage the tcas
				var otherPlane aviation.Plane
				for _, plane := range globalSimState.PlanesInFlight {
					if plane.Serial == tcasEngagement.engagement.OtherPlaneSerial {
						otherPlane = plane
						break
					}
				}

				// Check if otherPlane not found
				if otherPlane.Serial == "" {
					continue // plane has probably landed
				}

				// update the planes tcas records
				globalSimState.Mu.Lock()
				for i, plane := range globalSimState.PlanesInFlight {
					if plane.Serial == tcasEngagement.engagement.PlaneSerial {
						globalSimState.PlanesInFlight[i].TCASEngagementRecords = append(globalSimState.PlanesInFlight[i].TCASEngagementRecords, tcasEngagement.engagement)
					}
					if plane.Serial == tcasEngagement.engagement.OtherPlaneSerial {
						globalSimState.PlanesInFlight[i].TCASEngagementRecords = append(globalSimState.PlanesInFlight[i].TCASEngagementRecords, tcasEngagement.engagement)
					}
				}
				globalSimState.Mu.Unlock()

				if !tcasEngagement.engagement.WarningTriggered {
					// implement the TCAS early warning system
					log.Printf("TCAS: CRASH IMMINENT! Plane %s and Plane %s about to collide! ENGAGE EVASIVE MANEUVER NOW!!!\n\n",
						tcasEngagement.plane.Serial, otherPlane.Serial)
					fmt.Fprintf(tcasLog, "%s TCAS: CRASH IMMINENT! Plane %s and Plane %s about to collide! ENGAGE EVASIVE MANEUVER NOW!!!\n\n",
						time.Now().Format("2006-01-02 15:04:05"), tcasEngagement.plane.Serial, otherPlane.Serial)
					fmt.Fprintf(f, "%s TCAS: CRASH IMMINENT! Plane %s and Plane %s about to collide! ENGAGE EVASIVE MANEUVER NOW!!!\n\n",
						time.Now().Format("2006-01-02 15:04:05"), tcasEngagement.plane.Serial, otherPlane.Serial)

					// Carry out the corresponding actions depending of if the planes will successfully evade each orther or not
					if tcasEngagement.engagement.WillCrash {
						time.AfterFunc(3*time.Second, func() {
							log.Printf("DISASTER OCCURED!: Plane %s and Plane %s CRASHED\n\n",
								tcasEngagement.plane.Serial, otherPlane.Serial)
							fmt.Fprintf(tcasLog, "%s DISASTER OCCURED!: Plane %s and Plane %s CRASHED\n\n",
								time.Now().Format("2006-01-02 15:04:05"), tcasEngagement.plane.Serial, otherPlane.Serial)
							fmt.Fprintf(f, "%s DISASTER OCCURED!: Plane %s and Plane %s CRASHED\n\n",
								time.Now().Format("2006-01-02 15:04:05"), tcasEngagement.plane.Serial, otherPlane.Serial)

							// at this point, the simulation ends
							if simState.SimIsRunning {
								emergencyStop(simState)
							}

						})
					} else {
						time.AfterFunc(3*time.Second, func() {
							log.Printf("DISASTER AVERTED! Plane %s and Plane %s SUCCESSFULLY ENGAGED EVASIVE MANEUVER\n\n",
								tcasEngagement.plane.Serial, otherPlane.Serial)
							fmt.Fprintf(tcasLog, "%s DISASTER AVERTED! Plane %s and Plane %s SUCCESSFULLY ENGAGED EVASIVE MANEUVER\n\n",
								time.Now().Format("2006-01-02 15:04:05"), tcasEngagement.plane.Serial, otherPlane.Serial)
							fmt.Fprintf(f, "%s DISASTER AVERTED! Plane %s and Plane %s SUCCESSFULLY ENGAGED EVASIVE MANEUVER\n\n",
								time.Now().Format("2006-01-02 15:04:05"), tcasEngagement.plane.Serial, otherPlane.Serial)
						})
					}

					// update the plane to contain triggered warning so the monitor doesn't make multiple calls to print the warning
					globalSimState.Mu.Lock()
					for i, plane := range globalSimState.PlanesInFlight {
						if plane.Serial == tcasEngagement.engagement.PlaneSerial {
							for j, engagement := range globalSimState.PlanesInFlight[i].CurrentTCASEngagements {
								if engagement.EngagementID == tcasEngagement.engagement.EngagementID {
									globalSimState.PlanesInFlight[i].CurrentTCASEngagements[j].WarningTriggered = true
								}
							}
						}
					}
					globalSimState.Mu.Unlock()
				}

				if !globalSimState.SimIsRunning {
					break
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
