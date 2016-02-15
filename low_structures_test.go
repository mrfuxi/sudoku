package sudoku

import (
	"image"
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
		a        polarLine
		b        polarLine
		ok       bool
		solution image.Point
	}{
		{polarLine{Theta: 0.000000, Distance: 10}, polarLine{Theta: 1.570796, Distance: 10}, true, image.Point{10, 10}},
		{polarLine{Theta: 0.000000, Distance: 10}, polarLine{Theta: 0.785398, Distance: 148}, true, image.Point{10, 199}},
		{polarLine{Theta: 0.000000, Distance: 10}, polarLine{Theta: 0.453786, Distance: 184}, true, image.Point{10, 399}},
		{polarLine{Theta: 0.000000, Distance: 10}, polarLine{Theta: 1.117011, Distance: 184}, true, image.Point{10, 200}},
		{polarLine{Theta: 0.000000, Distance: 10}, polarLine{Theta: 0.785398, Distance: 290}, true, image.Point{10, 400}},
		{polarLine{Theta: 0.785398, Distance: 148}, polarLine{Theta: 1.117011, Distance: 184}, true, image.Point{9, 200}},
		{polarLine{Theta: 0.785398, Distance: 148}, polarLine{Theta: 0.785399, Distance: 290}, true, image.Point{-100409041, 100409284}}, // lines are almost parallel
		{polarLine{Theta: 0.000000, Distance: 10}, polarLine{Theta: 0.000000, Distance: 20}, false, image.Point{0, 0}},                   // no solution, lines are parallel
		{polarLine{Theta: 0.000000, Distance: 10}, polarLine{Theta: 0.000000, Distance: 10}, false, image.Point{0, 0}},                   // no solution, lines are parallel
		{polarLine{Theta: 0.785398, Distance: 148}, polarLine{Theta: 0.785398, Distance: 290}, false, image.Point{0, 0}},                 // no solution, lines are parallel
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
		a        image.Point
		b        image.Point
		distance float64
	}{
		{image.Point{0, 0}, image.Point{1, 1}, math.Sqrt(2)},
		{image.Point{1, 1}, image.Point{2, 2}, math.Sqrt(2)},
		{image.Point{0, 0}, image.Point{1, 2}, math.Sqrt(5)},
		{image.Point{-1, -1}, image.Point{1, 1}, math.Sqrt(8)},
	}

	for _, tt := range examples {
		distanceA := distanceBetweenPoints(tt.a, tt.b)
		distanceB := distanceBetweenPoints(tt.b, tt.a)
		assert.Equal(t, distanceA, distanceB)
		assert.Equal(t, tt.distance, distanceA)
		assert.Equal(t, tt.distance, distanceB)
	}
}

