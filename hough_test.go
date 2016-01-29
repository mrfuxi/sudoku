package main

import (
	"testing"

	"github.com/gonum/matrix/mat64"
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

func TestHoughLines(t *testing.T) {
	timg := mat64.NewDense(500, 700, nil)
	timg.Set(10, 10, 1.0)
	timg.Set(10, 200, 1.0)
	timg.Set(10, 400, 1.0)
	timg.Set(200, 10, 1.0)
	timg.Set(400, 10, 1.0)

	lines := HoughLines(timg, nil, 0, 10)
	if !assert.Len(t, lines, 6) {
		t.FailNow()
	}

	expectedLines := []Line{
		Line{Theta: 1.570796, Distance: 10, Count: 3},
		Line{Theta: 0.000000, Distance: 10, Count: 3},
		Line{Theta: 0.785398, Distance: 148, Count: 2},
		Line{Theta: 0.453786, Distance: 184, Count: 2},
		Line{Theta: 1.117011, Distance: 184, Count: 2},
		Line{Theta: 0.785398, Distance: 290, Count: 2},
	}
	for i, line := range lines {
		expected := expectedLines[i]
		thetaOk := assert.InDelta(t, expected.Theta, line.Theta, 0.0001)
		distanceOk := assert.Equal(t, expected.Distance, line.Distance)
		countOk := assert.Equal(t, expected.Count, line.Count)
		if !thetaOk || !distanceOk || !countOk {
			t.Fatalf("%v expected to equal %v", line, expected)
		}
	}
}
