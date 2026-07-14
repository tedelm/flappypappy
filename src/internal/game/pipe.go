package game

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Pipe struct {
	X       float64
	GapY    float64
	GapH    float64
	Scored  bool
}

type PipeManager struct {
	pipes      []*Pipe
	spawnTimer int
}

func NewPipeManager() *PipeManager {
	return &PipeManager{}
}

func (pm *PipeManager) Reset() {
	pm.pipes = nil
	pm.spawnTimer = SpawnInterval / 2
}

func (pm *PipeManager) spawn() {
	minGapY := 80.0
	maxGapY := float64(ScreenH-GroundHeight) - float64(PipeGap) - 80
	gapY := minGapY + rand.Float64()*(maxGapY-minGapY)

	pm.pipes = append(pm.pipes, &Pipe{
		X:    float64(ScreenW) + PipeWidth,
		GapY: gapY,
		GapH: float64(PipeGap),
	})
}

func (pm *PipeManager) Update() {
	pm.spawnTimer--
	if pm.spawnTimer <= 0 {
		pm.spawn()
		pm.spawnTimer = SpawnInterval
	}

	remaining := pm.pipes[:0]
	for _, p := range pm.pipes {
		p.X -= PipeSpeed
		if p.X+PipeWidth < 0 {
			continue
		}
		remaining = append(remaining, p)
	}
	pm.pipes = remaining
}

func (pm *PipeManager) CheckScore(birdX float64) int {
	scoreDelta := 0
	for _, p := range pm.pipes {
		if !p.Scored && birdX > p.X+PipeWidth/2 {
			p.Scored = true
			scoreDelta++
		}
	}
	return scoreDelta
}

func (pm *PipeManager) Collides(bx, by, bw, bh float64) bool {
	for _, p := range pm.pipes {
		if aabbOverlap(bx, by, bw, bh, p.X, 0, PipeWidth, p.GapY) {
			return true
		}
		gapBottom := p.GapY + p.GapH
		groundTop := float64(ScreenH - GroundHeight)
		if aabbOverlap(bx, by, bw, bh, p.X, gapBottom, PipeWidth, groundTop-gapBottom) {
			return true
		}
	}
	return false
}

func (pm *PipeManager) Draw(screen *ebiten.Image) {
	for _, p := range pm.pipes {
		x := float32(p.X)
		w := float32(PipeWidth)

		topH := float32(p.GapY)
		vector.DrawFilledRect(screen, x, 0, w, topH, ColorPipe, true)
		vector.StrokeRect(screen, x, 0, w, topH, 2, ColorPipeEdge, true)

		gapBottom := float32(p.GapY + p.GapH)
		bottomH := float32(ScreenH-GroundHeight) - gapBottom
		vector.DrawFilledRect(screen, x, gapBottom, w, bottomH, ColorPipe, true)
		vector.StrokeRect(screen, x, gapBottom, w, bottomH, 2, ColorPipeEdge, true)

		lipW := w + 8
		lipH := float32(24)
		vector.DrawFilledRect(screen, x-4, topH-lipH, lipW, lipH, ColorPipeEdge, true)
		vector.DrawFilledRect(screen, x-4, gapBottom, lipW, lipH, ColorPipeEdge, true)
	}
}

func aabbOverlap(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}
