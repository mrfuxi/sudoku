package main

import (
	"fmt"
	"math"

	"github.com/gonum/matrix/mat64"
)

func similarAngles(a, b float64) bool {
	minAngDiff := 0.5 // ~28deg

	angDiff := math.Abs(a - b)
	if angDiff < minAngDiff || angDiff > (math.Pi-minAngDiff) {
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
	x.Solve(A, b)
	fmt.Println(x.At(0, 0), x.At(1, 0))

	return false, 0, 0
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
