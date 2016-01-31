package sudoku

import (
	"errors"
	"sync"

	"github.com/gonum/matrix/mat64"
)

type thresholdType int

const (
	threshBinary thresholdType = iota
	threshBinaryInv
)

var (
	errBlockSize = errors.New("Incorrect block size. Has to be odd number greater than 1.")
)

func viewValues(start, ksize, total int) (newStart int, newNum int) {
	newNum = ksize
	newStart = start
	if start < 0 {
		newNum += start
		newStart = 0
	} else if start+ksize > total {
		newNum = total - start
	}
	return
}

func meanMat(row, rows, col, cols, delta, ksize int, src *mat64.Dense) (m float64) {
	startRow, rowNum := viewValues(row-delta, ksize, rows)
	startCol, colNum := viewValues(col-delta, ksize, cols)
	sub := src.View(startRow, startCol, rowNum, colNum)
	sum := mat64.Sum(sub)
	return sum / float64(rowNum*colNum)
}

func meanFilter(src *mat64.Dense, ksize int) (dst *mat64.Dense) {
	var wg sync.WaitGroup
	rows, cols := src.Dims()
	dst = mat64.NewDense(rows, cols, nil)
	delta := (ksize - 1) / 2
	for col := 0; col < cols; col++ {
		wg.Add(1)
		go func(col int) {
			for row := 0; row < rows; row++ {
				m := meanMat(row, rows, col, cols, delta, ksize, src)
				dst.Set(row, col, m)
			}
			wg.Done()
		}(col)
	}
	wg.Wait()
	return
}

func adaptiveThreshold(src *mat64.Dense, maxValue float64, threshold thresholdType, blockSize int, delta float64) (dst *mat64.Dense) {
	if blockSize%2 != 1 || blockSize < 1 {
		panic(errBlockSize)
	}

	if maxValue < 0 {
		return src
	}

	dst = meanFilter(src, blockSize)
	rows, cols := src.Dims()
	for col := 0; col < cols; col++ {
		for row := 0; row < rows; row++ {
			newVal := 0.0
			dstVal := dst.At(row, col) - delta
			srcVal := src.At(row, col)
			if threshold == threshBinary {
				if srcVal > dstVal {
					newVal = maxValue
				}
			} else if threshold == threshBinaryInv {
				if srcVal < dstVal {
					newVal = maxValue
				}
			}
			dst.Set(row, col, newVal)
		}
	}
	return dst
}
