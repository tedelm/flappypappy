package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed player_standing_ready_to_fight.png
var standingReadyPNG []byte

const (
	StandingReadyFrameCols  = 2
	StandingReadyFrameRows  = 1
	StandingReadyFrameCount = StandingReadyFrameCols * StandingReadyFrameRows
	StandingReadyFrameTicks = 20
)

var (
	standingReadyOnce   sync.Once
	standingReadyFrames [StandingReadyFrameCount]*ebiten.Image
)

func ensureStandingReadyLoaded() {
	standingReadyOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(standingReadyPNG))
		if err != nil {
			panic(err)
		}
		sheet := KeyBlackTransparent(ToRGBA(img))
		b := sheet.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()

		for i := 0; i < StandingReadyFrameCount; i++ {
			rect := standingReadyCellRect(sheetW, sheetH, i)
			cell := cropRGBA(sheet, rect)
			standingReadyFrames[i] = ebiten.NewImageFromImage(cell)
		}
	})
}

func standingReadyCellRect(sheetW, sheetH, i int) image.Rectangle {
	col := i % StandingReadyFrameCols
	row := i / StandingReadyFrameCols
	x0 := col * sheetW / StandingReadyFrameCols
	x1 := (col + 1) * sheetW / StandingReadyFrameCols
	y0 := row * sheetH / StandingReadyFrameRows
	y1 := (row + 1) * sheetH / StandingReadyFrameRows
	return image.Rect(x0, y0, x1, y1)
}

func StandingReadyFrame(i int) *ebiten.Image {
	ensureStandingReadyLoaded()
	return standingReadyFrames[i%StandingReadyFrameCount]
}

func StandingReadyFrameIndex(tick int64) int {
	return int(tick/int64(StandingReadyFrameTicks)) % StandingReadyFrameCount
}
