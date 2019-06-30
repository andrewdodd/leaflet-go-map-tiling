package mapimage

import (
	"gonum.org/v1/gonum/floats"
	"testing"
)

func TestConformalTransformationWithParams(t *testing.T) {
	k, theta, Tx, Ty := 1.25, RadFromDeg(-30), 164618.06, 1383016.332
	sut := NewConformalTransformation(k, theta, Tx, Ty)

	result := sut.Project(PointXY(100000, 200000))
	expect := PointEastingNorthing(397871.235, 1537022.683)

	if !floats.EqualWithinRel(result.Lat, expect.Lat, 0.00000001) ||
		!floats.EqualWithinRel(result.Lng, expect.Lng, 0.00000001) {
		t.Errorf("Projection incorrect, got: %v, want: %v.", result, expect)
	}
}

func TestBulkConformalTransformationWithParams(t *testing.T) {
	k, theta, Tx, Ty := 1.25, RadFromDeg(-30), 164618.06, 1383016.332
	sut := NewConformalTransformation(k, theta, Tx, Ty)

	results := sut.Projects(
		PointXY(100000, 200000),
		PointXY(104000, 204000),
		PointXY(104000, 200000),
		PointXY(100000, 204000),
	)

	for idx, expect := range []Point{
		PointEastingNorthing(397871.235, 1537022.683),
		PointEastingNorthing(404701.362, 1538852.81),
		PointEastingNorthing(402201.362, 1534522.683),
		PointEastingNorthing(400371.235, 1541352.8),
	} {

		result := results[idx]
		if !floats.EqualWithinRel(result.Lat, expect.Lat, 0.00000001) ||
			!floats.EqualWithinRel(result.Lng, expect.Lng, 0.00000001) {
			t.Errorf("Projection incorrect, got: %v, want: %v.", result, expect)
		}
	}
}

func TestBuildConformalTransformationFromPoints(t *testing.T) {
	standardPoints := []Point{
		PointEastingNorthing(397871.235, 1537022.683),
		PointEastingNorthing(404701.362, 1538852.81)}
	localPoints := []Point{
		PointXY(100000, 200000),
		PointXY(104000, 204000)}
	sut, _ := NewConformalTransformationFromPoints(
		standardPoints,
		localPoints,
	)

	result := sut.Project(PointXY(100000, 200000))
	expect := PointEastingNorthing(397871.235, 1537022.683)

	if !floats.EqualWithinRel(result.Lat, expect.Lat, 0.00000001) ||
		!floats.EqualWithinRel(result.Lng, expect.Lng, 0.00000001) {
		t.Errorf("Projection incorrect, got: %v, want: %v.", result, expect)
	}
}

func TestMineLevelProjectionsViaTheImplementation(t *testing.T) {

	minePoints := []Point{
		PointNorthingEasting(-4.072222, 137.136111),
		PointNorthingEasting(-4.080556, 137.141667)}

	imagePoints := []Point{
		PointNorthingEasting(8564.53042602539, 2377.5147705078125),
		PointNorthingEasting(1430.6239624023438, 7126.945892333984)}

	fromMap, err := NewConformalTransformationFromPoints(minePoints, imagePoints)
	if err != nil {
		panic(err)
	}
	fromImage, err := NewConformalTransformationFromPoints(imagePoints, minePoints)
	if err != nil {
		panic(err)
	}

	var pt Point
	var expect, result float64

	pt = fromImage.Project(PointEastingNorthing(137.136111, -4.072222))

	expect = 8564.53042602539
	result = pt.Lat
	if !floats.EqualWithinRel(result, expect, 0.00000001) {
		t.Errorf("LAT incorrect, expect %v, got %v", expect, result)
	}

	expect = 2377.5147705078125
	result = pt.Lng
	if !floats.EqualWithinRel(result, expect, 0.00001) {
		t.Errorf("LNG incorrect, expect %v, got %v", expect, result)
	}

	pt = fromMap.Project(PointEastingNorthing(0, 0))

	if !floats.EqualWithinRel(pt.Lat, -4.082232757389885, 0.00001) ||
		!floats.EqualWithinRel(pt.Lng, 137.1333390597705, 0.00001) {
		t.Errorf("Projection incorrect: %v", pt)
	}
}

