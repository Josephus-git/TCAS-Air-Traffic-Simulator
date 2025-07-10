package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
)

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

	// Bottom label
	simEndLabel := canvas.NewText("", color.RGBA{})
	simEndLabel.Alignment = fyne.TextAlignCenter
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
		simEndLabel,
		nil,
		nil,
		simulationArea,
	)
	simulationWindow.SetContent(simContent)
}
