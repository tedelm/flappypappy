package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed running.png
var runningPNG []byte

const (
	RunningFrameCols  = 4
	RunningFrameRows  = 1
	RunningFrameCount = RunningFrameCols * RunningFrameRows
	RunningFrameTicks = 6
	runningBlackKey   = 8
)

var (
	runningOnce     sync.Once
	runningFrames   [RunningFrameCount]*ebiten.Image
	runningContentH float64
	runningFeetPad  float64
)

func ensureRunningLoaded() {
	runningOnce.Do(func() {
		img, err := png.Decode(bytes.NewReader(runningPNG))
		if err != nil {
			panic(err)
		}
		sheet := keyBlackTransparent(ToRGBA(img))
		b := sheet.Bounds()
		sheetW, sheetH := b.Dx(), b.Dy()

		for i := 0; i < RunningFrameCount; i++ {
			rect := runningCellRect(sheetW, sheetH, i)
			cell := cropRGBA(sheet, rect)

			feetPad, contentH := MeasureRGBA(cell)
			if feetPad > runningFeetPad {
				runningFeetPad = feetPad
			}
			if contentH > runningContentH {
				runningContentH = contentH
			}

			runningFrames[i] = ebiten.NewImageFromImage(cell)
		}

		if runningContentH == 0 {
			runningContentH = float64(sheetH / RunningFrameRows)
		}
	})
}

func keyBlackTransparent(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.RGBAAt(x, y)
			if c.R <= runningBlackKey && c.G <= runningBlackKey && c.B <= runningBlackKey {
				src.SetRGBA(x, y, color.RGBA{})
			}
		}
	}
	return src
}

func runningCellRect(sheetW, sheetH, i int) image.Rectangle {
	col := i % RunningFrameCols
	row := i / RunningFrameCols
	x0 := col * sheetW / RunningFrameCols
	x1 := (col + 1) * sheetW / RunningFrameCols
	y0 := row * sheetH / RunningFrameRows
	y1 := (row + 1) * sheetH / RunningFrameRows
	return image.Rect(x0, y0, x1, y1)
}

func RunningFrame(i int) *ebiten.Image {
	ensureRunningLoaded()
	return runningFrames[i%RunningFrameCount]
}

func RunningFrameIndex(tick int64) int {
	return int(tick/int64(RunningFrameTicks)) % RunningFrameCount
}

func RunningContentH() float64 {
	ensureRunningLoaded()
	return runningContentH
}

func RunningFeetPad() float64 {
	ensureRunningLoaded()
	return runningFeetPad
}
