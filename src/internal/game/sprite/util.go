package sprite

import (
	"image"
	"image/color"
)

const AlphaThreshold = 16
const BlackKeyThreshold = 8

func ToRGBA(src image.Image) *image.RGBA {
	b := src.Bounds()
	out := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			out.SetRGBA(x, y, color.RGBAModel.Convert(src.At(x, y)).(color.RGBA))
		}
	}
	return out
}

func MeasureRGBA(img *image.RGBA) (feetPad, contentH float64) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	minY, maxY := h, -1
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if img.RGBAAt(x, y).A > AlphaThreshold {
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if maxY < 0 {
		return 0, float64(h)
	}
	return float64(h - 1 - maxY), float64(maxY - minY + 1)
}

func KeyBlackTransparent(src *image.RGBA) *image.RGBA {
	b := src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := src.RGBAAt(x, y)
			if c.R <= BlackKeyThreshold && c.G <= BlackKeyThreshold && c.B <= BlackKeyThreshold {
				src.SetRGBA(x, y, color.RGBA{})
			}
		}
	}
	return src
}
