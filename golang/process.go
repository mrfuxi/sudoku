package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path"
)

func getExampleImage(name string) (image.Image, error) {
	filePath := path.Join("../examples/", name)
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return png.Decode(reader)
}

func saveImage(image image.Image, name string) error {
	filePath := path.Join("examples_out/", name)

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

// Initial threshold to get binary image
func binarize(src image.Gray) (dst image.Gray, err error) {
	window := src.Bounds().Max.X / 10
	if window%2 == 0 {
		window += 1
	}

	return AdaptiveThreshold(src, 1, ThreshBinaryInv, window, 0)
}

func main() {
	img, err := getExampleImage("s2.png")
	if err != nil {
		fmt.Println(err)
	}
	gray := grayImage(img)

	bin, err := binarize(gray)
	if err != nil {
		fmt.Println(err)
	}

	err = saveImage(&bin, "t.png")
	if err != nil {
		fmt.Println(err)
	}
}
