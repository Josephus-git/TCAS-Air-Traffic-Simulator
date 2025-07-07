package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
)

// startInit parses the duration string and initializes the simulation,
// handles input validation, ensuring a positive integer for simulation duration.
func startInit(simState *aviation.SimulationState, durationMinutesString string) {

	logFilePath := "logs/console_log.txt"
	// Open the file in append mode. Create it if it doesn't exist.
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	logFilePath = "logs/tcasLog.txt"
	tcasLog, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	durationMinutes, err := strconv.Atoi(durationMinutesString)
	if err != nil {
		fmt.Println("usage: start <integer> (integer represents time in minute(s))")
		return
	}
	if durationMinutes < 1 {
		fmt.Println("Please input a valid integer greater than 0")
		return
	}
	simState.SimIsRunning = true
	simState.SimEndedTime = time.Time{}
	simState.SimStatusChannel = make(chan struct{})
	startSimulation(simState, time.Duration(durationMinutes), f, tcasLog)
}

// startAirports launches goroutines for each airport to handle takeoffs.
func startAirports(simState *aviation.SimulationState, ctx context.Context, wg *sync.WaitGroup, f, tcasLog *os.File) {
	log.Printf("--- Starting Airport Launch Operations ---")
	fmt.Fprintf(f, "%s--- Starting Airport Launch Operations ---\n",
		time.Now().Format("2006-01-02 15:04:05"))
	for i := range simState.Airports {
		ap := simState.Airports[i] // Get a pointer to the airport
		wg.Add(1)                  // Add to WaitGroup for each airport goroutine
		go func(airport *aviation.Airport) {
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
				}
			}
		}(ap) // Pass airport pointer
	}
}
