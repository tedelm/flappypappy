//go:build ignore

package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

var (
	gap   = color.RGBA{28, 18, 12, 255}
	dark  = color.RGBA{72, 44, 24, 255}
	mid   = color.RGBA{110, 68, 36, 255}
	light = color.RGBA{145, 92, 48, 255}
	hi    = color.RGBA{168, 108, 58, 255}
)

func main() {
	const W, H = 128, 32
	img := image.NewRGBA(image.Rect(0, 0, W, H))

	plankWidths := []int{28, 36, 32, 24, 40, 28, 36, 32}
	x := 0
	plank := 0
	for x < W {
		w := plankWidths[plank%len(plankWidths)]
		if x+w > W {
			w = W - x
		}
		drawPlank(img, x, 0, w, H, plank)
		x += w
		plank++
	}

	f, err := os.Create("src/internal/game/sprite/wood_floor.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func drawPlank(img *image.RGBA, x0, y0, w, h, seed int) {
	// gap line on left edge (except first plank)
	if x0 > 0 {
		for y := y0; y < y0+h; y++ {
			img.Set(x0, y, gap)
		}
	}

	for y := y0 + 1; y < y0+h-1; y++ {
		for x := x0 + 1; x < x0+w; x++ {
			c := plankColor(x, y, x0, w, h, seed)
			img.Set(x, y, c)
		}
	}

	// top highlight edge
	for x := x0 + 1; x < x0+w; x++ {
		img.Set(x, y0, hi)
	}
	// bottom shadow edge
	for x := x0 + 1; x < x0+w; x++ {
		img.Set(x, y0+h-1, dark)
	}
}

func plankColor(x, y, x0, w, h, seed int) color.RGBA {
	lx := x - x0
	ly := y

	// wavy grain like panels.png
	wave := ((lx*2 + ly*3 + seed*7) / 5) % 5
	var base color.RGBA
	switch wave {
	case 0:
		base = dark
	case 1:
		base = mid
	case 2:
		base = light
	case 3:
		base = hi
	default:
		base = mid
	}

	// subtle arc grain lines
	if (lx+ly*2+seed*3)%11 < 2 {
		base = dark
	}
	if (lx*3-ly+seed)%13 == 0 {
		base = hi
	}

	// knot per plank
	kx := x0 + w/3 + seed*5%8
	ky := h/2 + seed%3
	dx, dy := x-kx, ly-ky
	if dx*dx+dy*dy < 6 {
		return dark
	}

	return base
}
