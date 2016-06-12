package sudoku

import (
	"fmt"
	"image"
	"math"
	"sort"

	"github.com/mrfuxi/sudoku/digits"
)

type lineGrid struct {
	Horizontal []polarLine
	Vertical   []polarLine
	Score      float64
}

type lineGridByScore []lineGrid

func (a lineGridByScore) Len() int           { return len(a) }
func (a lineGridByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a lineGridByScore) Less(i, j int) bool { return a[i].Score > a[j].Score } // Reversed order most to least

type scoredLines struct {
	Lines []polarLine
	Score float64
}

type scoredLinesByScore []scoredLines

func (a scoredLinesByScore) Len() int           { return len(a) }
func (a scoredLinesByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a scoredLinesByScore) Less(i, j int) bool { return a[i].Score > a[j].Score } // Reversed order most to least

func (s *scoredLines) HashKey() string {
	return polarLineHash(s.Lines).HashKey()
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
func buildScoredLines(primary, secondary []polarLine, top uint) []scoredLines {
	lines := make(map[string]scoredLines, 0)
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

	scoredLn := make([]scoredLines, len(lines), len(lines))
	s := 0
	for hash, scoredLine := range lines {
		scoredLine.Score = scores[hash].Mean()
		scoredLn[s] = scoredLine
		s++
	}

	sort.Sort(scoredLinesByScore(scoredLn))

	return scoredLn[:minInt(int(top), len(scoredLn))]
}

func possibleGrids(horizontal, vertical []polarLine) []lineGrid {
	// Make sure lines are ordered correctly
	sort.Sort(polarLinesByDistance(vertical))
	sort.Sort(polarLinesByDistance(horizontal))

	linesH := buildScoredLines(horizontal, vertical, 3)
	linesV := buildScoredLines(vertical, horizontal, 3)

	var grids []lineGrid
	for _, h := range linesH {
		for _, v := range linesV {
			grid := lineGrid{
				Horizontal: h.Lines,
				Vertical:   v.Lines,
				Score:      h.Score * v.Score,
			}
			grids = append(grids, grid)
		}
	}

	sort.Sort(lineGridByScore(grids))

	return grids
}

func evaluateGrids(src image.Gray, grids []lineGrid) []lineGrid {
	for _, grid := range grids {
		hCount := len(grid.Horizontal)
		vCount := len(grid.Vertical)
		fragments := make([]lineFragment, hCount+vCount)

		firstVertLine := grid.Vertical[0]
		lastVertLine := grid.Vertical[vCount-1]
		for j, h := range grid.Horizontal {
			_, start := intersection(h, firstVertLine)
			_, end := intersection(h, lastVertLine)
			fragments[j] = lineFragment{start, end}
		}

		firstHorizLine := grid.Horizontal[0]
		lastHorizLine := grid.Horizontal[hCount-1]
		for j, h := range grid.Vertical {
			_, start := intersection(h, firstHorizLine)
			_, end := intersection(h, lastHorizLine)
			fragments[hCount+j] = lineFragment{start, end}
		}

		score := 0.0
		for _, fragment := range fragments {
			points := pointsOnLineFragment(fragment)
			value := 1.0 / fragment.Length()
			for _, point := range points {
				if src.Pix[src.PixOffset(point.X, point.Y)] != 0 {
					score += value
				}
			}
		}
		grid.Score = grid.Score * score / float64(len(fragments))
	}

	sort.Sort(lineGridByScore(grids))

	return grids
}

// Splits lines into groups of 10 with score of how much linearly distributed they are
func linearDistances(lines []polarLine, dividerLine polarLine) []scoredLines {
	// Lines have to be sorted correctly!
	var matches []scoredLines

	linesCount := len(lines)
	if linesCount < 10 {
		return matches
	}

	intersections := make([]image.Point, linesCount, linesCount)
	for i, line := range lines {
		_, point := intersection(line, dividerLine)
		intersections[i] = point
	}

	points := make([]float64, len(lines), len(lines))
	for i, point := range intersections {
		points[i] = distanceBetweenPoints(intersections[0], point)
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

			match := scoredLines{
				Score: score,
				Lines: make([]polarLine, 10, 10),
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
			posI++
		}
	}

	return closest
}

func pointSimilarities(expectedPoints, distances []float64) (float64, []float64) {
	fit := 0.0
	var matches []float64

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

func extractCells(grid lineGrid, img image.Image) (cells [9][9]image.Gray) {
	grayImg := grayImage(img)

	margin := 0
	size := 28.0 // Size of learning data set: MNIST
	dst := [4]pointF{
		pointF{0, 0},
		pointF{size, 0},
		pointF{size, size},
		pointF{0, size},
	}

	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			_, p1 := intersection(grid.Horizontal[row], grid.Vertical[col])
			_, p2 := intersection(grid.Horizontal[row], grid.Vertical[col+1])
			_, p3 := intersection(grid.Horizontal[row+1], grid.Vertical[col+1])
			_, p4 := intersection(grid.Horizontal[row+1], grid.Vertical[col])

			p1.Y -= margin
			p1.X -= margin
			p2.Y -= margin
			p2.X += margin
			p3.X += margin
			p3.Y += margin
			p4.X -= margin
			p4.Y += margin

			src := [4]pointF{
				newPointF(p1),
				newPointF(p2),
				newPointF(p3),
				newPointF(p4),
			}
			proj := newPerspective(src, dst)
			cells[row][col] = proj.warpPerspective(grayImg)

			cell := cells[row][col]
			digit, conf := digits.RecogniseDigit(cell, otsuValue(cell))
			fn := fmt.Sprintf("%v_%v-%v-%.2f.png", row, col, digit, conf)
			saveImage(&cells[row][col], fn)
		}
	}
	return
}
