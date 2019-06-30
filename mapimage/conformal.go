package mapimage

import (
	"gonum.org/v1/gonum/mat"
	"math"
)

type ConformalTransformation struct {
	trans *mat.Dense
}

func NewConformalTransformation(k, theta, Tx, Ty float64) ConformalTransformation {
	a := k * math.Cos(theta)
	b := k * math.Sin(theta)
	trans := mat.NewDense(4, 1, []float64{a, b, Tx, Ty})
	return ConformalTransformation{trans}
}

func NewConformalTransformationFromPoints(standardPoints []Point, localPoints []Point) (ConformalTransformation, error) {
	n1, e1 := standardPoints[0].Lat, standardPoints[0].Lng
	n2, e2 := standardPoints[1].Lat, standardPoints[1].Lng

	E := mat.NewDense(4, 1, []float64{e1, n1, e2, n2})
	y1, x1 := localPoints[0].Lat, localPoints[0].Lng
	y2, x2 := localPoints[1].Lat, localPoints[1].Lng
	X := mat.NewDense(4, 4, []float64{
		x1, -y1, 1, 0,
		y1, x1, 0, 1,
		x2, -y2, 1, 0,
		y2, x2, 0, 1,
	})

	var trans mat.Dense
	err := trans.Solve(X, E)
	if err != nil {
		return ConformalTransformation{}, err
	}

	// printMat("E", E)
	// printMat("X", X)
	// printMat("trans", &trans)

	return ConformalTransformation{&trans}, nil
}

func (t *ConformalTransformation) Project(p Point) Point {
	return t.Projects(p)[0]
}

func (t *ConformalTransformation) Projects(points ...Point) (results []Point) {
	if len(points) == 0 {
		return
	}

	X := mat.NewDense(2*len(points), 4, nil)
	for i, p := range points {
		y := p.Lat
		x := p.Lng
		X.SetRow(2*i, []float64{x, -y, 1, 0})
		X.SetRow(2*i+1, []float64{y, x, 0, 1})
	}

	// printMat("trans", t.trans)
	var transformed mat.Dense
	transformed.Mul(X, t.trans)

	for i, _ := range points {
		easting := transformed.At(2*i, 0)
		northing := transformed.At(2*i+1, 0)
		results = append(results, Point{Lat: northing, Lng: easting})
	}
	return
}
