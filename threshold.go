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

// Port from: https://en.wikipedia.org/wiki/Otsu%27s_method
func otsuValue(img image.Gray) uint8 {
	size := float64(len(img.Pix))
	histogram := make([]float64, 256, 256)
	for _, pix := range img.Pix {
		histogram[pix] += 1.0
	}

	sum := 0.0
	for i, val := range histogram {
		sum += float64(i) * val
	}

	sumB := 0.0
	wB := 0.0
	wF := 0.0
	mB := 0.0
	mF := 0.0
	max := 0.0
	between := 0.0
	threshold1 := 0.0
	threshold2 := 0.0

	for i, val := range histogram {
		wB += val
		if wB == 0 {
			continue
		}

		wF = size - wB
		if wF == 0 {
			break
		}
		sumB += float64(i) * val

		mB = sumB / wB
		mF = (sum - sumB) / wF
		between = wB * wF * (mB - mF) * (mB - mF)
		if between >= max {
			threshold1 = float64(i)
			if between > max {
				threshold2 = float64(i)
			}
			max = between
		}
	}

	return uint8((threshold1 + threshold2) / 2.0)
}
