package ui

import (
	"image"
	"image/color"
	"math"
	"os"
	"time"

	"github.com/disintegration/imaging"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
)

// PlaneRender represents a single plane with its properties for rendering.
type PlaneRender struct {
	ActualPlane    *aviation.Plane
	Image          *canvas.Image
	FlightPathLine *canvas.Line // To draw the flight path
	// We might need to store the initial and destination points of the line in canvas coordinates
	// so we can dynamically adjust the start point of the visible line.
	// For now, FlightPathLine's Position1 and Position2 will be updated directly.
}

// AddPlaneToRender adds a new PlaneRender object to the simulation area.
// This function will be called by the aviation package via a registered callback.
func (sa *SimulationArea) AddPlaneToRender(plane *aviation.Plane) {
	img := canvas.NewImageFromResource(sa.airplaneImage)
	img.SetMinSize(sa.initialAirplaneSize) // Set initial size
	img.Hidden = true                      // Start hidden, will be shown when position is updated

	// Create a faint flight path line
	line := canvas.NewLine(color.RGBA{R: 200, G: 200, B: 200, A: 100}) // Light grey, semi-transparent
	line.StrokeWidth = 1
	line.Hidden = true // Start hidden

	planeRender := &PlaneRender{
		ActualPlane:    plane,
		Image:          img,
		FlightPathLine: line,
	}

	sa.planesInFlight = append(sa.planesInFlight, planeRender)
	// Refresh renderer to include new objects
	sa.Refresh()
}

// RemovePlaneFromRender removes a PlaneRender object from the simulation area.
// This function will be called by the aviation package via a registered callback.
func (sa *SimulationArea) RemovePlaneFromRender(planeSerial string) {
	for i, p := range sa.planesInFlight {
		if p.ActualPlane.Serial == planeSerial {
			// Remove from slice
			sa.planesInFlight = append(sa.planesInFlight[:i], sa.planesInFlight[i+1:]...)
			// Ensure objects are removed from canvas. This happens implicitly on Refresh/Layout
			// if the object is no longer in the renderer's `objects` slice, but explicitly hiding
			// them immediately might be safer.
			p.Image.Hide()
			p.FlightPathLine.Hide()
			break
		}
	}
	sa.Refresh()
}

// planeCurrentPosition calculates the current position of a plane along its flight path.
// This is crucial for real-time animation.
func planeCurrentPosition(plane *aviation.Plane, simTime time.Time) (aviation.Coordinate, bool) {
	if len(plane.FlightLog) == 0 {
		return aviation.Coordinate{}, false
	}

	currentFlight := plane.FlightLog[len(plane.FlightLog)-1]

	if simTime.Before(currentFlight.TakeoffTime) {
		// Plane hasn't taken off yet, return its departure airport's location
		return currentFlight.FlightSchedule.Depature, true
	} else if simTime.After(currentFlight.DestinationArrivalTime) {
		// Plane has landed, return its destination airport's location
		// The UI should handle removing the plane if it's considered fully landed
		return currentFlight.FlightSchedule.Destination, true
	} else {
		// Plane is in transit
		totalDuration := float64(currentFlight.DestinationArrivalTime.Sub(currentFlight.TakeoffTime))
		elapsedDuration := float64(simTime.Sub(currentFlight.TakeoffTime))

		if totalDuration == 0 { // Avoid division by zero
			return currentFlight.FlightSchedule.Depature, true
		}

		// Interpolation factor (0.0 at takeoff, 1.0 at arrival)
		t := elapsedDuration / totalDuration

		// Linear interpolation for X, Y, Z
		x := currentFlight.FlightSchedule.Depature.X + t*(currentFlight.FlightSchedule.Destination.X-currentFlight.FlightSchedule.Depature.X)
		y := currentFlight.FlightSchedule.Depature.Y + t*(currentFlight.FlightSchedule.Destination.Y-currentFlight.FlightSchedule.Depature.Y)
		z := currentFlight.FlightSchedule.Depature.Z + t*(currentFlight.FlightSchedule.Destination.Z-currentFlight.FlightSchedule.Depature.Z)

		return aviation.Coordinate{X: x, Y: y, Z: z}, true
	}
}

// planeOrientation calculates the rotation angle for the plane image.
// Angle is in degrees, clockwise from positive Y-axis (Fyne's default).
func planeOrientation(dep, dest aviation.Coordinate) float64 {
	// Delta X and Delta Y
	dx := dest.X - dep.X
	dy := dest.Y - dep.Y

	// Adjust for Fyne's coordinate system (0 deg usually points up/North, increasing clockwise)
	// If Atan2 gives angle from positive X, then:
	//  - If dx > 0, dy = 0 (East): angle is 0 deg. We want 90 deg.
	//  - If dx = 0, dy > 0 (North): angle is 90 deg. We want 0 deg.
	//  - If dx < 0, dy = 0 (West): angle is 180 deg. We want 270 deg.
	//  - If dx = 0, dy < 0 (South): angle is -90 deg (or 270 deg). We want 180 deg.

	// A common way to get "heading" from Y-axis (North = 0) clockwise:
	// angle = 90 - (angle in degrees from X axis)
	// If Atan2(-dy, dx) is used, 0 deg is positive X, then 90 - result might work.
	// Let's try Atan2(dx, -dy) for angle from positive Y-axis clockwise.
	angleFromYClockwiseRad := float64(math.Atan2(dx, -dy))
	rotation := angleFromYClockwiseRad * (180 / math.Pi)

	return rotation
}

// RotateCanvasImage rotates the image associated with a canvas.Image by a specified angle
func RotateCanvasImage(img *canvas.Image, angle float64) (*canvas.Image, error) {
	var src image.Image

	switch v := img.Image.(type) {
	case image.Image:
		src = v
	default:
		// If the image wasn't already loaded (e.g. loaded from resource), load from the file path
		if img.File != "" {
			file, err := os.Open(img.File)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			decoded, _, err := image.Decode(file)
			if err != nil {
				return nil, err
			}
			src = decoded
		} else {
			return nil, nil // or return an error if you prefer
		}
	}

	// Rotate using the imaging package
	rotated := imaging.Rotate(src, angle, color.Transparent)

	// Create a new canvas.Image from the rotated image
	newImg := canvas.NewImageFromImage(rotated)
	newImg.SetMinSize(fyne.NewSize(float32(rotated.Bounds().Dx()), float32(rotated.Bounds().Dy())))
	newImg.FillMode = canvas.ImageFillContain

	return newImg, nil
}
