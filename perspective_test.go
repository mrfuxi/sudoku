package sudoku

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPerspectiveTransformation(t *testing.T) {
	src := [4]pointF{
		pointF{54, 64},
		pointF{368, 52},
		pointF{391, 391},
		pointF{27, 387},
	}

	dst := [4]pointF{
		pointF{0, 0},
		pointF{420, 0},
		pointF{420, 420},
		pointF{0, 420},
	}

	proj := newPerspective(src, dst)

	for i := range src {
		x, y := proj.Project(src[i].X, src[i].Y)
		assert.InDelta(t, x, dst[i].X, 0.001)
		assert.InDelta(t, y, dst[i].Y, 0.001)
	}
}
