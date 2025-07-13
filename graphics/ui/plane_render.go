package ui

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"

	"github.com/disintegration/imaging"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
)

// PlaneRender represents a single plane with its properties for rendering.
type PlaneRender struct {
	ActualPlane    *aviation.Plane
	Image          *canvas.Image
	ImageLocation  fyne.Position
	FlightPathLine *canvas.Line
	TCASCircle     *canvas.Circle
}

// AddPlaneToRender adds a new PlaneRender object to the simulation area.
// This function will be called by the aviation package via the registered callback in simState.OnPlaneTakeoff.
func (sa *SimulationArea) AddPlaneToRender(plane *aviation.Plane) {
	image := canvas.NewImageFromResource(sa.airplaneImage)
	image.Hidden = true                      // Start hidden, will be shown when position is updated
	image.SetMinSize(sa.initialAirplaneSize) // Set initial size

	currentFlight := plane.FlightLog[len(plane.FlightLog)-1]
	rotation := planeOrientation(currentFlight.FlightSchedule.Depature, currentFlight.FlightSchedule.Destination)
	rotatedImg, err := RotateCanvasImage(image, rotation)
	if err != nil {
		log.Printf("Failed to rotate plane image: %v", err)
		return // silently skip rendering this plane
	}
	canvas.Refresh(rotatedImg)

	var line *canvas.Line
	// Create a faint flight path line
	switch currentFlight.CruisingAltitude {
	case 11000.0:
		line = canvas.NewLine(color.RGBA{G: 200, A: 50}) // Light green, semi-transparent
	case 12000.0:
		line = canvas.NewLine(color.RGBA{B: 200, A: 50}) // Light blue, semi-transparent
	default:
		line = canvas.NewLine(color.RGBA{R: 200, G: 200, B: 200, A: 50}) // Light grey, semi-transparent
	}

	line.StrokeWidth = 1
	line.Hidden = true // Start hidden

	planeRender := &PlaneRender{
		ActualPlane:    plane,
		Image:          rotatedImg,
		FlightPathLine: line,
		TCASCircle:     canvas.NewCircle(color.Transparent),
	}

	planeRender.TCASCircle.StrokeWidth = 3 // Set a default stroke width
	planeRender.TCASCircle.Hidden = true   // Start hidden

	sa.planesInFlight = append(sa.planesInFlight, planeRender)
	// Refresh renderer to include new objects
	sa.Refresh()
}

// RemovePlaneFromRender removes a PlaneRender object from the simulation area.
// This function will be called by the aviation package via a registered callback in simState.OnPlaneLand.
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
			if p.TCASCircle != nil { // Hide the circle if it exists
				p.TCASCircle.Hide()
			}
			break
		}
	}
	sa.Refresh()
}

// planeOrientation calculates the rotation angle for the plane image.
// Angle is in degrees
func planeOrientation(dep, dest aviation.Coordinate) float64 {
	// Delta X and Delta Y
	dx := dep.X - dest.X
	dy := dep.Y - dest.Y

	angleFromYClockwiseRad := float64(math.Atan2(dx, dy))
	rotation := angleFromYClockwiseRad * (180 / math.Pi)

	return rotation
}

// RotateCanvasImage rotates a canvas.Image by a given angle (in degrees).
func RotateCanvasImage(img *canvas.Image, angle float64) (*canvas.Image, error) {
	if img.Resource == nil {
		return nil, errors.New("image has no resource")
	}

	// Load the image from the resource
	reader := bytes.NewReader(img.Resource.Content())

	srcImg, format, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	// Rotate using the imaging library
	rotated := imaging.Rotate(srcImg, angle, image.Transparent)

	// Convert back to image bytes
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, rotated, nil)
	default: // fallback to png
		err = png.Encode(&buf, rotated)
	}
	if err != nil {
		return nil, err
	}

	// Create a new StaticResource with rotated image
	res := fyne.NewStaticResource("rotated.png", buf.Bytes())

	// Update the canvas.Image
	img.Resource = res
	img.Refresh()

	return img, nil
}
