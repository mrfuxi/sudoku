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
	dst = *image.NewGray(image.Rect(0, 0, cols, rows))
	for col := 0; col < cols; col++ {
		for row := 0; row < rows; row++ {
			srcColor := src.At(row, col)
			dst.SetGray(col, row, color.Gray{uint8(srcColor)})
		}
	}
	return dst
}

// Initial threshold to get binary image
func binarize(src *mat64.Dense) (dst *mat64.Dense, err error) {
	rows, cols := src.Dims()
	max := rows
	if cols > max {
		max = cols
	}

	window := max / 10
	if window%2 == 0 {
		window += 1
	}

	return AdaptiveThreshold(src, 255, ThreshBinaryInv, window, 0)
}

func main() {
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

	t0 := time.Now()
	bin, err := binarize(mat)
	if err != nil {
		fmt.Println(err)
	}
	t1 := time.Now()
	fmt.Printf("The call took %v to run.\n", t1.Sub(t0))

	i := matrixToImage(bin)
	err = saveImage(&i, "t.png")
	if err != nil {
		fmt.Println(err)
	}
}
