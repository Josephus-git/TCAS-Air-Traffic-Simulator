package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/josephus-git/TCAS-simulation-Fyne/internal/aviation"
)

// logDetails logs specific simulation details (airports, airplanes, or flights) based on the provided argument.
// It prints usage instructions if an invalid option is given.
func logDetails(simState *aviation.SimulationState, argument2 string) {
	switch argument2 {
	case "airports":
		logAirportDetails(simState)
	case "airplanes":
		logAirplanesDetails(simState)
	case "flights":
		logFlightDetailsToFile(simState)
	case "all":
		logAirportDetails(simState)
		logAirplanesDetails(simState)
		logFlightDetailsToFile(simState)
	default:
		fmt.Println("usage: log <option>, options: airports, airplanes, flights, all")
	}
}

// logAirPlanesDetails prints selected details of all flights logged in all various planes
func logFlightDetailsToFile(simState *aviation.SimulationState) {
	logFilePath := "logs/flightDetails.txt"
	// Open the file in append mode. Create it if it doesn't exist.
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()

	var simTime time.Time
	if simState.SimIsRunning {
		simTime = time.Now()
	} else {
		simTime = simState.SimEndedTime
	}
	var flightLogs []aviation.Flight

	fmt.Fprintln(f, "\n--- Log of all recorded flights ---")

	for _, airport := range simState.Airports {
		for _, plane := range airport.Planes {
			if len(plane.FlightLog) == 0 {
				continue
			}
			flightLogs = append(flightLogs, plane.FlightLog...)
		}
	}

	for _, plane := range simState.PlanesInFlight {
		flightLogs = append(flightLogs, plane.FlightLog...)
	}

	if len(flightLogs) == 0 {
		fmt.Fprintln(f, "\n--- No flight recorded currently ---")
		return
	}
	sort.Slice(flightLogs, func(i, j int) bool {
		return flightLogs[i].FlightID < flightLogs[j].FlightID
	})
	for i, flight := range flightLogs {
		fmt.Fprintf(f, "\nflightLog %d:\n", i)
		logFlightDetails(flight, simTime, f)
	}
	fmt.Println("successfully logged all flights")
}

// logAirplanesDetails appends selected details of all airplanes from the simulation state to a log file.
// It includes serial, flight status, cruise speed, and a count of flights for each plane.
func logAirplanesDetails(simState *aviation.SimulationState) {
	logFilePath := "logs/airplaneDetails.txt"
	// Open the file in append mode. Create it if it doesn't exist.
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()

	var simTime time.Time
	if simState.SimIsRunning {
		simTime = time.Now()
	} else {
		simTime = simState.SimEndedTime
	}
	Planes := []*aviation.Plane{}

	for _, ap := range simState.Airports {
		Planes = append(Planes, ap.Planes...)
	}

	Planes = append(Planes, simState.PlanesInFlight...)
	sort.Slice(Planes, func(i, j int) bool {
		return Planes[i].Serial < Planes[j].Serial
	})

	fmt.Fprintln(f, "\n--- Logging selected fields for each plane in ---")
	if len(Planes) == 0 {
		fmt.Fprintln(f, "\n--- No plane recorded currently ---")
		return
	}
	for i, plane := range Planes {
		fmt.Fprintf(f, "Plane %d (Serial: %s):\n", i+1, plane.Serial)
		fmt.Fprintf(f, "  In Flight: %t\n", plane.PlaneInFlight)
		fmt.Fprintf(f, "  Cruise Speed: %.2f m/s\n", plane.CruiseSpeed)
		fmt.Fprintf(f, "  TCAS Capability: %s\n", func(capability aviation.TCASCapability) string {
			if capability == 1 {
				return "Working Perfectly"
			} else {
				return "Faulty"
			}
		}(plane.TCASCapability))
		fmt.Fprintln(f, "  Flight Log:")
		if len(plane.FlightLog) == 0 {
			fmt.Fprintln(f, "    No flights recorded for this plane.")
		} else {
			for _, flight := range plane.FlightLog { // Looping to count flights, but not printing content if 'flight' is empty
				logFlightDetails(flight, simTime, f)
			}
		}
		if len(plane.TCASEngagementRecords) == 0 {
			fmt.Fprintln(f, "    No TCAS engagement recorded for this plane.")
		} else {
			for _, engagement := range plane.TCASEngagementRecords {
				logEngagementDetails(engagement, f)
			}
		}
		if plane.CurrentTCASEngagement == nil {
			fmt.Fprintln(f, "    No current TCAS engagement recorded for this plane.")
		} else {
			logEngagementDetails(*plane.CurrentTCASEngagement, f)
		}

		fmt.Fprintln(f, "-------------------------------------------")
	}

	fmt.Println("Successfully logged airplanes")
}

