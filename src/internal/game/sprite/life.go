package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/draw"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed life.png
var lifePNG []byte

const (
	LifeFrameCount = 2
	LifeFrameW     = 32
)

var (
	lifeOnce   sync.Once
	lifeFrames [LifeFrameCount]*ebiten.Image
	lifeFrameW int
	lifeFrameH int
)

func ensureLifeLoaded() {
	lifeOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(lifePNG))
		if err != nil {
			panic(err)
		}
		rgba := ToRGBA(img)
		b := rgba.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()
		lifeFrameW = sheetW / LifeFrameCount
		lifeFrameH = sheetH
		if lifeFrameW == 0 {
			lifeFrameW = LifeFrameW
		}

		for i := 0; i < LifeFrameCount; i++ {
			rect := image.Rect(i*lifeFrameW, 0, (i+1)*lifeFrameW, lifeFrameH)
			sub := rgba.SubImage(rect)
			out := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
			draw.Draw(out, out.Bounds(), sub, rect.Min, draw.Src)
			lifeFrames[i] = ebiten.NewImageFromImage(out)
		}
	})
}

func LifeFull() *ebiten.Image {
	ensureLifeLoaded()
	return lifeFrames[0]
}

func LifeEmpty() *ebiten.Image {
	ensureLifeLoaded()
	return lifeFrames[1]
}

func LifeFrameWidth() int {
	ensureLifeLoaded()
	return lifeFrameW
}
