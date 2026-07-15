package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"flappy/internal/game/sprite"
)

func drawPubPanels(screen *ebiten.Image, scrollX float64) {
	wainscotingY := float32(WainscotingTopY)
	tile := sprite.Panels()
	tileW := float32(tile.Bounds().Dx())
	tileH := float32(tile.Bounds().Dy())
	scale := float32(WainscotingHeight) / tileH
	scaledW := tileW * scale
	rem := math.Mod(scrollX, float64(scaledW))

	for x := -rem - float64(scaledW); x < float64(ScreenW)+float64(scaledW); x += float64(scaledW) {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(x, float64(wainscotingY))
		screen.DrawImage(tile, op)
	}
}

func drawWoodenFloor(screen *ebiten.Image) {
	floorY := float32(FloorSurfaceY)
	tile := sprite.WoodFloor()
	tileW := float32(tile.Bounds().Dx())
	tileH := float32(tile.Bounds().Dy())
	scale := float32(FloorHeight) / tileH
	scaledW := tileW * scale

	for x := float32(0); x < ScreenW; x += scaledW {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(scale), float64(scale))
		op.GeoM.Translate(float64(x), float64(floorY))
		screen.DrawImage(tile, op)
	}
}
