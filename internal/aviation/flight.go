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

// CheckPlaneStatusAtTime checks the status of a plane at a specific time based on its flight log.
func (flight Flight) checkPlaneStatusAtTime(checkTime time.Time) string {
	if checkTime.After(flight.DestinationArrivalTime) {
		// If checkTime is after arrival, plane has landed for this flight
		return "landed or still landing"
	}
	if checkTime.After(flight.TakeoffTime) && checkTime.Before(flight.DestinationArrivalTime) {
		if flight.FlightStatus == "about to land" {
			return "about to land"
		}
		return "in transit"
	}
	if checkTime.Equal(flight.TakeoffTime) {
		return "taking off"
	}
	if checkTime.Equal(flight.DestinationArrivalTime) {
		return "arriving"
	}
	return "parked/unknown" // If no flight matches the time, assume parked or not in a known flight.
}

// GetClosestApproachDetails calculates the time and minimum Distance at which two planes will be closest during their respective flights.
func (f1 Flight) GetClosestApproachDetails(simState *SimulationState, f2 Flight) (distanceAtCA float64, closestTimeForPlane1 time.Time, otherPlaneStatusAtCATime string) {
	flight1Distance := Distance(f1.FlightSchedule.Depature, f1.FlightSchedule.Destination)
	if flight1Distance == 0.0 { // to avoid division by 0
		return 99999.99, f1.TakeoffTime, ""
	}

	// first check parralel situation i.e the depature of flight 1 equals arival of flight 2 and vice versa
	if (Distance(f1.FlightPath.Depature, f2.FlightPath.Destination) < Epsilon) && (Distance(f1.FlightPath.Destination, f2.FlightPath.Depature) < Epsilon) {
		flight2Progress := f2.getFlightProgress(simState.CurrentSimTime)
		flight2DistanceCovered := flight1Distance * flight2Progress
		// Calculating for time to collision
		flightToCover := flight1Distance - flight2DistanceCovered
		secondsFloat := (flightToCover / 2) / CruiseSpeed // Distance / Speed = Time

		// Convert float64 seconds to time.Duration
		timeToCollision := time.Duration(secondsFloat)

		closestTimeForPlane1 = f1.TakeoffTime.Add(timeToCollision)
		//get other planes status at this time
		otherPlaneStatusAtCATime = f2.checkPlaneStatusAtTime(closestTimeForPlane1)

		return 0.0, closestTimeForPlane1, otherPlaneStatusAtCATime
	}

	flight1ClosestCoord, _ := FindClosestApproachDuringTransit(f1.FlightSchedule, f2.FlightSchedule)

	distBtwDepatureAndClosestApproachForFlight1 := Distance(f1.FlightSchedule.Depature, flight1ClosestCoord)

	f1fractionofCA := distBtwDepatureAndClosestApproachForFlight1 / flight1Distance

	totalFlightDuration1 := f1.DestinationArrivalTime.Sub(f1.TakeoffTime)
	closestTimeForPlane1 = f1.TakeoffTime.Add(time.Duration(float64(totalFlightDuration1) * f1fractionofCA))

	Flight2CoordAtCATime1 := f2.getPlanePosition(closestTimeForPlane1)

	distanceAtCA = Distance(flight1ClosestCoord, Flight2CoordAtCATime1)

	otherPlaneStatusAtCATime = f2.checkPlaneStatusAtTime(closestTimeForPlane1)

	return distanceAtCA, closestTimeForPlane1, otherPlaneStatusAtCATime
}

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
