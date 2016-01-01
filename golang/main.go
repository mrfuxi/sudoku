package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	os.RemoveAll(SaveLocation)
	os.MkdirAll(SaveLocation, os.ModePerm)

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var debug = flag.Bool("debug", false, "prepare debug images")

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
	width, height := img.Bounds().Max.X, img.Bounds().Max.Y

	t0 := time.Now()
	preparedImg := PreProcess(img)
	lines := HoughLines(preparedImg, nil, 80, 100)
	lines = removeDuplicateLines(lines, width, height)
	t1 := time.Now()
	fmt.Printf("The call took %v to run.\n", t1.Sub(t0))

	if *debug {
		j := matrixToImage(preparedImg)
		saveImage(&j, "prepared.png")

		l := drawLines(img, lines)
		saveImage(l, "lines.png")
	}
}
