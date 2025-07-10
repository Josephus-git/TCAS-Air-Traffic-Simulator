package main

import "github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"

func emergencyStop(simState *aviation.SimulationState) {
	aviation.EmergencyStop(simState)
}
