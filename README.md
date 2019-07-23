# Leafet Map Tiling In Go

This repo is really just to demo how to do on-the-fly map tiling in Go, such that they work with LeafletJS.

The repo consists of:

 * A basic UI, written in React, using React-Leaflet as the binding to LeafletJS
 * A basic server, written in Go, that serves a pre-build UI from the ./ui/build directory, reads in some config about map images, and then serves those images as either whole images or as a tile layer.


## Caveats

 * This is a demo repo only, your milage may vary
 * This has only been tested on OS X
 * The results are "good enough"...i.e. the map images are not totally aligned, especially over large areas...but the images are just nice looking ones I found on the internet, not GIS produced images etc etc.

# Running the code

Because that's what you want to do right...

 1. Install libvips
    
   ```
> brew install vips
```

 2. Setup the necessary exports. I needed:
    
   ```
> export PKG_CONFIG_PATH="/usr/local/opt/libffi/lib/pkgconfig"
> export CGO_CFLAGS_ALLOW="-Xpreprocessor"
```

 3. Run the server

   ```
   > go run serverd.go
   ```
 4. Go to [http://localhost:8000](http://localhost:8000) for the UI (:6060 if you want to poke the profiler)

 
**NB:** Depending on how you use go, you might need to install the dependencies (you can see them in the go.mod) file

## Also running the UI in dev mode

If you would also like to run the UI, I usually (in another terminal):

 1. Change to the ui directory

 ```
 > cd ./ui
 ```
 
 2. Run yarn to start the server (I guess you could also use NPM instead...or whatever JS build tool is current right now)
 
 ```
 > yarn start
 ```
 
 3. Go to [http://localhost:3000](http://localhost:3000) for the "dev" UI

**NB:** The ui is configured to proxy `/api` to [http://localhost:8000](http://localhost:8000). This allows it to get the actual image info from the server. If you see something like "Unexpected token P in JSON at position 0" in your browser it probably means your server is not running.

# Explanation

I made this because it was interesting. For more info, you should be able to see a blog post on it soon at [adodd.net/post](http://adodd.net/post).
