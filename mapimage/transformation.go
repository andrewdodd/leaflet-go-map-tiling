package mapimage

import (
	"fmt"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
	"log"
	"math"
)

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (p Point) String() string {
	return fmt.Sprintf("{lat: %v, lng:%v}", p.Lat, p.Lng)
}

func PointXY(x, y float64) Point {
	return Point{Lat: y, Lng: x}
}

func (a Point) IsCloseTo(b Point) bool {
	tol := 0.0000001
	return floats.EqualWithinAbs(a.Lat, b.Lat, tol) && floats.EqualWithinAbs(a.Lng, b.Lng, tol)
}

func PointEastingNorthing(easting, northing float64) Point {
	return Point{Lat: northing, Lng: easting}
}

func PointNorthingEasting(northing, easting float64) Point {
	return Point{Lat: northing, Lng: easting}
}

func RadFromDeg(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

func DegFromRad(radians float64) float64 {
	return radians * (180 / math.Pi)
}

func printMat(name string, m *mat.Dense) {
	fc := mat.Formatted(m, mat.Prefix(" "), mat.Squeeze())
	log.Printf("%v =\n %v", name, fc)
}
