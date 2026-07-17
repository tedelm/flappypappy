package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed player_throws.png
var playerThrowsPNG []byte

const (
	PlayerThrowFrameCols  = 6
	PlayerThrowFrameRows  = 1
	PlayerThrowFrameCount = PlayerThrowFrameCols * PlayerThrowFrameRows
	PlayerThrowFrameTicks = 6
	PlayerThrowReleaseFrame = 3
)

var (
	playerThrowsOnce     sync.Once
	playerThrowsFrames   [PlayerThrowFrameCount]*ebiten.Image
	playerThrowsContentH float64
)

func ensurePlayerThrowsLoaded() {
	playerThrowsOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(playerThrowsPNG))
		if err != nil {
			panic(err)
		}
		sheet := KeyBlackTransparent(ToRGBA(img))
		b := sheet.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()

		for i := 0; i < PlayerThrowFrameCount; i++ {
			rect := playerThrowCellRect(sheetW, sheetH, i)
			cell := cropRGBA(sheet, rect)
			_, contentH := MeasureRGBA(cell)
			if contentH > playerThrowsContentH {
				playerThrowsContentH = contentH
			}
			playerThrowsFrames[i] = ebiten.NewImageFromImage(cell)
		}

		if playerThrowsContentH == 0 {
			playerThrowsContentH = float64(sheetH / PlayerThrowFrameRows)
		}
	})
}

func playerThrowCellRect(sheetW, sheetH, i int) image.Rectangle {
	col := i % PlayerThrowFrameCols
	row := i / PlayerThrowFrameCols
	x0 := col * sheetW / PlayerThrowFrameCols
	x1 := (col + 1) * sheetW / PlayerThrowFrameCols
	y0 := row * sheetH / PlayerThrowFrameRows
	y1 := (row + 1) * sheetH / PlayerThrowFrameRows
	return image.Rect(x0, y0, x1, y1)
}

func PlayerThrowFrame(i int) *ebiten.Image {
	ensurePlayerThrowsLoaded()
	return playerThrowsFrames[i%PlayerThrowFrameCount]
}

func PlayerThrowsContentH() float64 {
	ensurePlayerThrowsLoaded()
	return playerThrowsContentH
}
