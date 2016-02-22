package digits

import (
	"image"

	"github.com/fxsjy/gonn/gonn"
)

var nn *gonn.NeuralNetwork

func init() {
	nn = gonn.LoadNN("mnist.nn")
}

func argmax(A []float64) (int, float64) {
	x := 0
	v := -1.0
	for i, a := range A {
		if a > v {
			x = i
			v = a
		}
	}
	return x, v
}

// RecogniseDigit takes 28x28 gray image and tries to recognise a digit
// panics if image has wrong size
func RecogniseDigit(img image.Gray) (int, float64) {
	if img.Bounds().Max.X != 28 || img.Bounds().Max.Y != 28 {
		panic("Image size is invalid, use 28x28.")
	}

	input := make([]float64, 28*28, 28*28)
	for i, pix := range img.Pix {
		if pix < 128 {
			img.Pix[i] = 255
			input[i] = 255
		} else {
			img.Pix[i] = 0
			input[i] = 128
		}
	}
	digit, confidence := argmax(nn.Forward(input))
	return digit, confidence
}
