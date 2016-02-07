package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime/pprof"
	"strings"

	"github.com/mrfuxi/sudoku"
)

const (
	exampleDir   = "examples"
	saveLocation = "examples_out"
)

func getExampleImage(name string) (image.Image, error) {
	filePath := path.Join(exampleDir, name)
	reader, err := os.Open(filePath)
	if err != nil {
		log.Fatalln(err)
	}

	img, _, err := image.Decode(reader)
	return img, err
}

func findSudoku(filename string, debug bool) (sudoku.Sudoku, error) {
	img, err := getExampleImage(filename)
	if err != nil {
		log.Fatalln(err)
	}

	return sudoku.NewSudoku(img)
}

func main() {
	os.RemoveAll(saveLocation)
	os.MkdirAll(saveLocation, os.ModePerm)

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

	var s sudoku.Sudoku
	var err error

	if *file != "" {
		s, err = findSudoku(*file, *debug)
	} else {
		fileInfos, err := ioutil.ReadDir(exampleDir)
		if err != nil {
			log.Fatal(err)
		}
		for _, fileInfo := range fileInfos {
			if strings.HasSuffix(fileInfo.Name(), ".png") {
				s, err = findSudoku(fileInfo.Name(), *debug)
			}
		}
	}

	if err == nil && s != nil {
		fmt.Println("Sudoku has been found!")
	} else {
		fmt.Println("Could not find sudoku")
	}

}
