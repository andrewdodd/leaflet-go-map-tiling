import React, { Component } from 'react'
import logo from './logo.svg'
import './App.css'

import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

import {
  Map,
  Marker,
  Popup,
  TileLayer,
  LayersControl,
  ImageOverlay
} from 'react-leaflet'

const maps = [
  {
    id: 'new-york',
    text: 'New York Street Map',
    image: './newyork.jpg',
    bounds: [
      { lat: 40.981637441018464, lng: -74.07707825303079 },
      { lat: 40.537934245343585, lng: -73.70349347591402 }
    ]
  },
  {
    id: 'dardanelles',
    text: 'Dardanelles',
    image: './dardanelles.jpg',
    bounds: [
      { lat: 40.47835358455652, lng: 26.12436711788178 },
      { lat: 39.90604077881996, lng: 26.666804687500004 }
    ]
  }
]

class TodoApp extends React.Component {
  constructor (props) {
    super(props)
    this.state = { mapId: 'dardanelles' }
    this.selectMap = this.selectMap.bind(this)
  }

  render () {
    const map = maps.filter(m => m.id === this.state.mapId)[0]
    return (
      <div className='App'>
        <header className='App-header'>
          <h1 className='App-title'>Welcome to TODO</h1>
        </header>
        <div>
          <ul>
            {maps.map(({ id, text }) => (
              <button key={id} onClick={() => this.selectMap(id)}>
                {text}
              </button>
            ))}
          </ul>
          <Map
            onClick={e => {
              console.log('clicked', e.latlng)
            }}
            bounds={map.bounds}
            zoom={12}
            style={{
              height: '800px',
              width: '100%',
              position: 'relative',
              zIndex: 0
            }}
          >
            <LayersControl>
              <LayersControl.Overlay name='Open Street Map' checked>
                <TileLayer url='https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png' />
              </LayersControl.Overlay>
              <LayersControl.Overlay name={map.text} id={map.id} checked>
                <ImageOverlay
                  url={map.image}
                  bounds={map.bounds}
                  opacity={0.5}
                />
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

class TodoList extends React.Component {
  render () {
    return (
      <ul>
        {this.props.items.map(item => <li key={item.id}>{item.text}</li>)}
      </ul>
    )
  }
}

export default TodoApp
