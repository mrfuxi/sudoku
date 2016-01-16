package main

import (
	"bytes"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/gonum/matrix/mat64"
)

type Line struct {
	Theta    float64
	Distance int
	Count    uint64
}

func (l Line) String() string {
	return fmt.Sprintf("Line{Theta: %f, Distance: %d, Count: %d}", l.Theta, l.Distance, l.Count)
}

func (l Line) HashKey() string {
	return fmt.Sprintf("%0.8f:%d", l.Theta, l.Distance)
}

type LineHash []Line

func (l LineHash) HashKey() string {
	var buffer bytes.Buffer
	for _, line := range l {
		buffer.WriteString(line.String())
	}
	return buffer.String()
}

type ByCount []Line

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].Count > a[j].Count } // Reversed order most to least

func GenerateThetas(start, end, step float64) (thetas []float64) {
	count := int((end-start)/step) + 1
	thetas = make([]float64, count, count)
	theta := start
	for i := range thetas {
		thetas[i] = theta
		theta += step
	}
	return thetas
}

func HoughLines(src *mat64.Dense, thetas []float64, threshold uint64, limit int) []Line {
	if thetas == nil {
		thetas = GenerateThetas(-math.Pi/2, math.Pi/2, math.Pi/180.0)
	}
	rows, cols := src.Dims()
	maxR := 2 * math.Hypot(float64(cols), float64(rows))
	offset := maxR / 2

	hAcc := make([][]uint64, int(maxR), int(maxR))
	for i := range hAcc {
		hAcc[i] = make([]uint64, len(thetas), len(thetas))
	}

	sin := make([]float64, len(thetas), len(thetas))
	cos := make([]float64, len(thetas), len(thetas))
	for i, th := range thetas {
		sin[i] = math.Sin(th)
		cos[i] = math.Cos(th)
	}

	var wg sync.WaitGroup
	for col := 0; col < cols; col++ {
		wg.Add(1)
		go func(col int) {
			for row := 0; row < rows; row++ {
				val := src.At(row, col)
				if val == 0 {
					continue
				}

				for i := range thetas {
					r := float64(col)*cos[i] + float64(row)*sin[i]
					iry := int(r + offset)
					atomic.AddUint64(&hAcc[iry][i], 1)
				}
			}
			wg.Done()
		}(col)
	}
	wg.Wait()

	linesSet := make(map[string]bool)
	lines := make([]Line, 0)
	for i := range hAcc {
		r := i - int(offset)
		thetaOffset := 0.0
		if r < 0 {
			thetaOffset = math.Pi
			r *= -1
		}
		for j, count := range hAcc[i] {
			if count < 2 || count < threshold {
				continue
			}

			line := Line{
				Theta:    thetas[j] + thetaOffset,
				Distance: r,
				Count:    count,
			}
			if !linesSet[line.HashKey()] {
				linesSet[line.HashKey()] = true
				lines = append(lines, line)
			}
		}
	}

	sort.Sort(ByCount(lines))

	if limit > 0 && len(lines) > limit {
		lines = lines[:limit]
	}

	return lines
}
