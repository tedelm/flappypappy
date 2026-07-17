package sprite

import (
	_ "embed"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed dad_4_fighter.png
var dadFighterPNG []byte

//go:embed dad_4_fighter_beaten.png
var dadFighterBeatenPNG []byte

const (
	DadFighterFrameCols  = 5
	DadFighterFrameRows  = 2
	DadFighterFrameCount = DadFighterFrameCols * DadFighterFrameRows
	DadFighterFrameTicks = 5

	DadFighterIdle0         = 0
	DadFighterIdle1         = 1
	DadFighterStraightStart = 2
	DadFighterStraightCount = 3
	DadFighterDuckFrame     = 5
	DadFighterUppercutStart = 6
	DadFighterUppercutCount = 2
	DadFighterHitFrame      = 8
	DadFighterIdleAlt       = 9

	DadFighterBeatenFrameCols  = 5
	DadFighterBeatenFrameRows  = 1
	DadFighterBeatenFrameCount = DadFighterBeatenFrameCols * DadFighterBeatenFrameRows
)

var (
	dadFighterOnce     sync.Once
	dadFighterFrames   [DadFighterFrameCount]*ebiten.Image
	dadFighterContentH float64
	dadFighterFeetPad  float64

	dadFighterBeatenOnce   sync.Once
	dadFighterBeatenFrames [DadFighterBeatenFrameCount]*ebiten.Image
)

func ensureDadFighterLoaded() {
	dadFighterOnce.Do(func() {
		frames := make([]*ebiten.Image, DadFighterFrameCount)
		dadFighterContentH, dadFighterFeetPad = loadSheetFrames(
			dadFighterPNG, DadFighterFrameCols, DadFighterFrameRows, DadFighterFrameCount, frames,
		)
		for i := range frames {
			dadFighterFrames[i] = frames[i]
		}
	})
}

func ensureDadFighterBeatenLoaded() {
	dadFighterBeatenOnce.Do(func() {
		frames := make([]*ebiten.Image, DadFighterBeatenFrameCount)
		loadSheetFrames(
			dadFighterBeatenPNG, DadFighterBeatenFrameCols, DadFighterBeatenFrameRows, DadFighterBeatenFrameCount, frames,
		)
		for i := range frames {
			dadFighterBeatenFrames[i] = frames[i]
		}
	})
}

func DadFighterFrame(i int) *ebiten.Image {
	ensureDadFighterLoaded()
	if i < 0 {
		i = 0
	}
	return dadFighterFrames[i%DadFighterFrameCount]
}

func DadFighterBeatenFrame(i int) *ebiten.Image {
	ensureDadFighterBeatenLoaded()
	if i < 0 {
		i = 0
	}
	return dadFighterBeatenFrames[i%DadFighterBeatenFrameCount]
}

func DadFighterContentH() float64 {
	ensureDadFighterLoaded()
	return dadFighterContentH
}

func DadFighterFeetPad() float64 {
	ensureDadFighterLoaded()
	return dadFighterFeetPad
}

// DadFighterIdleFrameIndex returns a looping idle frame.
func DadFighterIdleFrameIndex(tick int) int {
	idle := [...]int{DadFighterIdle0, DadFighterIdle1, DadFighterIdleAlt}
	return idle[(tick/12)%len(idle)]
}
