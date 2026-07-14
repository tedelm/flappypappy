package game

import (
	"bytes"
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"

	"flappy/internal/game/sprite"
)

var titleFace *text.GoTextFace
var subtitleFace *text.GoTextFace
var labelFace *text.GoTextFace

func init() {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	titleFace = &text.GoTextFace{
		Source: source,
		Size:   48,
	}
	subtitleFace = &text.GoTextFace{
		Source: source,
		Size:   12,
	}
	labelFace = &text.GoTextFace{
		Source: source,
		Size:   18,
	}
}

const (
	difficultyLabelY    = 310
	difficultyGlassLift = 50
	difficultyTapPadX   = 8
	difficultyTapPadY   = 12
	glassTapPad         = 14
	restartBtnW         = 160
	restartBtnH         = 40
	restartBtnY         = ScreenH/2 + 24
	continueBtnW        = 160
	continueBtnH        = 40
	continueBtnY        = ScreenH/2 + 24
)

func difficultyLabelX(d Difficulty) float64 {
	positions := [3]float64{ScreenW * 0.2, ScreenW * 0.5, ScreenW * 0.8}
	return positions[d]
}

func difficultyColumnBounds(d Difficulty) (x, y, w, h float64) {
	cx := difficultyLabelX(d)
	w = ScreenW/3 - difficultyTapPadX
	h = difficultyGlassLift + 28 + difficultyTapPadY*2
	x = cx - w/2
	y = difficultyLabelY - difficultyGlassLift - BirdHeight/2 - difficultyTapPadY
	return x, y, w, h
}

func (g *Game) glassTapBounds() (x, y, w, h float64) {
	bx, by, bw, bh := g.bird.Bounds()
	return bx - glassTapPad, by - glassTapPad, bw + glassTapPad*2, bh + glassTapPad*2
}

func pointInRect(px, py, x, y, w, h float64) bool {
	return px >= x && px < x+w && py >= y && py < y+h
}

type State int

const (
	StateReady State = iota
	StatePlaying
	StateContinue
	StateGameOver
)

type Game struct {
	state            State
	rawScore         int
	difficulty       Difficulty
	lives            int
	invincibleFrames int
	bird             *Bird
	pipes            *PipeManager
	frames           int
	touchIDs         []ebiten.TouchID
}

func New() *Game {
	return &Game{
		state:      StateReady,
		difficulty: DifficultyEasy,
		bird:       NewBird(),
		pipes:      NewPipeManager(),
	}
}

func (g *Game) displayScore() int {
	return g.rawScore * g.difficulty.Config().ScoreMultiplier
}

func (g *Game) reset() {
	g.state = StateReady
	g.rawScore = 0
	g.frames = 0
	g.lives = 0
	g.invincibleFrames = 0
	g.bird.Reset()
	g.pipes.Reset()
}

func (g *Game) startGame() {
	cfg := g.difficulty.Config()
	g.bird.ApplyConfig(cfg)
	g.bird.Reset()
	g.pipes.ApplyConfig(cfg)
	g.pipes.Reset()
	g.lives = MaxLives
	g.invincibleFrames = 0
	g.state = StatePlaying
	g.bird.Flap()
}

func (g *Game) loseLife() {
	g.lives--
	if g.lives > 0 {
		g.state = StateContinue
	} else {
		g.state = StateGameOver
	}
}

func (g *Game) continueGame() {
	g.bird.Reset()
	g.bird.Flap()
	g.invincibleFrames = LifeInvincibleTicks
	g.state = StatePlaying
}

func (g *Game) flapInput() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return true
	}
	g.touchIDs = inpututil.AppendJustPressedTouchIDs(g.touchIDs[:0])
	return len(g.touchIDs) > 0
}

func (g *Game) cycleDifficultyPrev() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) ||
		inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.difficulty = g.difficulty.Prev()
	}
}

