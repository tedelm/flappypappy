package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed player_throws_glass.png
var thrownGlassPNG []byte

const (
	ThrownGlassFrameCols  = 2
	ThrownGlassFrameRows  = 1
	ThrownGlassFrameCount = ThrownGlassFrameCols * ThrownGlassFrameRows
)

var (
	thrownGlassOnce  sync.Once
	thrownGlassFrames [ThrownGlassFrameCount]*ebiten.Image
)

func ensureThrownGlassLoaded() {
	thrownGlassOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(thrownGlassPNG))
		if err != nil {
			panic(err)
		}
		sheet := KeyBlackTransparent(ToRGBA(img))
		b := sheet.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()

		for i := 0; i < ThrownGlassFrameCount; i++ {
			rect := thrownGlassCellRect(sheetW, sheetH, i)
			cell := cropRGBA(sheet, rect)
			thrownGlassFrames[i] = ebiten.NewImageFromImage(cell)
		}
	})
}

func thrownGlassCellRect(sheetW, sheetH, i int) image.Rectangle {
	col := i % ThrownGlassFrameCols
	row := i / ThrownGlassFrameCols
	x0 := col * sheetW / ThrownGlassFrameCols
	x1 := (col + 1) * sheetW / ThrownGlassFrameCols
	y0 := row * sheetH / ThrownGlassFrameRows
	y1 := (row + 1) * sheetH / ThrownGlassFrameRows
	return image.Rect(x0, y0, x1, y1)
}

func ThrownGlassFrame(i int) *ebiten.Image {
	ensureThrownGlassLoaded()
	return thrownGlassFrames[i%ThrownGlassFrameCount]
}
