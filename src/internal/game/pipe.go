package game

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

type Pipe struct {
	X      float64
	GapY   float64
	GapH   float64
	Scored bool
}

type PipeManager struct {
	pipes         []*Pipe
	spawnTimer    int
	pipeSpeed     float64
	pipeGap       float64
	spawnInterval int
}

func NewPipeManager() *PipeManager {
	pm := &PipeManager{}
	pm.ApplyConfig(DifficultyEasy.Config())
	return pm
}

func (pm *PipeManager) ApplyConfig(cfg DifficultyConfig) {
	pm.pipeSpeed = cfg.PipeSpeed
	pm.pipeGap = float64(cfg.PipeGap)
	pm.spawnInterval = cfg.SpawnInterval
}

func (pm *PipeManager) Reset() {
	pm.pipes = nil
	pm.spawnTimer = pm.spawnInterval / 2
}

func (pm *PipeManager) spawn() {
	minGapY := 80.0
	maxGapY := float64(ScreenH-GroundHeight) - pm.pipeGap - 80
	gapY := minGapY + rand.Float64()*(maxGapY-minGapY)

	pm.pipes = append(pm.pipes, &Pipe{
		X:    float64(ScreenW) + PipeWidth,
		GapY: gapY,
		GapH: pm.pipeGap,
	})
}

func (pm *PipeManager) Update() {
	pm.spawnTimer--
	if pm.spawnTimer <= 0 {
		pm.spawn()
		pm.spawnTimer = pm.spawnInterval
	}

	remaining := pm.pipes[:0]
	for _, p := range pm.pipes {
		p.X -= pm.pipeSpeed
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

		topH := float32(p.GapY)
		drawPoolLamp(screen, x, topH)

		gapBottom := float32(p.GapY + p.GapH)
		drawFatDad(screen, x, gapBottom)
	}
}

func aabbOverlap(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}
