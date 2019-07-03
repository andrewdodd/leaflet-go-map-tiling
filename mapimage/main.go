package mapimage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type MapImage interface {
	ImageContent() io.ReadSeeker
	MapTile(zoom, x, y int64) io.ReadSeeker
	Id() string
	Text() string
	GeoBounds() [2]LatLng
	PixelBounds() [2]LatLng
	MinZoom() int
	MaxZoom() int
	//ReferencePoints() []MapImagePair
}

type MapImagesSource interface {
	GetById(id string) (MapImage, error)
	ListAll() []MapImage
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type MapImagePair struct {
	Geographic LatLng `json:"geo"`
	Pixel      LatLng `json:"pixel"`
}

func (ll LatLng) toPoint() Point {
	return Point{Lat: ll.Lat, Lng: ll.Lng}
}

type ImageInfo struct {
	id      string
	text    string
	minZoom int
	maxZoom int
	//referencePoints []MapImagePair

	contents *os.File
	image    image.Image
	toGeo    AffineNoRotTransformation
	toPixel  AffineNoRotTransformation
}

func NewImageInfo(
	id,
	text string,
	referencePoints []MapImagePair,
	minZoom int,
	maxZoom int,
	contents *os.File) MapImage {
	//image, err := png.Decode(bytes.NewReader(fileContents))
	image, _, err := image.Decode(contents)
	if err != nil {
		panic(0)
	}

	geo := []Point{referencePoints[0].Geographic.toPoint(), referencePoints[1].Geographic.toPoint()}
	pixel := []Point{referencePoints[0].Pixel.toPoint(), referencePoints[1].Pixel.toPoint()}
	log.Println("geo", geo)
	log.Println("pixel", pixel)
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

	return &ImageInfo{
		id:      id,
		text:    text,
		minZoom: minZoom,
		maxZoom: maxZoom,
		//referencePoints: referencePoints,
		contents: contents,
		image:    image,
		toGeo:    toGeo,
		toPixel:  toPixel,
	}

}

func (i ImageInfo) Id() string {
	return i.id
}
func (i ImageInfo) Text() string {
	return i.text
}
func (i ImageInfo) GeoBounds() [2]LatLng {
	//imageSize := ml.Image.image.Bounds()
	imageSize := i.PixelBounds()

	min := i.GeoFromPixel(imageSize[0])
	max := i.GeoFromPixel(imageSize[1])

	return [2]LatLng{min, max}
}

func LatLngFromPoint(pt image.Point) LatLng {
	x := LatLng{Lat: float64(pt.Y), Lng: float64(pt.X)}
	return x
}

func (i ImageInfo) PixelBounds() [2]LatLng {
	return [2]LatLng{
		LatLngFromPoint(i.image.Bounds().Min),
		LatLngFromPoint(i.image.Bounds().Max),
	}
}
func (i ImageInfo) MinZoom() int {
	return i.minZoom
}
func (i ImageInfo) MaxZoom() int {
	return i.maxZoom
}

func (i ImageInfo) GeoFromPixel(p LatLng) LatLng {
	return LatLng(i.toGeo.Project(p.toPoint()))
}

func (i ImageInfo) PixelFromGeo(p LatLng) LatLng {
	return LatLng(i.toPixel.Project(p.toPoint()))
}

func (ii *ImageInfo) ImageContent() io.ReadSeeker {
	return ii.contents
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (ii ImageInfo) MapTile(zoom, x, y int64) io.ReadSeeker {

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

type ApiRepresentation struct {
	Id          string    `json:"id"`
	Text        string    `json:"text"`
	GeoBounds   [2]LatLng `json:"geo_bounds"`
	PixelBounds [2]LatLng `json:"pixel_bounds"`
	MinZoom     int       `json:"minZoom"`
	MaxZoom     int       `json:"maxZoom"`
	//ReferencePoints []MapImagePair `json:"referencePoints"`

	Image string `json:"image"`
	Tiled string `json:"tiled"`
}

func ToApi(imagePathBase string, i MapImage) ApiRepresentation {
	s := ApiRepresentation{
		Id:          i.Id(),
		Text:        i.Text(),
		Image:       fmt.Sprintf("api%s/raw/%s", imagePathBase, i.Id()),
		Tiled:       fmt.Sprintf("api%s/tms/%s/{z}/{x}/{y}", imagePathBase, i.Id()),
		GeoBounds:   i.GeoBounds(),
		PixelBounds: i.PixelBounds(),
		MinZoom:     i.MinZoom(),
		MaxZoom:     i.MaxZoom(),
		//ReferencePoints: i.ReferencePoints(),
	}

	return s
}

func AttachApi(source MapImagesSource, router *mux.Router, infoPath, imagePathBase string) {

	router.Handle(infoPath, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			items := make([]ApiRepresentation, 0)
			for _, i := range source.ListAll() {
				items = append(items, ToApi(imagePathBase, i))
			}
			b, err := json.Marshal(items)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		}))

	router.Handle(
		fmt.Sprintf("%s/{id}", infoPath), http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				vars := mux.Vars(r)
				id := strings.TrimSpace(vars["id"])
				if id == "" {
					http.Error(w, "empty id supplied", http.StatusBadRequest)
					return
				}

				if ii, err := source.GetById(id); err == nil {
					b, err := json.Marshal(ToApi(imagePathBase, ii))
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write(b)
				} else {
					http.Error(w, "Not found", http.StatusNotFound)
				}
			}))

	router.Handle(
		fmt.Sprintf("%s/raw/{id}", imagePathBase), http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				vars := mux.Vars(r)
				id := strings.TrimSpace(vars["id"])
				if id == "" {
					http.Error(w, "empty id supplied", http.StatusBadRequest)
					return
				}

				if ii, err := source.GetById(id); err == nil {
					http.ServeContent(w, r, "temp.png", time.Time{}, ii.ImageContent())
					//http.ServeFile(w, r, fmt.Sprintf("./ui/public/%s", ii.Image))
				} else {
					http.Error(w, "Not found", http.StatusNotFound)
				}
			}))

	router.Handle(
		fmt.Sprintf("%s/{tileFmt}/{id}/{z}/{x}/{y}", imagePathBase), http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				vars := mux.Vars(r)
				id := strings.TrimSpace(vars["id"])
				tileFmt := strings.TrimSpace(vars["tileFmt"])
				if id == "" {
					http.Error(w, "empty id supplied", http.StatusBadRequest)
					return
				}
				ii, err := source.GetById(id)
				if err != nil {
					http.Error(w, "Not found", http.StatusNotFound)
					return
				}

				zoom, _ := strconv.ParseInt(vars["z"], 10, 64)
				x, _ := strconv.ParseInt(vars["x"], 10, 64)
				y, _ := strconv.ParseInt(vars["y"], 10, 64)

				if tileFmt == "tms" {
					x, y, zoom = GoogleTile(x, y, zoom)
				}

				//tile := colorTile(zoom, x, y)
				tile := ii.MapTile(zoom, x, y)

				w.Header().Set("Expires", "Sun, 17 Jan 2038 19:14:07 GMT")
				w.Header().Set("Content-Type", "image/png")
				http.ServeContent(w, r, "huh.png", time.Time{}, tile)
			}))
}
