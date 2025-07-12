package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
)

// simulationAreaRenderer implements fyne.WidgetRenderer for SimulationArea.
type simulationAreaRenderer struct {
	simulationArea *SimulationArea
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
	currentAirplaneDisplaySize := fyne.NewSize(
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
		labelY := displayY - labelSize.Height // Position label directly above the airport image

		airport.IDLabel.Resize(labelSize)
		airport.IDLabel.Move(fyne.NewPos(labelX, labelY))
	}

	if r.simulationArea.simState == nil {
		return
	}

	// Use the simulation's current time for all calculations
	simTime := r.simulationArea.simState.CurrentSimTime

	// Iterate through planes and apply rendering logic
	for _, planeRender := range r.simulationArea.planesInFlight {
		plane := planeRender.ActualPlane
		if len(plane.FlightLog) == 0 {
			continue
		}
		currentFlight := plane.FlightLog[len(plane.FlightLog)-1]

		planeCoord, ok := planeCurrentPosition(plane, simTime)
		if !ok {
			planeRender.Image.Hidden = true
			planeRender.FlightPathLine.Hidden = true
			if planeRender.TCASCircle != nil {
				planeRender.TCASCircle.Hidden = true
			}
			continue
		}

		displayX := (float32(planeCoord.X) * scale) + r.simulationArea.offsetX
		displayY := (float32(planeCoord.Y) * scale) + r.simulationArea.offsetY

		planeRender.Image.Resize(currentAirplaneDisplaySize)
		planeRender.Image.Move(fyne.NewPos(displayX-currentAirplaneDisplaySize.Width/2, displayY-currentAirplaneDisplaySize.Height/2))
		planeRender.Image.Hidden = false

		destX := (float32(currentFlight.FlightSchedule.Destination.X) * scale) + r.simulationArea.offsetX
		destY := (float32(currentFlight.FlightSchedule.Destination.Y) * scale) + r.simulationArea.offsetY

		planeRender.FlightPathLine.Position1 = fyne.NewPos(displayX, displayY)
		planeRender.FlightPathLine.Position2 = fyne.NewPos(destX, destY)
		planeRender.FlightPathLine.Hidden = false

		// --- TCAS Circle Logic ---
		// Get the relevant engagement for *this* plane

		for _, tcasE := range plane.CurrentTCASEngagements {

			// Helper function to apply circle properties
			applyTCASCircle := func(pr *PlaneRender, pCoord aviation.Coordinate, engagement aviation.TCASEngagement) {

				engagementTime := engagement.TimeOfEngagement
				preEngagementStart := engagementTime.Add(-3 * time.Second)
				postEngagementEnd := engagementTime.Add(2 * time.Second)

				if simTime.After(preEngagementStart) && simTime.Before(postEngagementEnd) {
					circleRadiusDisplay := 50.0 * scale // 50 units radius in simulation coordinates, scaled to display
					circleSize := fyne.NewSize(circleRadiusDisplay*2, circleRadiusDisplay*2)

					displayPX := (float32(pCoord.X) * scale) + r.simulationArea.offsetX
					displayPY := (float32(pCoord.Y) * scale) + r.simulationArea.offsetY

					circleX := displayPX - circleRadiusDisplay
					circleY := displayPY - circleRadiusDisplay
					pr.TCASCircle.Move(fyne.NewPos(circleX, circleY))
					pr.TCASCircle.Resize(circleSize)
					pr.TCASCircle.Hidden = false

					if simTime.Before(engagementTime) { // Pre-engagement phase
						pr.TCASCircle.StrokeColor = color.RGBA{R: 255, G: 165, A: 255} // Orange stroke
						pr.TCASCircle.FillColor = color.Transparent
					} else { // At or after TimeOfEngagement
						pr.TCASCircle.StrokeColor = color.Transparent
						if engagement.WillCrash {
							pr.TCASCircle.FillColor = color.RGBA{R: 255, A: 200} // Red fill, semi-transparent
						} else {
							pr.TCASCircle.FillColor = color.RGBA{G: 255, A: 200} // Green fill, semi-transparent
						}
					}
				} else {
					pr.TCASCircle.Hidden = true
				}

				pr.TCASCircle.Refresh()

			}

			// Apply circle logic for the current plane (planeA)
			applyTCASCircle(planeRender, planeCoord, tcasE)

			// Apply circle logic for the other engaged plane (planeB), if any

			for _, otherPlaneRender := range r.simulationArea.planesInFlight {
				if tcasE.OtherPlaneSerial == otherPlaneRender.ActualPlane.Serial {
					otherPlane := otherPlaneRender.ActualPlane

					// Find the current position of the other plane
					otherPlaneCoord, otherOk := planeCurrentPosition(otherPlane, simTime)
					if otherOk {
						applyTCASCircle(otherPlaneRender, otherPlaneCoord, tcasE)
					} else {
						// If other plane is not in transit, hide its circle
						if otherPlaneRender.TCASCircle != nil {
							otherPlaneRender.TCASCircle.Hidden = true
						}
					}
				}

			}

		}
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
		objects = append(objects,
			planeRender.FlightPathLine,
			planeRender.TCASCircle,
			planeRender.Image,
		)
	}

	// Add status label last so it's always on top
	objects = append(objects, r.simulationArea.statusLabel)

	return objects
}

func (r *simulationAreaRenderer) Destroy() {
	// Clean up any resources if necessary
	r.simulationArea.animationTicker.Stop() // Stop the animation ticker
	r.simulationArea.ClearAllPlanes()
}

func (r *simulationAreaRenderer) Refresh() {
	zoomText := fmt.Sprintf("Zoom: %.1fx", r.simulationArea.zoomScales[r.simulationArea.zoomLevel])

	r.simulationArea.statusLabel.Text = fmt.Sprintf(
		"Offset: %.0f, %.0f | %s | Drag to pan | Planes: %d",
		r.simulationArea.offsetX, r.simulationArea.offsetY, zoomText, len(r.simulationArea.planesInFlight),
	)
	r.simulationArea.statusLabel.Refresh()

	for _, airport := range r.simulationArea.airports {
		airport.IDLabel.TextSize = 8 * r.simulationArea.zoomScales[r.simulationArea.zoomLevel]
		airport.IDLabel.Refresh()
	}

	r.Layout(r.simulationArea.Size()) // force refresh
}
