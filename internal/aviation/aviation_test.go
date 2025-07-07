package aviation

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// Test helpers for float comparison:
// FloatEqualityThreshold defines the tolerance for comparing floating-point numbers.
const FloatEqualityThreshold = 1e-5

// FloatEquals compares two float64 numbers for approximate equality.
func FloatEquals(a, b float64) bool {
	return math.Abs(a-b) < FloatEqualityThreshold
}

// CoordEquals compares two Coordinate structs for approximate equality.
func CoordEquals(c1, c2 Coordinate) bool {
	return FloatEquals(c1.X, c2.X) && FloatEquals(c1.Y, c2.Y) && FloatEquals(c1.Z, c2.Z)
}

// TestFindClosestApproachDuringTransit verifies the accuracy of FindClosestApprachDuringTransit
// by testing various flight path configurations, including intersecting, parallel, and skew paths.
func TestFindClosestApproachDuringTransit(t *testing.T) {
	tests := []struct {
		name    string
		fp1     FlightPath
		fp2     FlightPath
		wantFp1 Coordinate
		wantFp2 Coordinate
	}{
		{
			name: "Intersecting Paths",
			fp1: FlightPath{
				Depature:    Coordinate{X: 0, Y: 0, Z: 0},
				Destination: Coordinate{X: 10, Y: 0, Z: 0},
			},
			fp2: FlightPath{
				Depature:    Coordinate{X: 5, Y: -5, Z: 0},
				Destination: Coordinate{X: 5, Y: 5, Z: 0},
			},
			wantFp1: Coordinate{X: 5, Y: 0, Z: 0},
			wantFp2: Coordinate{X: 5, Y: 0, Z: 0},
		},
		{
			name: "Parallel Paths (non-overlapping)",
			fp1: FlightPath{
				Depature:    Coordinate{X: 0, Y: 0, Z: 0},
				Destination: Coordinate{X: 10, Y: 0, Z: 0},
			},
			fp2: FlightPath{
				Depature:    Coordinate{X: 0, Y: 1, Z: 0},
				Destination: Coordinate{X: 10, Y: 1, Z: 0},
			},
			wantFp1: Coordinate{X: 0, Y: 0, Z: 0},
			wantFp2: Coordinate{X: 0, Y: 1, Z: 0},
		},
		{
			name: "Skew Paths (non-intersecting, 3D)",
			fp1: FlightPath{
				Depature:    Coordinate{X: 0, Y: 0, Z: 0},
				Destination: Coordinate{X: 10, Y: 0, Z: 0},
			},
			fp2: FlightPath{
				Depature:    Coordinate{X: 0, Y: 10, Z: 10},
				Destination: Coordinate{X: 10, Y: 10, Z: 0},
			},
			wantFp1: Coordinate{X: 10, Y: 0, Z: 0},
			wantFp2: Coordinate{X: 10, Y: 10, Z: 0},
		},
		{
			name: "Endpoint to Endpoint (closest is an end point)",
			fp1: FlightPath{
				Depature:    Coordinate{X: 0, Y: 0, Z: 0},
				Destination: Coordinate{X: 1, Y: 0, Z: 0},
			},
			fp2: FlightPath{
				Depature:    Coordinate{X: 10, Y: 0, Z: 0},
				Destination: Coordinate{X: 11, Y: 0, Z: 0},
			},
			wantFp1: Coordinate{X: 1, Y: 0, Z: 0},
			wantFp2: Coordinate{X: 10, Y: 0, Z: 0},
		},
		{
			name: "Identical Paths (should return start points)",
			fp1: FlightPath{
				Depature:    Coordinate{X: 0, Y: 0, Z: 0},
				Destination: Coordinate{X: 10, Y: 0, Z: 0},
			},
			fp2: FlightPath{
				Depature:    Coordinate{X: 0, Y: 0, Z: 0},
				Destination: Coordinate{X: 10, Y: 0, Z: 0},
			},
			wantFp1: Coordinate{X: 0, Y: 0, Z: 0},
			wantFp2: Coordinate{X: 0, Y: 0, Z: 0},
		},
		{
			name: "Collinear, Overlapping (Segment 1 contains Segment 2)",
			fp1: FlightPath{
				Depature:    Coordinate{X: 0, Y: 0, Z: 0},
				Destination: Coordinate{X: 10, Y: 0, Z: 0},
			},
			fp2: FlightPath{
				Depature:    Coordinate{X: 2, Y: 0, Z: 0},
				Destination: Coordinate{X: 8, Y: 0, Z: 0},
			},
			wantFp1: Coordinate{X: 2, Y: 0, Z: 0},
			wantFp2: Coordinate{X: 2, Y: 0, Z: 0},
		},
		{
			name: "Collinear, Non-overlapping (Segment 1 before Segment 2)",
			fp1: FlightPath{
				Depature:    Coordinate{X: 0, Y: 0, Z: 0},
				Destination: Coordinate{X: 5, Y: 0, Z: 0},
			},
			fp2: FlightPath{
				Depature:    Coordinate{X: 7, Y: 0, Z: 0},
				Destination: Coordinate{X: 10, Y: 0, Z: 0},
			},
			wantFp1: Coordinate{X: 5, Y: 0, Z: 0},
			wantFp2: Coordinate{X: 7, Y: 0, Z: 0},
		},
		{
			name: "Perpendicular, not intersecting, one endpoint is closest",
			fp1: FlightPath{
				Depature:    Coordinate{X: 0, Y: 0, Z: 0},
				Destination: Coordinate{X: 5, Y: 0, Z: 0},
			},
			fp2: FlightPath{
				Depature:    Coordinate{X: 0, Y: 5, Z: 0},
				Destination: Coordinate{X: 0, Y: 10, Z: 0},
			},
			wantFp1: Coordinate{X: 0, Y: 0, Z: 0},
			wantFp2: Coordinate{X: 0, Y: 5, Z: 0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotFp1, gotFp2 := FindClosestApproachDuringTransit(test.fp1, test.fp2)

			if !CoordEquals(gotFp1, test.wantFp1) {
				t.Errorf("\nFor %s\n fp1Closest: got %v, want %v\n", test.name, gotFp1, test.wantFp1)
			} else {
				fmt.Printf("\n\nSuccess For %s\n fp1Closest: got %v, want %v\n", test.name, gotFp1, test.wantFp1)
			}

			if !CoordEquals(gotFp2, test.wantFp2) {
				t.Errorf("\nFor %s\n fp2Closest: got %v, want %v\n", test.name, gotFp2, test.wantFp2)
			} else {
				fmt.Printf("\nSuccess For %s\n fp2Closest: got %v, want %v\n", test.name, gotFp2, test.wantFp2)
			}
		})
	}

}

