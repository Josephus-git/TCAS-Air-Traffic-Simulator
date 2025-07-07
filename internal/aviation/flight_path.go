package aviation

// FindClosestApproachDuringTransit returns closest points between flightpath 1 and flightpath
func FindClosestApproachDuringTransit(fp1, fp2 FlightPath) (fp1Closest, fp2Closest Coordinate) {
	p1 := fp1.Depature
	p2 := fp2.Depature
	q1 := fp1.Destination
	q2 := fp2.Destination
	// Segment 1: P1 + t*D1 (from p1 to q1)
	// Segment 2: P2 + u*D2 (from p2 to q2)
	D1 := q1.subtract(p1)
	D2 := q2.subtract(p2)
	R := p1.subtract(p2) // Vector from P2 to P1

	a := D1.dot(D1) // Squared length of D1
	e := D2.dot(D2) // Squared length of D2
	f := D2.dot(R)  // Dot product of D2 and R

	// Parallel or nearly parallel lines check
	const epsilon = 1e-6 // A small value to check for near-parallelism
	if a <= epsilon && e <= epsilon {
		// Both segments are points
		return p1, p2
	}
	if a <= epsilon {
		// First segment is a point
		s := clamp(f/e, 0, 1)
		return p1, p2.add(D2.mulScalar(s))
	}
	if e <= epsilon {
		// Second segment is a point
		s := clamp(-R.dot(D1)/a, 0, 1)
		return p1.add(D1.mulScalar(s)), p2
	}

	// General case for non-parallel lines/segments
	b := D1.dot(D2)
	c := D1.dot(R)
	denom := a*e - b*b

	var s, t float64

	if denom < epsilon { // Lines are nearly parallel
		t = 0.0 // Default to s=0
		s = clamp(-c/a, 0.0, 1.0)
	} else {
		s = clamp((b*f-c*e)/denom, 0.0, 1.0)
		t = (b*s + f) / e
	}

	// Clamp t if it falls outside [0,1] or if the lines are parallel.
	// This part is crucial for line *segments*
	if t < 0.0 {
		t = 0.0
		s = clamp(-c/a, 0.0, 1.0)
	} else if t > 1.0 {
		t = 1.0
		s = clamp((b-c)/a, 0.0, 1.0)
	}

	fp1Closest = p1.add(D1.mulScalar(s))
	fp2Closest = p2.add(D2.mulScalar(t))

	return fp1Closest, fp2Closest
}
