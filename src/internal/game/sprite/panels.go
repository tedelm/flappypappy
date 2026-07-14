package sprite

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed panels.png
var panelsPNG []byte

var (
	panelsOnce  sync.Once
	panelsImage *ebiten.Image
)

func ensurePanelsLoaded() {
	panelsOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(panelsPNG))
		if err != nil {
			panic(err)
		}
		panelsImage = ebiten.NewImageFromImage(ToRGBA(img))
	})
}

func Panels() *ebiten.Image {
	ensurePanelsLoaded()
	return panelsImage
}
