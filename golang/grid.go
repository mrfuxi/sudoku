package main

import (
	"math"
	"sort"
)

type Grid struct {
	Horizontal []Line
	Vertical   []Line
	Score      float64
}

type GridByScore []Grid

func (a GridByScore) Len() int           { return len(a) }
func (a GridByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a GridByScore) Less(i, j int) bool { return a[i].Score > a[j].Score } // Reversed order most to least

type meanAcc struct {
	values []float64
}

func (m *meanAcc) Add(value float64) {
	m.values = append(m.values, value)
}

func (m *meanAcc) Mean() float64 {
	res := 0.0
	count := float64(len(m.values))

	for _, v := range m.values {
		res += v
	}

	return res / count
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func possibleGrids(horizontal, vertical []Line) []Grid {
	if !sort.IsSorted(ByCount(vertical)) {
		panic("should be sorted")
	}
	if !sort.IsSorted(ByCount(horizontal)) {
		panic("should be sorted")
	}

	linesV := make(map[string][]Line, 0)
	scoresV := make(map[string]*meanAcc, 0)
	for _, h := range horizontal {
		scores, lineGroups := linearDistances(vertical, h)
		for i, lines := range lineGroups {
			hash := LineHash(lines).HashKey()
			if scoresV[hash] == nil {
				scoresV[hash] = new(meanAcc)
			}
			scoresV[hash].Add(scores[i])
			linesV[hash] = lines
		}
	}

	linesH := make(map[string][]Line, 0)
	scoresH := make(map[string]*meanAcc, 0)
	for _, v := range vertical {
		scores, lineGroups := linearDistances(horizontal, v)
		for i, lines := range lineGroups {
			hash := LineHash(lines).HashKey()
			if scoresH[hash] == nil {
				scoresH[hash] = new(meanAcc)
			}
			scoresH[hash].Add(scores[i])
			linesH[hash] = lines
		}
	}

	grids := make([]Grid, 0)
	for hHash, h := range linesH {
		for vHash, v := range linesV {
			grid := Grid{
				Horizontal: h,
				Vertical:   v,
				Score:      scoresH[hHash].Mean() * scoresV[vHash].Mean(),
			}
			grids = append(grids, grid)
		}
	}

	sort.Sort(GridByScore(grids))

	return grids[:minInt(9, len(grids))]
}

// Splits lines into groups of 10 with score of how much linearly distributed they are
func linearDistances(lines []Line, dividerLine Line) (scores []float64, matches [][]Line) {
	scores = make([]float64, 0)
	matches = make([][]Line, 0)
	linesCount := len(lines)
	if linesCount < 10 {
		return
	}

	intersections := make([]Point, linesCount, linesCount)
	for i, line := range lines {
		_, point := intersection(line, dividerLine)
		intersections[i] = point
	}

	points := make([]float64, len(lines), len(lines))
	for i, point := range intersections {
		points[i] = intersections[0].DistanceTo(point)
	}

	distances := preparePointDistances(points)

	expectedPoints := make([]float64, 10, 10)

	for i := range points[:linesCount-10+1] {
		dI := i + 10 - 1
		for j := range points[dI:] {
			start, end := points[i], points[j+dI]
			step := (end - start) / 9.0
			for k := range expectedPoints {
				expectedPoints[k] = start + step*float64(k)
			}
			score, selectedPoints := pointSimilarities(expectedPoints, distances)

			if len(selectedPoints) != 10 {
				continue
			}

			selectedLines := make([]Line, 10, 10)
			searchablePoints := sort.Float64Slice(points)
			for l := range selectedLines {
				selectedLines[l] = lines[searchablePoints.Search(selectedPoints[l])]
			}

			scores = append(scores, score)
			matches = append(matches, selectedLines)
		}
	}
	return
}

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
