package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateThetas(t *testing.T) {
	var examples = []struct {
		start  float64
		end    float64
		step   float64
		thetas []float64
	}{
		{0, 1, 0.1, []float64{0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0}},
		{0, 1, 0.3, []float64{0, 0.3, 0.6, 0.9}},
		{0, 1, 1, []float64{0, 1}},
		{-1, 1, 1, []float64{-1, 0, 1}},
		{1, 0, -0.5, []float64{1, 0.5, 0}},
		{1, 2, 0.3, []float64{1, 1.3, 1.6, 1.9}},
	}

	for _, tt := range examples {
		thetas := GenerateThetas(tt.start, tt.end, tt.step)
		if !assert.InDeltaSlice(t, tt.thetas, thetas, 0.01) {
			t.Logf("For theta: %v, %v, %v", tt.start, tt.end, tt.step)
		}
	}
}
