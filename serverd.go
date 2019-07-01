package main

import (
	//"encoding/json"
	"fmt"
	//"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	//"image"
	//_ "image/jpeg"
	//"image/png"
	//"io/ioutil"
	"errors"
	"log"
	"net/http"
	"os"
	//"strconv"
	"main/mapimage"
)

// var MapById map[string]ImageInfo
// var MapList []ImageInfo
//
// func init() {
// 	raw, err := ioutil.ReadFile("./ui/public/imageinfo.json")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	json.Unmarshal(raw, &MapList)
//
// 	MapById = make(map[string]ImageInfo, len(MapList))
// 	log.Printf("Found %d infos\n", len(MapList))
// 	for _, info := range MapList {
// 		path := fmt.Sprintf("./ui/public/%s", info.Filename)
// 		log.Println("Reading path:", path)
// 		reader, err := os.Open(path)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Println("Got reader:", reader)
// 		im, t, err := image.Decode(reader)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		log.Printf("Attaching image %v to %v\n", t, info)
// 		info.image = im
// 		info.Filename = path
// 		MapById[info.Id] = info
// 		reader.Close()
// 	}
// }

type db struct {
	images []mapimage.ImageInfo
}

var maps db

func init() {
	maps.images = make([]mapimage.ImageInfo, 0)
	f, err := os.Open(fmt.Sprintf("./ui/public/%s", "newyork.jpg"))
	if err != nil {
		fmt.Println(err)
	}
	// TOP LEFT   {lat: 40.98092266414473, lng: -74.0760523080826}
	// TOP RIGHT  {lat: 40.97884915992417, lng: -73.70595037937166}
	// BOT LEFT   {lat: 40.54070031600866, lng: -74.0774041414261}
	// BOT RIGHT  {lat: 40.53913486041921, lng: -73.70524227619173}
	maps.images = append(maps.images, mapimage.NewImageInfo(
		"new-york",
		"New York Street Map",
		[]mapimage.MapImagePair{
			mapimage.MapImagePair{
				Geographic: mapimage.LatLng{Lat: 40.981637441018464, Lng: -74.07707825303079},
				Pixel:      mapimage.LatLng{Lat: 0, Lng: 0},
			},
			mapimage.MapImagePair{
				Geographic: mapimage.LatLng{Lat: 40.537934245343585, Lng: -73.70349347591402},
				Pixel:      mapimage.LatLng{Lat: 10760, Lng: 7500},
			},
		},
		20, 0,
		f,
	))
	f, err = os.Open(fmt.Sprintf("./ui/public/%s", "dardanelles.jpg"))
	if err != nil {
		fmt.Println(err)
	}
	maps.images = append(maps.images, mapimage.NewImageInfo(
		"dardanelles",
		"Dardanelles",
		[]mapimage.MapImagePair{
			mapimage.MapImagePair{
				Geographic: mapimage.LatLng{Lat: 40.47835358455652, Lng: 26.12436711788178},
				Pixel:      mapimage.LatLng{Lat: 0, Lng: 0},
			},
			mapimage.MapImagePair{
				Geographic: mapimage.LatLng{Lat: 39.90604077881996, Lng: 26.666804687500004},
				Pixel:      mapimage.LatLng{Lat: 15328, Lng: 10507},
			},
		},
		20, 0,
		f,
	))
}

func (obj db) ListAll() []mapimage.MapImage {
	items := make([]mapimage.MapImage, 0)
	for idx, _ := range obj.images {
		items = append(items, &obj.images[idx])
	}
	return items
}

func (obj db) GetById(id string) (mapimage.MapImage, error) {
	for _, ii := range obj.images {
		if ii.Id() == id {
			return &ii, nil
		}
	}
	return nil, errors.New("not found")
}

// our main function
func main() {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	mapimage.AttachApi(maps, api, "/imageinfo", "/file")

	//router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "./ui/build/index.html") }))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./ui/build")))

	log.Fatal(http.ListenAndServe(":8000", router))
}
