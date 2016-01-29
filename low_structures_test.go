package main

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

const thetaDelta = 0.00001

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
		{2 * math.Pi, 0.0, true},
		{2 * math.Pi, 2 * math.Pi, true},
		{2 * math.Pi, 0.49, true},
		{2 * math.Pi, 0.50, false},
		{2 * math.Pi, -0.49, true},
		{2 * math.Pi, -0.50, false},
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
		a        Line
		b        Line
		ok       bool
		solution Point
	}{
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 1.570796, Distance: 10}, true, Point{10, 10}},
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 0.785398, Distance: 148}, true, Point{10, 199}},
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 0.453786, Distance: 184}, true, Point{10, 399}},
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 1.117011, Distance: 184}, true, Point{10, 200}},
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 0.785398, Distance: 290}, true, Point{10, 400}},
		{Line{Theta: 0.785398, Distance: 148}, Line{Theta: 1.117011, Distance: 184}, true, Point{9, 200}},
		{Line{Theta: 0.785398, Distance: 148}, Line{Theta: 0.785399, Distance: 290}, true, Point{-100409041, 100409284}}, // lines are almost parallel
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 0.000000, Distance: 20}, false, Point{0, 0}},                   // no solution, lines are parallel
		{Line{Theta: 0.000000, Distance: 10}, Line{Theta: 0.000000, Distance: 10}, false, Point{0, 0}},                   // no solution, lines are parallel
		{Line{Theta: 0.785398, Distance: 148}, Line{Theta: 0.785398, Distance: 290}, false, Point{0, 0}},                 // no solution, lines are parallel
	}
	for _, tt := range examples {
		ok, point := intersection(tt.a, tt.b)
		format := "Intersection between %v and %v"
		assert.Equal(t, tt.ok, ok, format, tt.a, tt.b)
		assert.Equal(t, tt.solution.X, point.X, format, tt.a, tt.b)
		assert.Equal(t, tt.solution.Y, point.Y, format, tt.a, tt.b)
	}
}

