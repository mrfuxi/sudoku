package sudoku

import (
	"math"
	"sort"

	"github.com/gonum/matrix/mat64"
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

type ScoredLines struct {
	Lines []Line
	Score float64
}

type ScoredLinesByScore []ScoredLines

func (a ScoredLinesByScore) Len() int           { return len(a) }
func (a ScoredLinesByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ScoredLinesByScore) Less(i, j int) bool { return a[i].Score > a[j].Score } // Reversed order most to least

func (s *ScoredLines) HashKey() string {
	return LineHash(s.Lines).HashKey()
}

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

// Builds possible line grouppings by using muiltple "cutting" lines
func buildScoredLines(primary, secondary []Line, top uint) []ScoredLines {
	lines := make(map[string]ScoredLines, 0)
	scores := make(map[string]*meanAcc, 0)
	for _, s := range secondary {
		matches := linearDistances(primary, s)
		for _, match := range matches {
			hash := match.HashKey()
			if scores[hash] == nil {
				scores[hash] = new(meanAcc)
				lines[hash] = match
			}
			scores[hash].Add(match.Score)
		}
	}

	scoredLines := make([]ScoredLines, len(lines), len(lines))
	s := 0
	for hash, scoredLine := range lines {
		scoredLine.Score = scores[hash].Mean()
		scoredLines[s] = scoredLine
		s++
	}

	sort.Sort(ScoredLinesByScore(scoredLines))

	return scoredLines[:minInt(int(top), len(scoredLines))]
}

func possibleGrids(horizontal, vertical []Line) []Grid {
	// Make sure lines are ordered correctly
	sort.Sort(ByDistance(vertical))
	sort.Sort(ByDistance(horizontal))

	linesH := buildScoredLines(horizontal, vertical, 3)
	linesV := buildScoredLines(vertical, horizontal, 3)

	grids := make([]Grid, 0)
	for _, h := range linesH {
		for _, v := range linesV {
			grid := Grid{
				Horizontal: h.Lines,
				Vertical:   v.Lines,
				Score:      h.Score * v.Score,
			}
			grids = append(grids, grid)
		}
	}

	sort.Sort(GridByScore(grids))

	return grids
}

func evaluateGrids(image *mat64.Dense, grids []Grid) []Grid {
	for _, grid := range grids {
		hCount := len(grid.Horizontal)
		vCount := len(grid.Vertical)
		fragments := make([]Fragment, hCount+vCount)

		firstVertLine := grid.Vertical[0]
		lastVertLine := grid.Vertical[vCount-1]
		for j, h := range grid.Horizontal {
			_, start := intersection(h, firstVertLine)
			_, end := intersection(h, lastVertLine)
			fragments[j] = Fragment{start, end}
		}

		firstHorizLine := grid.Horizontal[0]
		lastHorizLine := grid.Horizontal[hCount-1]
		for j, h := range grid.Vertical {
			_, start := intersection(h, firstHorizLine)
			_, end := intersection(h, lastHorizLine)
			fragments[hCount+j] = Fragment{start, end}
		}

		score := 0.0
		for _, fragment := range fragments {
			points := PointsOnLineFragment(fragment)
			value := 1.0 / fragment.Length()
			for _, point := range points {
				if image.At(point.Y, point.X) != 0 {
					score += value
				}
			}
		}
		grid.Score = grid.Score * score / float64(len(fragments))
	}

	sort.Sort(GridByScore(grids))

	return grids
}

// Splits lines into groups of 10 with score of how much linearly distributed they are
func linearDistances(lines []Line, dividerLine Line) []ScoredLines {
	// Lines have to be sortd correctly!
	matches := make([]ScoredLines, 0)

	linesCount := len(lines)
	if linesCount < 10 {
		return matches
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

			match := ScoredLines{
				Score: score,
				Lines: make([]Line, 10, 10),
			}

			searchablePoints := sort.Float64Slice(points)
			for l := range match.Lines {
				match.Lines[l] = lines[searchablePoints.Search(selectedPoints[l])]
			}

			matches = append(matches, match)
		}
	}

	return matches
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
