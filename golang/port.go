package main

import (
	"errors"
	"image"
)

const (
	ThreshBinaryInv = iota
)

var (
	ErrBlockSize = errors.New("Incorrect block size. Has to be odd number greater than 1.")
)

func boxFilter(src image.Gray, ksize int) (dst image.Gray) {
	// bounds := src.Bounds()
	// for x:=0; x<bounds.Max.X
	return src
}

func AdaptiveThreshold(src image.Gray, maxValue float64, typeOf int, blockSize int, delta float64) (dst image.Gray, err error) {
	if blockSize%2 != 1 || blockSize < 1 {
		return dst, ErrBlockSize
	}

	dst = *image.NewGray(src.Bounds())

	if maxValue < 0 {
		return src, nil
	}

	boxFilter(src, blockSize)

	return dst, nil
}
