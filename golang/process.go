package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"path"
	"runtime/pprof"
	"time"

	"github.com/gonum/matrix/mat64"
)

const (
	ExampleDir   = "../examples/"
	SaveLocation = "examples_out"
)

func getExampleImage(name string) (image.Image, error) {
	filePath := path.Join(ExampleDir, name)
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return png.Decode(reader)
}

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
func binarize(src *mat64.Dense) (dst *mat64.Dense) {
	rows, cols := src.Dims()
	max := rows
	if cols > max {
		max = cols
	}

	window := max / 10
	if window%2 == 0 {
		window += 1
	}

	return AdaptiveThreshold(src, 1, ThreshBinaryInv, window, 0)
}

// Removes body of regions over 1/20 of image width/height
func removeBlobsBody(src *mat64.Dense) (dst *mat64.Dense) {
	_, cols := src.Dims()
	max := cols
	if cols > max {
		max = cols
	}

	window := max / 20
	if window%2 == 0 {
		window += 1
	}

	return AdaptiveThreshold(src, 1, ThreshBinary, window, -0.5)
}

func main() {
	os.RemoveAll(SaveLocation)
	os.MkdirAll(SaveLocation, os.ModePerm)

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	img, err := getExampleImage("s2.png")
	if err != nil {
		fmt.Println(err)
	}
	gray := grayImage(img)
	mat := imageToMatrix(gray)

	binary := binarize(mat)
	deblobbed := removeBlobsBody(binary)

	t0 := time.Now()
	lines := HoughLines(deblobbed, nil, 80)
	t1 := time.Now()
	fmt.Printf("The call took %v to run.\n", t1.Sub(t0))
	l := drawLines(img, lines)

	i := matrixToImage(binary)
	saveImage(&i, "b.png")

	j := matrixToImage(deblobbed)
	saveImage(&j, "d.png")

	saveImage(l, "l.png")
}
