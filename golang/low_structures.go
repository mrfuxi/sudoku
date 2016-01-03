package main

import (
	"fmt"
	"math"

	"github.com/gonum/matrix/mat64"
)

const pi2 = 2 * math.Pi

type Bucket struct {
	Start float64
	End   float64
}

func (b Bucket) String() string {
	return fmt.Sprintf("Bucket{%.0f, %.0f}", b.Start*180/math.Pi, b.End*180/math.Pi)
}

func similarAngles(a, b float64) bool {
	minAngDiff := 0.5 // ~28deg

	if a > pi2 || a < -pi2 {
		a -= math.Floor(a/pi2) * pi2
	}
	if b > pi2 || b < -pi2 {
		b -= math.Floor(b/pi2) * pi2
	}

	angDiff := math.Abs(a - b)
	if angDiff < minAngDiff || angDiff > (pi2-minAngDiff) {
		return true
	}
	return false
}

// Intersection calculates intersection of two lines
//
// Solve:
// x*cos(thetaA) + y*sin(thetaA) = rA
// x*cos(thetaB) + y*sin(thetaB) = rB
//
// As matrix:
// A*X = b
func intersection(lineA, lineB Line) (bool, int, int) {
	A := mat64.NewDense(2, 2, []float64{
		math.Cos(lineA.Theta), math.Sin(lineA.Theta),
		math.Cos(lineB.Theta), math.Sin(lineB.Theta),
	})
	b := mat64.NewDense(2, 1, []float64{
		float64(lineA.Distance), float64(lineB.Distance),
	})
	x := mat64.NewDense(2, 1, nil)
	err := x.Solve(A, b)

	ok := err == nil
	// Using 0.5 to force round to nearest int rather than Floor
	return ok, int(x.At(0, 0) + 0.5), int(x.At(1, 0) + 0.5)
}

// duplicates: crosses in view at low angle
func removeDuplicateLines(lines []Line, width, height int) []Line {
	minDist := 3.0

	scope := 2
	minX := 0 - width/scope
	minY := 0 - height/scope
	maxX := width + width/scope
	maxY := height + height/scope

	toRemove := make(map[int]bool, len(lines))
	for i, lineA := range lines {
		for j, lineB := range lines[i+1:] {
			k := j + i + 1

			similar := similarAngles(lineA.Theta, lineB.Theta)
			if !similar {
				continue
			}

			if math.Abs(float64(lineA.Distance-lineB.Distance)) < minDist {
				toRemove[k] = true
				continue
			}

			ok, x, y := intersection(lineA, lineB)
			if !ok {
				continue
			}

			in_view := (minX <= x && x <= maxX &&
				minY <= y && y <= maxY)

			if in_view {
				toRemove[k] = true
				continue
			}
		}
	}

	deDuped := make([]Line, len(lines)-len(toRemove))
	j := 0
	for i, line := range lines {
		if toRemove[i] {
			continue
		}
		deDuped[j] = line
		j++
	}
	return deDuped
}

// generateAngleBuckets creates a map with ortogonal (if required) ranges for angles (in radians)
//     Both bucketSize and step are taken in deg (it's easier to reason about)
//     Angles between 0 and 180 deg
//
//     Example output (bucket_size=20, step=5) - all values in deg:
//     {
//         45: [(35, 55), (125, 145)],
//         50: [(40, 60), (130, 150)],
//     }
func generateAngleBuckets(bucketSize uint, step uint, ortogonal bool) map[float64][]Bucket {
	const DegToRad float64 = math.Pi / 180

	window := DegToRad * float64(bucketSize)
	stepSize := DegToRad * float64(step)

	window_2 := window / 2.0
	pos := 0.0
	div := 1.0
	if ortogonal {
		div = 2.0
	}
	max_pos := math.Pi/div - stepSize
	buckets := make(map[float64][]Bucket, 0)

	for {
		b1 := Bucket{pos - window_2, pos + window_2}
		bucket := []Bucket{b1}

		if b1.Start < 0 {
			b1_prim := Bucket{math.Pi + b1.Start, math.Pi}
			bucket = append(bucket, b1_prim)
		}

		if b1.End > math.Pi {
			b1_bis := Bucket{0, b1.End - math.Pi}
			bucket = append(bucket, b1_bis)
		}

		if ortogonal {
			b2 := Bucket{b1.Start + math.Pi/2, b1.End + math.Pi/2}
			bucket = append(bucket, b2)

			if b2.End > math.Pi {
				b2_prim := Bucket{0, b2.End - math.Pi}
				bucket = append(bucket, b2_prim)
			}
		}

		buckets[pos] = bucket

		pos += stepSize
		if pos >= max_pos {
			break
		}
	}

	return buckets
}
