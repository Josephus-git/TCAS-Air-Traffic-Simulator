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
			// Reset current engagement if plane is not in flight
			plane.CurrentTCASEngagement = nil
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
		// Reset current engagement for this frame unless an active one is found
		plane.CurrentTCASEngagement = nil

		for _, otherPlaneRender := range r.simulationArea.planesInFlight {
			// Don't compare a plane with itself
			if planeRender.ActualPlane.Serial == otherPlaneRender.ActualPlane.Serial {
				continue
			}

			otherPlane := otherPlaneRender.ActualPlane

			otherPlaneCoord, otherOk := planeCurrentPosition(otherPlane, simTime)
			if !otherOk {
				continue // Skip if other plane is not in flight
			}

			distanceBetweenPlanes := aviation.Distance(planeCoord, otherPlaneCoord)

			// Find if there's an existing engagement between these two planes
			var existingEngagement *aviation.TCASEngagement
			for i := range plane.TCASEngagementRecords {
				rec := &plane.TCASEngagementRecords[i]
				if (rec.PlaneSerial == plane.Serial && rec.OtherPlaneSerial == otherPlane.Serial) ||
					(rec.PlaneSerial == otherPlane.Serial && rec.OtherPlaneSerial == plane.Serial) {
					existingEngagement = rec
					break
				}
			}

			if distanceBetweenPlanes < TriggerEngageTCAS {
				// Engagement Zone (Green/Red)
				if existingEngagement == nil || !existingEngagement.Engaged { // Only trigger if not already engaged
					// Trigger TCAS core logic
					newEngagement := tcasCore(r.simulationArea.simState, plane, otherPlane)
					// Update the current engagement for both planes for rendering this frame
					plane.CurrentTCASEngagement = &newEngagement
					otherPlane.CurrentTCASEngagement = &newEngagement // Ensure the other plane also has this engagement
				} else {
					// If already engaged, just re-apply the existing engagement
					plane.CurrentTCASEngagement = existingEngagement
				}
			} else if distanceBetweenPlanes < TriggerTCAS {
				// Warning Zone (Orange)
				if existingEngagement == nil || (!existingEngagement.WarningTriggered && !existingEngagement.Engaged) {
					// Create a temporary engagement record for warning if not already warned or engaged
					tempEngagement := aviation.TCASEngagement{
						EngagementID:     fmt.Sprintf("W-%s-%s-%d", plane.Serial, otherPlane.Serial, time.Now().UnixNano()),
						PlaneSerial:      plane.Serial,
						OtherPlaneSerial: otherPlane.Serial,
						TimeOfEngagement: simTime,
						WillCrash:        false, // Not determined yet, just a warning
						WarningTriggered: true,  // Mark warning as triggered
						Engaged:          false, // Not yet in engagement phase
					}
					// Assign to current engagement for rendering
					plane.CurrentTCASEngagement = &tempEngagement
					otherPlane.CurrentTCASEngagement = &tempEngagement

					// Persist the warning state in the TCASEngagementRecords
					if existingEngagement == nil {
						plane.TCASEngagementRecords = append(plane.TCASEngagementRecords, tempEngagement)
						otherPlane.TCASEngagementRecords = append(otherPlane.TCASEngagementRecords, tempEngagement)
					} else {
						existingEngagement.WarningTriggered = true
					}

				} else if existingEngagement.WarningTriggered && !existingEngagement.Engaged {
					// Continue showing orange if already warned but not yet engaged
					plane.CurrentTCASEngagement = existingEngagement
				} else {
					// If already engaged (green/red), suppress orange
					plane.CurrentTCASEngagement = existingEngagement // Might still be needed if it just exited engagement
				}
			}
			// If distance > TriggerTCAS, ensure TCAS circle is hidden, unless there's an active engagement for the other plane
		}

		// After checking all pairs, apply the determined current engagement to the plane's rendering
		r.applyTCASCircle(planeRender, planeCoord, plane.CurrentTCASEngagement, scale)
	}
}

// applyTCASCircle Helper function to apply circle properties based on TCASEngagement
func (r *simulationAreaRenderer) applyTCASCircle(pr *PlaneRender, pCoord aviation.Coordinate,
	engagement *aviation.TCASEngagement, scale float32) {

	if engagement == nil {
		pr.TCASCircle.Hidden = true
		return
	}

	circleRadiusDisplay := 50.0 * scale // 50 units radius in simulation coordinates, scaled to display
	circleSize := fyne.NewSize(circleRadiusDisplay*2, circleRadiusDisplay*2)

	displayPX := (float32(pCoord.X) * scale) + r.simulationArea.offsetX
	displayPY := (float32(pCoord.Y) * scale) + r.simulationArea.offsetY

	circleX := displayPX - circleRadiusDisplay
	circleY := displayPY - circleRadiusDisplay
	pr.TCASCircle.Move(fyne.NewPos(circleX, circleY))
	pr.TCASCircle.Resize(circleSize)
	pr.TCASCircle.Hidden = false

	if engagement.Engaged { // Green or Red state
		pr.TCASCircle.StrokeColor = color.Transparent // No stroke for filled circles
		if engagement.WillCrash {
			pr.TCASCircle.FillColor = color.RGBA{R: 255, A: 200} // Red fill, semi-transparent
		} else {
			pr.TCASCircle.FillColor = color.RGBA{G: 255, A: 200} // Green fill, semi-transparent
		}
	} else if engagement.WarningTriggered { // Orange warning state
		pr.TCASCircle.StrokeColor = color.RGBA{R: 255, G: 165, A: 255} // Orange stroke
		pr.TCASCircle.FillColor = color.Transparent                    // No fill for warning
	} else {
		// Should not happen if engagement is not nil, but as a fallback, hide it
		pr.TCASCircle.Hidden = true
	}

	pr.TCASCircle.Refresh()
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
			planeRender.TCASCircle, // Draw circle before plane image so plane is on top
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
