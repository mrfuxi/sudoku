package sudoku

import (
	"image"
	"math"
	"sync"

	"github.com/gonum/matrix/mat64"
)

type pointF struct {
	X float64
	Y float64
}

func newPointF(pt xyPoint) pointF {
	return pointF{
		X: float64(pt.X),
		Y: float64(pt.Y),
	}
}

type perspectiveTrasnformation struct {
	H11       float64 // Deconstructed homography matrix
	H12       float64
	H13       float64
	H21       float64
	H22       float64
	H23       float64
	H31       float64
	H32       float64
	H33       float64
	srcPoints [4]pointF
	dstPoints [4]pointF
}

func (p *perspectiveTrasnformation) Project(x, y float64) (float64, float64) {
	scale := (p.H31*x + p.H32*y + p.H33)
	X := (p.H11*x + p.H12*y + p.H13) / scale
	Y := (p.H21*x + p.H22*y + p.H23) / scale
	return X, Y
}

func newPerspective(src [4]pointF, dst [4]pointF) *perspectiveTrasnformation {
	b := make([]float64, 8, 8)
	A := mat64.NewDense(8, 8, nil)

	for i := 0; i < 4; i++ {
		A.Set(i, 0, src[i].X)
		A.Set(i+4, 3, src[i].X)
		A.Set(i, 1, src[i].Y)
		A.Set(i+4, 4, src[i].Y)

		A.Set(i, 2, 1)
		A.Set(i+4, 5, 1)

		A.Set(i, 3, 0)
		A.Set(i, 4, 0)
		A.Set(i, 5, 0)
		A.Set(i+4, 0, 0)
		A.Set(i+4, 1, 0)
		A.Set(i+4, 2, 0)

		A.Set(i, 6, -src[i].X*dst[i].X)
		A.Set(i, 7, -src[i].Y*dst[i].X)
		A.Set(i+4, 6, -src[i].X*dst[i].Y)
		A.Set(i+4, 7, -src[i].Y*dst[i].Y)
		b[i] = dst[i].X
		b[i+4] = dst[i].Y
	}

	B := mat64.NewDense(8, 1, b)
	homography := mat64.NewDense(8, 1, nil)
	homography.Solve(A, B)

	projection := &perspectiveTrasnformation{
		srcPoints: src,
		dstPoints: dst,
		H11:       homography.At(0, 0),
		H12:       homography.At(1, 0),
		H13:       homography.At(2, 0),
		H21:       homography.At(3, 0),
		H22:       homography.At(4, 0),
		H23:       homography.At(5, 0),
		H31:       homography.At(6, 0),
		H32:       homography.At(7, 0),
		H33:       1,
	}

	return projection
}

func (p *perspectiveTrasnformation) warpPerspective(src *image.Gray) *image.Gray {
	var wg sync.WaitGroup
	maxX := 0.0
	maxY := 0.0
	for _, p := range p.dstPoints {
		maxX = math.Max(maxX, p.X)
		maxY = math.Max(maxY, p.Y)
	}
	dst := image.NewGray(image.Rect(0, 0, int(maxX), int(maxY)))
	mask := make([]bool, len(dst.Pix), len(dst.Pix))

	srcWidth := src.Bounds().Max.X
	srcHeight := src.Bounds().Max.Y
	for x := 0; x < srcWidth; x++ {
		wg.Add(1)
		go func(x int) {
			for y := 0; y < srcHeight; y++ {
				newX, newY := p.Project(float64(x), float64(y))
				if newX < 0 || newX >= maxX || newY < 0 || newY >= maxY {
					continue
				}

				g := src.Pix[src.PixOffset(x, y)]

				dstPos := dst.PixOffset(int(newX), int(newY))
				dst.Pix[dstPos] = g
				mask[dstPos] = true
			}
			wg.Done()
		}(x)
	}

	wg.Wait()
	interpolateMissingPixels(dst, mask)
	return dst
}

// Fill in missing pixels
func interpolateMissingPixels(img *image.Gray, mask []bool) {
	var wg sync.WaitGroup
	newMask := make([]bool, len(mask), len(mask))
	copy(newMask, mask)
	couldNotFill := false

	imgWidth := img.Bounds().Max.X
	imgHeight := img.Bounds().Max.Y
	for x := 0; x < imgWidth; x++ {
		wg.Add(1)
		go func(x int) {
			for y := 0; y < imgHeight; y++ {
				imgPos := img.PixOffset(x, y)
				if mask[imgPos] == true {
					continue
				}

				var sum int
				cnt := 0

				if x > 0 && mask[imgPos-1] {
					sum += int(img.Pix[imgPos-1])
					cnt++
				}
				if x < imgWidth-1 && mask[imgPos+1] {
					sum += int(img.Pix[imgPos+1])
					cnt++
				}
				if y > 0 && mask[imgPos-img.Stride] {
					sum += int(img.Pix[imgPos-img.Stride])
					cnt++
				}
				if y < imgHeight-1 && mask[imgPos+img.Stride] {
					sum += int(img.Pix[imgPos+img.Stride])
					cnt++
				}
				if cnt != 0 {
					img.Pix[imgPos] = uint8(sum / cnt)
					newMask[imgPos] = true
				} else {
					couldNotFill = true
				}
			}
			wg.Done()
		}(x)
	}

	wg.Wait()
	if couldNotFill {
		interpolateMissingPixels(img, newMask)
	}
}