func (g *Game) cycleDifficultyNext() {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) ||
		inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.difficulty = g.difficulty.Next()
	}
}

func (g *Game) readyPointerJustPressed() (float64, float64, bool) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return float64(x), float64(y), true
	}
	g.touchIDs = inpututil.AppendJustPressedTouchIDs(g.touchIDs[:0])
	if len(g.touchIDs) > 0 {
		x, y := ebiten.TouchPosition(g.touchIDs[0])
		return float64(x), float64(y), true
	}
	return 0, 0, false
}

func difficultyAtPointer(px, py float64) (Difficulty, bool) {
	labelBandTop := float64(difficultyLabelY - difficultyTapPadY)
	for d := DifficultyEasy; d <= DifficultyInsane; d++ {
		x, y, w, h := difficultyColumnBounds(d)
		if py >= labelBandTop && pointInRect(px, py, x, y, w, h) {
			return d, true
		}
	}
	return 0, false
}

func (g *Game) readyStartInputAt(px, py float64) bool {
	gx, gy, gw, gh := g.glassTapBounds()
	return pointInRect(px, py, gx, gy, gw, gh)
}

func restartButtonBounds() (x, y, w, h float64) {
	return (ScreenW - restartBtnW) / 2, restartBtnY, restartBtnW, restartBtnH
}

func continueButtonBounds() (x, y, w, h float64) {
	return (ScreenW - continueBtnW) / 2, continueBtnY, continueBtnW, continueBtnH
}

func (g *Game) gameOverRestartInput() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	px, py, ok := g.readyPointerJustPressed()
	if !ok {
		return false
	}
	x, y, w, h := restartButtonBounds()
	return pointInRect(px, py, x, y, w, h)
}

func (g *Game) continueInput() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	px, py, ok := g.readyPointerJustPressed()
	if !ok {
		return false
	}
	x, y, w, h := continueButtonBounds()
	return pointInRect(px, py, x, y, w, h)
}

func (g *Game) Update() error {
	g.frames++

	switch g.state {
	case StateReady:
		g.bird.X = difficultyLabelX(g.difficulty)
		g.bird.Y = difficultyLabelY - difficultyGlassLift + sinBob(g.frames)*4
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.startGame()
		} else if px, py, ok := g.readyPointerJustPressed(); ok {
			if d, hit := difficultyAtPointer(px, py); hit {
				g.difficulty = d
			} else if g.readyStartInputAt(px, py) {
				g.startGame()
			}
		} else {
			g.cycleDifficultyPrev()
			g.cycleDifficultyNext()
		}

	case StatePlaying:
		g.bird.Update()
		g.pipes.Update()
		g.rawScore += g.pipes.CheckScore(g.bird.X)

		if g.invincibleFrames > 0 {
			g.invincibleFrames--
		} else {
			bx, by, bw, bh := g.bird.Bounds()
			if g.pipes.Collides(bx, by, bw, bh) || g.hitBounds(bx, by, bw, bh) {
				g.loseLife()
			}
		}

		if g.flapInput() {
			g.bird.Flap()
		}

	case StateContinue:
		if g.continueInput() {
			g.continueGame()
		}

	case StateGameOver:
		if g.gameOverRestartInput() {
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

	g.pipes.Draw(screen)
	drawPubForeground(screen)

	if g.state == StateReady {
		drawTitle(screen, "Flappy Pappy", ScreenW/2, 140)
		drawSubtitle(screen, "Brought to you by PappaPuben", ScreenW/2, ScreenH/2-100)
		drawDifficultySelector(screen, g.difficulty)
		g.bird.Draw(screen)
		ebitenutil.DebugPrintAt(screen, "Tap the beer to Start", ScreenW/2-70, ScreenH/2+90)
		ebitenutil.DebugPrintAt(screen, "Tap a level to choose", ScreenW/2-60, ScreenH/2+110)
	} else {
		g.bird.Draw(screen)
		switch g.state {
		case StateContinue:
			drawContinuePrompt(screen)
		case StateGameOver:
			ebitenutil.DebugPrintAt(screen, "GAME OVER", ScreenW/2-40, ScreenH/2-40)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.displayScore()), ScreenW/2-35, ScreenH/2-20)
			drawRestartButton(screen)
		}
	}

	if g.state == StatePlaying || g.state == StateContinue {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", g.displayScore()), ScreenW/2-10, 20)
	}

	if g.state == StatePlaying || g.state == StateContinue || g.state == StateGameOver {
		drawLifeHUD(screen, g.lives)
	}
}