func TestRemoveDuplicateLines(t *testing.T) {
	var examples = []struct {
		pre    []polarLine
		post   []polarLine
		width  int
		height int
	}{
		{
			pre:   []polarLine{polarLine{0.000000, 10, 0}, polarLine{1.570796, 10, 0}},
			post:  []polarLine{polarLine{0.000000, 10, 0}, polarLine{1.570796, 10, 0}},
			width: 300, height: 300,
		}, // Angle too different
		{
			pre:   []polarLine{polarLine{1.570796, 100, 0}, polarLine{1.50000, 102, 0}, polarLine{1.50000, 98, 0}},
			post:  []polarLine{polarLine{1.570796, 100, 0}},
			width: 300, height: 300,
		}, // Similar angle, close to each other (middle one)
		{
			pre:   []polarLine{polarLine{1.570796, 100, 0}, polarLine{1.605703, 104, 0}},
			post:  []polarLine{polarLine{1.570796, 100, 0}},
			width: 300, height: 300,
		}, // Similar angles, crossing somewhere in view (-115, 100) vs [(-150, 450), (-150, 450)]
		{
			pre:   []polarLine{polarLine{1.570796, 100, 0}, polarLine{1.605703, 104, 0}},
			post:  []polarLine{polarLine{1.570796, 100, 0}, polarLine{1.605703, 104, 0}},
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
		expected   map[int][]angleBucket
	}{
		{
			60, 30, false,
			map[int][]angleBucket{
				0:   {angleBucket{-30, 30}, angleBucket{150, 180}},
				30:  {angleBucket{0, 60}},
				60:  {angleBucket{30, 90}},
				90:  {angleBucket{60, 120}},
				120: {angleBucket{90, 150}},
				150: {angleBucket{120, 180}},
			},
		},
		{
			60, 30, true,
			map[int][]angleBucket{
				0:  {angleBucket{-30, 30}, angleBucket{150, 180}, angleBucket{60, 120}},
				30: {angleBucket{0, 60}, angleBucket{90, 150}},
				60: {angleBucket{30, 90}, angleBucket{120, 180}},
			},
		},
		{
			20, 5, true,
			map[int][]angleBucket{
				0:  {angleBucket{-10, 10}, angleBucket{170, 180}, angleBucket{80, 100}},
				5:  {angleBucket{-5, 15}, angleBucket{175, 180}, angleBucket{85, 105}},
				10: {angleBucket{0, 20}, angleBucket{90, 110}},
				15: {angleBucket{5, 25}, angleBucket{95, 115}},
				20: {angleBucket{10, 30}, angleBucket{100, 120}},
				25: {angleBucket{15, 35}, angleBucket{105, 125}},
				30: {angleBucket{20, 40}, angleBucket{110, 130}},
				35: {angleBucket{25, 45}, angleBucket{115, 135}},
				40: {angleBucket{30, 50}, angleBucket{120, 140}},
				45: {angleBucket{35, 55}, angleBucket{125, 145}},
				50: {angleBucket{40, 60}, angleBucket{130, 150}},
				55: {angleBucket{45, 65}, angleBucket{135, 155}},
				60: {angleBucket{50, 70}, angleBucket{140, 160}},
				65: {angleBucket{55, 75}, angleBucket{145, 165}},
				70: {angleBucket{60, 80}, angleBucket{150, 170}},
				75: {angleBucket{65, 85}, angleBucket{155, 175}},
				80: {angleBucket{70, 90}, angleBucket{160, 180}},
				85: {angleBucket{75, 95}, angleBucket{165, 185}, angleBucket{0, 5}},
			},
		},
		{
			20, 5, false,
			map[int][]angleBucket{
				0:   {angleBucket{-10, 10}, angleBucket{170, 180}},
				5:   {angleBucket{-5, 15}, angleBucket{175, 180}},
				10:  {angleBucket{0, 20}},
				15:  {angleBucket{5, 25}},
				20:  {angleBucket{10, 30}},
				25:  {angleBucket{15, 35}},
				30:  {angleBucket{20, 40}},
				35:  {angleBucket{25, 45}},
				40:  {angleBucket{30, 50}},
				45:  {angleBucket{35, 55}},
				50:  {angleBucket{40, 60}},
				55:  {angleBucket{45, 65}},
				60:  {angleBucket{50, 70}},
				65:  {angleBucket{55, 75}},
				70:  {angleBucket{60, 80}},
				75:  {angleBucket{65, 85}},
				80:  {angleBucket{70, 90}},
				85:  {angleBucket{75, 95}},
				90:  {angleBucket{80, 100}},
				95:  {angleBucket{85, 105}},
				100: {angleBucket{90, 110}},
				105: {angleBucket{95, 115}},
				110: {angleBucket{100, 120}},
				115: {angleBucket{105, 125}},
				120: {angleBucket{110, 130}},
				125: {angleBucket{115, 135}},
				130: {angleBucket{120, 140}},
				135: {angleBucket{125, 145}},
				140: {angleBucket{130, 150}},
				145: {angleBucket{135, 155}},
				150: {angleBucket{140, 160}},
				155: {angleBucket{145, 165}},
				160: {angleBucket{150, 170}},
				165: {angleBucket{155, 175}},
				170: {angleBucket{160, 180}},
				175: {angleBucket{165, 185}, angleBucket{0, 5}},
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
		lines   []polarLine
		similar []polarLine
		other   []polarLine
	}{
		{
			angle:   0,
			lines:   []polarLine{polarLine{Theta: 0, Distance: 1}, polarLine{Theta: 0, Distance: 1000}, polarLine{Theta: 0.49}, polarLine{Theta: 0.5}, polarLine{Theta: -0.49}, polarLine{Theta: -0.5}},
			similar: []polarLine{polarLine{Theta: 0, Distance: 1}, polarLine{Theta: 0, Distance: 1000}, polarLine{Theta: 0.49}, polarLine{Theta: -0.49}},
			other:   []polarLine{polarLine{Theta: 0.5}, polarLine{Theta: -0.5}},
		},
		{
			angle:   2 * math.Pi,
			lines:   []polarLine{polarLine{Theta: 0, Distance: 1}, polarLine{Theta: 0, Distance: 1000}, polarLine{Theta: 0.49}, polarLine{Theta: 0.5}, polarLine{Theta: -0.49}, polarLine{Theta: -0.5}},
			similar: []polarLine{polarLine{Theta: 0, Distance: 1}, polarLine{Theta: 0, Distance: 1000}, polarLine{Theta: 0.49}, polarLine{Theta: -0.49}},
			other:   []polarLine{polarLine{Theta: 0.5}, polarLine{Theta: -0.5}},
		},
		{
			angle:   math.Pi,
			lines:   []polarLine{polarLine{Theta: math.Pi + 0, Distance: 1}, polarLine{Theta: math.Pi + 0, Distance: 1000}, polarLine{Theta: math.Pi + 0.49}, polarLine{Theta: math.Pi + 0.5}, polarLine{Theta: math.Pi - 0.49}, polarLine{Theta: math.Pi - 0.5}},
			similar: []polarLine{polarLine{Theta: math.Pi + 0, Distance: 1}, polarLine{Theta: math.Pi + 0, Distance: 1000}, polarLine{Theta: math.Pi + 0.49}, polarLine{Theta: math.Pi - 0.49}},
			other:   []polarLine{polarLine{Theta: math.Pi + 0.5}, polarLine{Theta: math.Pi - 0.5}},
		},
		{
			angle:   math.Pi,
			lines:   []polarLine{polarLine{Theta: 0, Distance: 1}, polarLine{Theta: 0, Distance: 1000}, polarLine{Theta: 0.49}, polarLine{Theta: 0.5}, polarLine{Theta: -0.49}, polarLine{Theta: -0.5}},
			similar: []polarLine{},
			other:   []polarLine{polarLine{Theta: 0, Distance: 1}, polarLine{Theta: 0, Distance: 1000}, polarLine{Theta: 0.49}, polarLine{Theta: 0.5}, polarLine{Theta: -0.49}, polarLine{Theta: -0.5}},
		},
		{
			angle:   0,
			lines:   []polarLine{polarLine{Theta: 0, Distance: 1}, polarLine{Theta: 0, Distance: 1000}, polarLine{Theta: 0.49}, polarLine{Theta: -0.49}},
			similar: []polarLine{polarLine{Theta: 0, Distance: 1}, polarLine{Theta: 0, Distance: 1000}, polarLine{Theta: 0.49}, polarLine{Theta: -0.49}},
			other:   []polarLine{},
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
	buckets := map[float64][]angleBucket{
		0.0: {angleBucket{-0.1, 0.1}, angleBucket{math.Pi - 0.1, math.Pi + 0.1}},
		1.0: {angleBucket{0.9, 1.1}},
	}

	lines := []polarLine{
		polarLine{Theta: 0},
		polarLine{Theta: -0.1},
		polarLine{Theta: 0.1},
		polarLine{Theta: math.Pi},
		polarLine{Theta: 1.1},
		polarLine{Theta: 100},
		polarLine{Theta: -0.11},
		polarLine{Theta: 0.11},
	}

	expected := map[float64][]polarLine{
		0.0: {polarLine{Theta: 0}, polarLine{Theta: -0.1}, polarLine{Theta: 0.1}, polarLine{Theta: math.Pi}},
		1.0: {polarLine{Theta: 1.1}},
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
	buckets := map[float64][]angleBucket{
		1.0: {angleBucket{0, 2}},
		2.0: {angleBucket{1, 3}},
	}

	lines := []polarLine{
		polarLine{Theta: 1},
		polarLine{Theta: 1.1},
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
	buckets := map[float64][]angleBucket{
		1.0: {angleBucket{0, 2}},
		2.0: {angleBucket{1, 3}},
	}

	lines := []polarLine{
		polarLine{Theta: 1},
		polarLine{Theta: 1.1},
		polarLine{Theta: 2.1},
	}

	expected := map[float64][]polarLine{
		1.0: {polarLine{Theta: 1}, polarLine{Theta: 1.1}},
		2.0: {polarLine{Theta: 1}, polarLine{Theta: 1.1}, polarLine{Theta: 2.1}},
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
		fragment lineFragment
		points   []image.Point
	}{
		{
			lineFragment{Start: image.Point{0, 0}, End: image.Point{5, 5}},
			[]image.Point{{0, 0}, {1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},
		},
		{
			lineFragment{Start: image.Point{0, 1}, End: image.Point{6, 4}},
			[]image.Point{{0, 1}, {1, 1}, {2, 2}, {3, 2}, {4, 3}, {5, 3}, {6, 4}},
		},
	}
	for _, tt := range examples {
		points := pointsOnLineFragment(tt.fragment)
		assert.EqualValues(t, tt.points, points)
	}
}
