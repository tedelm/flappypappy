package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"flappy/internal/game/sprite"
)

func drawPoolLamp(screen *ebiten.Image, x, gapY float32) {
	zoneH := gapY
	img := sprite.Lamp()
	imgW := float32(img.Bounds().Dx())
	imgH := float32(img.Bounds().Dy())

	contentH := float32(sprite.LampContentH())
	if contentH <= 0 {
		contentH = imgH
	}
	feetPad := float32(sprite.LampFeetPad())
	scale := zoneH / contentH

	targetW := imgW * scale
	targetH := imgH * scale

	destX := x + (float32(PipeWidth)-targetW)/2
	destY := gapY - targetH + feetPad*scale

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(destX), float64(destY))
	screen.DrawImage(img, op)
}
