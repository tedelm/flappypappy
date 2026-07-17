package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed player_standing.png
var playerStandingPNG []byte

const (
	PlayerStandingFrameCols  = 5
	PlayerStandingFrameRows  = 1
	PlayerStandingFrameCount = PlayerStandingFrameCols * PlayerStandingFrameRows
	// Right-facing profile — clearest idle for a left-side fighter.
	PlayerStandingIdleFrame = 2
)

var (
	playerStandingOnce     sync.Once
	playerStandingFrames   [PlayerStandingFrameCount]*ebiten.Image
	playerStandingContentH float64
)

func ensurePlayerStandingLoaded() {
	playerStandingOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(playerStandingPNG))
		if err != nil {
			panic(err)
		}
		sheet := KeyBlackTransparent(ToRGBA(img))
		b := sheet.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()

		for i := 0; i < PlayerStandingFrameCount; i++ {
			rect := playerStandingCellRect(sheetW, sheetH, i)
			cell := cropRGBA(sheet, rect)
			_, contentH := MeasureRGBA(cell)
			if contentH > playerStandingContentH {
				playerStandingContentH = contentH
			}
			playerStandingFrames[i] = ebiten.NewImageFromImage(cell)
		}

		if playerStandingContentH == 0 {
			playerStandingContentH = float64(sheetH / PlayerStandingFrameRows)
		}
	})
}

func playerStandingCellRect(sheetW, sheetH, i int) image.Rectangle {
	col := i % PlayerStandingFrameCols
	row := i / PlayerStandingFrameCols
	x0 := col * sheetW / PlayerStandingFrameCols
	x1 := (col + 1) * sheetW / PlayerStandingFrameCols
	y0 := row * sheetH / PlayerStandingFrameRows
	y1 := (row + 1) * sheetH / PlayerStandingFrameRows
	return image.Rect(x0, y0, x1, y1)
}

func PlayerStandingFrame(i int) *ebiten.Image {
	ensurePlayerStandingLoaded()
	return playerStandingFrames[i%PlayerStandingFrameCount]
}

func PlayerStanding() *ebiten.Image {
	return PlayerStandingFrame(PlayerStandingIdleFrame)
}

func PlayerStandingContentH() float64 {
	ensurePlayerStandingLoaded()
	return playerStandingContentH
}
