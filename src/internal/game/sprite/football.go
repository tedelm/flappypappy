package sprite

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed football.png
var footballPNG []byte

var (
	footballOnce  sync.Once
	footballImage *ebiten.Image
)

func ensureFootballLoaded() {
	footballOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(footballPNG))
		if err != nil {
			panic(err)
		}
		footballImage = ebiten.NewImageFromImage(KeyBlackTransparent(ToRGBA(img)))
	})
}

func Football() *ebiten.Image {
	ensureFootballLoaded()
	return footballImage
}
