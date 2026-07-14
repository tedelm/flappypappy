package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"flappy/internal/game/sprite"
)

const tableHashSalt = 0xABCDEF

func drawPubBackground(screen *ebiten.Image, scrollX float64, seed int) {
	playH := ScreenH - GroundHeight
	screen.Fill(ColorPubBase)

	tile := sprite.PubWall()
	tileW := tile.Bounds().Dx()
	tileH := tile.Bounds().Dy()
	rem := math.Mod(scrollX, float64(tileW))

	for y := 0; y < playH; y += tileH {
		for x := -rem - float64(tileW); x < float64(ScreenW+tileW); x += float64(tileW) {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x, float64(y))
			screen.DrawImage(tile, op)
		}
	}

	drawPubDecor(screen, scrollX, seed)
}

func decorHash(seed, worldX int) uint32 {
	v := uint32(seed) ^ uint32(worldX)*2654435761
	v ^= v >> 16
	v *= 0x7feb352d
	v ^= v >> 15
	v *= 0x846ca68b
	v ^= v >> 16
	return v
}

func decorShouldSpawn(seed, worldX int) bool {
	return int(decorHash(seed, worldX)%100) < DecorPaintChance
}

func tableDecorHash(seed, worldX int) uint32 {
	return decorHash(seed, worldX^tableHashSalt)
}

func tableShouldSpawn(seed, worldX int) bool {
	return int(tableDecorHash(seed, worldX)%100) < DecorTableChance
}

func frontStoolJitter(seed, worldX int) float64 {
	h := tableDecorHash(seed, worldX^0x12345)
	return float64(int(h>>8)%16) - 8
}

func drawPubDecor(screen *ebiten.Image, scrollX float64, seed int) {
	drawPubPaintings(screen, scrollX, seed)
	drawPubTableClusters(screen, scrollX, seed)
}

func decorGridRange(scrollX float64, margin, step int) (start, end int) {
	s := int(scrollX)
	start = ((s - ScreenW - margin) / step) * step
	end = s + ScreenW + margin + step
	return start, end
}

func drawPubPaintings(screen *ebiten.Image, scrollX float64, seed int) {
	worldStart, worldEnd := decorGridRange(scrollX, DecorChunkMin, DecorChunkMin)

	for worldX := worldStart; worldX < worldEnd; worldX += DecorChunkMin {
		if !decorShouldSpawn(seed, worldX) {
			continue
		}
		centerScreenX := float64(worldX) - scrollX
		drawPubPainting(screen, centerScreenX, float64(DecorPaintingY))
	}
}

func drawPubTableClusters(screen *ebiten.Image, scrollX float64, seed int) {
	worldStart, worldEnd := decorGridRange(scrollX, DecorTableChunkMin, DecorTableChunkMin)

	for worldX := worldStart; worldX < worldEnd; worldX += DecorTableChunkMin {
		if !tableShouldSpawn(seed, worldX) {
			continue
		}
		centerScreenX := float64(worldX) - scrollX
		drawPubTable(screen, centerScreenX, float64(DecorTableY))

		stoolImg := sprite.PubStool()
		stoolW := float64(stoolImg.Bounds().Dx()) * DecorStoolScale
		feetY := float64(DecorStoolY)

		drawPubStool(screen, centerScreenX-DecorStoolAroundX-stoolW/2, feetY)
		drawPubStool(screen, centerScreenX+DecorStoolAroundX-stoolW/2, feetY)
		drawPubStool(screen, centerScreenX+frontStoolJitter(seed, worldX)-stoolW/2, feetY)
	}
}

func drawPubTable(screen *ebiten.Image, centerX, topY float64) {
	x := float32(centerX - float64(DecorTableW)/2)
	y := float32(topY)
	w := float32(DecorTableW)
	h := float32(DecorTableH)
	vector.DrawFilledRect(screen, x, y, w, h, ColorKeg, true)
	vector.StrokeRect(screen, x, y, w, h, 2, ColorKegEdge, true)
}

func drawPubStool(screen *ebiten.Image, x, feetY float64) {
	img := sprite.PubStool()
	scale := DecorStoolScale
	targetH := float64(img.Bounds().Dy()) * scale
	feetPad := sprite.PubStoolFeetPad() * scale
	destY := feetY - targetH + feetPad

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, destY)
	screen.DrawImage(img, op)
}

func drawPubPainting(screen *ebiten.Image, centerX, topY float64) {
	img := sprite.PubPainting()
	scale := DecorPaintingScale
	scaledW := float64(img.Bounds().Dx()) * scale
	drawX := centerX - scaledW/2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(drawX, topY)
	screen.DrawImage(img, op)
}
