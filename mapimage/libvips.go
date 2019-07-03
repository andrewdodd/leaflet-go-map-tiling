package mapimage

import (
	"bytes"
	"fmt"
	"github.com/h2non/bimg"
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

type libvipsImage struct {
	id      string
	text    string
	minZoom int
	maxZoom int

	contents    *os.File
	fileBuf     []byte
	imageConfig image.Config
	imageFormat string
	toGeo       Transformation
	toPixel     Transformation
}

func NewVIPSImageInfo(
	id,
	text string,
	referencePoints []MapImagePair,
	minZoom int,
	maxZoom int,
	contents *os.File) MapImage {
	//image, err := png.Decode(bytes.NewReader(fileContents))
	imageConfig, format, err := image.DecodeConfig(contents)
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

	contents.Seek(0, 0)
	buf, err := ioutil.ReadAll(contents)
	if err != nil {
		panic(0)
	}

	return &libvipsImage{
		id:      id,
		text:    text,
		minZoom: minZoom,
		maxZoom: maxZoom,
		//referencePoints: referencePoints,
		contents:    contents,
		fileBuf:     buf,
		imageConfig: imageConfig,
		imageFormat: format,
		toGeo:       &toGeo,
		toPixel:     &toPixel,
	}

}

func (i libvipsImage) Id() string {
	return i.id
}
func (i libvipsImage) Text() string {
	return i.text
}
func (i libvipsImage) GeoBounds() [2]LatLng {
	imageSize := i.PixelBounds()

	min := i.GeoFromPixel(imageSize[0])
	max := i.GeoFromPixel(imageSize[1])

	return [2]LatLng{min, max}
}

func (i libvipsImage) PixelBounds() [2]LatLng {
	return [2]LatLng{LatLng{}, LatLng{
		Lat: float64(i.imageConfig.Height),
		Lng: float64(i.imageConfig.Width),
	},
	}
}

func (i libvipsImage) MinZoom() int {
	return i.minZoom
}
func (i libvipsImage) MaxZoom() int {
	return i.maxZoom
}

func (i libvipsImage) GeoFromPixel(p LatLng) LatLng {
	return LatLng(i.toGeo.Project(p.toPoint()))
}

func (i libvipsImage) PixelFromGeo(p LatLng) LatLng {
	return LatLng(i.toPixel.Project(p.toPoint()))
}

func (ii *libvipsImage) ImageContent() io.ReadSeeker {
	return ii.contents
}

func (ii libvipsImage) MapTile(zoom, x, y int64) io.ReadSeeker {
	tileMin, tileMax := TileLatLonBounds(x, y, zoom)
	pxlMin := ii.PixelFromGeo(LatLng(tileMin))
	pxlMax := ii.PixelFromGeo(LatLng(tileMax))

	tileSize := image.Rect(0, 0, 256, 256)
	tileRect := image.Rect(
		int(math.Round(pxlMax.Lng)),
		int(math.Round(pxlMax.Lat)),
		int(math.Round(pxlMin.Lng)),
		int(math.Round(pxlMin.Lat)),
	)

	imgBounds := image.Rect(0, 0, ii.imageConfig.Width, ii.imageConfig.Height)
	if !imgBounds.Overlaps(tileRect) {
		img := image.NewRGBA(tileSize)
		draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)
		w := bytes.Buffer{}
		png.Encode(&w, img)
		return bytes.NewReader(w.Bytes())
	}

	// Check if file already exists
	path := fmt.Sprintf("./media/%s/%d/%d", ii.id, zoom, x)
	filename := fmt.Sprintf("%s/%d", path, y)
	if buf, err := ioutil.ReadFile(filename); err == nil {
		return bytes.NewReader(buf)
	}

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

	imgObj := bimg.NewImage(ii.fileBuf)
	_, err := imgObj.Extract(srcRect.Min.Y, srcRect.Min.X, srcRect.Dx(), srcRect.Dy())

	if err != nil {
		log.Println("extract image", srcRect, err)
		return colorTile(zoom, x, y)
	}

	newImage, err := imgObj.ForceResize(dstRect.Dx(), dstRect.Dy())
	if err != nil {
		log.Println("resize image", err)
		return colorTile(zoom, x, y)
	}

	srcImage, _, err := image.Decode(bytes.NewReader(newImage))
	if err != nil {
		log.Println("decode image", err)
		return colorTile(zoom, x, y)
	}

	img := image.NewRGBA(tileSize)
	draw.Draw(img, img.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

	scaler := draw.ApproxBiLinear
	scaler.Scale(img, dstRect, srcImage, srcImage.Bounds(), draw.Over, nil)

	err = os.MkdirAll(path, 0777)
	if err != nil {
		log.Println("path", err)
		return colorTile(zoom, x, y)
	}

	w := bytes.Buffer{}
	png.Encode(&w, img)

	ioutil.WriteFile(filename, w.Bytes(), 0777)

	return bytes.NewReader(w.Bytes())
}
