package main

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimilarAngles(t *testing.T) {
	var examples = []struct {
		a       float64
		b       float64
		similar bool
	}{
		{0.0, 0.0, true},
		{0.0, 0.49, true},
		{0.0, 0.50, false},
		{0.0, -0.49, true},
		{0.0, -0.50, false},
		{0.0, 2 * math.Pi, true},
		{0.0, 2*math.Pi + 0.49, true},
		{0.0, 2*math.Pi + 0.50, false},
		{0.0, 2*math.Pi - 0.49, true},
		{0.0, 2*math.Pi - 0.50, false},
		{0.0, 3 * math.Pi, false},
		{0.0, 4 * math.Pi, true},
		{0.0, 4*math.Pi + 0.49, true},
		{0.0, 4*math.Pi + 0.50, false},
		{0.0, 4*math.Pi - 0.49, true},
		{0.0, 4*math.Pi - 0.50, false},
		{1.0, 1.0, true},
		{1.0, 1.49, true},
		{1.0, 1.50, false},
		{1.0, 1 - 0.49, true},
		{1.0, 1 - 0.50, false},
		{2 * math.Pi, 4 * math.Pi, true},
		{math.Pi, math.Pi, true},
		{math.Pi / 2, math.Pi, false},
	}

	for _, tt := range examples {
		isSimilar := similarAngles(tt.a, tt.b)
		assert.Equal(t, tt.similar, isSimilar, "Angles: %f and %f. Exect %v got %v instead", tt.a, tt.b, tt.similar, isSimilar)
	}
}

func TestIntersections(t *testing.T) {
	var examples = []struct {
		a  Line
		b  Line
		ok bool
		x  int
		y  int
	}{
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 1.570796, Distance: 10}, true, 10, 10},
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 0.785398, Distance: 148}, true, 10, 199},
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 0.453786, Distance: 184}, true, 10, 399},
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 1.117011, Distance: 184}, true, 10, 200},
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 0.785398, Distance: 290}, true, 10, 400},
		{Line{Theta: 0.785398, Distance: 148}, Line{Theta: 1.117011, Distance: 184}, true, 9, 200},
		{Line{Theta: 0.785398, Distance: 148}, Line{Theta: 0.785399, Distance: 290}, true, -100409041, 100409284}, // lines are almost parallel
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 0.000000, Distance: 20}, false, 0, 0},                   // no solution, lines are parallel
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 0.000000, Distance: 10}, false, 0, 0},                   // no solution, lines are parallel
		{Line{Theta: 0.785398, Distance: 148}, Line{Theta: 0.785398, Distance: 290}, false, 0, 0},                 // no solution, lines are parallel
	}
	for _, tt := range examples {
		ok, x, y := intersection(tt.a, tt.b)
		format := "Intersection between %v and %v"
		assert.Equal(t, tt.ok, ok, format, tt.a, tt.b)
		assert.Equal(t, tt.x, x, format, tt.a, tt.b)
		assert.Equal(t, tt.y, y, format, tt.a, tt.b)
	}
}

func TestRemoveDuplicateLines(t *testing.T) {
	var examples = []struct {
		pre    []Line
		post   []Line
		width  int
		height int
	}{
		{
			pre:   []Line{Line{0.000000, 10, 0}, Line{1.570796, 10, 0}},
			post:  []Line{Line{0.000000, 10, 0}, Line{1.570796, 10, 0}},
			width: 300, height: 300,
		}, // Angle too different
		{
			pre:   []Line{Line{1.570796, 100, 0}, Line{1.50000, 102, 0}, Line{1.50000, 98, 0}},
			post:  []Line{Line{1.570796, 100, 0}},
			width: 300, height: 300,
		}, // Similar angle, close to each other (middle one)
		{
			pre:   []Line{Line{1.570796, 100, 0}, Line{1.605703, 104, 0}},
			post:  []Line{Line{1.570796, 100, 0}},
			width: 300, height: 300,
		}, // Similar angles, crossing somewhere in view (-115, 100) vs [(-150, 450), (-150, 450)]
		{
			pre:   []Line{Line{1.570796, 100, 0}, Line{1.605703, 104, 0}},
			post:  []Line{Line{1.570796, 100, 0}, Line{1.605703, 104, 0}},
			width: 200, height: 200,
		}, // Similar angles, crossing outside in view (-115, 100) vs [(-100, 300), (-100, 300)]
	}

	for _, tt := range examples {
		post := removeDuplicateLines(tt.pre, tt.width, tt.height)
		assert.Len(t, post, len(tt.post))
	}
}
