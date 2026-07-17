package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed neighbour_hit.png
var neighbourHitPNG []byte

const (
	NeighbourHitFrameCols  = 2
	NeighbourHitFrameRows  = 1
	NeighbourHitFrameCount = NeighbourHitFrameCols * NeighbourHitFrameRows
	// Front view with surprise marks.
	NeighbourHitReactFrame = 1
)

var (
	neighbourHitOnce     sync.Once
	neighbourHitFrames   [NeighbourHitFrameCount]*ebiten.Image
	neighbourHitContentH float64
)

func ensureNeighbourHitLoaded() {
	neighbourHitOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(neighbourHitPNG))
		if err != nil {
			panic(err)
		}
		sheet := KeyBlackTransparent(ToRGBA(img))
		b := sheet.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()

		for i := 0; i < NeighbourHitFrameCount; i++ {
			rect := neighbourHitCellRect(sheetW, sheetH, i)
			cell := cropRGBA(sheet, rect)
			_, contentH := MeasureRGBA(cell)
			if contentH > neighbourHitContentH {
				neighbourHitContentH = contentH
			}
			neighbourHitFrames[i] = ebiten.NewImageFromImage(cell)
		}

		if neighbourHitContentH == 0 {
			neighbourHitContentH = float64(sheetH / NeighbourHitFrameRows)
		}
	})
}

func neighbourHitCellRect(sheetW, sheetH, i int) image.Rectangle {
	col := i % NeighbourHitFrameCols
	row := i / NeighbourHitFrameCols
	x0 := col * sheetW / NeighbourHitFrameCols
	x1 := (col + 1) * sheetW / NeighbourHitFrameCols
	y0 := row * sheetH / NeighbourHitFrameRows
	y1 := (row + 1) * sheetH / NeighbourHitFrameRows
	return image.Rect(x0, y0, x1, y1)
}

func NeighbourHitFrame(i int) *ebiten.Image {
	ensureNeighbourHitLoaded()
	return neighbourHitFrames[i%NeighbourHitFrameCount]
}

func NeighbourHit() *ebiten.Image {
	return NeighbourHitFrame(NeighbourHitReactFrame)
}

func NeighbourHitContentH() float64 {
	ensureNeighbourHitLoaded()
	return neighbourHitContentH
}
