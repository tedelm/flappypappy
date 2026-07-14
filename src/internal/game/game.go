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
	state  State
	score  int
	bird   *Bird
	pipes  *PipeManager
	frames int
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
	return inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
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

	groundY := float32(ScreenH - GroundHeight)
	vector.DrawFilledRect(screen, 0, groundY, ScreenW, GroundHeight, ColorGround, true)
	vector.StrokeLine(screen, 0, groundY, ScreenW, groundY, 2, ColorPipeEdge, true)

	g.pipes.Draw(screen)
	g.bird.Draw(screen)

	switch g.state {
	case StateReady:
		ebitenutil.DebugPrintAt(screen, "Press SPACE or Click to Start", ScreenW/2-110, ScreenH/2-40)
	case StateGameOver:
		ebitenutil.DebugPrintAt(screen, "GAME OVER", ScreenW/2-40, ScreenH/2-40)
		ebitenutil.DebugPrintAt(screen, "Press SPACE or Click to Restart", ScreenW/2-120, ScreenH/2-10)
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", g.score), ScreenW/2-10, 20)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}
