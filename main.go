package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"
)

func grid(filename string, debug bool) {
	img, err := getExampleImage(filename)
	if err != nil {
		fmt.Println(err)
	}
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y

	t0 := time.Now()
	preparedImg := PreProcess(img)
	lines := HoughLines(preparedImg, nil, 80, 200)
	lines = removeDuplicateLines(lines, width, height)
	bucketSize := 90 / 5
	buckets := generateAngleBuckets(uint(bucketSize), uint(bucketSize/2.0), true)
	bucketedLines := putLinesIntoBuckets(buckets, lines)

	grids := make([]Grid, 0, 0)
	for angle, line_class := range bucketedLines {
		// don't even bother doing any more work
		// it's not a 9x9 grid
		if len(line_class) < 20 {
			continue
		}

		vertical, horizontal := linesWithSimilarAngle(line_class, angle)

		if len(vertical) < 10 || len(horizontal) < 10 {
			continue
		}

		grids = append(grids, possibleGrids(horizontal, vertical)...)
	}

	evaluateGrids(preparedImg, grids)
	if len(grids) != 0 {
		bestGrid := grids[0]
		l := drawLines(img, append(bestGrid.Horizontal, bestGrid.Vertical...))
		saveImage(l, filename)
	}

	t1 := time.Now()
	fmt.Printf("The call took %v to run.\n", t1.Sub(t0))

	if debug {
		j := matrixToImage(preparedImg)
		saveImage(&j, "prepared.png")

		l := drawLines(img, lines)
		saveImage(l, "lines.png")
	}
}

func main() {
	os.RemoveAll(SaveLocation)
	os.MkdirAll(SaveLocation, os.ModePerm)

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var debug = flag.Bool("debug", false, "prepare debug images")
	var file = flag.String("file", "", "file to process")

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *file != "" {
		grid(*file, *debug)
	} else {
		fileInfos, err := ioutil.ReadDir(ExampleDir)
		if err != nil {
			log.Fatal(err)
		}
		for _, fileInfo := range fileInfos {
			if strings.HasSuffix(fileInfo.Name(), ".png") {
				grid(fileInfo.Name(), *debug)
			}
		}
	}
}
