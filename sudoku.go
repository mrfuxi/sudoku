package sudoku

import (
	"errors"
	"fmt"
	"image"
	"time"
)

var NotRecognisedErr = errors.New("Could not find sudoku on the image")

type Sudoku struct {
	Image image.Image
}

func NewSudoku(image image.Image) (s Sudoku, err error) {
	width, height := image.Bounds().Max.X, image.Bounds().Max.Y

	t0 := time.Now()
	preparedImg := PreProcess(image)
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
		l := drawLines(image, append(bestGrid.Horizontal, bestGrid.Vertical...))
		s.Image = l
	} else {
		err = NotRecognisedErr
	}

	t1 := time.Now()
	fmt.Printf("The call took %v to run.\n", t1.Sub(t0))

	// if debug {
	// 	j := matrixToImage(preparedImg)
	// 	saveImage(&j, "prepared.png")

	// 	l := drawLines(image, lines)
	// 	saveImage(l, "lines.png")
	// }

	return
}
