package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// simulationAreaRenderer implements fyne.WidgetRenderer for SimulationArea.
type simulationAreaRenderer struct {
	simulationArea *SimulationArea
	objects        []fyne.CanvasObject
	background     *canvas.Rectangle // Background is separate to ensure it's drawn first
}

func (r *simulationAreaRenderer) MinSize() fyne.Size {
	return fyne.NewSize(600, 400) // Default minimum size for the simulation area
}

func (r *simulationAreaRenderer) Layout(size fyne.Size) {
	// Layout the background to fill the widget
	r.background.Resize(size)
	r.background.Move(fyne.NewPos(0, 0))

	// Layout the status label in the center
	r.simulationArea.statusLabel.Resize(size)
	r.simulationArea.statusLabel.Move(fyne.NewPos(0, 0))

	// Get current scale factor
	scale := r.simulationArea.zoomScales[r.simulationArea.zoomLevel]
	currentAirportDisplaySize := fyne.NewSize(
		r.simulationArea.initialAirportSize.Width*scale,
		r.simulationArea.initialAirportSize.Height*scale,
	)

	// Layout each airport and its serial number label
	for _, airport := range r.simulationArea.airports {
		// Apply pan and zoom to airport position
		displayX := (airport.X * scale) + r.simulationArea.offsetX
		displayY := (airport.Y * scale) + r.simulationArea.offsetY

		// Position the airport image
		airport.Image.Resize(currentAirportDisplaySize)
		airport.Image.Move(fyne.NewPos(displayX, displayY))

		// Position the serial number label just above the airport image
		// Adjust label position based on its size and airport's size
		labelSize := airport.IDLabel.MinSize() // Get the minimum size required for the label text
		labelX := displayX + (currentAirportDisplaySize.Width / 2) - (labelSize.Width / 2)
		labelY := displayY - labelSize.Height - 2 // 2 units above the airport image

		airport.IDLabel.Resize(labelSize)
		airport.IDLabel.Move(fyne.NewPos(labelX, labelY))
	}
}

func (r *simulationAreaRenderer) Objects() []fyne.CanvasObject {
	// Return the background first, then all other objects (status label, airports)
	return append([]fyne.CanvasObject{r.background}, r.objects...)
}

func (r *simulationAreaRenderer) Destroy() {
	// Clean up any resources if necessary
}

func (r *simulationAreaRenderer) Refresh() {
	// Update the status label to show current offset and zoom level
	zoomText := fmt.Sprintf("Zoom: %.1fx", r.simulationArea.zoomScales[r.simulationArea.zoomLevel])
	r.simulationArea.statusLabel.Text = fmt.Sprintf("Offset: %.0f, %.0f | %s | Drag to pan", r.simulationArea.offsetX, r.simulationArea.offsetY, zoomText)
	r.simulationArea.statusLabel.Refresh()

	// Request the widget itself to redraw, which will trigger Layout()
}
