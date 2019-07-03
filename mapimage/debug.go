package mapimage

import (
	"bytes"
	"fmt"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/color/palette"
	"image/png"
	"io"
)

func addLabel(img *image.RGBA, x, y int, label string, col color.Color) {
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

func chooseColor(z, x, y int64) color.Color {
	idx := x % 128
	if y%2 == 1 {
		idx += 127
	}
	return palette.Plan9[idx]
}

func colorTile(zoom, x, y int64) io.ReadSeeker {
	tileSize := image.Rect(0, 0, 256, 256)
	img := image.NewRGBA(tileSize)
	col := chooseColor(zoom, x, y)
	draw.Draw(img, img.Bounds(), &image.Uniform{col}, image.ZP, draw.Src)

	a, b := TileLatLonBounds(x, y, zoom)

	addLabel(img, 20, 20, fmt.Sprintf("zoom=%v", zoom), color.Black)
	addLabel(img, 20, 40, fmt.Sprintf("x=%v, y=%v", x, y), color.Black)
	addLabel(img, 10, 60, fmt.Sprintf("a.lat=%v", a.Lat), color.Black)
	addLabel(img, 10, 80, fmt.Sprintf("a.lng=%v", a.Lng), color.Black)
	addLabel(img, 10, 100, fmt.Sprintf("b.lat=%v", b.Lat), color.Black)
	addLabel(img, 10, 120, fmt.Sprintf("b.lng=%v", b.Lng), color.Black)

	w := bytes.Buffer{}
	png.Encode(&w, img)
	return bytes.NewReader(w.Bytes())
}
