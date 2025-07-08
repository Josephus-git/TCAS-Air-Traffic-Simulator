package ui

import (
	"image/color"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func StartFyne() {
	a := app.New()

	// --- Initial Input Window ---
	inputWindow := a.NewWindow("Airport Simulation Setup")
	inputWindow.Resize(fyne.NewSize(400, 200)) // Smaller initial window

	numAirportsEntry := widget.NewEntry()
	numAirportsEntry.SetPlaceHolder("Enter number of airports (min 2)")

	errorMessage := canvas.NewText("", color.RGBA{R: 255, A: 255}) // Red text for errors
	errorMessage.Alignment = fyne.TextAlignCenter

	var simulationWindow fyne.Window

	startSimulationButton := widget.NewButton("Start Simulation", func() {
		simulationWindow = a.NewWindow("Airport Simulation")
		numStr := numAirportsEntry.Text
		num, err := strconv.Atoi(numStr)
		if err != nil || num < 2 {
			errorMessage.Text = "Please enter a valid number (minimum 2)."
			errorMessage.Refresh()
			return
		}

		errorMessage.Text = "" // Clear error message
		errorMessage.Refresh()

		// Create and show the simulation window
		simulationInit(simulationWindow, num, inputWindow)

		inputWindow.Hide()
		simulationWindow.Show()
		log.Printf("Starting simulation with %d airports.", num)
	})

	// Layout for the input window
	inputContent := container.NewVBox(
		widget.NewLabel("Welcome to Airport Simulation!"),
		numAirportsEntry,
		errorMessage,
		layout.NewSpacer(),
		container.NewHBox(
			layout.NewSpacer(),
			startSimulationButton,
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
	)

	inputWindow.SetContent(container.NewCenter(inputContent))
	inputWindow.ShowAndRun()
}

func simulationInit(simulationWindow fyne.Window, num int, inputWindow fyne.Window) {
	simulationWindow.Resize(fyne.NewSize(800, 600)) // Larger window for simulation

	simulationArea := NewSimulationArea(num, inputWindow) // Pass inputWindow reference

	// Controls for the simulation window

	homeButton := widget.NewButtonWithIcon("", theme.HomeIcon(), func() {
		simulationArea.Home()
	})
	zoomInButton := widget.NewButtonWithIcon("", theme.ZoomInIcon(), func() {
		simulationArea.ZoomIn()
	})
	zoomOutButton := widget.NewButtonWithIcon("", theme.ZoomOutIcon(), func() {
		simulationArea.ZoomOut()
	})
	quitButton := widget.NewButtonWithIcon("Quit", theme.CancelIcon(), func() {
		simulationWindow.Hide()
		inputWindow.Show() // Show the input window again
		log.Println("Returned to setup screen.")
	})

	// Arrange controls at the top
	simControls := container.NewHBox(
		layout.NewSpacer(),
		homeButton,
		zoomInButton,
		zoomOutButton,
		quitButton,
		layout.NewSpacer(),
	)

	// Main content layout for simulation window: controls at top, simulation area fills rest
	simContent := container.NewBorder(
		simControls,
		nil,
		nil,
		nil,
		simulationArea,
	)
	simulationWindow.SetContent(simContent)
}
