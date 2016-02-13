package sudoku

import (
	"fmt"
	"image"
	"math"

	"github.com/gonum/matrix/mat64"
)

const pi2 = 2 * math.Pi

type angleBucket struct {
	Start float64
	End   float64
}

func (b angleBucket) String() string {
	return fmt.Sprintf("Bucket{%.0f, %.0f}", b.Start*180/math.Pi, b.End*180/math.Pi)
}

type lineFragment struct {
	Start image.Point
	End   image.Point
}

func (f lineFragment) Length() float64 {
	return distanceBetweenPoints(f.Start, f.End)
}

func distanceBetweenPoints(a, b image.Point) float64 {
	return math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y))
}

func similarAngles(a, b float64) bool {
	minAngDiff := 0.5 // ~28deg

	if a >= pi2 || a <= -pi2 {
		a -= math.Floor(a/pi2) * pi2
	}
	if b >= pi2 || b <= -pi2 {
		b -= math.Floor(b/pi2) * pi2
	}

	angDiff := math.Abs(a - b)
	if angDiff < minAngDiff || angDiff > (pi2-minAngDiff) {
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
func intersection(lineA, lineB polarLine) (bool, image.Point) {
	A := mat64.NewDense(2, 2, []float64{
		math.Cos(lineA.Theta), math.Sin(lineA.Theta),
		math.Cos(lineB.Theta), math.Sin(lineB.Theta),
	})
	b := mat64.NewDense(2, 1, []float64{
		float64(lineA.Distance), float64(lineB.Distance),
	})
	x := mat64.NewDense(2, 1, nil)
	err := x.Solve(A, b)

	ok := err == nil
	// Using 0.5 to force round to nearest int rather than Floor
	point := image.Point{
		X: int(x.At(0, 0) + 0.5),
		Y: int(x.At(1, 0) + 0.5),
	}
	return ok, point
}

// duplicates: crosses in view at low angle
func removeDuplicateLines(lines []polarLine, width, height int) []polarLine {
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

			ok, point := intersection(lineA, lineB)
			if !ok {
				continue
			}

			in_view := (minX <= point.X && point.X <= maxX &&
				minY <= point.Y && point.Y <= maxY)

			if in_view {
				toRemove[k] = true
				continue
			}
		}
	}

	deDuped := make([]polarLine, len(lines)-len(toRemove))
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

// generateAngleBuckets creates a map with ortogonal (if required) ranges for angles (in radians)
//     Both bucketSize and step are taken in deg (it's easier to reason about)
//     Angles between 0 and 180 deg
//
//     Example output (bucket_size=20, step=5) - all values in deg:
//     {
//         45: [(35, 55), (125, 145)],
//         50: [(40, 60), (130, 150)],
//     }
func generateAngleBuckets(bucketSize uint, step uint, ortogonal bool) map[float64][]angleBucket {
	const DegToRad float64 = math.Pi / 180

	window := DegToRad * float64(bucketSize)
	stepSize := DegToRad * float64(step)

	window2 := window / 2.0
	maxPos := math.Pi
	if ortogonal {
		maxPos = math.Pi / 2
	}
	maxPos = maxPos - stepSize
	buckets := make(map[float64][]angleBucket, 0)

	pos := 0.0
	for {
		b1 := angleBucket{pos - window2, pos + window2}
		bucket := []angleBucket{b1}

		if b1.Start < 0 {
			b1Prim := angleBucket{math.Pi + b1.Start, math.Pi}
			bucket = append(bucket, b1Prim)
		}

		if b1.End > math.Pi {
			b1Bis := angleBucket{0, b1.End - math.Pi}
			bucket = append(bucket, b1Bis)
		}

		if ortogonal {
			b2 := angleBucket{b1.Start + math.Pi/2, b1.End + math.Pi/2}
			bucket = append(bucket, b2)

			if b2.End > math.Pi {
				b2Prim := angleBucket{0, b2.End - math.Pi}
				bucket = append(bucket, b2Prim)
			}
		}

		buckets[pos] = bucket

		pos += stepSize
		if pos >= maxPos {
			break
		}
	}

	return buckets
}

// Splits lines into two groups one that are similar to given angle,
// and the rest of lines
func linesWithSimilarAngle(lines []polarLine, angle float64) ([]polarLine, []polarLine) {
	similar := make([]polarLine, 0)
	other := make([]polarLine, 0)

	for _, line := range lines {
		if similarAngles(line.Theta, angle) {
			similar = append(similar, line)
		} else {
			other = append(other, line)
		}
	}

	return similar, other
}

func putLinesIntoBuckets(buckets map[float64][]angleBucket, lines []polarLine) map[float64][]polarLine {
	bucketed := make(map[float64][]polarLine, 0)
	alreadyMatched := make(map[string]bool, 0)

	for angle, bucket := range buckets {
		matches := make([]polarLine, 0)

		for _, line := range lines {
			for _, b := range bucket {
				if b.Start <= line.Theta && line.Theta <= b.End {
					matches = append(matches, line)
					break
				}
			}
		}

		matchesKey := polarLineHash(matches).HashKey()
		if len(matches) > 0 && !alreadyMatched[matchesKey] {
			alreadyMatched[matchesKey] = true
			bucketed[angle] = matches
		}
	}
	return bucketed
}

// Bresenham's line algorithm
func pointsOnLineFragment(fragment lineFragment) []image.Point {
	x0, x1 := fragment.Start.X, fragment.End.X
	y0, y1 := fragment.Start.Y, fragment.End.Y

	dx := float64(x1 - x0)
	sx := 1
	if dx < 0 {
		sx = -1
		dx = -dx
	}

	dy := float64(y1 - y0)
	sy := 1
	if dy < 0 {
		sy = -1
		dy = -dy
	}

	points := make([]image.Point, 0)

	var err float64
	x, y := x0, y0
	if dx > dy {
		err = dx / 2.0
		for {
			points = append(points, image.Point{x, y})
			if x == x1 {
				break
			}

			err -= dy
			if err < 0 {
				y += sy
				err += dx
			}
			x += sx
		}
	} else {
		err = dy / 2.0
		for {
			points = append(points, image.Point{x, y})
			if y == y1 {
				break
			}

			err -= dx
			if err < 0 {
				x += sx
				err += dy
			}
			y += sy
		}
	}

	return points
}
