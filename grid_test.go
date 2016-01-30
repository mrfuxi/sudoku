package sudoku

import (
	"math"
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

func TestLinearDistances(t *testing.T) {
	lines := []Line{
		Line{Theta: 0, Distance: -10}, // odd
		Line{Theta: 0, Distance: 10},
		Line{Theta: 0, Distance: 15}, // odd
		Line{Theta: 0, Distance: 20},
		Line{Theta: 0, Distance: 30},
		Line{Theta: 0, Distance: 40},
		Line{Theta: 0, Distance: 50},
		Line{Theta: 0, Distance: 53}, // odd
		Line{Theta: 0, Distance: 55}, // odd
		Line{Theta: 0, Distance: 60},
		Line{Theta: 0, Distance: 70},
		Line{Theta: 0, Distance: 80},
		Line{Theta: 0, Distance: 90},
		Line{Theta: 0, Distance: 101}, // slightly off
		Line{Theta: 0, Distance: 111}, // slightly off
		Line{Theta: 0, Distance: 120},
		Line{Theta: 0, Distance: 130},
	}
	dividerLine := Line{Theta: math.Pi / 2, Distance: 0}

	expectedScoredLines := []ScoredLines{
		{
			Lines: []Line{
				Line{Theta: 0, Distance: 10},
				Line{Theta: 0, Distance: 20},
				Line{Theta: 0, Distance: 30},
				Line{Theta: 0, Distance: 40},
				Line{Theta: 0, Distance: 50},
				Line{Theta: 0, Distance: 60},
				Line{Theta: 0, Distance: 70},
				Line{Theta: 0, Distance: 80},
				Line{Theta: 0, Distance: 90},
				Line{Theta: 0, Distance: 101},
			},
			Score: 0.9804,
		},
		{
			Lines: []Line{
				Line{Theta: 0, Distance: 20},
				Line{Theta: 0, Distance: 30},
				Line{Theta: 0, Distance: 40},
				Line{Theta: 0, Distance: 50},
				Line{Theta: 0, Distance: 60},
				Line{Theta: 0, Distance: 70},
				Line{Theta: 0, Distance: 80},
				Line{Theta: 0, Distance: 90},
				Line{Theta: 0, Distance: 101},
				Line{Theta: 0, Distance: 111},
			},
			Score: 0.9804,
		},
		{
			Lines: []Line{
				Line{Theta: 0, Distance: 30},
				Line{Theta: 0, Distance: 40},
				Line{Theta: 0, Distance: 50},
				Line{Theta: 0, Distance: 60},
				Line{Theta: 0, Distance: 70},
				Line{Theta: 0, Distance: 80},
				Line{Theta: 0, Distance: 90},
				Line{Theta: 0, Distance: 101},
				Line{Theta: 0, Distance: 111},
				Line{Theta: 0, Distance: 120},
			},
			Score: 0.9777,
		},
		{
			Lines: []Line{
				Line{Theta: 0, Distance: 40},
				Line{Theta: 0, Distance: 50},
				Line{Theta: 0, Distance: 60},
				Line{Theta: 0, Distance: 70},
				Line{Theta: 0, Distance: 80},
				Line{Theta: 0, Distance: 90},
				Line{Theta: 0, Distance: 101},
				Line{Theta: 0, Distance: 111},
				Line{Theta: 0, Distance: 120},
				Line{Theta: 0, Distance: 130},
			},
			Score: 0.9777,
		},
	}

	matches := linearDistances(lines, dividerLine)
	assert.Len(t, matches, len(expectedScoredLines))

	for i, match := range matches {
		assert.Len(t, match.Lines, 10)
		assert.InDelta(t, expectedScoredLines[i].Score, match.Score, 0.0001)
		assert.EqualValues(t, expectedScoredLines[i].Lines, match.Lines)
	}
}

func TestPossibleGrids(t *testing.T) {
	linesH := []Line{
		Line{Theta: 0, Distance: -10}, // odd
		Line{Theta: 0, Distance: 10},
		Line{Theta: 0, Distance: 15}, // odd
		Line{Theta: 0, Distance: 20},
		Line{Theta: 0, Distance: 30},
		Line{Theta: 0, Distance: 40},
		Line{Theta: 0, Distance: 50},
		Line{Theta: 0, Distance: 53}, // odd
		Line{Theta: 0, Distance: 55}, // odd
		Line{Theta: 0, Distance: 60},
		Line{Theta: 0, Distance: 70},
		Line{Theta: 0, Distance: 80},
		Line{Theta: 0, Distance: 90},
		Line{Theta: 0, Distance: 101}, // slightly off
		Line{Theta: 0, Distance: 110},
		Line{Theta: 0, Distance: 120},
		Line{Theta: 0, Distance: 130},
	}
	linesV := []Line{
		Line{Theta: math.Pi / 2, Distance: -10}, // odd
		Line{Theta: math.Pi / 2, Distance: 10},
		Line{Theta: math.Pi / 2, Distance: 15}, // odd
		Line{Theta: math.Pi / 2, Distance: 20},
		Line{Theta: math.Pi / 2, Distance: 30},
		Line{Theta: math.Pi / 2, Distance: 40},
		Line{Theta: math.Pi / 2, Distance: 50},
		Line{Theta: math.Pi / 2, Distance: 53}, // odd
		Line{Theta: math.Pi / 2, Distance: 55}, // odd
		Line{Theta: math.Pi / 2, Distance: 60},
		Line{Theta: math.Pi / 2, Distance: 70},
		Line{Theta: math.Pi / 2, Distance: 80},
		Line{Theta: math.Pi / 2, Distance: 90},
		Line{Theta: math.Pi / 2, Distance: 101}, // slightly off
		Line{Theta: math.Pi / 2, Distance: 111}, // slightly off
		Line{Theta: math.Pi / 2, Distance: 120},
		Line{Theta: math.Pi / 2, Distance: 130},
	}

	firstExpectedGrid := Grid{
		Horizontal: []Line{
			Line{Theta: 0, Distance: 10},
			Line{Theta: 0, Distance: 20},
			Line{Theta: 0, Distance: 30},
			Line{Theta: 0, Distance: 40},
			Line{Theta: 0, Distance: 50},
			Line{Theta: 0, Distance: 60},
			Line{Theta: 0, Distance: 70},
			Line{Theta: 0, Distance: 80},
			Line{Theta: 0, Distance: 90},
			Line{Theta: 0, Distance: 101},
		},
		Vertical: []Line{
			Line{Theta: math.Pi / 2, Distance: 10},
			Line{Theta: math.Pi / 2, Distance: 20},
			Line{Theta: math.Pi / 2, Distance: 30},
			Line{Theta: math.Pi / 2, Distance: 40},
			Line{Theta: math.Pi / 2, Distance: 50},
			Line{Theta: math.Pi / 2, Distance: 60},
			Line{Theta: math.Pi / 2, Distance: 70},
			Line{Theta: math.Pi / 2, Distance: 80},
			Line{Theta: math.Pi / 2, Distance: 90},
			Line{Theta: math.Pi / 2, Distance: 101},
		},
		Score: 0.98046 * 0.98046,
	}

	grids := possibleGrids(linesH, linesV)
	assert.Len(t, grids, 9)
	assert.EqualValues(t, grids[0].Horizontal, firstExpectedGrid.Horizontal)
	assert.EqualValues(t, grids[0].Vertical, firstExpectedGrid.Vertical)
	assert.InDelta(t, grids[0].Score, firstExpectedGrid.Score, 0.0001)
}
