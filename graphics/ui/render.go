package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

// simulationAreaRenderer implements fyne.WidgetRenderer for SimulationArea.
type simulationAreaRenderer struct {
	simulationArea *SimulationArea
	//objects        object
	background *canvas.Rectangle // Background is separate to ensure it's drawn first
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
	currentAirplaneDisplaySize := fyne.NewSize( // NEW: Airplane display size
		r.simulationArea.initialAirplaneSize.Width*scale,
		r.simulationArea.initialAirplaneSize.Height*scale,
	)

	// Layout each airport and its serial number label
	for _, airport := range r.simulationArea.airports {
		// Apply pan and zoom to airport position
		displayX := (float32(airport.ActualAirport.Location.X) * scale) + r.simulationArea.offsetX
		displayY := (float32(airport.ActualAirport.Location.Y) * scale) + r.simulationArea.offsetY

		// Position the airport image
		airport.Image.Resize(currentAirportDisplaySize)
		airport.Image.Move(fyne.NewPos(displayX, displayY))

		// Position the serial number label just above the airport image
		// Adjust label position based on its size and airport's size
		labelSize := airport.IDLabel.MinSize() // Get the minimum size required for the label text
		labelX := displayX + (currentAirportDisplaySize.Width / 2) - (labelSize.Width / 2)
		labelY := displayY - labelSize.Height // 0 units above the airport image

		airport.IDLabel.Resize(labelSize)
		airport.IDLabel.Move(fyne.NewPos(labelX, labelY))
	}

	// NEW: Layout each plane and its flight path
	simTime := time.Now() // Use current real time for animation

	for _, planeRender := range r.simulationArea.planesInFlight {
		plane := planeRender.ActualPlane
		currentFlight := plane.FlightLog[len(plane.FlightLog)-1] // Get the active flight

		// Calculate current plane position
		planeCoord, ok := planeCurrentPosition(plane, simTime)
		if !ok {
			planeRender.Image.Hidden = true
			planeRender.FlightPathLine.Hidden = true
			continue
		}

		// Apply pan and zoom to plane position
		displayX := (float32(planeCoord.X) * scale) + r.simulationArea.offsetX
		displayY := (float32(planeCoord.Y) * scale) + r.simulationArea.offsetY

		// Position the plane image
		planeRender.Image.Resize(currentAirplaneDisplaySize)
		planeRender.Image.Move(fyne.NewPos(displayX-currentAirplaneDisplaySize.Width/2, displayY-currentAirplaneDisplaySize.Height/2)) // Center image
		planeRender.Image.Hidden = false                                                                                               // Make sure plane is visible

		// Calculate and apply plane orientation
		rotation := planeOrientation(currentFlight.FlightSchedule.Depature, currentFlight.FlightSchedule.Destination)
		RotateCanvasImage(planeRender.Image, rotation)
		// NOTE: canvas.Image does not support rotation directly. To visually rotate, use a canvas.Raster or custom widget.

		// Update flight path line
		//depX := (float32(currentFlight.FlightSchedule.Depature.X) * scale) + r.simulationArea.offsetX
		//depY := (float32(currentFlight.FlightSchedule.Depature.Y) * scale) + r.simulationArea.offsetY
		destX := (float32(currentFlight.FlightSchedule.Destination.X) * scale) + r.simulationArea.offsetX
		destY := (float32(currentFlight.FlightSchedule.Destination.Y) * scale) + r.simulationArea.offsetY

		// The line should start from the current plane position to the destination
		// This achieves the "covered section should not be displayed" effect.
		planeRender.FlightPathLine.Position1 = fyne.NewPos(displayX, displayY) // Start from current plane position
		planeRender.FlightPathLine.Position2 = fyne.NewPos(destX, destY)
		planeRender.FlightPathLine.Hidden = false
	}
}

// Objects is where the dynamic object list creation happens.
func (r *simulationAreaRenderer) Objects() []fyne.CanvasObject {
	// Dynamically build the list of objects every time this is called
	var objects []fyne.CanvasObject

	// Always add the background first
	objects = append(objects, r.background)

	// Add airport images and their labels
	for _, airport := range r.simulationArea.airports {
		objects = append(objects, airport.Image, airport.IDLabel)
	}

	// Add plane flight paths and images (order matters, paths usually behind planes)
	for _, planeRender := range r.simulationArea.planesInFlight {
		objects = append(objects, planeRender.FlightPathLine, planeRender.Image)
	}

	// Add status label last so it's always on top
	objects = append(objects, r.simulationArea.statusLabel)

	return objects
}

func (r *simulationAreaRenderer) Destroy() {
	// Clean up any resources if necessary
	r.simulationArea.animationTicker.Stop() // Stop the animation ticker
}

func (r *simulationAreaRenderer) Refresh() {
	zoomText := fmt.Sprintf("Zoom: %.1fx", r.simulationArea.zoomScales[r.simulationArea.zoomLevel])
	r.simulationArea.statusLabel.Text = fmt.Sprintf(
		"Offset: %.0f, %.0f | %s | Drag to pan | Planes: %d",
		r.simulationArea.offsetX, r.simulationArea.offsetY, zoomText, len(r.simulationArea.planesInFlight), // NEW: Add plane count
	)
	r.simulationArea.statusLabel.Refresh()

	// Update label text sizes according to zoom scale
	for _, airport := range r.simulationArea.airports {
		airport.IDLabel.TextSize = 8 * r.simulationArea.zoomScales[r.simulationArea.zoomLevel]
		airport.IDLabel.Refresh()
	}

	// Force layout and redraw
	r.Layout(r.simulationArea.Size())
	canvas.Refresh(r.simulationArea) // Request a redraw of the widget
}
