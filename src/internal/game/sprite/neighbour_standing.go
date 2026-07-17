package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed neighbour_standing.png
var neighbourStandingPNG []byte

const (
	NeighbourStandingFrameCols  = 4
	NeighbourStandingFrameRows  = 1
	NeighbourStandingFrameCount = NeighbourStandingFrameCols * NeighbourStandingFrameRows
	// 3/4 facing right; draw path flips horizontally to face the player.
	NeighbourStandingIdleFrame = 0
)

var (
	neighbourStandingOnce     sync.Once
	neighbourStandingFrames   [NeighbourStandingFrameCount]*ebiten.Image
	neighbourStandingContentH float64
)

func ensureNeighbourStandingLoaded() {
	neighbourStandingOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(neighbourStandingPNG))
		if err != nil {
			panic(err)
		}
		sheet := KeyBlackTransparent(ToRGBA(img))
		b := sheet.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()

		for i := 0; i < NeighbourStandingFrameCount; i++ {
			rect := neighbourStandingCellRect(sheetW, sheetH, i)
			cell := cropRGBA(sheet, rect)
			_, contentH := MeasureRGBA(cell)
			if contentH > neighbourStandingContentH {
				neighbourStandingContentH = contentH
			}
			neighbourStandingFrames[i] = ebiten.NewImageFromImage(cell)
		}

		if neighbourStandingContentH == 0 {
			neighbourStandingContentH = float64(sheetH / NeighbourStandingFrameRows)
		}
	})
}

func neighbourStandingCellRect(sheetW, sheetH, i int) image.Rectangle {
	col := i % NeighbourStandingFrameCols
	row := i / NeighbourStandingFrameCols
	x0 := col * sheetW / NeighbourStandingFrameCols
	x1 := (col + 1) * sheetW / NeighbourStandingFrameCols
	y0 := row * sheetH / NeighbourStandingFrameRows
	y1 := (row + 1) * sheetH / NeighbourStandingFrameRows
	return image.Rect(x0, y0, x1, y1)
}

func NeighbourStandingFrame(i int) *ebiten.Image {
	ensureNeighbourStandingLoaded()
	return neighbourStandingFrames[i%NeighbourStandingFrameCount]
}

func NeighbourStanding() *ebiten.Image {
	return NeighbourStandingFrame(NeighbourStandingIdleFrame)
}

func NeighbourStandingContentH() float64 {
	ensureNeighbourStandingLoaded()
	return neighbourStandingContentH
}
