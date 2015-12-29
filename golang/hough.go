package main

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/gonum/matrix/mat64"
)

type Line struct {
	Theta    float64
	Distance int
	Count    uint64
}

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

func HoughLines(src *mat64.Dense, thetas []float64, threshold uint64) []Line {
	if thetas == nil {
		thetas = GenerateThetas(-math.Pi/2.0, math.Pi/2, math.Pi/180.0)
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

	lines := make([]Line, 0)
	for i := range hAcc {
		r := i - int(offset)
		for j, count := range hAcc[i] {
			if count < threshold {
				continue
			}
			line := Line{
				Theta:    thetas[j],
				Distance: r,
				Count:    count,
			}
			lines = append(lines, line)
		}
	}

	return lines
}
