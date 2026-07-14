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
		drawKegStack(screen, x, 0, w, topH)

		gapBottom := float32(p.GapY + p.GapH)
		bottomH := float32(ScreenH-GroundHeight) - gapBottom
		drawKegStack(screen, x, gapBottom, w, bottomH)

		rimW := w + 8
		rimH := float32(24)
		drawKegRim(screen, x-4, topH-rimH, rimW, rimH)
		drawKegRim(screen, x-4, gapBottom, rimW, rimH)
	}
}

func drawKegStack(screen *ebiten.Image, x, y, w, h float32) {
	segH := float32(KegSegmentH)
	remaining := h
	cy := y + h - segH
	for remaining > 0 {
		drawH := segH
		if remaining < segH {
			drawH = remaining
			cy = y
		}
		drawKegSegment(screen, x, cy, w, drawH)
		remaining -= segH
		cy -= segH
	}
}

func drawKegSegment(screen *ebiten.Image, x, y, w, h float32) {
	inset := float32(2)
	vector.DrawFilledRect(screen, x+inset, y, w-inset*2, h, ColorKeg, true)
	vector.StrokeRect(screen, x+inset, y, w-inset*2, h, 1.5, ColorKegEdge, true)

	bandY1 := y + h*0.3
	bandY2 := y + h*0.7
	vector.StrokeLine(screen, x+inset, bandY1, x+w-inset, bandY1, 2, ColorKegBand, true)
	vector.StrokeLine(screen, x+inset, bandY2, x+w-inset, bandY2, 2, ColorKegBand, true)

	cx := x + w/2
	vector.DrawFilledCircle(screen, cx, y, 4, ColorKegRim, true)
	vector.DrawFilledCircle(screen, cx, y+h, 4, ColorKegRim, true)
}

func drawKegRim(screen *ebiten.Image, x, y, w, h float32) {
	vector.DrawFilledRect(screen, x, y, w, h, ColorKegRim, true)
	vector.StrokeRect(screen, x, y, w, h, 2, ColorKegEdge, true)
	vector.StrokeLine(screen, x+4, y+h/2, x+w-4, y+h/2, 2, ColorKegBand, true)
}

func aabbOverlap(ax, ay, aw, ah, bx, by, bw, bh float64) bool {
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}
