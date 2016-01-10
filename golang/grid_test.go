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
