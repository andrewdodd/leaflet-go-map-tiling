package mapimage

import (
	"testing"
)

/*

    Coordinates 1                       Coords 2

    ^                                    ^
    |                                    |
   5+    B---------C                     |
    |    |         |                  70 +     b---------------c
    |    |         |            =>       |     |               |
    |    |         |                     |     |               |
   1+    A---------D                  50 +     a---------------d
    |                                    |
    +----+---------+----->               +-----+---------------+----->
         1         5                           20             120

	Scale-X = 4:100                     Scale-X = 100:4
	Scale-Y = 4:20                      Scale-Y = 20:4
	Translation-X = 19                  Translation-X = -19
	Translation-Y = 49                  Translation-Y = -49

	A(1,1), C(5,5)             =>       a(20,50), c(120, 70)
*/

func testPointsInBothProjections(t *testing.T, fromOne, fromTwo AffineNoRotTransformation) {

	var testPoints = []struct {
		coords1Desc string
		coords1Pt   Point
		coords2Desc string
		coords2Pt   Point
	}{
		{"A(1,1)", PointXY(1, 1), "a(20,50)", PointXY(20, 50)},
		{"C(5,5)", PointXY(5, 5), "c(120,70)", PointXY(120, 70)},
		{"Origin(0,0)", PointXY(0, 0), "pt(-5, 45)", PointXY(-5, 45)},
		{"Centre(3,3)", PointXY(3, 3), "Centre(70, 60)", PointXY(70, 60)},
	}
	for _, tt := range testPoints {
		var inputDesc, expectDesc string
		var result, expect Point
		result = fromOne.Project(tt.coords1Pt)
		inputDesc = tt.coords1Desc
		expect = tt.coords2Pt
		expectDesc = tt.coords2Desc

		if !result.IsCloseTo(expect) {
			t.Errorf("incorrect for %v to %v, got: %v, want: %v.", inputDesc, expectDesc, result, expect)
		}

		result = fromTwo.Project(tt.coords2Pt)
		inputDesc = tt.coords2Desc
		expect = tt.coords1Pt
		expectDesc = tt.coords1Desc

		if !result.IsCloseTo(expect) {
			t.Errorf("incorrect for %v to %v, got: %v, want: %v.", inputDesc, expectDesc, result, expect)
		}
	}
}

func TestViaConstructor(t *testing.T) {
	fromOne := NewAffineNoRotTransformation(100.0/4, 20.0/4, -5, 45)
	fromTwo := NewAffineNoRotTransformation(4.0/100, 4.0/20, -20*4.0/100+1, -50*4.0/20+1)

	testPointsInBothProjections(t, fromOne, fromTwo)
}

func TestViaTwoPoints(t *testing.T) {
	coords1 := []Point{PointXY(1, 1), PointXY(5, 5)}
	coords2 := []Point{PointXY(20, 50), PointXY(120, 70)}
	fromTwo, _ := NewAffineNoRotTransformationFromPoints(coords1, coords2)
	fromOne, _ := NewAffineNoRotTransformationFromPoints(coords2, coords1)

	testPointsInBothProjections(t, fromOne, fromTwo)
}
