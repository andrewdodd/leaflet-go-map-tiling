import React from 'react'
import Control from 'react-leaflet-control'
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
  state = { mapId: null, mapImages: [], opacity: 0.7}
  constructor (props) {
    super(props)
    this.selectMap = this.selectMap.bind(this)
    this.updateOpacity = this.updateOpacity.bind(this)
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
    const { mapImages, mapId , opacity } = this.state
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
          <h1 className='App-title'>Leaflet Map Tiling in Go</h1>
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
            key={mapId} // force minZoom/maxZoom etc to be reset by forcing React to create a new map (NB: not very efficient)
            bounds={mapImage.geo_bounds}
            maxBounds={mapImage.geo_bounds}
            // Allow 2 extra zooms, and prevent the last 2 extra zoom outs
            minZoom={mapImage.minZoom + 2}
            maxZoom={mapImage.maxZoom + 2}
            style={{
              height: '800px',
              width: '100%',
              position: 'relative',
              zIndex: 0
            }}
          >
            <LayersControl>
              <LayersControl.Overlay
                name={'Open Street Map'}
                checked
                key={'open-street-map'}
              >
                <TileLayer url='https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png' />
              </LayersControl.Overlay>

              <LayersControl.Overlay
                name={'Single Image'}
                key={'single-image'}
              >
                <ImageOverlay
                  url={mapImage.image}
                  bounds={mapImage.geo_bounds}
                  opacity={opacity}
                />
              </LayersControl.Overlay>

              <LayersControl.Overlay
                name={'Tiled'}
                key={'tiled'}
                checked
              >
                <TileLayer url={mapImage.tiled} tms={mapImage.tiled.indexOf("tms") !== -1} opacity={opacity}/>
              </LayersControl.Overlay>
              <Control>
                <span style={{backgroundColor: 'white'}}>
                  Opacity: <input
                    type="range"
                    value={opacity}
                    min={0.0}
                    max={1.0}
                    onChange={this.updateOpacity}
                    name="myslider"
                    step={0.01}
                  />
                </span>
              </Control>
            </LayersControl>
          </Map>
        </div>
      </div>
    )
  }

  selectMap (e) {
    this.setState({ mapId: e })
  }

  updateOpacity (e) {
    this.setState({ opacity: e.target.value })
  }
}

export default TodoApp
