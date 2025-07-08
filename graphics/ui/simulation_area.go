package ui

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Airport represents a single airport with its properties
type Airport struct {
	ID      int     // Serial number (e.g., 1 for "001")
	X, Y    float32 // Base coordinates (before pan/zoom)
	Image   *canvas.Image
	IDLabel *canvas.Text
}

// SimulationArea is the custom widget where the simulation (airport map) will be rendered.
type SimulationArea struct {
	widget.BaseWidget               // Embed BaseWidget for core widget functionality
	offsetX, offsetY  float32       // Current pan offset for drawing simulation elements
	lastPanPos        fyne.Position // Last mouse position for drag calculations
	statusLabel       *canvas.Text  // Label to display current status or offset

	airports           []*Airport    // Slice of all airports to draw
	airportImage       fyne.Resource // The base airport image resource
	initialAirportSize fyne.Size     // Base size of an airport image (5x5)

	zoomLevel  int       // 0.25, 0.5, 1, 2, 3
	zoomScales []float32 // Scale factors for each zoom level

	// Reference to the main window to return to it
	mainWindow fyne.Window

	stopTicker chan struct{}
}

// Ensure SimulationArea implements the necessary interfaces for a widget,
// mouse interactions, and dragging.
var _ fyne.Widget = (*SimulationArea)(nil)
var _ desktop.Mouseable = (*SimulationArea)(nil)
var _ fyne.Draggable = (*SimulationArea)(nil)

// NewSimulationArea creates a new SimulationArea widget.
func NewSimulationArea(numAirports int, mainWindow fyne.Window) *SimulationArea {
	airportImage, err := fyne.LoadResourceFromPath("assets/whiteAirport.png")
	if err != nil {
		log.Fatalf("Error loading airport image: %v. Make sure 'assets/whiteAirport.png' exists.", err)
	}

	sa := &SimulationArea{
		offsetX:            0,
		offsetY:            0,
		lastPanPos:         fyne.Position{},
		statusLabel:        canvas.NewText("Drag to pan | Zoom: 1x", color.RGBA{R: 0, G: 0, B: 0, A: 0}),
		airportImage:       airportImage,
		initialAirportSize: fyne.NewSize(70, 56),                // Base size
		zoomLevel:          2,                                   // Start at the base zoom level
		zoomScales:         []float32{0.25, 0.5, 1.0, 2.0, 3.0}, // Scales for 5x5, 10x10, 15x15 (relative to initial size)
		mainWindow:         mainWindow,
		stopTicker:         make(chan struct{}),
	}
	sa.statusLabel.Alignment = fyne.TextAlignCenter
	sa.statusLabel.TextSize = 8

	// Initialize the BaseWidget part of SimulationArea
	sa.BaseWidget.ExtendBaseWidget(sa) // This is how you initialize the embedded BaseWidget

	sa.generateAirports(numAirports)

	// Start a goroutine for continuous refresh
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond) // Refresh every 0.5 seconds
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fyne.Do(func() {
					sa.Refresh() // Call the widget's Refresh method
				})

			case <-sa.stopTicker:
				log.Println("SimulationArea refresh ticker stopped.")
				return
			}
		}
	}()

	return sa
}

// generateAirports creates the airport objects based on the input number.
func (sa *SimulationArea) generateAirports(num int) {
	const spacing = 250.0    // 250 units apart
	const airportsPerRow = 8 // Arbitrary number of airports per row for grid layout

	sa.airports = make([]*Airport, num)
	for i := 0; i < num; i++ {
		id := i + 1
		x := float32((i % airportsPerRow) * spacing)
		y := float32((i / airportsPerRow) * spacing)

		// Create a canvas.Image for each airport
		img := canvas.NewImageFromResource(sa.airportImage)
		img.SetMinSize(sa.initialAirportSize) // Set initial size

		// Create a label for the serial number
		label := canvas.NewText(fmt.Sprintf("%03d", id), color.White)
		label.TextSize = 8 // Small text for serial number

		sa.airports[i] = &Airport{
			ID:      id,
			X:       x,
			Y:       y,
			Image:   img,
			IDLabel: label,
		}
	}
}

// CreateRenderer is part of the fyne.Widget interface.
// It defines how the widget is drawn.
func (sa *SimulationArea) CreateRenderer() fyne.WidgetRenderer {
	// The objects slice will contain all airport images and their labels
	// and the status label. The background is handled separately.
	var objects []fyne.CanvasObject
	objects = append(objects, sa.statusLabel) // Status label is always on top

	// Add all airport images and labels to the renderable objects
	for _, airport := range sa.airports {
		objects = append(objects, airport.Image, airport.IDLabel)
	}

	return &simulationAreaRenderer{
		simulationArea: sa,
		objects:        objects,
		background:     canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 0}), // Transparent background
	}
}

// MouseDown captures the initial position for panning.
func (sa *SimulationArea) MouseDown(ev *desktop.MouseEvent) {
	sa.lastPanPos = ev.Position
	sa.Refresh() // Refresh to update status label
}

// Dragged updates the pan offset based on mouse movement.
// This implements the fyne.Draggable interface.
func (sa *SimulationArea) Dragged(ev *fyne.DragEvent) {
	dx := ev.Position.X - sa.lastPanPos.X
	dy := ev.Position.Y - sa.lastPanPos.Y

	sa.offsetX += dx
	sa.offsetY += dy

	sa.lastPanPos = ev.Position
	sa.Refresh()
}

// DragEnd is called when a drag operation finishes.
// This implements the fyne.Draggable interface.
func (sa *SimulationArea) DragEnd() {
	// Optional: Any cleanup or final state update after a drag
	sa.lastPanPos = fyne.Position{} // Reset last pan position
	sa.Refresh()
}

// MouseUp resets the last pan position.
// This implements the desktop.Mouseable interface.
func (sa *SimulationArea) MouseUp(ev *desktop.MouseEvent) {
	// For panning, MouseUp might not be strictly necessary if DragEnd handles the reset.
	// However, it's good to keep it if you have other non-drag mouse up interactions.
	// For now, we'll just log and refresh.
	sa.Refresh() // Refresh to update status label
}

// MouseMoved is implemented as part of desktop.Mouseable <<<<<<<<<< to be updated
func (sa *SimulationArea) MouseMoved(ev *desktop.MouseEvent) {
	// Implement show coordinate or plane on hover
}

// Home resets the view to the origin (0,0) with current zoom.
func (sa *SimulationArea) Home() {
	sa.offsetX = 0
	sa.offsetY = 0
	sa.Refresh()
	log.Println("Moved to origin (0,0)")
}

// Zoom increases sizes of object on render screen.
func (sa *SimulationArea) ZoomIn() {
	if sa.zoomLevel < 4 {
		sa.zoomLevel++
	}
	sa.Refresh()
	log.Printf("Zoom level changed to: %d (Scale: %.1fx)", sa.zoomLevel, sa.zoomScales[sa.zoomLevel])
}

// Zoom increases sizes of object on render screen.
func (sa *SimulationArea) ZoomOut() {
	if sa.zoomLevel > 0 {
		sa.zoomLevel--
	}
	sa.Refresh()
	log.Printf("Zoom level changed to: %d (Scale: %.1fx)", sa.zoomLevel, sa.zoomScales[sa.zoomLevel])
}

// Destroy is called when the widget is no longer needed.
func (sa *SimulationArea) Destroy() {
	close(sa.stopTicker) // Signal the goroutine to stop
	sa.BaseWidget.Hide() // <<<<<<<< return here to check functionality
}
