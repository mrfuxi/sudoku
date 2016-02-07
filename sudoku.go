package sudoku

import (
	"errors"
	"fmt"
	"image"
	"time"

	"github.com/gonum/matrix/mat64"
)

// ErrNotRecognised is reported when sudoku could not be localized on image
var ErrNotRecognised = errors.New("Could not find sudoku on the image")

// Sudoku interface describes access to recognised sudoku puzzle
type Sudoku interface {
	Overlay() image.Image
}

type lineSudoku struct {
	BaseImage    image.Image
	PreProcessed *mat64.Dense
	Grid         lineGrid
	Recognised   bool
}

func (l *lineSudoku) Overlay() image.Image {
	if !l.Recognised {
		return nil
	}
	return drawLines(l.BaseImage, append(l.Grid.Horizontal, l.Grid.Vertical...))
}

// NewSudoku processes given image in order to find sudoku puzzle on the image
func NewSudoku(image image.Image) (s Sudoku, err error) {
	sudoku := &lineSudoku{
		BaseImage: image,
	}
	width, height := sudoku.BaseImage.Bounds().Max.X, sudoku.BaseImage.Bounds().Max.Y

	t0 := time.Now()
	sudoku.PreProcessed = preProcess(sudoku.BaseImage)
	t1 := time.Now()
	lines := houghLines(sudoku.PreProcessed, nil, 80, 200)
	lines = removeDuplicateLines(lines, width, height)
	bucketSize := 90 / 5
	buckets := generateAngleBuckets(uint(bucketSize), uint(bucketSize/2.0), true)
	bucketedLines := putLinesIntoBuckets(buckets, lines)

	grids := make([]lineGrid, 0, 0)
	for angle, lineClass := range bucketedLines {
		// don't even bother doing any more work
		// it's not a 9x9 grid
		if len(lineClass) < 20 {
			continue
		}

		vertical, horizontal := linesWithSimilarAngle(lineClass, angle)

		if len(vertical) < 10 || len(horizontal) < 10 {
			continue
		}

		grids = append(grids, possibleGrids(horizontal, vertical)...)
	}

	evaluateGrids(sudoku.PreProcessed, grids)
	if len(grids) != 0 {
		sudoku.Grid = grids[0] // Best grid
		sudoku.Recognised = true
	} else {
		err = ErrNotRecognised
	}

	t2 := time.Now()
	fmt.Printf("Time to find Sudoku %v. PreProcessing: %v. Success: %v\n", t2.Sub(t0), t1.Sub(t0), sudoku.Recognised)

	return sudoku, err
}
