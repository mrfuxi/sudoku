package sudoku

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path"

	"github.com/gonum/matrix/mat64"
)

func saveImage(image image.Image, name string) error {
	filePath := path.Join("examples_out", name)

	outfile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outfile.Close()
	png.Encode(outfile, image)
	return nil
}

func grayImage(src image.Image) (dst image.Gray) {
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	dst = *image.NewGray(bounds)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			srcColor := src.At(x, y)
			dstColor := color.GrayModel.Convert(srcColor)
			dst.Set(x, y, dstColor)
		}
	}
	return dst
}

func windowSize(img image.Image, divider int) int {
	bounds := img.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y

	max := h
	if w > max {
		max = h
	}

	window := max / divider
	if window%2 == 0 {
		window++
	}
	return window
}

func imageToMatrix(src image.Gray) (dst *mat64.Dense) {
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	dst = mat64.NewDense(h, w, nil)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			srcColor := src.GrayAt(x, y)
			dst.Set(y, x, float64(srcColor.Y))
		}
	}
	return dst
}

func matrixToImage(src *mat64.Dense) (dst image.Gray) {
	rows, cols := src.Dims()
	mult := 1.0
	if mat64.Max(src) <= 1.0 {
		mult = 255.0
	}
	dst = *image.NewGray(image.Rect(0, 0, cols, rows))
	for col := 0; col < cols; col++ {
		for row := 0; row < rows; row++ {
			srcColor := src.At(row, col)
			dst.SetGray(col, row, color.Gray{uint8(srcColor * mult)})
		}
	}
	return dst
}

// Initial threshold to get binary image
func binarize(src image.Gray) (dst image.Gray) {
	window := windowSize(&src, 10)
	return adaptiveThreshold(src, 255, threshBinaryInv, (window-1)/2, 0)
}

// Removes body of regions over 1/20 of image width/height
func removeBlobsBody(src image.Gray) (dst image.Gray) {
	window := windowSize(&src, 20)
	return adaptiveThreshold(src, 255, threshBinary, (window-1)/2, -128)
}

// PreProcess prepares original image for actual work
// - coverts to gray scale
// - threshold to produce binary image
// - removes some of big areas/blobs
func preProcess(img image.Image) *mat64.Dense {
	grayImg := grayImage(img)
	binary := binarize(grayImg)
	deblobbed := removeBlobsBody(binary)
	mat := imageToMatrix(deblobbed)
	return mat
}
