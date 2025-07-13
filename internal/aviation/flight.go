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
	FlightPath             FlightPath
}

// FlightPath to store the movement of plane from one location to the other
type FlightPath struct {
	Depature    Coordinate
	Destination Coordinate
}

// GetFlightProgress calculates Progress made by plane in transit
func (f Flight) GetFlightProgressString(simTime time.Time) string {
	if f.TakeoffTime.IsZero() {
		return "N/A (Flight not yet initiated)"
	}

	switch {
	case simTime.After(f.DestinationArrivalTime) && f.FlightStatus == "landed":
		return "100% (Landed)"
	case simTime.After(f.DestinationArrivalTime) && f.FlightStatus == "about to land":
		return "100% (About to land)"
	case simTime.After(f.TakeoffTime) && simTime.Before(f.DestinationArrivalTime):
		pct := f.getFlightProgress(simTime)
		return fmt.Sprintf("%.2f%% (As at %s)", pct, simTime.Format("15:04:05"))
	default:
		return "0% (Plane about to take off or still taking off)"
	}

}

// getFlightProgress calculates the percentage of a flight's duration covered at a given simulation time.
func (flight Flight) getFlightProgress(simTime time.Time) (percentageCovered float64) {
	total := flight.DestinationArrivalTime.Sub(flight.TakeoffTime)
	elapsed := simTime.Sub(flight.TakeoffTime)
	if total > 0 {
		return (float64(elapsed) / float64(total)) * 100
	} else if simTime.After(flight.DestinationArrivalTime) {
		return 100.00
	} else {
		return 0.0
	}
}
