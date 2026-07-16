package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

const NumPubPaintings = 12

//go:embed pub_paintings.png
var pubPaintingsPNG []byte

var Paintings = map[string]SpriteRect{
	"painting_1":  {0, 0, 170, 212},
	"painting_2":  {180, 0, 190, 150},
	"painting_3":  {380, 0, 190, 150},
	"painting_4":  {580, 0, 170, 150},
	"painting_5":  {0, 225, 170, 122},
	"painting_6":  {180, 164, 190, 160},
	"painting_7":  {380, 164, 190, 160},
	"painting_8":  {580, 164, 170, 160},
	"painting_9":  {0, 360, 170, 122},
	"painting_10": {180, 335, 190, 150},
	"painting_11": {380, 335, 190, 150},
	"painting_12": {580, 335, 170, 145},
}

var paintingKeys = []string{
	"painting_1",
	"painting_2",
	"painting_3",
	"painting_4",
	"painting_5",
	"painting_6",
	"painting_7",
	"painting_8",
	"painting_9",
	"painting_10",
	"painting_11",
	"painting_12",
}

var (
	pubPaintingsOnce  sync.Once
	pubPaintingImages []*ebiten.Image
)

func ensurePubPaintingsLoaded() {
	pubPaintingsOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(pubPaintingsPNG))
		if err != nil {
			panic(err)
		}
		sheet := ToRGBA(img)
		pubPaintingImages = make([]*ebiten.Image, NumPubPaintings)
		for i, key := range paintingKeys {
			rect := paintingImageRect(Paintings[key])
			pubPaintingImages[i] = ebiten.NewImageFromImage(cropRGBA(sheet, rect))
		}
	})
}

func paintingImageRect(r SpriteRect) image.Rectangle {
	return image.Rect(r.X, r.Y, r.X+r.W, r.Y+r.H)
}

func PubPainting(index int) *ebiten.Image {
	ensurePubPaintingsLoaded()
	if index < 0 {
		index = 0
	}
	if index >= NumPubPaintings {
		index = NumPubPaintings - 1
	}
	return pubPaintingImages[index]
}
