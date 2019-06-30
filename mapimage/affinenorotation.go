package mapimage

import (
	"gonum.org/v1/gonum/mat"
)

type AffineNoRotTransformation struct {
	trans *mat.Dense
}

func NewAffineNoRotTransformation(Sx, Sy, Tx, Ty float64) AffineNoRotTransformation {
	trans := mat.NewDense(4, 1, []float64{Sx, Sy, Tx, Ty})
	//printMat("trans", trans)
	return AffineNoRotTransformation{trans}
}

func NewAffineNoRotTransformationFromPoints(standardPoints []Point, localPoints []Point) (AffineNoRotTransformation, error) {
	n1, e1 := standardPoints[0].Lat, standardPoints[0].Lng
	n2, e2 := standardPoints[1].Lat, standardPoints[1].Lng

	E := mat.NewDense(4, 1, []float64{e1, n1, e2, n2})
	y1, x1 := localPoints[0].Lat, localPoints[0].Lng
	y2, x2 := localPoints[1].Lat, localPoints[1].Lng
	X := mat.NewDense(4, 4, []float64{
		x1, 0, 1, 0,
		0, y1, 0, 1,
		x2, 0, 1, 0,
		0, y2, 0, 1,
	})

	var invX mat.Dense
	err := invX.Inverse(X)
	if err != nil {
		return AffineNoRotTransformation{}, err
	}

	var trans mat.Dense
	trans.Mul(&invX, E)

	// printMat("E", E)
	// printMat("X", X)
	// printMat("invX", &invX)
	//printMat("trans", &trans)

	return AffineNoRotTransformation{&trans}, nil
}

func (t *AffineNoRotTransformation) Project(p Point) Point {
	return t.Projects(p)[0]
}

func (t *AffineNoRotTransformation) Projects(points ...Point) (results []Point) {
	if len(points) == 0 {
		return
	}

	X := mat.NewDense(2*len(points), 4, nil)
	for i, p := range points {
		y := p.Lat
		x := p.Lng
		X.SetRow(2*i, []float64{x, 0, 1, 0})
		X.SetRow(2*i+1, []float64{0, y, 0, 1})
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
