package game

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"

	"flappy/internal/game/sprite"
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
	spawned       int
	spawnLimit    int
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

func (pm *PipeManager) SetSpawnLimit(n int) {
	pm.spawnLimit = n
}

func (pm *PipeManager) Reset() {
	pm.pipes = nil
	pm.spawned = 0
	pm.spawnTimer = pm.spawnInterval / 2
}

func (pm *PipeManager) spawn() {
	minGapY := 80.0
	maxGapY := float64(FloorSurfaceY) - pm.pipeGap - 80
	gapY := minGapY + rand.Float64()*(maxGapY-minGapY)

	pm.pipes = append(pm.pipes, &Pipe{
		X:    float64(ScreenW) + PipeWidth,
		GapY: gapY,
		GapH: pm.pipeGap,
	})
	pm.spawned++
}

func (pm *PipeManager) Speed() float64 {
	return pm.pipeSpeed
}

func (pm *PipeManager) Spawned() int {
	return pm.spawned
}

func (pm *PipeManager) Pipes() []*Pipe {
	return pm.pipes
}

func (pm *PipeManager) Update() {
	pm.spawnTimer--
	if pm.spawnTimer <= 0 {
		if pm.spawnLimit <= 0 || pm.spawned < pm.spawnLimit {
			pm.spawn()
		}
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

// LastDadFullyCleared reports whether the bird has passed the right edge of every
// on-screen pipe (used to gate boss entry after the final dad is scored).
func (pm *PipeManager) LastDadFullyCleared(birdX float64) bool {
	for _, p := range pm.pipes {
		if birdX <= p.X+PipeWidth {
			return false
		}
	}
	return true
}

// ResetBeforeLastDad clears pipes and places one unscored dad ahead of the bird
// for a final-dad retry. spawnLimit should already be set to the level target.
func (pm *PipeManager) ResetBeforeLastDad() {
	minGapY := 80.0
	maxGapY := float64(FloorSurfaceY) - pm.pipeGap - 80
	gapY := minGapY + rand.Float64()*(maxGapY-minGapY)

	pm.pipes = []*Pipe{{
		X:    float64(ScreenW),
		GapY: gapY,
		GapH: pm.pipeGap,
	}}
	if pm.spawnLimit > 0 {
		pm.spawned = pm.spawnLimit
	} else {
		pm.spawned = 1
	}
	pm.spawnTimer = pm.spawnInterval
}

func (pm *PipeManager) Collides(bx, by, bw, bh float64, dadVariant int) bool {
	return pm.collidesAt(bx, by, bw, bh, dadVariant, sprite.FrameIndex(ebiten.Tick()))
}

func (pm *PipeManager) collidesAt(bx, by, bw, bh float64, dadVariant, frame int) bool {
	for _, p := range pm.pipes {
		lx, ly, lw, lh := lampHitbox(p.X, p.GapY)
		if aabbOverlap(bx, by, bw, bh, lx, ly, lw, lh) {
			return true
		}
		gapBottom := p.GapY + p.GapH
		dx, dy, dw, dh := dadHitbox(p.X, gapBottom, dadVariant, frame)
		if aabbOverlap(bx, by, bw, bh, dx, dy, dw, dh) {
			return true
		}
	}
	return false
}

func (pm *PipeManager) DrawLamps(screen *ebiten.Image) {
	for _, p := range pm.pipes {
		x := float32(p.X)
		topH := float32(p.GapY)
		drawPoolLamp(screen, x, topH)
	}
}

func (pm *PipeManager) DrawDads(screen *ebiten.Image, variant int) {
	for _, p := range pm.pipes {
		x := float32(p.X)
		gapBottom := float32(p.GapY + p.GapH)
		drawFatDad(screen, x, gapBottom, variant)
	}
}

func aabbOverlap(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}
