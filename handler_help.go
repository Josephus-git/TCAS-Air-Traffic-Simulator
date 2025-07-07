package main

import (
	"fmt"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/config"
)

// helpFunc displays a welcome message and lists all available commands with their descriptions.
func helpFunc(cfg *config.Config, simState *aviation.SimulationState, argument2 string) {
	fmt.Print("Welcome to TCAS-simulator!\nUsage\n\n")
	for key := range getCommand(cfg, simState, argument2) {
		fmt.Printf("%s: %s\n", getCommand(cfg, simState, argument2)[key].name, getCommand(cfg, simState, argument2)[key].description)
	}
}