func drawTitle(screen *ebiten.Image, str string, centerX, y float64) {
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignStart
	op.GeoM.Translate(centerX, y)
	op.ColorScale.ScaleWithColor(ColorKegRim)
	text.Draw(screen, str, titleFace, op)
}

func drawSubtitle(screen *ebiten.Image, str string, centerX, y float64) {
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignStart
	op.GeoM.Translate(centerX, y)
	op.ColorScale.ScaleWithColor(ColorKegRim)
	text.Draw(screen, str, subtitleFace, op)
}

/* func drawText(screen *ebiten.Image, str string, centerX, y float64) {
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignStart
	op.GeoM.Translate(centerX, y)
	op.ColorScale.ScaleWithColor(ColorKegRim)
	labelFace.Size = 18
	text.Draw(screen, str, labelFace, op)
} */

func drawDifficultySelector(screen *ebiten.Image, selected Difficulty) {
	labels := []string{DifficultyEasy.Name(), DifficultyHard.Name(), DifficultyInsane.Name()}

	for i, label := range labels {
		col := ColorKegEdge
		if Difficulty(i) == selected {
			col = ColorKegRim
		}
		drawLabel(screen, label, difficultyLabelX(Difficulty(i)), difficultyLabelY, col)
	}

	cfg := selected.Config()
	multLabel := fmt.Sprintf("%dx score", cfg.ScoreMultiplier)
	drawLabel(screen, multLabel, ScreenW/2, difficultyLabelY+28, ColorKegBand)
}

func drawLabel(screen *ebiten.Image, str string, centerX, y float64, col color.Color) {
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignStart
	op.GeoM.Translate(centerX, y)
	op.ColorScale.ScaleWithColor(col)
	text.Draw(screen, str, labelFace, op)
}

func drawLifeHUD(screen *ebiten.Image, lives int) {
	full := sprite.LifeFull()
	empty := sprite.LifeEmpty()
	scale := float64(LifeGlassW) / float64(sprite.LifeFrameWidth())

	totalW := LifeGlassW*MaxLives + LifeGlassGap*(MaxLives-1)
	baseX := ScreenW - LifeHUDMargin - totalW
	scaledH := int(float64(full.Bounds().Dy()) * scale)
	baseY := ScreenH - GroundHeight + (GroundHeight-scaledH)/2

	for i := 0; i < MaxLives; i++ {
		img := full
		if i >= lives {
			img = empty
		}
		x := baseX + i*(LifeGlassW+LifeGlassGap)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(x), float64(baseY))
		screen.DrawImage(img, op)
	}
}

func drawPubButton(screen *ebiten.Image, label string, x, y, w, h float64) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), ColorKegBand, true)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 2, ColorKegEdge, true)
	drawLabel(screen, label, x+w/2, y+h/2-9, ColorKegRim)
}

func drawContinuePrompt(screen *ebiten.Image) {
	drawLabel(screen, "You beer was spilled!", ScreenW/2, continueBtnY-28, ColorKegRim)
	x, y, w, h := continueButtonBounds()
	drawPubButton(screen, "CONTINUE", x, y, w, h)
}

func drawRestartButton(screen *ebiten.Image) {
	x, y, w, h := restartButtonBounds()
	drawPubButton(screen, "RESTART", x, y, w, h)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}
