package main

import (
	"errors"
	"sync"

	"github.com/gonum/matrix/mat64"
)

type ThresholdType int

const (
	ThreshBinary ThresholdType = iota
	ThreshBinaryInv
)

var (
	ErrBlockSize = errors.New("Incorrect block size. Has to be odd number greater than 1.")
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

func boxFilter(src *mat64.Dense, ksize int) (dst *mat64.Dense) {
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

func AdaptiveThreshold(src *mat64.Dense, maxValue float64, thresholdType ThresholdType, blockSize int, delta float64) (dst *mat64.Dense, err error) {
	if blockSize%2 != 1 || blockSize < 1 {
		return dst, ErrBlockSize
	}

	if maxValue < 0 {
		return src, nil
	}

	dst = boxFilter(src, blockSize)
	rows, cols := src.Dims()
	for col := 0; col < cols; col++ {
		for row := 0; row < rows; row++ {
			newVal := 0.0
			dstVal := dst.At(row, col)
			srcVal := src.At(row, col)
			if dstVal > srcVal {
				newVal = maxValue
			}
			dst.Set(row, col, newVal)
		}
	}
	return dst, nil
}
