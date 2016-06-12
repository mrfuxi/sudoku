package digits

import (
	"image"
	"image/color"
	"log"
	"os"

	"github.com/mrfuxi/neural"
)

var nn neural.Evaluator

// LoadNetwork intializes global neural network used to process digits
func LoadNetwork(fileName string) {
	inputSize := 28 * 28

	activator := neural.NewSigmoidActivator()
	outActivator := neural.NewSoftmaxActivator()
	nn = neural.NewNeuralNetwork(
		[]int{inputSize, 100, 10},
		neural.NewFullyConnectedLayer(activator),
		neural.NewFullyConnectedLayer(outActivator),
	)

	fn, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	if err := neural.Load(nn, fn); err != nil {
		log.Fatal(err)
	}
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
func RecogniseDigit(img image.Gray, threshold uint8) (int, float64) {
	if img.Bounds().Max.X != 28 || img.Bounds().Max.Y != 28 {
		panic("Image size is invalid, use 28x28.")
	}

	input := make([]float64, 28*28, 28*28)
	pos := 0
	for x := 0; x < 28; x++ {
		for y := 0; y < 28; y++ {
			val := img.GrayAt(x, y).Y
			if val < threshold {
				val = 255 - val
			} else {
				val = 0
			}

			img.SetGray(x, y, color.Gray{Y: val})
			input[pos] = float64(val) / 255
			pos++
		}
	}
	digit, confidence := argmax(nn.Evaluate(input))
	return digit, confidence
}
