package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed AngryWoman.png
var angryWomanPNG []byte

const (
	AngryWomanFrameCols  = 5
	AngryWomanFrameRows  = 2
	AngryWomanFrameCount = AngryWomanFrameCols * AngryWomanFrameRows
	AngryWomanFrameTicks = 6
)

var (
	angryWomanOnce     sync.Once
	angryWomanFrames   [AngryWomanFrameCount]*ebiten.Image
	angryWomanContentH float64
)

func ensureAngryWomanLoaded() {
	angryWomanOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(angryWomanPNG))
		if err != nil {
			panic(err)
		}
		sheet := ToRGBA(img)
		b := sheet.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()

		for i := 0; i < AngryWomanFrameCount; i++ {
			rect := angryWomanCellRect(sheetW, sheetH, i)
			cell := cropRGBA(sheet, rect)

			_, contentH := MeasureRGBA(cell)
			if contentH > angryWomanContentH {
				angryWomanContentH = contentH
			}

			angryWomanFrames[i] = ebiten.NewImageFromImage(cell)
		}

		if angryWomanContentH == 0 {
			angryWomanContentH = float64(sheetH / AngryWomanFrameRows)
		}
	})
}

func angryWomanCellRect(sheetW, sheetH, i int) image.Rectangle {
	col := i % AngryWomanFrameCols
	row := i / AngryWomanFrameCols
	x0 := col * sheetW / AngryWomanFrameCols
	x1 := (col + 1) * sheetW / AngryWomanFrameCols
	y0 := row * sheetH / AngryWomanFrameRows
	y1 := (row + 1) * sheetH / AngryWomanFrameRows
	return image.Rect(x0, y0, x1, y1)
}

func AngryWomanFrame(i int) *ebiten.Image {
	ensureAngryWomanLoaded()
	return angryWomanFrames[i%AngryWomanFrameCount]
}

func AngryWomanFrameIndex(tick int64) int {
	return int(tick/int64(AngryWomanFrameTicks)) % AngryWomanFrameCount
}

func AngryWomanContentH() float64 {
	ensureAngryWomanLoaded()
	return angryWomanContentH
}
