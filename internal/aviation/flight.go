package aviation

import (
	"fmt"
	"time"
)

// Flight represents a single flight from departure to arrival
type Flight struct {
	FlightID               string
	FlightSchedule         FlightPath
	TakeoffTime            time.Time
	DestinationArrivalTime time.Time
	CruisingAltitude       float64 // Meters
	DepatureAirPort        string
	ArrivalAirPort         string
	FlightStatus           string
	ActualLandingTime      time.Time
}

// FlightPath to store the movement of plane from one location to the other
type FlightPath struct {
	Depature    Coordinate
	Destination Coordinate
}

// GetFlightProgress calculates Progress made by plane in transit
func (f Flight) GetFlightProgress(simTime time.Time) string {

	if simTime.After(f.DestinationArrivalTime) && f.FlightStatus == "landed" {
		return "100% (Landed)"
	} else if simTime.After(f.DestinationArrivalTime) && f.FlightStatus == "about to land" {
		return "100% (About to land)"
	} else if simTime.After(f.TakeoffTime) && simTime.Before(f.DestinationArrivalTime) {
		totalDuration := f.DestinationArrivalTime.Sub(f.TakeoffTime)
		elapsedDuration := simTime.Sub(f.TakeoffTime)

		// Ensure totalDuration is not zero to prevent division by zero
		if totalDuration > 0 {
			completionPercentage := (float64(elapsedDuration) / float64(totalDuration)) * 100
			return fmt.Sprintf("%.2f%% (As at %s)", completionPercentage, simTime.Format("15:04:05"))
		} else {
			return "0% (Invalid flight duration)"
		}
	} else {
		return "0% (Plane about to take off or still taking off)"
	}
}
