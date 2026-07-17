package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed neighbour_throw_ball.png
var neighbourThrowPNG []byte

const (
	NeighbourThrowFrameCols  = 7
	NeighbourThrowFrameRows  = 1
	NeighbourThrowFrameCount = NeighbourThrowFrameCols * NeighbourThrowFrameRows
	NeighbourThrowFrameTicks = 6
	NeighbourThrowReleaseFrame = 4
)

var (
	neighbourThrowOnce     sync.Once
	neighbourThrowFrames   [NeighbourThrowFrameCount]*ebiten.Image
	neighbourThrowContentH float64
)

func ensureNeighbourThrowLoaded() {
	neighbourThrowOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(neighbourThrowPNG))
		if err != nil {
			panic(err)
		}
		sheet := KeyBlackTransparent(ToRGBA(img))
		b := sheet.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()

		for i := 0; i < NeighbourThrowFrameCount; i++ {
			rect := neighbourThrowCellRect(sheetW, sheetH, i)
			cell := cropRGBA(sheet, rect)
			_, contentH := MeasureRGBA(cell)
			if contentH > neighbourThrowContentH {
				neighbourThrowContentH = contentH
			}
			neighbourThrowFrames[i] = ebiten.NewImageFromImage(cell)
		}

		if neighbourThrowContentH == 0 {
			neighbourThrowContentH = float64(sheetH / NeighbourThrowFrameRows)
		}
	})
}

func neighbourThrowCellRect(sheetW, sheetH, i int) image.Rectangle {
	col := i % NeighbourThrowFrameCols
	row := i / NeighbourThrowFrameCols
	x0 := col * sheetW / NeighbourThrowFrameCols
	x1 := (col + 1) * sheetW / NeighbourThrowFrameCols
	y0 := row * sheetH / NeighbourThrowFrameRows
	y1 := (row + 1) * sheetH / NeighbourThrowFrameRows
	return image.Rect(x0, y0, x1, y1)
}

func NeighbourThrowFrame(i int) *ebiten.Image {
	ensureNeighbourThrowLoaded()
	return neighbourThrowFrames[i%NeighbourThrowFrameCount]
}

func NeighbourThrowContentH() float64 {
	ensureNeighbourThrowLoaded()
	return neighbourThrowContentH
}
