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

//go:embed dad.png
var dadPNG []byte

const (
	FrameCols  = 5
	FrameRows  = 2
	FrameCount = FrameCols * FrameRows
	FrameTicks = 6
)

var (
	loadOnce    sync.Once
	dadFrames   [FrameCount]*ebiten.Image
	dadFeetPad  float64
	dadContentH float64
)

func ensureLoaded() {
	loadOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(dadPNG))
		if err != nil {
			panic(err)
		}
		sheet := ToRGBA(img)
		b := sheet.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()

		for i := 0; i < FrameCount; i++ {
			rect := cellRect(sheetW, sheetH, i)
			cell := cropRGBA(sheet, rect)

			feetPad, contentH := MeasureRGBA(cell)
			if feetPad > dadFeetPad {
				dadFeetPad = feetPad
			}
			if contentH > dadContentH {
				dadContentH = contentH
			}

			dadFrames[i] = ebiten.NewImageFromImage(cell)
		}

		if dadContentH == 0 {
			dadContentH = float64(sheetH / FrameRows)
		}
	})
}

func cropRGBA(src *image.RGBA, rect image.Rectangle) *image.RGBA {
	sub := src.SubImage(rect)
	out := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(out, out.Bounds(), sub, rect.Min, draw.Src)
	return out
}

func cellRect(sheetW, sheetH, i int) image.Rectangle {
	col := i % FrameCols
	row := i / FrameCols
	x0 := col * sheetW / FrameCols
	x1 := (col + 1) * sheetW / FrameCols
	y0 := row * sheetH / FrameRows
	y1 := (row + 1) * sheetH / FrameRows
	return image.Rect(x0, y0, x1, y1)
}

func Frame(i int) *ebiten.Image {
	ensureLoaded()
	return dadFrames[i%FrameCount]
}

func FrameIndex(tick int64) int {
	return int(tick/int64(FrameTicks)) % FrameCount
}

func FeetPad() float64 {
	ensureLoaded()
	return dadFeetPad
}

func ContentH() float64 {
	ensureLoaded()
	return dadContentH
}