func TestPointDistince(t *testing.T) {
	var examples = []struct {
		a        Point
		b        Point
		distance float64
	}{
		{Point{0, 0}, Point{1, 1}, math.Sqrt(2)},
		{Point{1, 1}, Point{2, 2}, math.Sqrt(2)},
		{Point{0, 0}, Point{1, 2}, math.Sqrt(5)},
		{Point{-1, -1}, Point{1, 1}, math.Sqrt(8)},
	}

	for _, tt := range examples {
		distanceA := tt.a.DistanceTo(tt.b)
		distanceB := tt.b.DistanceTo(tt.a)
		assert.Equal(t, distanceA, distanceB)
		assert.Equal(t, tt.distance, distanceA)
		assert.Equal(t, tt.distance, distanceB)
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

func TestGenerateAngleBuckets(t *testing.T) {
	var examples = []struct {
		bucketSize uint
		step       uint
		ortogonal  bool
		expected   map[int][]Bucket
	}{
		{
			60, 30, false,
			map[int][]Bucket{
				0:   {Bucket{-30, 30}, Bucket{150, 180}},
				30:  {Bucket{0, 60}},
				60:  {Bucket{30, 90}},
				90:  {Bucket{60, 120}},
				120: {Bucket{90, 150}},
				150: {Bucket{120, 180}},
			},
		},
		{
			60, 30, true,
			map[int][]Bucket{
				0:  {Bucket{-30, 30}, Bucket{150, 180}, Bucket{60, 120}},
				30: {Bucket{0, 60}, Bucket{90, 150}},
				60: {Bucket{30, 90}, Bucket{120, 180}},
			},
		},
		{
			20, 5, true,
			map[int][]Bucket{
				0:  {Bucket{-10, 10}, Bucket{170, 180}, Bucket{80, 100}},
				5:  {Bucket{-5, 15}, Bucket{175, 180}, Bucket{85, 105}},
				10: {Bucket{0, 20}, Bucket{90, 110}},
				15: {Bucket{5, 25}, Bucket{95, 115}},
				20: {Bucket{10, 30}, Bucket{100, 120}},
				25: {Bucket{15, 35}, Bucket{105, 125}},
				30: {Bucket{20, 40}, Bucket{110, 130}},
				35: {Bucket{25, 45}, Bucket{115, 135}},
				40: {Bucket{30, 50}, Bucket{120, 140}},
				45: {Bucket{35, 55}, Bucket{125, 145}},
				50: {Bucket{40, 60}, Bucket{130, 150}},
				55: {Bucket{45, 65}, Bucket{135, 155}},
				60: {Bucket{50, 70}, Bucket{140, 160}},
				65: {Bucket{55, 75}, Bucket{145, 165}},
				70: {Bucket{60, 80}, Bucket{150, 170}},
				75: {Bucket{65, 85}, Bucket{155, 175}},
				80: {Bucket{70, 90}, Bucket{160, 180}},
				85: {Bucket{75, 95}, Bucket{165, 185}, Bucket{0, 5}},
			},
		},
		{
			20, 5, false,
			map[int][]Bucket{
				0:   {Bucket{-10, 10}, Bucket{170, 180}},
				5:   {Bucket{-5, 15}, Bucket{175, 180}},
				10:  {Bucket{0, 20}},
				15:  {Bucket{5, 25}},
				20:  {Bucket{10, 30}},
				25:  {Bucket{15, 35}},
				30:  {Bucket{20, 40}},
				35:  {Bucket{25, 45}},
				40:  {Bucket{30, 50}},
				45:  {Bucket{35, 55}},
				50:  {Bucket{40, 60}},
				55:  {Bucket{45, 65}},
				60:  {Bucket{50, 70}},
				65:  {Bucket{55, 75}},
				70:  {Bucket{60, 80}},
				75:  {Bucket{65, 85}},
				80:  {Bucket{70, 90}},
				85:  {Bucket{75, 95}},
				90:  {Bucket{80, 100}},
				95:  {Bucket{85, 105}},
				100: {Bucket{90, 110}},
				105: {Bucket{95, 115}},
				110: {Bucket{100, 120}},
				115: {Bucket{105, 125}},
				120: {Bucket{110, 130}},
				125: {Bucket{115, 135}},
				130: {Bucket{120, 140}},
				135: {Bucket{125, 145}},
				140: {Bucket{130, 150}},
				145: {Bucket{135, 155}},
				150: {Bucket{140, 160}},
				155: {Bucket{145, 165}},
				160: {Bucket{150, 170}},
				165: {Bucket{155, 175}},
				170: {Bucket{160, 180}},
				175: {Bucket{165, 185}, Bucket{0, 5}},
			},
		},
	}

	for _, tt := range examples {
		buckets := generateAngleBuckets(tt.bucketSize, tt.step, tt.ortogonal)
		if !assert.Len(t, buckets, len(tt.expected)) {
			t.FailNow()
		}
		for k, v := range buckets {
			kDeg := int(k*180/math.Pi + 0.5)
			assert.Contains(t, tt.expected, kDeg)
			if !assert.Len(t, v, len(tt.expected[kDeg])) {
				t.FailNow()
			}
			for i, bucket := range v {
				assert.InDelta(t, tt.expected[kDeg][i].Start, bucket.Start*180/math.Pi, thetaDelta)
				assert.InDelta(t, tt.expected[kDeg][i].End, bucket.End*180/math.Pi, thetaDelta)
			}
		}
	}
}

func TestLinesWithSimilarAngle(t *testing.T) {
	examples := []struct {
		angle   float64
		lines   []Line
		similar []Line
		other   []Line
	}{
		{
			angle:   0,
			lines:   []Line{Line{Theta: 0, Distance: 1}, Line{Theta: 0, Distance: 1000}, Line{Theta: 0.49}, Line{Theta: 0.5}, Line{Theta: -0.49}, Line{Theta: -0.5}},
			similar: []Line{Line{Theta: 0, Distance: 1}, Line{Theta: 0, Distance: 1000}, Line{Theta: 0.49}, Line{Theta: -0.49}},
			other:   []Line{Line{Theta: 0.5}, Line{Theta: -0.5}},
		},
		{
			angle:   2 * math.Pi,
			lines:   []Line{Line{Theta: 0, Distance: 1}, Line{Theta: 0, Distance: 1000}, Line{Theta: 0.49}, Line{Theta: 0.5}, Line{Theta: -0.49}, Line{Theta: -0.5}},
			similar: []Line{Line{Theta: 0, Distance: 1}, Line{Theta: 0, Distance: 1000}, Line{Theta: 0.49}, Line{Theta: -0.49}},
			other:   []Line{Line{Theta: 0.5}, Line{Theta: -0.5}},
		},
		{
			angle:   math.Pi,
			lines:   []Line{Line{Theta: math.Pi + 0, Distance: 1}, Line{Theta: math.Pi + 0, Distance: 1000}, Line{Theta: math.Pi + 0.49}, Line{Theta: math.Pi + 0.5}, Line{Theta: math.Pi - 0.49}, Line{Theta: math.Pi - 0.5}},
			similar: []Line{Line{Theta: math.Pi + 0, Distance: 1}, Line{Theta: math.Pi + 0, Distance: 1000}, Line{Theta: math.Pi + 0.49}, Line{Theta: math.Pi - 0.49}},
			other:   []Line{Line{Theta: math.Pi + 0.5}, Line{Theta: math.Pi - 0.5}},
		},
		{
			angle:   math.Pi,
			lines:   []Line{Line{Theta: 0, Distance: 1}, Line{Theta: 0, Distance: 1000}, Line{Theta: 0.49}, Line{Theta: 0.5}, Line{Theta: -0.49}, Line{Theta: -0.5}},
			similar: []Line{},
			other:   []Line{Line{Theta: 0, Distance: 1}, Line{Theta: 0, Distance: 1000}, Line{Theta: 0.49}, Line{Theta: 0.5}, Line{Theta: -0.49}, Line{Theta: -0.5}},
		},
		{
			angle:   0,
			lines:   []Line{Line{Theta: 0, Distance: 1}, Line{Theta: 0, Distance: 1000}, Line{Theta: 0.49}, Line{Theta: -0.49}},
			similar: []Line{Line{Theta: 0, Distance: 1}, Line{Theta: 0, Distance: 1000}, Line{Theta: 0.49}, Line{Theta: -0.49}},
			other:   []Line{},
		},
	}

	for _, tt := range examples {
		similar, other := linesWithSimilarAngle(tt.lines, tt.angle)
		if !assert.Len(t, similar, len(tt.similar)) {
			t.Logf("Got:\n%v\nexpecting:\n%v", similar, tt.similar)
			t.FailNow()
		}
		for i, line := range similar {
			assert.InDelta(t, tt.similar[i].Theta, line.Theta, thetaDelta)
			assert.Equal(t, tt.similar[i].Distance, line.Distance)
		}

		if !assert.Len(t, other, len(tt.other)) {
			t.Logf("Got:\n%v\nexpecting:\n%v", similar, tt.similar)
			t.FailNow()
		}
		for i, line := range other {
			assert.InDelta(t, tt.other[i].Theta, line.Theta, thetaDelta)
			assert.Equal(t, tt.other[i].Distance, line.Distance)
		}
	}
}

func TestPutLinesIntoBuckets(t *testing.T) {
	buckets := map[float64][]Bucket{
		0.0: {Bucket{-0.1, 0.1}, Bucket{math.Pi - 0.1, math.Pi + 0.1}},
		1.0: {Bucket{0.9, 1.1}},
	}

	lines := []Line{
		Line{Theta: 0},
		Line{Theta: -0.1},
		Line{Theta: 0.1},
		Line{Theta: math.Pi},
		Line{Theta: 1.1},
		Line{Theta: 100},
		Line{Theta: -0.11},
		Line{Theta: 0.11},
	}

	expected := map[float64][]Line{
		0.0: {Line{Theta: 0}, Line{Theta: -0.1}, Line{Theta: 0.1}, Line{Theta: math.Pi}},
		1.0: {Line{Theta: 1.1}},
	}

	bucketed := putLinesIntoBuckets(buckets, lines)
	assert.Len(t, bucketed, len(expected))
	for angle, expected_lines := range expected {
		if !assert.Len(t, bucketed[angle], len(expected_lines)) {
			t.FailNow()
		}

		for i, line := range bucketed[angle] {
			assert.EqualValues(t, expected_lines[i], line)
		}
	}
}

func TestPutLinesIntoBucketsDontReuseLinesIfBucketsHaveTheSame(t *testing.T) {
	buckets := map[float64][]Bucket{
		1.0: {Bucket{0, 2}},
		2.0: {Bucket{1, 3}},
	}

	lines := []Line{
		Line{Theta: 1},
		Line{Theta: 1.1},
	}

	bucketed := putLinesIntoBuckets(buckets, lines)
	assert.Len(t, bucketed, 1) // it will be either 1.0 or 2.0, not both

	angle := 1.0
	if len(bucketed[angle]) == 0 {
		angle = 2.0
	}

	for i, line := range bucketed[angle] {
		assert.EqualValues(t, lines[i], line)
	}
}

func TestPutLinesIntoBucketsReuseLineIfBucketsHaveSlightlyDifferent(t *testing.T) {
	buckets := map[float64][]Bucket{
		1.0: {Bucket{0, 2}},
		2.0: {Bucket{1, 3}},
	}

	lines := []Line{
		Line{Theta: 1},
		Line{Theta: 1.1},
		Line{Theta: 2.1},
	}

	expected := map[float64][]Line{
		1.0: {Line{Theta: 1}, Line{Theta: 1.1}},
		2.0: {Line{Theta: 1}, Line{Theta: 1.1}, Line{Theta: 2.1}},
	}

	bucketed := putLinesIntoBuckets(buckets, lines)
	assert.Len(t, bucketed, len(expected))
	for angle, expected_lines := range expected {
		if !assert.Len(t, bucketed[angle], len(expected_lines)) {
			t.FailNow()
		}

		for i, line := range bucketed[angle] {
			assert.EqualValues(t, expected_lines[i], line)
		}
	}
}

func TestPointsOnFragment(t *testing.T) {
	examples := []struct {
		fragment Fragment
		points   []Point
	}{
		{
			Fragment{Start: Point{0, 0}, End: Point{5, 5}},
			[]Point{{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},
		},
		{
			Fragment{Start: Point{0, 1}, End: Point{6, 4}},
			[]Point{{0, 1}, {1, 1}, {2, 2}, {3, 2}, {4, 3}, {5, 3}, {6, 4}},
		},
	}
	for _, tt := range examples {
		points := PointsOnLineFragment(tt.fragment)
		assert.EqualValues(t, tt.points, points)
	}
}
