package mapimage

import (
	"bytes"
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
)

type cached struct {
	mi MapImage
}

func FilesystemCachedImage(mi MapImage) MapImage {
	return &cached{mi: mi}
}

func (i cached) Id() string {
	return i.mi.Id()
}
func (i cached) Text() string {
	return i.mi.Text()
}
func (i cached) GeoBounds() [2]LatLng {
	return i.mi.GeoBounds()
}

func (i cached) PixelBounds() [2]LatLng {
	return i.mi.PixelBounds()
}

func (i cached) MinZoom() int {
	return i.mi.MinZoom()
}
func (i cached) MaxZoom() int {
	return i.mi.MaxZoom()
}

func (i cached) GeoFromPixel(p LatLng) LatLng {
	return i.mi.GeoFromPixel(p)
}

func (i cached) PixelFromGeo(p LatLng) LatLng {
	return i.mi.PixelFromGeo(p)
}

func (i cached) ImageContent() io.ReadSeeker {
	return i.mi.ImageContent()
}

func (i cached) MapTile(zoom, x, y int64) io.ReadSeeker {
	tileMin, tileMax := TileLatLonBounds(x, y, zoom)
	pxlMin := i.mi.PixelFromGeo(LatLng(tileMin))
	pxlMax := i.mi.PixelFromGeo(LatLng(tileMax))

	tileSize := image.Rect(0, 0, 256, 256)
	tileRect := image.Rect(
		int(math.Round(pxlMax.Lng)),
		int(math.Round(pxlMax.Lat)),
		int(math.Round(pxlMin.Lng)),
		int(math.Round(pxlMin.Lat)),
	)

	pixelBounds := i.mi.PixelBounds()
	imgBounds := image.Rect(
		int(pixelBounds[0].Lng), int(pixelBounds[0].Lat),
		int(pixelBounds[1].Lng), int(pixelBounds[1].Lat),
	)
	// If the requested area is not inside the map image,
	// then just return a black square from ram
	if !imgBounds.Overlaps(tileRect) {
		img := image.NewRGBA(tileSize)
		draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)
		w := bytes.Buffer{}
		png.Encode(&w, img)
		return bytes.NewReader(w.Bytes())
	}

	// Check if file already exists
	path := fmt.Sprintf("./media/%s/%d/%d", i.mi.Id(), zoom, x)
	filename := fmt.Sprintf("%s/%d", path, y)
	if buf, err := ioutil.ReadFile(filename); err == nil {
		return bytes.NewReader(buf)
	}

	// Produce the image with the underlying MapImage implementation
	img := i.mi.MapTile(zoom, x, y)

	err := os.MkdirAll(path, 0777)
	if err != nil {
		log.Println("path", err)
		return colorTile(zoom, x, y)
	}

	// Pay the cost of putting on the filesystem now
	w := bytes.Buffer{}
	w.ReadFrom(img)

	ioutil.WriteFile(filename, w.Bytes(), 0777)

	return bytes.NewReader(w.Bytes())
}
