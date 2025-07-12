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
	if f.TakeoffTime.IsZero() {
		return "N/A (Flight not yet initiated)"
	}

	switch {
	case simTime.After(f.DestinationArrivalTime) && f.FlightStatus == "landed":
		return "100% (Landed)"
	case simTime.After(f.DestinationArrivalTime) && f.FlightStatus == "about to land":
		return "100% (About to land)"
	case simTime.After(f.TakeoffTime) && simTime.Before(f.DestinationArrivalTime):
		total := f.DestinationArrivalTime.Sub(f.TakeoffTime)
		elapsed := simTime.Sub(f.TakeoffTime)
		if total > 0 {
			pct := (float64(elapsed) / float64(total)) * 100
			return fmt.Sprintf("%.2f%% (As at %s)", pct, simTime.Format("15:04:05"))
		}
		return "0% (Invalid flight duration)"
	default:
		return "0% (Plane about to take off or still taking off)"
	}

}

// GetClosestApproachDetails calculates the time and minimum Distance at which two planes will be closest during their respective flights.
func (f1 Flight) GetClosestApproachDetails(f2 Flight) (closestTime time.Time, distanceBetweenPlanesatCA float64) {
	flight1ClosestCoord, flight2ClosestCoord := FindClosestApproachDuringTransit(f1.FlightSchedule, f2.FlightSchedule)

	flight1Distance := Distance(f1.FlightSchedule.Depature, f1.FlightSchedule.Destination)
	if flight1Distance == 0 {
		return f1.TakeoffTime, Distance(f1.FlightSchedule.Depature, f2.FlightSchedule.Depature)
	}

	distBtwDepatureAndClosestApproachForFlight1 := Distance(f1.FlightSchedule.Depature, flight1ClosestCoord)

	f1fractionofCA := distBtwDepatureAndClosestApproachForFlight1 / flight1Distance

	totalFlightDuration1 := f1.DestinationArrivalTime.Sub(f1.TakeoffTime)
	closestTime = f1.TakeoffTime.Add(time.Duration(float64(totalFlightDuration1) * f1fractionofCA))

	distanceBetweenPlanesatCA = Distance(flight1ClosestCoord, flight2ClosestCoord)

	return closestTime, distanceBetweenPlanesatCA
}