/*
import { Transformation, MineLevelMapImageProjection } from '../projection'
import { toBeDeepCloseTo, toMatchCloseTo } from 'jest-matcher-deep-close-to'

expect.extend({ toBeDeepCloseTo, toMatchCloseTo })

const radFromDeg = degrees => degrees * Math.PI / 180

describe('Transformation', () => {
  const [k, theta, Tx, Ty] = [1.25, radFromDeg(-30), 164618.06, 1383016.332]
  describe('transformations when parameters are known (example 1)', () => {
    const sut = new Transformation(k, theta, Tx, Ty)

    it('calculates correctly if transformations parameters are known (example 1)', () => {
      const result = sut.transform([100000, 200000])
      expect(result[0]).toBeCloseTo(397871.235, 3)
      expect(result[1]).toBeCloseTo(1537022.683, 3)
    })

    it('calculates a list of points', () => {
      const inputs = [
        [100000, 200000],
        [104000, 204000],
        [104000, 200000],
        [100000, 204000]
      ]
      const expected = [
        [397871.235, 1537022.683],
        [404701.362, 1538852.81],
        [402201.362, 1534522.683],
        [400371.235, 1541352.81]
      ]

      const results = sut.transform(inputs)
      expect(results.length).toBe(4)
      results.forEach((result, idx) => {
        expect(result[0]).toBeCloseTo(expected[idx][0], 3)
        expect(result[1]).toBeCloseTo(expected[idx][1], 3)
      })
    })
  })

  describe('calculating transformation from four points', () => {
    it('calculates parameters from four points (example 2)', () => {
      const standardPts = [[397871.235, 1537022.683], [404701.362, 1538852.81]]
      const localPts = [[100000, 200000], [104000, 204000]]
      const result = Transformation.fromPoints(standardPts, localPts)

      localPts.forEach((pt, idx) => {
        result.transform(pt).forEach((coord, i) => {
          expect(coord).toBeCloseTo(standardPts[idx][i], 3)
        })
      })
    })
  })
})

describe('MineLevelMapImageProjection', () => {
  const mineCoords = [[0, 0], [457, 0]] // NB: using Northing, Easting format
  const imageCoords = [[-281, 217], [-69, 385]] // NB: using Northing, Easting format

  const sut = MineLevelMapImageProjection.fromPoints(mineCoords, imageCoords)

  it('projects mine coordinates to image pixels', () => {
    expect(sut.imageFromMine([0, 0])).toBeDeepCloseTo([-281, 217], 0)
    expect(sut.imageFromMine([457, 300])).toBeDeepCloseTo([-179, 524], 0)
  })
  it('projects image pixels to mine coordinates', () => {
    expect(sut.mineFromImage([-69, 385])).toBeDeepCloseTo([457, 0], 0)
    expect(sut.mineFromImage([-176, 522])).toBeDeepCloseTo([459, 293], 0)
  })
  it('handles objects with lat and lng properties', () => {
    const minePt = { lat: 457, lng: 300 }
    const imagePt = { lat: -179, lng: 524 }
    expect(sut.mineFromImage(imagePt)).toMatchCloseTo(minePt, 0)
    expect(sut.imageFromMine(minePt)).toMatchCloseTo(imagePt, 0)
  })

  it('handles list of points', () => {
    const minePoints = [[457, 0], [457, 300]]
    const imagePoints = [[-69, 385], [-179, 524]]
    expect(sut.mineFromImage(imagePoints)).toBeDeepCloseTo(minePoints, 0)
    expect(sut.imageFromMine(minePoints)).toBeDeepCloseTo(imagePoints, 0)
  })
  it('handles list of objects with lat and lng props', () => {
    const minePoints = [{ lat: 457, lng: 0 }, { lat: 457, lng: 300 }]
    const imagePoints = [{ lat: -69, lng: 385 }, { lat: -179, lng: 524 }]
    expect(sut.mineFromImage(imagePoints)).toMatchCloseTo(minePoints, 0)
    expect(sut.imageFromMine(minePoints)).toMatchCloseTo(imagePoints, 0)
  })
})
*/
