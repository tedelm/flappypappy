package sprite

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed pub_painting.png
var pubPaintingPNG []byte

var (
	pubPaintingOnce  sync.Once
	pubPaintingImage *ebiten.Image
)

func ensurePubPaintingLoaded() {
	pubPaintingOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(pubPaintingPNG))
		if err != nil {
			panic(err)
		}
		pubPaintingImage = ebiten.NewImageFromImage(ToRGBA(img))
	})
}

func PubPainting() *ebiten.Image {
	ensurePubPaintingLoaded()
	return pubPaintingImage
}
