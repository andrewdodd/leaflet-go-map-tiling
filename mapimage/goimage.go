package mapimage

import (
	"bytes"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"os"
)

type goImage struct {
	id              string
	text            string
	minZoom         int
	maxZoom         int
	referencePoints []MapImagePair
	contents        *os.File
	image           image.Image
	toGeo           Transformation
	toPixel         Transformation
}

func NewImageInfo(
	id,
	text string,
	referencePoints []MapImagePair,
	contents *os.File) MapImage {
	image, _, err := image.Decode(contents)
	if err != nil {
		panic(0)
	}

	geo := []Point{referencePoints[0].Geographic.toPoint(), referencePoints[1].Geographic.toPoint()}
	pixel := []Point{referencePoints[0].Pixel.toPoint(), referencePoints[1].Pixel.toPoint()}
	toGeo, err := NewAffineNoRotTransformationFromPoints(geo, pixel) //, local)
	if err != nil {
		log.Println(err)
		panic(0)
	}

	toPixel, err := NewAffineNoRotTransformationFromPoints(pixel, geo) //, local)
	if err != nil {
		log.Println(err)
		panic(0)
	}

	i := goImage{
		id:              id,
		text:            text,
		minZoom:         0,
		maxZoom:         0,
		referencePoints: referencePoints,
		contents:        contents,
		image:           image,
		toGeo:           &toGeo,
		toPixel:         &toPixel,
	}

	i.minZoom = calculateMinZoom(&i)
	i.maxZoom = calculateMaxZoom(&i)

	return &i
}

func (i goImage) Id() string {
	return i.id
}

func (i goImage) Text() string {
	return i.text
}

func (i goImage) GeoBounds() [2]LatLng {
	imageSize := i.PixelBounds()

	min := i.GeoFromPixel(imageSize[0])
	max := i.GeoFromPixel(imageSize[1])

	return [2]LatLng{min, max}
}

func LatLngFromPoint(pt image.Point) LatLng {
	x := LatLng{Lat: float64(pt.Y), Lng: float64(pt.X)}
	return x
}

func (i goImage) PixelBounds() [2]LatLng {
	return [2]LatLng{
		LatLngFromPoint(i.image.Bounds().Min),
		LatLngFromPoint(i.image.Bounds().Max),
	}
}
func (i goImage) MinZoom() int {
	return i.minZoom
}
func (i goImage) MaxZoom() int {
	return i.maxZoom
}

func (i goImage) GeoFromPixel(p LatLng) LatLng {
	return LatLng(i.toGeo.Project(p.toPoint()))
}

func (i goImage) PixelFromGeo(p LatLng) LatLng {
	return LatLng(i.toPixel.Project(p.toPoint()))
}

func (ii *goImage) ImageContent() io.ReadSeeker {
	return ii.contents
}

func (ii goImage) MapTile(zoom, x, y int64) io.ReadSeeker {
	tileMin, tileMax := TileLatLonBounds(x, y, zoom)
	pxlMin := ii.PixelFromGeo(LatLng(tileMin))
	pxlMax := ii.PixelFromGeo(LatLng(tileMax))

	tileSize := image.Rect(0, 0, 256, 256)
	img := image.NewRGBA(tileSize)
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	tileRect := image.Rect(
		int(math.Round(pxlMax.Lng)),
		int(math.Round(pxlMax.Lat)),
		int(math.Round(pxlMin.Lng)),
		int(math.Round(pxlMin.Lat)),
	)

	imgBounds := ii.image.Bounds()
	if imgBounds.Overlaps(tileRect) {
		srcRect := image.Rect(
			max(tileRect.Min.X, imgBounds.Min.X),
			max(tileRect.Min.Y, imgBounds.Min.Y),
			min(tileRect.Max.X, imgBounds.Max.X),
			min(tileRect.Max.Y, imgBounds.Max.Y),
		)

		dstRect := tileSize
		if srcRect.Max.X != tileRect.Max.X {
			// Reduce the right hand side of dstRect by the same ratio
			dstRect.Max.X -= int(float64(tileSize.Dx()) * (float64(tileRect.Max.X-srcRect.Max.X) / float64(tileRect.Dx())))
		}

		if srcRect.Min.X != tileRect.Min.X {
			// Increase the left hand side of dstRect by the same ratio
			dstRect.Min.X += int(float64(tileSize.Dx()) * (float64(srcRect.Min.X-tileRect.Min.X) / float64(tileRect.Dx())))
		}

		if srcRect.Max.Y != tileRect.Max.Y {
			dstRect.Max.Y -= int(float64(tileSize.Dy()) * (float64(tileRect.Max.Y-srcRect.Max.Y) / float64(tileRect.Dy())))
		}

		if srcRect.Min.Y != tileRect.Min.Y {
			dstRect.Min.Y += int(float64(tileSize.Dy()) * (float64(srcRect.Min.Y-tileRect.Min.Y) / float64(tileRect.Dy())))
		}

		//scaler := draw.BiLinear
		//scaler := draw.NearestNeighbor
		scaler := draw.ApproxBiLinear
		scaler.Scale(img, dstRect, ii.image, srcRect, draw.Over, nil)
	}

	w := bytes.Buffer{}
	png.Encode(&w, img)
	return bytes.NewReader(w.Bytes())
}
