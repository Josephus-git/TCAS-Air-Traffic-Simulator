package main

import "github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"

// emergencyStop safely halts the simulation by calling the core aviation emergency stop function.
func emergencyStop(simState *aviation.SimulationState) {
	aviation.EmergencyStop(simState)
}
