package sudoku

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"sort"
	"sync"
	"sync/atomic"
)

type polarLine struct {
	Theta    float64
	Distance int
	Count    uint64
}

func (l polarLine) String() string {
	return fmt.Sprintf("Line{Theta: %f, Distance: %d, Count: %d}", l.Theta, l.Distance, l.Count)
}

func (l polarLine) HashKey() string {
	return fmt.Sprintf("%0.8f:%d", l.Theta, l.Distance)
}

type polarLineHash []polarLine

func (l polarLineHash) HashKey() string {
	var buffer bytes.Buffer
	for _, line := range l {
		buffer.WriteString(line.String())
	}
	return buffer.String()
}

type polarLinesByCount []polarLine

func (a polarLinesByCount) Len() int           { return len(a) }
func (a polarLinesByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a polarLinesByCount) Less(i, j int) bool { return a[i].Count > a[j].Count } // Reversed order most to least

type polarLinesByDistance []polarLine

func (a polarLinesByDistance) Len() int           { return len(a) }
func (a polarLinesByDistance) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a polarLinesByDistance) Less(i, j int) bool { return a[i].Distance < a[j].Distance }

func generateThetas(start, end, step float64) (thetas []float64) {
	count := int((end-start)/step) + 1
	thetas = make([]float64, count, count)
	theta := start
	for i := range thetas {
		thetas[i] = theta
		theta += step
	}
	return thetas
}

func houghLines(src image.Gray, thetas []float64, threshold uint64, limit int) []polarLine {
	if thetas == nil {
		thetas = generateThetas(-math.Pi/2, math.Pi/2, math.Pi/180.0)
	}
	maxY, maxX := src.Bounds().Max.Y, src.Bounds().Max.X
	maxR := 2 * math.Hypot(float64(maxX), float64(maxY))
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
	for y := 0; y < maxY; y++ {
		wg.Add(1)
		go func(y int) {
			for x := 0; x < maxX; x++ {
				val := src.Pix[src.PixOffset(x, y)]
				if val == 0 {
					continue
				}

				for i := range thetas {
					r := float64(x)*cos[i] + float64(y)*sin[i]
					iry := int(r + offset)
					atomic.AddUint64(&hAcc[iry][i], 1)
				}
			}
			wg.Done()
		}(y)
	}
	wg.Wait()

	linesSet := make(map[string]bool)
	var lines []polarLine
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

			line := polarLine{
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

	sort.Sort(polarLinesByCount(lines))

	if limit > 0 && len(lines) > limit {
		lines = lines[:limit]
	}

	return lines
}
