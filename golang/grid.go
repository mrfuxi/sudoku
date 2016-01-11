package main

import "math"

func preparePointDistances(positions []float64) []float64 {
	maxPos := positions[len(positions)-1]
	ld := int(maxPos) + 1

	closest := make([]float64, ld, ld)

	posI := 0
	for c := range closest {
		d1 := math.Abs(positions[posI] - float64(c))
		d2 := float64(ld)
		if posI+1 < len(positions) {
			d2 = math.Abs(positions[posI+1] - float64(c))
		}

		if d1 <= d2 {
			closest[c] = positions[posI]
		} else {
			closest[c] = positions[posI+1]
			posI += 1
		}
	}

	return closest
}

func pointSimilarities(expectedPoints, distances []float64) (float64, []float64) {
	fit := 0.0
	matches := make([]float64, 0)

	step := expectedPoints[1] - expectedPoints[0]
	for _, expected := range expectedPoints {
		point := distances[int(expected)]
		if len(matches) > 0 {
			f := math.Abs(math.Abs(point-matches[len(matches)-1])-step) / step
			if f >= 0.2 {
				break
			}
			fit += f / 9.0
		}

		matches = append(matches, point)
	}

	return (1 - fit), matches
}
