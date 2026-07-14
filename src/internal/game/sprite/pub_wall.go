package sprite

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed pub_wall.png
var pubWallPNG []byte

var (
	pubWallOnce  sync.Once
	pubWallImage *ebiten.Image
)

func ensurePubWallLoaded() {
	pubWallOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(pubWallPNG))
		if err != nil {
			panic(err)
		}
		pubWallImage = ebiten.NewImageFromImage(ToRGBA(img))
	})
}

func PubWall() *ebiten.Image {
	ensurePubWallLoaded()
	return pubWallImage
}
