package sudoku

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path"
	"sync"
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
	var wg sync.WaitGroup
	bounds := src.Bounds()
	w, h := bounds.Max.X, bounds.Max.Y
	dst = *image.NewGray(bounds)
	for x := 0; x < w; x++ {
		wg.Add(1)
		go func(x int) {
			for y := 0; y < h; y++ {
				srcColor := src.At(x, y)
				dstColor := color.GrayModel.Convert(srcColor)
				dst.Set(x, y, dstColor)
			}
			wg.Done()
		}(x)
	}
	wg.Wait()
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
func preProcess(img image.Image) image.Gray {
	grayImg := grayImage(img)
	binary := binarize(grayImg)
	deblobbed := removeBlobsBody(binary)
	return deblobbed
}
