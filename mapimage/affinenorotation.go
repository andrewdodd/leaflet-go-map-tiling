package mapimage

import (
	"gonum.org/v1/gonum/mat"
)

type AffineNoRotTransformation struct {
	trans *mat.Dense
}

func NewAffineNoRotTransformation(Sx, Sy, Tx, Ty float64) AffineNoRotTransformation {
	//   --
	// | Sx |
	// | Sy |
	// | Tx |
	// | Ty |
	//   --

	trans := mat.NewDense(4, 1, []float64{Sx, Sy, Tx, Ty})
	return AffineNoRotTransformation{trans}
}

/*

    System 1                             System 2

    ^                                    ^
    |                                    |
    |    B---------C                     |
    |    |         |                     |     b---------------c
    |    |         |            =        |     |               |
    |    |         |                     |     |               |
    |    A---------D                     |     a---------------d
    |                                    |
    +-------------------->               +--------------------------->

	Supply matching points from both coordinate systems. I.e:

	proj := NewAffineNoRotTransformationFromPoints([A, C], [a, c])
	proj.Project(b) => returns B

	 or for the opposite projection:

	proj := NewAffineNoRotTransformationFromPoints([a, c], [A, C])
	proj.Project(B) => returns b
*/
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

	//          X         *    t    =    e
	//   --------------       ----      ----
	//  | x1  0  1  0  |     | Sx |    | e1 |
	//  |  0 y1  0  1  |  *  | Sy | =  | n1 |
	//  | x2  0  1  0  |     | Tx |    | e2 |
	//  |  0 y2  0  1  |     | Ty |    | n2 |
	//   --------------       ----      ----

	var trans mat.Dense
	// https://en.wikipedia.org/wiki/System_of_linear_equations
	// Expressed in the form Ax=b (or Xt=e)
	// I.e. solve for T
	// T = X^-1 * E
	err := trans.Solve(X, E)
	if err != nil {
		return AffineNoRotTransformation{}, err
	}

	return AffineNoRotTransformation{&trans}, nil
}

func (t *AffineNoRotTransformation) Project(p Point) Point {
	// Project just one point and retrieve it from the returned slice
	return t.Projects(p)[0]
}

func (t *AffineNoRotTransformation) Projects(points ...Point) (results []Point) {
	if len(points) == 0 {
		return
	}

	// Build "X" such that there are two rows for each input point:
	//
	//          X         *    t    =    e
	//   --------------       ----      ----
	//  | x1  0  1  0  |     | Sx |    | e1 |
	//  |  0 y1  0  1  |  *  | Sy | =  | n1 |
	//  | x2  0  1  0  |     | Tx |    | e2 |
	//  |  0 y2  0  1  |     | Ty |    | n2 |
	//  |  .  .  .  .  |      ----|    |  . |
	//  |  .  .  .  .  |               |  . |
	//  |  .  .  .  .  |               |  . |
	//  | xN  0  1  0  |               | eN |
	//  |  0 yN  0  1  |               | nN |
	//   --------------                 ----
	X := mat.NewDense(2*len(points), 4, nil)
	for i, p := range points {
		y := p.Lat
		x := p.Lng
		X.SetRow(2*i, []float64{x, 0, 1, 0})
		X.SetRow(2*i+1, []float64{0, y, 0, 1})
	}

	var transformed mat.Dense
	transformed.Mul(X, t.trans)

	for i, _ := range points {
		easting := transformed.At(2*i, 0)
		northing := transformed.At(2*i+1, 0)
		results = append(results, Point{Lat: northing, Lng: easting})
	}
	return
}
