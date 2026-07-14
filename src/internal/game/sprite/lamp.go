package sprite

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed lamp.png
var lampPNG []byte

var (
	lampOnce     sync.Once
	lampImage    *ebiten.Image
	lampFeetPad  float64
	lampContentH float64
)

func ensureLampLoaded() {
	lampOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(lampPNG))
		if err != nil {
			panic(err)
		}
		rgba := ToRGBA(img)
		lampFeetPad, lampContentH = MeasureRGBA(rgba)
		if lampContentH == 0 {
			lampContentH = float64(rgba.Bounds().Dy())
		}
		lampImage = ebiten.NewImageFromImage(rgba)
	})
}

func Lamp() *ebiten.Image {
	ensureLampLoaded()
	return lampImage
}

func LampFeetPad() float64 {
	ensureLampLoaded()
	return lampFeetPad
}

func LampContentH() float64 {
	ensureLampLoaded()
	return lampContentH
}