// logAirportDetails appends selected details of all airports from the simulation state to a log file.
// It includes serial, location, plane capacity, runway information, and a list of associated plane serials.
func logAirportDetails(simState *aviation.SimulationState) {
	logFilePath := "logs/airportDetails.txt"
	// Open the file in append mode. Create it if it doesn't exist.
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	defer f.Close()
	fmt.Fprintln(f, "\n--- Logging selected fields for each airport ---")
	if len(simState.Airports) == 0 {
		fmt.Fprintln(f, "\n--- No airport recorded currently ---")
		return
	}
	for i, ap := range simState.Airports {
		fmt.Fprintf(f, "Airport %d (Serial: %s):\n", i+1, ap.Serial)
		fmt.Fprintf(f, "  Location: %v\n", ap.Location)
		fmt.Fprintf(f, "  Runway: %v\n", ap.Runway)
		fmt.Fprintln(f, "  Planes:")
		if len(ap.Planes) == 0 {
			fmt.Fprintln(f, "    No Planes currently.")
		} else {
			for j, p := range ap.Planes {
				fmt.Fprintf(f, "    %d. Serial: %s\n", j+1, p.Serial)
			}
		}
		fmt.Fprintln(f, "-------------------------------------------")
	}
	fmt.Println("Successfully logged airports")
}

// logFlightDetails logs all details for a given Flight struct,
func logFlightDetails(flight aviation.Flight, simTime time.Time, f *os.File) {
	fmt.Fprintln(f, "    --- Flight Details ---")
	fmt.Fprintf(f, "    Flight ID: %s\n", flight.FlightID)
	fmt.Fprintf(f, "    Takeoff Time: %s\n", flight.TakeoffTime.Format("15:04:05"))
	fmt.Fprintf(f, "    Destination Arrival Time: %s\n", flight.DestinationArrivalTime.Format("15:04:05"))
	fmt.Fprintf(f, "    Cruising Altitude: %.2f meters\n", flight.CruisingAltitude)
	fmt.Fprintf(f, "    Depature Airport: %s\n", flight.DepatureAirPort)
	fmt.Fprintf(f, "    Destination Airport: %s\n", flight.ArrivalAirPort)
	var actualLandingTime string
	if flight.ActualLandingTime.IsZero() {
		actualLandingTime = "Plane is yet to land"
	} else {
		actualLandingTime = flight.ActualLandingTime.Format("15:04:05")
	}
	fmt.Fprintf(f, "    Actual Landing Time: %s\n", actualLandingTime)

	// calculate progress
	progress := flight.GetFlightProgressString(simTime)

	fmt.Fprintf(f, "    Progress: %s\n", progress)
	fmt.Fprintln(f, "    ---------------------------------------")
}

// logEngagementDetails writes the detailed information of a TCASEngagement to the provided file.
func logEngagementDetails(engagement aviation.TCASEngagement, f *os.File) {
	fmt.Fprintln(f, "    --- Engagement Details ---")
	fmt.Fprintf(f, "    Engagement ID: %s\n", engagement.EngagementID)
	fmt.Fprintf(f, "    Flight ID: %s\n", engagement.FlightID)
	fmt.Fprintf(f, "    Plane Serial: %s\n", engagement.PlaneSerial)
	fmt.Fprintf(f, "    Other Plane Serial: %s\n", engagement.OtherPlaneSerial)
	fmt.Fprintf(f, "    Time Of Engagement: %s\n", engagement.TimeOfEngagement.Format("15:04:05"))
	fmt.Fprintf(f, "    Will Crash: %s\n", func(willCrash bool) string {
		if engagement.WillCrash {
			return "yes"
		} else {
			return "no"
		}
	}(engagement.WillCrash))
}
