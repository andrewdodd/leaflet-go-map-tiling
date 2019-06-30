package mapimage

import (
	//	"log"
	"math"
)

// Based on globalmaptiles.py : https://gist.github.com/unorthodox123/5944793

type TileBounds struct {
	minX, minY, maxX, maxY float64
}

const tileSize int64 = 256

func Resolution(zoom int64) float64 {
	//"Resolution (meters/pixel) for given zoom level (measured at Equator)"
	// return (2 * math.pi * 6378137) / (self.tileSize * 2**zoom)
	initialResolution := 2.0 * math.Pi * 6378137.0 / float64(tileSize)
	return initialResolution / math.Pow(2, float64(zoom))
}

func PixelsToMeters(px, py, zoom int64) (float64, float64) {
	//"Converts pixel coordinates in given zoom level of pyramid to EPSG:900913"
	originShift := 2.0 * math.Pi * 6378137.0 / 2.0

	res := Resolution(zoom)
	mx := float64(px)*res - originShift
	my := float64(py)*res - originShift
	return mx, my
}

func (b *TileBounds) FromTile(tx, ty, zoom int64) {
	//"Returns bounds of the given tile in EPSG:900913 coordinates"
	b.minX, b.minY = PixelsToMeters(tx*tileSize, ty*tileSize, zoom)
	b.maxX, b.maxY = PixelsToMeters((tx+1)*tileSize, (ty+1)*tileSize, zoom)
}

func MetersToLatLon(mx, my float64) (float64, float64) {
	//"Converts XY point from Spherical Mercator EPSG:900913 to lat/lon in WGS84 Datum"

	originShift := 2.0 * math.Pi * 6378137.0 / 2.0
	lon := (mx / originShift) * 180.0
	lat := (my / originShift) * 180.0
	//log.Println("PRE lat = ", lat)

	lat = 180 / math.Pi * (2*math.Atan(math.Exp(lat*math.Pi/180.0)) - math.Pi/2.0)
	//log.Println("POST lat = ", lat)
	lat *= -1
	//lat =  -44.6484375
	//lat =  -40.71395582628603
	return lat, lon
}

func GoogleTile(tx, ty, tzoom int64) (x, y, zoom int64) {
	//"Converts TMS tile coordinates to Google Tile coordinates"

	// # coordinate origin is moved from bottom-left to top-left corner of the extent
	return tx, int64(math.Pow(2, float64(tzoom))) - 1 - ty, tzoom
}

func TileLatLonBounds(tx, ty, zoom int64) (a, b Point) {
	//"Returns bounds of the given tile in latutude/longitude using WGS84 datum"

	bounds := TileBounds{}
	bounds.FromTile(tx, ty, zoom)
	minLat, minLon := MetersToLatLon(bounds.minX, bounds.minY)
	maxLat, maxLon := MetersToLatLon(bounds.maxX, bounds.maxY)

	a = Point{Lat: minLat, Lng: minLon}
	b = Point{Lat: maxLat, Lng: maxLon}

	//	log.Println("TileLatLonBounds", a, b, c, d)

	return
}
