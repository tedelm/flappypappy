package sprite

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed soccer_stadium.png
var soccerStadiumPNG []byte

var (
	soccerStadiumOnce  sync.Once
	soccerStadiumImage *ebiten.Image
)

func ensureSoccerStadiumLoaded() {
	soccerStadiumOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(soccerStadiumPNG))
		if err != nil {
			panic(err)
		}
		soccerStadiumImage = ebiten.NewImageFromImage(ToRGBA(img))
	})
}

func SoccerStadium() *ebiten.Image {
	ensureSoccerStadiumLoaded()
	return soccerStadiumImage
}
