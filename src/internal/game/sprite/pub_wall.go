package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

const NumPubWalls = 6

//go:embed pub_wall_sprite.png
var pubWallSpritePNG []byte

var pubWallRects = []SpriteRect{
	{0, 0, 128, 128},
	{128, 0, 128, 128},
	{256, 0, 128, 128},
	{384, 0, 128, 128},
	{512, 0, 128, 128},
	{640, 0, 128, 128},
}

var (
	pubWallOnce   sync.Once
	pubWallImages []*ebiten.Image
)

func ensurePubWallsLoaded() {
	pubWallOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(pubWallSpritePNG))
		if err != nil {
			panic(err)
		}
		sheet := ToRGBA(img)
		pubWallImages = make([]*ebiten.Image, NumPubWalls)
		for i, r := range pubWallRects {
			rect := pubWallImageRect(r)
			pubWallImages[i] = ebiten.NewImageFromImage(cropRGBA(sheet, rect))
		}
	})
}

func pubWallImageRect(r SpriteRect) image.Rectangle {
	return image.Rect(r.X, r.Y, r.X+r.W, r.Y+r.H)
}

func PubWall(index int) *ebiten.Image {
	ensurePubWallsLoaded()
	if index < 0 {
		index = 0
	}
	if index >= NumPubWalls {
		index = NumPubWalls - 1
	}
	return pubWallImages[index]
}
