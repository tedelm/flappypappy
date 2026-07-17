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
	lampContentW float64
)

func ensureLampLoaded() {
	lampOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(lampPNG))
		if err != nil {
			panic(err)
		}
		rgba := KeyBlackTransparent(ToRGBA(img))
		lampFeetPad, lampContentH, lampContentW = MeasureRGBASize(rgba)
		if lampContentH == 0 {
			lampContentH = float64(rgba.Bounds().Dy())
		}
		if lampContentW == 0 {
			lampContentW = float64(rgba.Bounds().Dx())
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

func LampContentW() float64 {
	ensureLampLoaded()
	return lampContentW
}
