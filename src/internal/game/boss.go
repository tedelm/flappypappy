package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"flappy/internal/game/sprite"
)

const (
	bossDisplayH    = 140.0
	bossMarginX     = 20.0
	bossMoveSpeed   = 2.5
	bossMoveMinY    = 60.0
	projectileSpeed = 8.0
	projectileSize  = 20.0
	bossPlayerY     = float64(FloorSurfaceY+80) / 2
	bossHitboxInset = 0.15
)

type Projectile struct {
	X, Y float64
}

type Boss struct {
	X, Y          float64
	width, height float64
	moveDir       float64
}

type BossFight struct {
	throwsUsed  int
	hits        int
	projectiles []*Projectile
	boss        Boss
}

func NewBossFight() *BossFight {
	bf := &BossFight{}
	bf.Reset()
	return bf
}

func (bf *BossFight) Reset() {
	frame := sprite.AngryWomanFrame(0)
	contentH := sprite.AngryWomanContentH()
	scale := bossDisplayH / contentH
	bossW := float64(frame.Bounds().Dx()) * scale

	bf.throwsUsed = 0
	bf.hits = 0
	bf.projectiles = nil
	bf.boss = Boss{
		X:       float64(ScreenW) - bossMarginX - bossW,
		Y:       bossPlayerY,
		width:   bossW,
		height:  bossDisplayH,
		moveDir: 1,
	}
}

func (b *Boss) Bounds() (x, y, w, h float64) {
	insetX := b.width * bossHitboxInset
	insetY := b.height * bossHitboxInset
	return b.X + insetX, b.Y + insetY, b.width - insetX*2, b.height - insetY*2
}

func (b *Boss) Update() {
	b.Y += b.moveDir * bossMoveSpeed
	maxY := float64(FloorSurfaceY) - b.height - 20
	if b.Y <= bossMoveMinY {
		b.Y = bossMoveMinY
		b.moveDir = 1
	} else if b.Y >= maxY {
		b.Y = maxY
		b.moveDir = -1
	}
}

func (b *Boss) Draw(screen *ebiten.Image) {
	frame := sprite.AngryWomanFrame(sprite.AngryWomanFrameIndex(ebiten.Tick()))
	contentH := sprite.AngryWomanContentH()
	scale := b.height / contentH

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(b.X, b.Y)
	screen.DrawImage(frame, op)
}

func (p *Projectile) Bounds() (x, y, w, h float64) {
	half := projectileSize / 2
	return p.X - half, p.Y - half, projectileSize, projectileSize
}

func drawProjectile(screen *ebiten.Image, p *Projectile) {
	half := float32(projectileSize / 2)
	cx := float32(p.X)
	cy := float32(p.Y)

	vector.DrawFilledRect(screen, cx-half+2, cy-half, half*2-4, half*2, ColorGlass, true)
	vector.DrawFilledRect(screen, cx-half+4, cy-half+4, half*2-8, half*2-8, ColorBeer, true)
	vector.StrokeRect(screen, cx-half+2, cy-half, half*2-4, half*2, 1.5, ColorGlassEdge, true)
	vector.DrawFilledCircle(screen, cx, cy-half+3, 4, ColorFoam, true)
}

func (bf *BossFight) CanThrow() bool {
	return bf.throwsUsed < BossThrowsAllowed && len(bf.projectiles) == 0
}

func (bf *BossFight) Throw() {
	if !bf.CanThrow() {
		return
	}
	bf.projectiles = append(bf.projectiles, &Projectile{
		X: float64(BirdStartX) + BirdWidth/2,
		Y: bossPlayerY,
	})
	bf.throwsUsed++
}

func (bf *BossFight) Update() (won, lost bool) {
	bf.boss.Update()

	remaining := bf.projectiles[:0]
	for _, p := range bf.projectiles {
		p.X += projectileSpeed
		if p.X > float64(ScreenW)+projectileSize {
			continue
		}
		px, py, pw, ph := p.Bounds()
		bx, by, bw, bh := bf.boss.Bounds()
		if aabbOverlap(px, py, pw, ph, bx, by, bw, bh) {
			bf.hits++
			continue
		}
		remaining = append(remaining, p)
	}
	bf.projectiles = remaining

	if bf.hits >= BossHitsRequired {
		return true, false
	}
	if bf.throwsUsed >= BossThrowsAllowed && len(bf.projectiles) == 0 && bf.hits < BossHitsRequired {
		return false, true
	}
	return false, false
}

func (bf *BossFight) Draw(screen *ebiten.Image) {
	bf.boss.Draw(screen)
	for _, p := range bf.projectiles {
		drawProjectile(screen, p)
	}
}

func drawBossPlayer(screen *ebiten.Image) {
	cx := float64(BirdStartX)
	cy := bossPlayerY
	halfH := BirdHeight / 2.0
	topHW := BirdWidth/2.0 - 2
	botHW := BirdWidth/2.0 - 6
	inset := 3.0

	glass := toWorldQuad([4][2]float64{
		{-topHW, -halfH},
		{topHW, -halfH},
		{botHW, halfH},
		{-botHW, halfH},
	}, cx, cy, 0)

	beer := toWorldQuad([4][2]float64{
		{localEdgeX(topHW, botHW, halfH, -halfH+16, true), -halfH + 16},
		{localEdgeX(topHW, botHW, halfH, -halfH+16, false), -halfH + 16},
		{localEdgeX(topHW, botHW, halfH, halfH-inset, false), halfH - inset},
		{localEdgeX(topHW, botHW, halfH, halfH-inset, true), halfH - inset},
	}, cx, cy, 0)

	drawFilledQuad(screen, glass, ColorGlass)
	drawFilledQuad(screen, beer, ColorBeer)
	drawStrokedQuad(screen, glass, 2, ColorGlassEdge)

	foamCX, foamCY := toWorld(0, -halfH+10, cx, cy, 0)
	vector.DrawFilledCircle(screen, float32(foamCX), float32(foamCY), 5, ColorFoam, true)
}
