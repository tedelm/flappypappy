package sprite

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed boss_1_bg.png
var boss1BGPNG []byte

var (
	boss1BGOnce  sync.Once
	boss1BGImage *ebiten.Image
)

func ensureBoss1BGLoaded() {
	boss1BGOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(boss1BGPNG))
		if err != nil {
			panic(err)
		}
		boss1BGImage = ebiten.NewImageFromImage(ToRGBA(img))
	})
}

func Boss1BG() *ebiten.Image {
	ensureBoss1BGLoaded()
	return boss1BGImage
}
