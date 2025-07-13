package main

import (
	"bufio"
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"
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

func StartFyne(cfg *config.Config, simState *aviation.SimulationState) {
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

		information1 := canvas.NewText("Light grey is cruise altitude: 10000.0", color.RGBA{R: 200, G: 200, B: 200, A: 255})
		information1.TextSize = 10
		information2 := canvas.NewText("Light green is cruise altitude: 11000.0", color.RGBA{G: 255, A: 255})
		information2.TextSize = 10
		information3 := canvas.NewText("Light blue is cruise altitude: 12000.0", color.RGBA{B: 255, A: 255})
		information3.TextSize = 10

		// Input entry for Number of Planes
		numPlanesEntry := widget.NewEntry()

		numPlanesEntry.Validator = func(s string) error {
			_, err := strconv.Atoi(s)
			if err != nil {
				if s == "" && cfg.FirstRun {
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

			// update the form so the number of planes can be updated
			simulationWindow.SetOnClosed(func() {
				numPlanesEntry.Show()
				numPlanesEntry.SetPlaceHolder("")
				varyingAltitudeCheckbox.Show()
				inputWindow.Show()
				if simState.SimIsRunning {
					aviation.EmergencyStop(simState)
				}
				aviation.CloseLogFiles(simState)
			})

			var numAirPlanes int

			if !simState.SimWindowOpened {
				numAirPlanes = cfg.NoOfAirplanes
			} else {
				numAirPlanesV, err := strconv.Atoi(numPlanesEntry.Text)
				if err != nil || numAirPlanesV < 4 {
					errorMessage.Text = "Please enter a valid number of airplanes (minimum 4)."
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

			if !cfg.FirstRun && simState.SimWindowOpened {
				cfg.DifferentAltitudes = varyingAltitudeCheckbox.Checked
			}

			errorMessage.Text = "" // Clear error message
			errorMessage.Refresh()

			simState.Airports = []*aviation.Airport{}
			simState.PlanesInFlight = []*aviation.Plane{}

			// Initialize the airports
			cfg.NoOfAirplanes = numAirPlanes
			aviation.InitializeAirports(cfg, simState)

			// run the simulation
			go aviation.StartSimulation(simState, time.Duration(durationOfSimulation))

			// Create and show the simulation window
			if !cfg.FirstRun {
				ui.GraphicsSimulationInit(simState, simulationWindow, inputWindow)
			}

			simulationWindow.Show()
			inputWindow.Hide()
			aviation.OpenLogFiles(cfg, simState)

			simState.SimWindowOpened = true
			simState.SimIsRunning = true
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
			information1,
			information2,
			information3,
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
	cfg.FirstRun = false
	scanner := bufio.NewScanner(os.Stdin)

	for {
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
