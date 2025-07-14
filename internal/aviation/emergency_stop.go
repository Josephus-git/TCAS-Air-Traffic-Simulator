package aviation

import (
	"log"
)

// EmergencyStop immediately halts the simulation, canceling all active goroutines and resetting the simulation state.
func EmergencyStop(simState *SimulationState) {
	if simulationCancelFunc != nil {
		log.Println("\n--- EMERGENCY STOP ACTIVATED! Signaling all goroutines to stop... ---")
		simulationCancelFunc() // Trigger cancellation
		// Reset the cancel func to indicate no active simulation,
		// and prevent multiple calls to a potentially nil context if Start() finished.
		if stopTrigger.Stop() {
		}
		simulationCancelFunc = nil

	} else {
		log.Println("EmergencyStop: Simulation not running")
	}
}
