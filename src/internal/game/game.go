package game

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type State int

const (
	StateReady State = iota
	StatePlaying
	StateGameOver
)

type Game struct {
	state    State
	score    int
	bird     *Bird
	pipes    *PipeManager
	frames   int
	touchIDs []ebiten.TouchID
}

func New() *Game {
	return &Game{
		state: StateReady,
		bird:  NewBird(),
		pipes: NewPipeManager(),
	}
}

func (g *Game) reset() {
	g.state = StateReady
	g.score = 0
	g.frames = 0
	g.bird.Reset()
	g.pipes.Reset()
}

func (g *Game) flapInput() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return true
	}
	g.touchIDs = inpututil.AppendJustPressedTouchIDs(g.touchIDs[:0])
	return len(g.touchIDs) > 0
}

func (g *Game) Update() error {
	g.frames++

	switch g.state {
	case StateReady:
		g.bird.Y = float64(ScreenH-GroundHeight)/2 + float64(sinBob(g.frames)*4)
		if g.flapInput() {
			g.state = StatePlaying
			g.bird.Flap()
		}

	case StatePlaying:
		g.bird.Update()
		g.pipes.Update()
		g.score += g.pipes.CheckScore(g.bird.X)

		bx, by, bw, bh := g.bird.Bounds()
		if g.pipes.Collides(bx, by, bw, bh) || g.hitBounds(bx, by, bw, bh) {
			g.state = StateGameOver
		}

		if g.flapInput() {
			g.bird.Flap()
		}

	case StateGameOver:
		if g.flapInput() {
			g.reset()
		}
	}

	return nil
}

func (g *Game) hitBounds(bx, by, bw, bh float64) bool {
	if by < 0 {
		return true
	}
	groundY := float64(ScreenH - GroundHeight)
	return by+bh > groundY
}

func sinBob(frame int) float64 {
	return math.Sin(float64(frame) * 0.08)
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(ColorSky)

	drawFoamGround(screen)

	g.pipes.Draw(screen)
	g.bird.Draw(screen)

	switch g.state {
	case StateReady:
		ebitenutil.DebugPrintAt(screen, "Tap or press SPACE to Start", ScreenW/2-110, ScreenH/2-40)
	case StateGameOver:
		ebitenutil.DebugPrintAt(screen, "GAME OVER", ScreenW/2-40, ScreenH/2-40)
		ebitenutil.DebugPrintAt(screen, "Tap or press SPACE to Restart", ScreenW/2-120, ScreenH/2-10)
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", g.score), ScreenW/2-10, 20)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

func drawFoamGround(screen *ebiten.Image) {
	groundY := float32(ScreenH - GroundHeight)
	vector.DrawFilledRect(screen, 0, groundY, ScreenW, GroundHeight, ColorFoam, true)

	waveRadii := []float32{12, 10, 14, 9, 13, 11, 10, 14, 12, 9, 13, 10, 14, 11, 12, 10, 13, 9}
	for i, r := range waveRadii {
		cx := float32(i*23 + 8)
		if cx > ScreenW+14 {
			break
		}
		vector.DrawFilledCircle(screen, cx, groundY+2, r, ColorFoamLight, true)
	}

	bubbleData := [][3]float32{
		{18, 28, 3}, {45, 45, 2}, {72, 22, 4}, {98, 52, 2}, {125, 35, 3},
		{158, 48, 2}, {185, 25, 3}, {212, 55, 2}, {245, 38, 4}, {278, 20, 2},
		{305, 50, 3}, {332, 30, 2}, {358, 42, 3}, {30, 60, 2}, {110, 65, 3},
		{200, 58, 2}, {290, 62, 3}, {370, 55, 2},
	}
	for _, b := range bubbleData {
		vector.DrawFilledCircle(screen, b[0], groundY+b[1], b[2], ColorBubble, true)
	}
}
