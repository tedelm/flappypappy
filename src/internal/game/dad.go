package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"flappy/internal/game/sprite"
)

func drawFatDad(screen *ebiten.Image, x, gapBottom float32) {
	groundY := float32(FloorSurfaceY)
	zoneH := groundY - gapBottom

	frame := sprite.Frame(sprite.FrameIndex(ebiten.Tick()))
	frameW := float32(frame.Bounds().Dx())
	frameH := float32(frame.Bounds().Dy())

	contentH := float32(sprite.ContentH())
	if contentH <= 0 {
		contentH = frameH
	}
	feetPad := float32(sprite.FeetPad())
	scale := zoneH / contentH

	targetW := frameW * scale
	targetH := frameH * scale

	destX := x + (float32(PipeWidth)-targetW)/2
	destY := groundY - targetH + feetPad*scale

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(-float64(scale), float64(scale))
	op.GeoM.Translate(float64(destX+targetW), float64(destY))
	screen.DrawImage(frame, op)
}
