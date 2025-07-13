package ui

import (
	"fmt"
	"image/color"
	"log"
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

// MinSize returns the default minimum dimensions (600x400) required for the simulation area renderer.
func (r *simulationAreaRenderer) MinSize() fyne.Size {
	return fyne.NewSize(600, 400)
}

// Layout positions and renders all simulation elements (
// background, status, airports, planes, flight paths, and TCAS circles
// ) within the given size, applying pan, zoom, and TCAS logic based on the current simulation state.
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

	// Pre-calculate plane positions and reset their current engagement state for this frame.
	// This map helps avoid re-calculating positions multiple times and ensures all planes start clean.
	planeStates := make(map[*aviation.Plane]struct {
		Coord aviation.Coordinate
		OK    bool
	})

	for _, pr := range r.simulationArea.planesInFlight {
		coord, ok := planeCurrentPosition(pr.ActualPlane, simTime)
		planeStates[pr.ActualPlane] = struct {
			Coord aviation.Coordinate
			OK    bool
		}{Coord: coord, OK: ok}
		// IMPORTANT: Reset the plane's CurrentTCASEngagement at the start of each frame.
		// It will be set again below if an active interaction is found.
		pr.ActualPlane.CurrentTCASEngagement = nil
	}

	// Iterate through planes and apply rendering logic
	for _, planeRender := range r.simulationArea.planesInFlight {
		plane := planeRender.ActualPlane
		planeState := planeStates[plane]

		if !planeState.OK {
			planeRender.Image.Hidden = true
			planeRender.FlightPathLine.Hidden = true
			if planeRender.TCASCircle != nil {
				planeRender.TCASCircle.Hidden = true
			}
			// plane.CurrentTCASEngagement was already reset in the pre-calculation loop above.
			continue
		}

		planeCoord := planeState.Coord

		// Update plane image position and visibility
		displayX := (float32(planeCoord.X) * scale) + r.simulationArea.offsetX
		displayY := (float32(planeCoord.Y) * scale) + r.simulationArea.offsetY

		planeRender.Image.Resize(currentAirplaneDisplaySize)
		planeRender.Image.Move(fyne.NewPos(displayX-currentAirplaneDisplaySize.Width/2, displayY-currentAirplaneDisplaySize.Height/2))
		planeRender.Image.Hidden = false

		// Update flight path line
		currentFlight := plane.FlightLog[len(plane.FlightLog)-1]
		destX := (float32(currentFlight.FlightSchedule.Destination.X) * scale) + r.simulationArea.offsetX
		destY := (float32(currentFlight.FlightSchedule.Destination.Y) * scale) + r.simulationArea.offsetY

		planeRender.FlightPathLine.Position1 = fyne.NewPos(displayX, displayY)
		planeRender.FlightPathLine.Position2 = fyne.NewPos(destX, destY)
		planeRender.FlightPathLine.Hidden = false

		// Determine the most critical engagement for *this* plane in *this* frame
		var mostCriticalEngagement *aviation.TCASEngagement = nil

		// --- TCAS Circle Logic ---
		// Loop through all other planes to find potential interactions
		for _, otherPlaneRender := range r.simulationArea.planesInFlight {
			// Don't compare a plane with itself
			if plane.Serial == otherPlaneRender.ActualPlane.Serial {
				continue
			}
			otherPlane := otherPlaneRender.ActualPlane // Get the actual plane object

			// If both planes are not in the same Cruise altitude, skip
			if plane.FlightLog[len(plane.FlightLog)-1].CruisingAltitude != otherPlane.FlightLog[len(otherPlane.FlightLog)-1].CruisingAltitude {
				continue
			}

			otherPlaneState := planeStates[otherPlane] // Use pre-calculated state
			if !otherPlaneState.OK {
				continue // Skip if other plane is not in flight
			}
			otherPlaneCoord := otherPlaneState.Coord // Use pre-calculated position

			distanceBetweenPlanes := aviation.Distance(planeCoord, otherPlaneCoord)

			if distanceBetweenPlanes < TriggerEngageTCAS {
				// This is a full engagement: Highest priority.
				// Call tcasCore which now handles finding/creating the persistent record.
				engagement := tcasCore(r.simulationArea.simState, plane, otherPlane)
				mostCriticalEngagement = &engagement // Set this as the display engagement for this plane
				break                                // Found a full engagement, no need to check other planes for *this* 'plane' anymore
			} else if distanceBetweenPlanes < TriggerTCAS {
				// This is a warning zone: Lower priority than full engagement.
				// Only set if we haven't already found a full engagement for 'plane'.
				if mostCriticalEngagement == nil { // If no full engagement found yet
					// Create a temporary engagement record for warning display.
					// This warning is *transient* for display only and not persisted in TCASEngagementRecords.
					tempEngagement := aviation.TCASEngagement{
						EngagementID:     fmt.Sprintf("W-Disp-%s-%s-%d", plane.Serial, otherPlane.Serial, simTime.UnixNano()), // Transient ID for display
						PlaneSerial:      plane.Serial,
						OtherPlaneSerial: otherPlane.Serial,
						TimeOfEngagement: simTime,
						WillCrash:        false,
						WarningTriggered: true,
						Engaged:          false,
					}
					mostCriticalEngagement = &tempEngagement
				}
			}
		} // End of inner loop (otherPlaneRender)

		// After checking all other planes for 'plane', set its CurrentTCASEngagement
		// based on the most critical interaction found (or nil if none).
		plane.CurrentTCASEngagement = mostCriticalEngagement

		// Apply the determined current engagement to the plane's rendering (show/hide circle)
		r.applyTCASCircle(planeRender, planeCoord, plane.CurrentTCASEngagement, scale)
	} // End of outer loop (planeRender)
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
			r.simulationArea.CrashedPlanes = append(r.simulationArea.CrashedPlanes, engagement.PlaneSerial, engagement.OtherPlaneSerial)
			pr.TCASCircle.FillColor = color.RGBA{R: 255, A: 255} // Red fill, plane destroyed
			pr.Image.Hide()
			r.simulationArea.planeCrash = true
		} else {
			pr.TCASCircle.FillColor = color.RGBA{G: 255, A: 200} // Green fill, semi-transparent
		}
	} else if engagement.WarningTriggered { // Orange warning state
		pr.TCASCircle.StrokeColor = color.RGBA{R: 255, G: 165, A: 255} // Orange stroke
		pr.TCASCircle.FillColor = color.Transparent                    // No fill for warning
	} else {
		// This else block should theoretically not be reached if engagement is not nil,
		// but as a failsafe, ensure it's hidden if something unexpected occurs.
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

// Destroy stops the animation ticker and clears all associated resources, performing necessary cleanup for the renderer.
func (r *simulationAreaRenderer) Destroy() {
	// Clean up any resources if necessary
	r.simulationArea.animationTicker.Stop() // Stop the animation ticker
	r.simulationArea.ClearAllResource()
}

func (r *simulationAreaRenderer) Refresh() {
	zoomText := fmt.Sprintf("Zoom: %.1fx", r.simulationArea.zoomScales[r.simulationArea.zoomLevel])

	if !r.simulationArea.planeCrash {
		r.simulationArea.statusLabel.Text = fmt.Sprintf(
			"Offset: %.0f, %.0f | %s | Drag to pan | Planes: %d",
			r.simulationArea.offsetX, r.simulationArea.offsetY, zoomText, len(r.simulationArea.planesInFlight),
		)
	} else {
		if len(r.simulationArea.CrashedPlanes) > 1 {
			tcasLog := r.simulationArea.simState.TCASLog
			f := r.simulationArea.simState.ConsoleLog
			planeSerial := r.simulationArea.CrashedPlanes[0]
			otherPlaneSerial := r.simulationArea.CrashedPlanes[1]
			r.simulationArea.statusLabel.Text = fmt.Sprintf("PLANE: %s AND PLANE: %s HAVE CRASHED !!!", planeSerial, otherPlaneSerial)
			r.simulationArea.statusLabel.Color = color.RGBA{R: 255, A: 255}
			r.simulationArea.statusLabel.TextSize = 30
			r.simulationArea.statusLabel.TextStyle.Bold = true
			if !r.simulationArea.crashTrigger {
				// Carry out the corresponding actions depending of if the planes will successfully evade each orther or not
				time.AfterFunc(3*time.Second, func() {
					log.Printf("DISASTER OCCURED!: Plane %s and Plane %s CRASHED\n\n",
						planeSerial, otherPlaneSerial)
					fmt.Fprintf(tcasLog, "%s DISASTER OCCURED!: Plane %s and Plane %s CRASHED\n\n",
						time.Now().Format("2006-01-02 15:04:05"), planeSerial, otherPlaneSerial)
					fmt.Fprintf(f, "%s DISASTER OCCURED!: Plane %s and Plane %s CRASHED\n\n",
						time.Now().Format("2006-01-02 15:04:05"), planeSerial, otherPlaneSerial)

					// at this point, the simulation ends
					if r.simulationArea.simState.SimIsRunning {
						aviation.EmergencyStop(r.simulationArea.simState)
					}

				})
				r.simulationArea.crashTrigger = true

			}
		}

	}
	r.simulationArea.statusLabel.Refresh()

	for _, airport := range r.simulationArea.airports {
		airport.IDLabel.TextSize = 8 * r.simulationArea.zoomScales[r.simulationArea.zoomLevel]
		airport.IDLabel.Refresh()
	}

	r.Layout(r.simulationArea.Size()) // force refresh
}
