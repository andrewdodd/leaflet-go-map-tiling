package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"main/mapimage"
	"net/http"
	_ "net/http/pprof"
	"os"
)

type db struct {
	images []mapimage.MapImage
}

var maps db

type ImageConfig struct {
	Id              string                  `json:"id"`
	Name            string                  `json:"name"`
	ReferencePoints []mapimage.MapImagePair `json:"referencePoints"`
	Filename        string                  `json:"filename"`
}

func DisallowUnknownFields(d *json.Decoder) *json.Decoder {
	d.DisallowUnknownFields()
	return d
}

func init() {
	config, err := os.Open("./images/config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer config.Close()

	buf, err := ioutil.ReadAll(config)
	if err != nil {
		log.Fatal(err)
	}

	var loadedImages []ImageConfig
	err = yaml.UnmarshalStrict(buf, &loadedImages, DisallowUnknownFields)
	if err != nil {
		log.Fatal(err)
	}

	maps.images = make([]mapimage.MapImage, 0)
	for _, loadedImage := range loadedImages {
		f, err := os.Open(fmt.Sprintf("./images/%s", loadedImage.Filename))
		if err == nil {
			if fileConfig, format, err := image.DecodeConfig(f); err == nil {
				f.Seek(0, 0)
				approxSize := (3 * fileConfig.Width * fileConfig.Height) / 1024 / 1024

				log.Printf("%v is approx %v MB, in format %v\n", loadedImage.Filename, approxSize, format)
				var mi mapimage.MapImage
				if approxSize > 1000 {
					log.Printf("Using VIPS for %v\n", loadedImage.Filename)
					mi = mapimage.NewVIPSImageInfo(
						loadedImage.Id,
						loadedImage.Name,
						loadedImage.ReferencePoints,
						20, 0,
						f)

				} else {
					log.Printf("Using Go Image for %v\n", loadedImage.Filename)
					mi = mapimage.NewImageInfo(
						loadedImage.Id,
						loadedImage.Name,
						loadedImage.ReferencePoints,
						20, 0,
						f)
					log.Printf(" >> MinZoom: %v MaxZoom: %v\n", mi.MinZoom(), mi.MaxZoom())
				}
				maps.images = append(maps.images, mi)
			}
		} else {
			f.Close()
		}
	}

	log.Println("Running")
}

func (obj db) ListAll() []mapimage.MapImage {
	items := make([]mapimage.MapImage, 0)
	for idx, _ := range obj.images {
		items = append(items, obj.images[idx])
	}
	return items
}

func (obj db) GetById(id string) (mapimage.MapImage, error) {
	for _, ii := range obj.images {
		if ii.Id() == id {
			return ii, nil
		}
	}
	return nil, errors.New("not found")
}

// our main function
func main() {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	mapimage.AttachApi(maps, api, "/imageinfo", "/file")

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./ui/build")))

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	log.Fatal(http.ListenAndServe(":8000", router))

}
