package sprite

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed wood_floor.png
var woodFloorPNG []byte

var (
	woodFloorOnce  sync.Once
	woodFloorImage *ebiten.Image
)

func ensureWoodFloorLoaded() {
	woodFloorOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(woodFloorPNG))
		if err != nil {
			panic(err)
		}
		woodFloorImage = ebiten.NewImageFromImage(ToRGBA(img))
	})
}

func WoodFloor() *ebiten.Image {
	ensureWoodFloorLoaded()
	return woodFloorImage
}
