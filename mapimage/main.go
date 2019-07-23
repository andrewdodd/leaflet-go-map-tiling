package mapimage

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"math"
	"net/http"
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
	// The zoom where whole image is on a single tile?
	MinZoom() int
	// The zoom where image is being stretched by more than half?
	MaxZoom() int
	//ReferencePoints() []MapImagePair
	GeoFromPixel(p LatLng) LatLng
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

func calculateMinZoom(i MapImage) int {
	// The zoom where whole image is on a single tile?
	geoBounds := i.GeoBounds()
	geoBoundsMin := geoBounds[0]
	geoBoundsMax := geoBounds[1]

	for z := 20; z > 0; z-- {
		minX, minY := LatLonToPixels(geoBoundsMin.Lat, geoBoundsMin.Lng, z)
		maxX, maxY := LatLonToPixels(geoBoundsMax.Lat, geoBoundsMax.Lng, z)
		if math.Abs(minX-maxX) < 256 {
			return z
		}
		if math.Abs(minY-maxY) < 256 {
			return z
		}
	}

	return 0
}

func calculateMaxZoom(i MapImage) int {
	// The zoom where tiles start to stretch (i.e. pixel density limit)
	singleTileMin := i.GeoFromPixel(LatLng{Lat: 0, Lng: 0})
	singleTileMax := i.GeoFromPixel(LatLng{Lat: 256, Lng: 256})

	for z := 0; z < 21; z++ {
		minX, minY := LatLonToPixels(singleTileMin.Lat, singleTileMin.Lng, z)
		maxX, maxY := LatLonToPixels(singleTileMax.Lat, singleTileMax.Lng, z)
		if math.Abs(minX-maxX) > 256 {
			return z
		}
		if math.Abs(minY-maxY) > 256 {
			return z
		}
	}

	return 20
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
					w.Header().Set("Expires", "Sun, 17 Jan 2038 19:14:07 GMT")
					http.ServeContent(w, r, id, time.Time{}, ii.ImageContent())
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
