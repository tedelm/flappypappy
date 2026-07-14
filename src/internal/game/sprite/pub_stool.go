package sprite

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed pub_stool.png
var pubStoolPNG []byte

var (
	pubStoolOnce     sync.Once
	pubStoolImage    *ebiten.Image
	pubStoolFeetPad  float64
	pubStoolContentH float64
)

func ensurePubStoolLoaded() {
	pubStoolOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(pubStoolPNG))
		if err != nil {
			panic(err)
		}
		rgba := ToRGBA(img)
		pubStoolFeetPad, pubStoolContentH = MeasureRGBA(rgba)
		if pubStoolContentH == 0 {
			pubStoolContentH = float64(rgba.Bounds().Dy())
		}
		pubStoolImage = ebiten.NewImageFromImage(rgba)
	})
}

func PubStool() *ebiten.Image {
	ensurePubStoolLoaded()
	return pubStoolImage
}

func PubStoolFeetPad() float64 {
	ensurePubStoolLoaded()
	return pubStoolFeetPad
}

func PubStoolContentH() float64 {
	ensurePubStoolLoaded()
	return pubStoolContentH
}
