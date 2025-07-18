package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
)

// GraphicsSimulationInit initializes and sets up the main simulation window, including UI controls for navigation,
// zoom, and quitting, alongside the core simulation area display.
func GraphicsSimulationInit(simState *aviation.SimulationState, simulationWindow fyne.Window, inputWindow fyne.Window) {
	simulationWindow.Resize(fyne.NewSize(800, 600)) // Larger window for simulation

	simulationArea := NewSimulationArea(simState, inputWindow) // Pass inputWindow reference

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
		simulationWindow.Close()
		inputWindow.Show()
		if simState.SimIsRunning {
			aviation.EmergencyStop(simState)
		}
		aviation.CloseLogFiles(simState)

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

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if !simState.SimIsRunning {
				fyne.Do(func() {
					simulationWindow.Close()

				})

				return
			}
		}
	}()

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
