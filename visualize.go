package sudoku

import (
	"image"
	"image/color"
	"math"

	"github.com/llgcode/draw2d/draw2dimg"
)

func drawLines(src image.Image, lines []polarLine) image.Image {
	dst := image.NewRGBA(src.Bounds())

	gc := draw2dimg.NewGraphicContext(dst)
	gc.SetFillColor(color.RGBA{0x0, 0x0, 0x0, 0xff})
	gc.SetStrokeColor(color.RGBA{0x0, 0xff, 0x0, 0xff})
	gc.SetLineWidth(2)
	gc.Clear()
	gc.DrawImage(src)

	for _, line := range lines {
		a := math.Cos(line.Theta)
		b := math.Sin(line.Theta)
		x0 := a * float64(line.Distance)
		y0 := b * float64(line.Distance)
		x1 := (x0 + 10000*(-b))
		y1 := (y0 + 10000*(a))
		x2 := (x0 - 10000*(-b))
		y2 := (y0 - 10000*(a))

		gc.MoveTo(x1, y1)
		gc.LineTo(x2, y2)
		gc.Close()
	}

	gc.FillStroke()
	return dst
}
