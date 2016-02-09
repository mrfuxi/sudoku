package sudoku

import "github.com/gonum/matrix/mat64"

type pointF struct {
	X float64
	Y float64
}

type perspectiveTrasnformation struct {
	H *mat64.Dense
}

func (p *perspectiveTrasnformation) Project(x, y float64) (float64, float64) {
	v := mat64.NewDense(3, 1, []float64{x, y, 1})
	v.Mul(p.H, v)
	scale := v.At(2, 0)
	X := v.At(0, 0) / scale
	Y := v.At(1, 0) / scale
	return X, Y
}

func newPerspective(src [4]pointF, dst [4]pointF) *perspectiveTrasnformation {
	b := make([]float64, 8, 8)
	A := mat64.NewDense(8, 8, nil)

	for i := 0; i < 4; i++ {
		A.Set(i, 0, src[i].X)
		A.Set(i+4, 3, src[i].X)
		A.Set(i, 1, src[i].Y)
		A.Set(i+4, 4, src[i].Y)

		A.Set(i, 2, 1)
		A.Set(i+4, 5, 1)

		A.Set(i, 3, 0)
		A.Set(i, 4, 0)
		A.Set(i, 5, 0)
		A.Set(i+4, 0, 0)
		A.Set(i+4, 1, 0)
		A.Set(i+4, 2, 0)

		A.Set(i, 6, -src[i].X*dst[i].X)
		A.Set(i, 7, -src[i].Y*dst[i].X)
		A.Set(i+4, 6, -src[i].X*dst[i].Y)
		A.Set(i+4, 7, -src[i].Y*dst[i].Y)
		b[i] = dst[i].X
		b[i+4] = dst[i].Y
	}

	B := mat64.NewDense(8, 1, b)
	X := mat64.NewDense(8, 1, nil)
	X.Solve(A, B)

	data := mat64.Col(nil, 0, X)
	data = append(data, 1)

	projection := &perspectiveTrasnformation{
		H: mat64.NewDense(3, 3, data),
	}

	return projection
}
