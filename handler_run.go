package main

import (
	"bufio"
	"context"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/josephus-git/TCAS-simulation-Fyne/graphics/ui"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/config"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/util"
)

var a fyne.App
var inputWindow fyne.Window

func StartFyne(cfg *config.Config, simState *aviation.SimulationState, f, tcasLog *os.File) {
	if cfg.FirstRun {
		// Create a new Fyne application
		a = app.NewWithID("tcas.app")
		a.Settings().SetTheme(ui.CustomDarkTheme{})

		// --- Initial Input Window ---
		if inputWindow == nil {
			inputWindow = a.NewWindow("TCAS Simulation Setup")
			inputWindow.Resize(fyne.NewSize(400, 600)) // Smaller initial window
		}

		// A close interceptor for the main window
		inputWindow.SetCloseIntercept(func() {
			inputWindow.Hide()
			cfg.FirstRun = false
		})

		title := canvas.NewText("TCAS Simulation Setup", color.White)
		title.TextSize = 24
		title.TextStyle.Bold = true
		title.Alignment = fyne.TextAlignCenter

		errorMessage := canvas.NewText("", color.RGBA{R: 255, A: 255}) // Red text for errors
		errorMessage.Alignment = fyne.TextAlignCenter
		errorMessage.TextStyle.Italic = true

		// Input entry for Number of Planes
		numPlanesEntry := widget.NewEntry()
		if cfg.NoOfAirplanes >= 2 {
			numPlanesEntry.SetPlaceHolder(fmt.Sprintf("Currently set to: %d", cfg.NoOfAirplanes))
		}

		numPlanesEntry.Validator = func(s string) error {
			_, err := strconv.Atoi(s)
			if err != nil {
				if s == "" && cfg.FirstRun {
					return nil
				} else if s == "" && !simState.SimWindowOpened {
					return nil
				}

				return fmt.Errorf("please input a valid integer")
			}
			return nil
		}
		numPlanesEntry.Hide()
		numPlanesFormItem := widget.NewFormItem("Number of Planes:", numPlanesEntry)

		// Input entry Duration of Simulation
		durationEntry := widget.NewEntry()
		durationEntry.SetPlaceHolder("Enter duration of simulation")
		durationEntry.Validator = func(s string) error {
			num, err := strconv.Atoi(s)
			{
				if err != nil {
					return fmt.Errorf("please input a valid integer")
				}

			}
			if num < 1 {
				return fmt.Errorf("1 minute minimum")
			}
			return nil
		}
		durationFormItem := widget.NewFormItem("Duration (minutes):", durationEntry)

		//  checkbox for Varying Altitude
		varyingAltitudeCheckbox := widget.NewCheck("Yes", func(b bool) {})
		varyingAltitudeCheckbox.SetChecked(simState.DifferentAltitudes)
		varyingAltitudeCheckbox.Hide()

		// A form to group the input fields
		inputForm := widget.NewForm(
			numPlanesFormItem,
			durationFormItem,
			widget.NewFormItem("Varying Altitude:", varyingAltitudeCheckbox),
		)

		var simulationWindow fyne.Window

		// The simulation button
		startSimulationButton := widget.NewButton("Start Simulation", func() {
			simulationWindow = a.NewWindow("Airport Simulation")

			// <<<<<<<<<<<check here
			simulationWindow.SetOnClosed(func() {
				numPlanesEntry.Show()
				numPlanesEntry.SetPlaceHolder("")
				varyingAltitudeCheckbox.Show()
				inputWindow.Show()
			})

			var numAirPlanes int

			if !simState.SimWindowOpened {
				numAirPlanes = cfg.NoOfAirplanes
			} else {
				numAirPlanesV, err := strconv.Atoi(numPlanesEntry.Text)
				if err != nil || numAirPlanesV < 2 {
					errorMessage.Text = "Please enter a valid number of airplanes (minimum 2)."
					errorMessage.Refresh()
					return
				} else {
					numAirPlanes = numAirPlanesV
				}
			}

			durationOfSimulation, err := strconv.Atoi(durationEntry.Text)
			if err != nil || durationOfSimulation < 1 {
				errorMessage.Text = "Please enter a valid duration of simulation in minutes"
				errorMessage.Refresh()
				return
			}

			varyingAltitude := varyingAltitudeCheckbox.Checked

			errorMessage.Text = "" // Clear error message
			errorMessage.Refresh()

			fmt.Printf("Simulation Started!\n")
			fmt.Printf("Number of Planes: %d\n", numAirPlanes)
			fmt.Printf("Duration: %d minute(s)\n", durationOfSimulation)
			fmt.Printf("Varying Altitude: %v\n", varyingAltitude)

			// Create and show the simulation window
			ui.GraphicsSimulationInit(simState, simulationWindow, inputWindow)

			// <<<<< modify when correct
			//startSimulation(simState, time.Duration(durationOfSimulation), f, tcasLog)

			simulationWindow.Show()
			inputWindow.Hide()
			cfg.FirstRun = false
			simState.SimWindowOpened = true
			log.Printf("Starting simulation with %d airplanes.", numAirPlanes)
		})

		// Set content
		background := canvas.NewRectangle(color.RGBA{})
		inputContent := container.NewVBox(
			layout.NewSpacer(), // Pushes content towards the center
			title,
			layout.NewSpacer(),
			inputForm,
			layout.NewSpacer(),
			startSimulationButton,
			layout.NewSpacer(),
			errorMessage,
			layout.NewSpacer(),
		)

		inputWindow.SetContent(container.NewStack(background, inputContent))

		// Show input window
		inputWindow.Show()

		go startPartition(cfg, simState)
		a.Run()

	} else {
		fyne.Do(func() { inputWindow.Show() })
	}

}

func startPartition(cfg *config.Config, simState *aviation.SimulationState) {
	scanner := bufio.NewScanner(os.Stdin)

	for i := 0; cfg.IsRunning; i++ {
		fmt.Print("TCAS-simulator > ")
		scanner.Scan()
		input := util.CleanInput(scanner.Text())
		argument2 := ""
		if len(input) > 1 {
			argument2 = input[1]
		}

		if len(input) == 0 {
			fmt.Println("")
			continue
		}

		cmd, ok := getCommand(cfg, simState, argument2)[input[0]]
		if !ok {
			fmt.Println("Unknown command, type <help> for usage")
			continue
		}
		cmd.callback()

		println("")
	}
}

// startInit parses the duration string and initializes the simulation,
// handles input validation, ensuring a positive integer for simulation duration.
func runInit(cfg *config.Config, simState *aviation.SimulationState) {

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

	simState.SimIsRunning = true
	simState.SimEndedTime = time.Time{}

	StartFyne(cfg, simState, f, tcasLog)

}

// <<<<<<<<< change back when done
// startAirports launches goroutines for each airport to handle takeoffs.
func StartAirports(simState *aviation.SimulationState, ctx context.Context, wg *sync.WaitGroup, f, tcasLog *os.File) {
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
