package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreparePointDistances(t *testing.T) {
	var examples = []struct {
		positions []float64
		expected  []float64
	}{
		{
			[]float64{0, 5.5},
			[]float64{0, 0, 0, 5.5, 5.5, 5.5},
		},
		{
			[]float64{0, 6},
			[]float64{0, 0, 0, 0, 6, 6, 6},
		},
		{
			[]float64{0, 5},
			[]float64{0, 0, 0, 5, 5, 5},
		},
		{
			[]float64{2, 5},
			[]float64{2, 2, 2, 2, 5, 5},
		},
		{
			[]float64{2, 5, 6, 10},
			[]float64{2, 2, 2, 2, 5, 5, 6, 6, 6, 10, 10},
		},
	}

	for _, tt := range examples {
		distances := preparePointDistances(tt.positions)
		assert.EqualValues(t, tt.expected, distances)
	}
}

func TestPointSimilarities(t *testing.T) {
	var examples = []struct {
		closestPoints   []float64
		idealPoints     []float64
		expectedMatches []float64
		fit             float64
	}{
		{
			closestPoints:   preparePointDistances([]float64{2.5, 12, 21.5, 32.5}),
			idealPoints:     []float64{2, 12, 22},
			expectedMatches: []float64{2.5, 12, 21.5},
			fit:             0.9888,
		},
		{
			closestPoints:   preparePointDistances([]float64{2.5, 12, 21.5, 32.5}),
			idealPoints:     []float64{12, 22, 32},
			expectedMatches: []float64{12, 21.5, 32.5},
			fit:             0.9833,
		},
		{
			closestPoints:   preparePointDistances([]float64{0, 5, 10, 15, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 140}),
			idealPoints:     []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			expectedMatches: []float64{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			fit:             1.0,
		},
	}

	for _, tt := range examples {
		fit, matchedPoints := pointSimilarities(tt.idealPoints, tt.closestPoints)
		assert.InDelta(t, tt.fit, fit, 0.0001)
		assert.EqualValues(t, tt.expectedMatches, matchedPoints)
	}
}
