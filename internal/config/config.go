package config

// Config holds the simulation's configuration parameters.
type Config struct {
	NoOfAirplanes      int
	DifferentAltitudes bool
	FirstRun           bool // must be true only in the first oppening of the application, otherwise trying to open another instance of the fyne application will crash the program
}