// TestGetClosestApproachDetails is the main test function for your logic.
func TestGetClosestApproachDetails(t *testing.T) {
	baseTime := time.Date(2025, time.June, 19, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name                            string
		flight1                         Flight
		flight2                         Flight
		expectedClosestTime             time.Time
		expectedDistanceBetweenPlanesCA float64
		expectError                     bool // Use this if your function returns errors
		// Add expectedPanic bool if you expect a panic for certain inputs
	}{
		{
			name: "Sceaviation.Fario 1: Direct Intersection (Mid-flight)",
			flight1: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{0, 0, 0}, Destination: Coordinate{100, 0, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			flight2: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{50, -50, 0}, Destination: Coordinate{50, 50, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			expectedClosestTime:             baseTime.Add(5 * time.Minute),
			expectedDistanceBetweenPlanesCA: 0.0,
		},
		{
			name: "Scenario 2: Parallel Paths (Constant Distance)",
			flight1: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{0, 0, 0}, Destination: Coordinate{100, 0, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			flight2: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{0, 10, 0}, Destination: Coordinate{100, 10, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			expectedClosestTime:             baseTime.Add(0 * time.Minute),
			expectedDistanceBetweenPlanesCA: 10.0,
		},
		{
			name: "Scenario 3: Closest Approach at Departure (Flight 1 starts near Flight 2)",
			flight1: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{0, 0, 0}, Destination: Coordinate{100, 0, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			flight2: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{5, 0, 0}, Destination: Coordinate{101, 0, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			expectedClosestTime:             baseTime.Add(30 * time.Second),
			expectedDistanceBetweenPlanesCA: 0.0,
		},
		{
			name: "Scenario 4: Closest Approach at Destination (Flight 1 ends near Flight 2)",
			flight1: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{0, 0, 0}, Destination: Coordinate{100, 0, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			flight2: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{-100, 0, 0}, Destination: Coordinate{0, 0, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			expectedClosestTime:             baseTime.Add(0 * time.Minute),
			expectedDistanceBetweenPlanesCA: 100.0, // distance between (100,0,0) and (0,0,0)
		},
		{
			name: "Scenario 5: Different Start Times, Same Intersection Point Geometrically",
			flight1: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{0, 0, 0}, Destination: Coordinate{100, 0, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			flight2: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{50, -50, 0}, Destination: Coordinate{50, 50, 0}},
				TakeoffTime:            baseTime.Add(2 * time.Minute),
				DestinationArrivalTime: baseTime.Add(12 * time.Minute),
			},
			expectedClosestTime:             baseTime.Add(5 * time.Minute),
			expectedDistanceBetweenPlanesCA: 0.0,
		},
		{
			name: "Scenario 7: Closest Approach Asymmetric (Near Start F1, Near End F2)",
			flight1: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{0, 0, 0}, Destination: Coordinate{100, 0, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			flight2: Flight{
				FlightSchedule:         FlightPath{Depature: Coordinate{0, 100, 0}, Destination: Coordinate{100, 100, 0}},
				TakeoffTime:            baseTime,
				DestinationArrivalTime: baseTime.Add(10 * time.Minute),
			},
			expectedClosestTime:             baseTime,
			expectedDistanceBetweenPlanesCA: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closestTime, distanceBetweenPlanesatCA := tt.flight1.GetClosestApproachDetails(tt.flight2)
			// Assertions for each returned value.
			if !closestTime.Equal(tt.expectedClosestTime) {
				t.Errorf("%s: unexpected closestTime.\nExpected: %v\nActual:   %v", tt.name, tt.expectedClosestTime, closestTime)
			}

			if !FloatEquals(distanceBetweenPlanesatCA, tt.expectedDistanceBetweenPlanesCA) {
				t.Errorf("%s: unexpected distanceBetweenPlanesatCA.\nExpected: %v\nActual:   %v", tt.name, tt.expectedDistanceBetweenPlanesCA, distanceBetweenPlanesatCA)
			}
		})
	}
}
