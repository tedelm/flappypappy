package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"flappy/internal/game/sprite"
)

func drawFatDad(screen *ebiten.Image, x, gapBottom float32, variant int) {
	groundY := float32(FloorSurfaceY)
	zoneH := groundY - gapBottom
	if zoneH <= 0 {
		return
	}

	idx := sprite.FrameIndex(ebiten.Tick())
	frame := sprite.Frame(variant, idx)
	frameW := float32(frame.Bounds().Dx())
	frameH := float32(frame.Bounds().Dy())

	contentH := float32(sprite.ContentH(variant, idx))
	if contentH <= 0 {
		contentH = frameH
	}
	feetPad := float32(sprite.FeetPad(variant, idx))
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

// dadHitbox returns the centered bottom-zone collision rect for the current dad frame.
func dadHitbox(pipeX, gapBottom float64, variant, frame int) (x, y, w, h float64) {
	groundTop := float64(FloorSurfaceY)
	zoneH := groundTop - gapBottom
	if zoneH <= 0 {
		return pipeX, gapBottom, PipeWidth, 0
	}
	contentH := sprite.ContentH(variant, frame)
	if contentH <= 0 {
		contentH = 1
	}
	contentW := sprite.ContentW(variant, frame)
	scale := zoneH / contentH
	hitW := contentW * scale
	if hitW <= 0 || hitW > PipeWidth {
		hitW = PipeWidth
	}
	hitX := pipeX + (PipeWidth-hitW)/2
	return hitX, gapBottom, hitW, zoneH
}

// lampHitbox returns the centered top-zone collision rect for the lamp.
func lampHitbox(pipeX, gapY float64) (x, y, w, h float64) {
	if gapY <= 0 {
		return pipeX, 0, PipeWidth, 0
	}
	contentH := sprite.LampContentH()
	if contentH <= 0 {
		contentH = 1
	}
	contentW := sprite.LampContentW()
	scale := gapY / contentH
	hitW := contentW * scale
	if hitW <= 0 || hitW > PipeWidth {
		hitW = PipeWidth
	}
	hitX := pipeX + (PipeWidth-hitW)/2
	return hitX, 0, hitW, gapY
}
