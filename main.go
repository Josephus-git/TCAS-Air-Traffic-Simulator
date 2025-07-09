package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/config"
	"github.com/josephus-git/TCAS-simulation-Fyne/internal/util"
)

func main() {
	util.ResetLog()
	start()
}

// start initializes the TCAS simulator, loads configurations, and enters a continuous command-line interaction loop.
func start() {
	scanner := bufio.NewScanner(os.Stdin)
	initialize := &config.Config{
		IsRunning: true,
		FirstRun:  true,
	}
	simState := &aviation.SimulationState{}

	aviation.GetNumberOfPlanes(initialize)
	aviation.InitializeAirports(initialize, simState)

	for i := 0; initialize.IsRunning; i++ {
		fmt.Print("TCAS-simulator > ")
		scanner.Scan()
		input := util.CleanInput(scanner.Text())
		argument2 := ""
		if len(input) > 1 {
			argument2 = input[1]
		}

		if len(input) == 0 {
			fmt.Println("")
			continue
		}

		cmd, ok := getCommand(initialize, simState, argument2)[input[0]]
		if !ok {
			fmt.Println("Unknown command, type <help> for usage")
			continue
		}
		cmd.callback()

		println("")
	}
	restartApplication()
}
