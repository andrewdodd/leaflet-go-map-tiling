import React from 'react'
import './App.css'
import 'leaflet/dist/leaflet.css'
import {ImageOverlay, LayersControl, Map, TileLayer} from 'react-leaflet'

// { "id":,
//   "bounds": [{lat:,lng}, {lat:lng}],
//   "name":,
//   "minZoom",
//   "maxZoom",
//   "referencePoints": [...],
//   "image":
//   "tileed":
//   }

// const maps = [
//   {
//     id: 'new-york',
//     text: 'New York Street Map',
//     image: './newyork.jpg',
//     bounds: [
//       { lat: 40.981637441018464, lng: -74.07707825303079 },
//       { lat: 40.537934245343585, lng: -73.70349347591402 }
//     ]
//   },
//   {
//     id: 'dardanelles',
//     text: 'Dardanelles',
//     image: './dardanelles.jpg',
//     bounds: [
//       { lat: 40.47835358455652, lng: 26.12436711788178 },
//       { lat: 39.90604077881996, lng: 26.666804687500004 }
//     ]
//   }
// ]
// 
// class TodoApp extends React.Component {
//   constructor (props) {
//     super(props)
//     this.state = { mapId: 'dardanelles' }
//     this.selectMap = this.selectMap.bind(this)
//   }
// 
//   render () {
//     const map = maps.filter(m => m.id === this.state.mapId)[0]

class TodoApp extends React.Component {
  state = { mapId: null, mapImages: [] }
  constructor (props) {
    super(props)
    this.selectMap = this.selectMap.bind(this)
  }

  componentDidMount () {
    var that = this
    fetch('/api/imageinfo')
      .then(function (response) {
        return response.json()
      })
      .then(function (mapImages) {
        that.setState({ mapId: mapImages[0].id, mapImages })
      })
  }

  render () {
    const { mapImages, mapId } = this.state
    const mapImage = mapImages.filter(m => m.id === mapId)[0]
    if (!mapImage) {
      return (
        <div className='App'>
          <header className='App-header'>
            <h1 className='App-title'>Loading...</h1>
          </header>
        </div>
      )
    }
    return (
      <div className='App'>
        <header className='App-header'>
          <h1 className='App-title'>Welcome to TODO</h1>
        </header>
        <div>
          <ul>
            {mapImages.map(({ id, text }) => (
              <button key={id} onClick={() => this.selectMap(id)}>
                {text}
              </button>
            ))}
          </ul>
          <Map
            onClick={e => {
              console.log('clicked', e.latlng)
            }}
            bounds={mapImage.geo_bounds}
            zoom={12}
            style={{
              height: '800px',
              width: '100%',
              position: 'relative',
              zIndex: 0
            }}
          >
            <LayersControl>
              <LayersControl.Overlay
                name='Open Street Map'
                checked
                key='open-street-map'
              >
                <TileLayer url='https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png' />
              </LayersControl.Overlay>
              <LayersControl.Overlay
                name='Single Image'
                id={mapImage.id}
                checked
                key='single-image'
              >
                <ImageOverlay
                  url={mapImage.image}
                  bounds={mapImage.geo_bounds}
                />
              </LayersControl.Overlay>
              <LayersControl.Overlay
                name='Tiled'
                key='tiled'
              >
                <TileLayer url={mapImage.tiled} tms={mapImage.tiled.indexOf("tms") !== -1} />
              </LayersControl.Overlay>
            </LayersControl>
          </Map>
        </div>
      </div>
    )
  }

  selectMap (e) {
    this.setState({ mapId: e })
  }
}

export default TodoApp
