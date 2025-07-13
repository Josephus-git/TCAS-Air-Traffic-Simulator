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
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
)

// Airport represents a single airport with its properties
type AirportRender struct {
	ActualAirport *aviation.Airport
	Image         *canvas.Image
	IDLabel       *canvas.Text
}

// SimulationArea is the custom widget where the simulation (airport map) will be rendered.
type SimulationArea struct {
	widget.BaseWidget               // Embed BaseWidget for core widget functionality
	offsetX, offsetY  float32       // Current pan offset for drawing simulation elements
	lastPanPos        fyne.Position // Last mouse position for drag calculations
	statusLabel       *canvas.Text  // Label to display current status or offset

	airports           []*AirportRender // Slice of all airports to draw
	airportImage       fyne.Resource    // The base airport image resource
	initialAirportSize fyne.Size        // Base size of an airport image (5x5)

	planesInFlight      []*PlaneRender // NEW: Slice of all planes currently in flight to draw
	airplaneImage       fyne.Resource  // NEW: The base airplane image resource
	initialAirplaneSize fyne.Size      // NEW: Base size of an airplane image

	zoomLevel  int       // 0.25, 0.5, 1, 2, 3
	zoomScales []float32 // Scale factors for each zoom level

	// Reference to the main window to return to it
	mainWindow fyne.Window

	// Callbacks for plane take-off and landing notifications from aviation package
	PlaneTakeOffCallback func(*aviation.Plane)
	PlaneLandCallback    func(string) // Pass plane serial for removal

	// Timer for updating plane positions
	animationTicker *time.Ticker
	simState        *aviation.SimulationState
}

// Ensure SimulationArea implements the necessary interfaces for a widget,
// mouse interactions, and dragging.
var _ fyne.Widget = (*SimulationArea)(nil)
var _ desktop.Mouseable = (*SimulationArea)(nil)
var _ fyne.Draggable = (*SimulationArea)(nil)

// NewSimulationArea creates a new SimulationArea widget.
func NewSimulationArea(simState *aviation.SimulationState, mainWindow fyne.Window) *SimulationArea {
	airportImage, err := fyne.LoadResourceFromPath("assets/whiteAirport.png")
	if err != nil {
		log.Fatalf("Error loading airport image: %v. Make sure 'assets/whiteAirport.png' exists.", err)
	}

	// NEW: Load airplane image
	airplaneImage, err := fyne.LoadResourceFromPath("assets/whiteAirplane.png")
	if err != nil {
		log.Fatalf("Error loading airplane image: %v. Make sure 'assets/whiteAirplane.png' exists.", err)
	}

	sa := &SimulationArea{
		offsetX:             700,
		offsetY:             300,
		lastPanPos:          fyne.Position{},
		statusLabel:         canvas.NewText("Drag to pan | Zoom: 1x", color.RGBA{R: 0, G: 0, B: 0, A: 0}),
		airportImage:        airportImage,
		initialAirportSize:  fyne.NewSize(70, 56),                // Base size
		airplaneImage:       airplaneImage,                       // NEW
		initialAirplaneSize: fyne.NewSize(30, 24),                // NEW: Smaller initial size for planes
		zoomLevel:           2,                                   // Start at the base zoom level
		zoomScales:          []float32{0.25, 0.5, 1.0, 2.0, 3.0}, // Scales (relative to initial size)
		mainWindow:          mainWindow,
		planesInFlight:      []*PlaneRender{}, // Initialize empty slice
		simState:            simState,
	}
	sa.statusLabel.Alignment = fyne.TextAlignCenter
	sa.statusLabel.TextSize = 8

	// Initialize the BaseWidget part of SimulationArea
	sa.BaseWidget.ExtendBaseWidget(sa) // This is how you initialize the embedded BaseWidget

	sa.generateAirportsToRender(simState)

	// NEW: Register callbacks with the simulation state
	simState.OnPlaneTakeOffCallback = sa.AddPlaneToRender
	simState.OnPlaneLandCallback = sa.RemovePlaneFromRender

	// NEW: Start a ticker for continuous animation updates
	sa.animationTicker = time.NewTicker(50 * time.Millisecond) // Update 20 times per second
	go func() {
		for range sa.animationTicker.C {
			simState.Mu.Lock()
			simState.CurrentSimTime = time.Now()
			simState.Mu.Unlock()
			if sa.Size().IsZero() { // Don't refresh if widget hasn't been laid out yet
				continue
			}
			// This Do() ensures UI updates happen on the main goroutine
			fyne.Do(func() {
				sa.Refresh()
			})
		}
	}()

	return sa
}

// generateAirports creates the airport objects based on the input number.
func (sa *SimulationArea) generateAirportsToRender(simState *aviation.SimulationState) {

	airportNum := len(simState.Airports)

	sa.airports = make([]*AirportRender, airportNum)
	for i, actualAirport := range simState.Airports {

		// Create a canvas.Image for each airport
		img := canvas.NewImageFromResource(sa.airportImage)
		img.SetMinSize(sa.initialAirportSize) // Set initial size

		// Create a label for the serial number
		label := canvas.NewText(fmt.Sprintf("%s (%.1f,%.1f)",
			actualAirport.Serial, actualAirport.Location.X, actualAirport.Location.Y), color.White)
		label.TextSize = 8 * sa.zoomScales[sa.zoomLevel] // Small text for serial number

		sa.airports[i] = &AirportRender{
			ActualAirport: actualAirport,
			Image:         img,
			IDLabel:       label,
		}
	}
}

// CreateRenderer is part of the fyne.Widget interface.
// It defines how the widget is drawn.
func (sa *SimulationArea) CreateRenderer() fyne.WidgetRenderer {

	return &simulationAreaRenderer{
		simulationArea: sa,
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
	sa.offsetX = 700
	sa.offsetY = 300
	sa.Refresh()
}

// Zoom increases sizes of object on render screen.
func (sa *SimulationArea) ZoomIn() {
	if sa.zoomLevel < 4 {
		sa.zoomLevel++
	}
	sa.Refresh()
}

// Zoom increases sizes of object on render screen.
func (sa *SimulationArea) ZoomOut() {
	if sa.zoomLevel > 0 {
		sa.zoomLevel--
	}
	sa.Refresh()
}

func (sa *SimulationArea) ClearAllResource() {
	for _, p := range sa.planesInFlight {
		if p.Image != nil {
			p.Image.Hide()
		}
		if p.FlightPathLine != nil {
			p.FlightPathLine.Hide()
		}
		if p.TCASCircle != nil {
			p.TCASCircle.Hide()
		}
	}
	sa.planesInFlight = []*PlaneRender{} // Reset the slice
	sa.airports = []*AirportRender{}

	sa.Refresh()
}
