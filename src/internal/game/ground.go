package game

import (
	"github.com/hajimehoshi/ebiten/v2"

	"flappy/internal/game/sprite"
)

func drawPubForeground(screen *ebiten.Image) {
	groundY := float32(ScreenH - GroundHeight)
	tile := sprite.Panels()
	tileW := float32(tile.Bounds().Dx())
	tileH := float32(tile.Bounds().Dy())
	scale := float32(GroundHeight) / tileH
	scaledW := tileW * scale

	for x := float32(0); x < ScreenW; x += scaledW {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(x), float64(groundY))
		screen.DrawImage(tile, op)
	}
}
