package sudoku

import (
	"image"
	"sync"
)

type thresholdType int

const (
	threshBinary thresholdType = iota
	threshBinaryInv
)

func inRange(val, max int) int {
	if val < 0 {
		return 0
	}
	if val > max-1 {
		return max - 1
	}
	return val
}

func meanHorizontal(src image.Gray, radius int) (dst image.Gray) {
	var wg sync.WaitGroup
	norm := float64(radius*2 + 1)
	dst = *image.NewGray(src.Bounds())
	width, height := src.Bounds().Max.X, src.Bounds().Max.Y

	for y := 0; y < height; y++ {
		wg.Add(1)
		go func(y int) {
			total := 0.0

			for kx := -radius; kx <= radius; kx++ {
				total += float64(src.Pix[src.PixOffset(inRange(kx, width), y)])
			}
			dst.Pix[dst.PixOffset(0, y)] = uint8(total / norm)

			for x := 1; x < width; x++ {
				total -= float64(src.Pix[src.PixOffset(inRange(x-radius-1, width), y)])
				total += float64(src.Pix[src.PixOffset(inRange(x+radius, width), y)])

				dst.Pix[dst.PixOffset(x, y)] = uint8(total / norm)
			}
			wg.Done()
		}(y)
	}
	wg.Wait()
	return
}

func meanVertical(src image.Gray, radius int) (dst image.Gray) {
	var wg sync.WaitGroup
	norm := float64(radius*2 + 1)
	dst = *image.NewGray(src.Bounds())
	width, height := src.Bounds().Max.X, src.Bounds().Max.Y

	for x := 0; x < width; x++ {
		wg.Add(1)
		go func(x int) {

			total := 0.0

			for ky := -radius; ky <= radius; ky++ {
				total += float64(src.Pix[src.PixOffset(x, inRange(ky, height))])
			}
			dst.Pix[dst.PixOffset(x, 0)] = uint8(total / norm)

			for y := 1; y < height; y++ {
				total -= float64(src.Pix[src.PixOffset(x, inRange(y-radius-1, height))])
				total += float64(src.Pix[src.PixOffset(x, inRange(y+radius, height))])

				dst.Pix[dst.PixOffset(x, y)] = uint8(total / norm)
			}
			wg.Done()
		}(x)
	}
	wg.Wait()
	return
}

func mean(src image.Gray, radius int) image.Gray {
	return meanVertical(meanHorizontal(src, radius), radius)
}

func adaptiveThreshold(src image.Gray, maxValue uint8, threshold thresholdType, radius int, delta int) image.Gray {
	dst := mean(src, radius)

	var newVal uint8
	for i, srcVal := range src.Pix {
		newVal = 0
		dstVal := int(dst.Pix[i]) - (delta)

		if threshold == threshBinary {
			if int(srcVal) > dstVal {
				newVal = maxValue
			}
		} else if threshold == threshBinaryInv {
			if int(srcVal) < dstVal {
				newVal = maxValue
			}
		}

		dst.Pix[i] = newVal
	}

	return dst
}
