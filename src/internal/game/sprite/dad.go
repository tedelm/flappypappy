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

//go:embed dad_2.png
var dad2PNG []byte

//go:embed dad_3.png
var dad3PNG []byte

//go:embed dad_4.png
var dad4PNG []byte

const (
	DadVariantCount = 4

	FrameCols  = 5
	FrameRows  = 2
	FrameCount = FrameCols * FrameRows
	FrameTicks = 6
)

var (
	loadOnce     sync.Once
	dadFrames    [DadVariantCount][FrameCount]*ebiten.Image
	dadFeetPad   [DadVariantCount][FrameCount]float64
	dadContentH  [DadVariantCount][FrameCount]float64
	dadContentW  [DadVariantCount][FrameCount]float64
	dadSheetPNGs = [][]byte{dadPNG, dad2PNG, dad3PNG, dad4PNG}
)

func ensureLoaded() {
	loadOnce.Do(func() {
		for v, pngBytes := range dadSheetPNGs {
			img, err := png.Decode(bytes.NewReader(pngBytes))
			if err != nil {
				panic(err)
			}
			sheet := KeyBlackTransparent(ToRGBA(img))
			b := sheet.Bounds()
			sheetW, sheetH := b.Dx(), b.Dy()

			for i := 0; i < FrameCount; i++ {
				rect := cellRect(sheetW, sheetH, i)
				cell := cropRGBA(sheet, rect)

				fp, ch, cw := MeasureRGBASize(cell)
				if ch == 0 {
					ch = float64(sheetH / FrameRows)
				}
				if cw == 0 {
					cw = float64(sheetW / FrameCols)
				}
				dadFeetPad[v][i] = fp
				dadContentH[v][i] = ch
				dadContentW[v][i] = cw
				dadFrames[v][i] = ebiten.NewImageFromImage(cell)
			}
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

func clampDadVariant(variant int) int {
	if variant < 0 || variant >= DadVariantCount {
		return 0
	}
	return variant
}

func clampDadFrame(frame int) int {
	if frame < 0 {
		return 0
	}
	return frame % FrameCount
}

func Frame(variant, i int) *ebiten.Image {
	ensureLoaded()
	v := clampDadVariant(variant)
	return dadFrames[v][clampDadFrame(i)]
}

func FrameIndex(tick int64) int {
	return int(tick/int64(FrameTicks)) % FrameCount
}

func FeetPad(variant, frame int) float64 {
	ensureLoaded()
	return dadFeetPad[clampDadVariant(variant)][clampDadFrame(frame)]
}

func ContentH(variant, frame int) float64 {
	ensureLoaded()
	return dadContentH[clampDadVariant(variant)][clampDadFrame(frame)]
}

func ContentW(variant, frame int) float64 {
	ensureLoaded()
	return dadContentW[clampDadVariant(variant)][clampDadFrame(frame)]
}
