package game

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"unicode"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"

	"flappy/internal/game/font"
	"flappy/internal/game/sound"
	"flappy/internal/game/sprite"
)

var titleFace *text.GoTextFace
var subtitleFace *text.GoTextFace
var labelFace *text.GoTextFace

func init() {
	titleSource, err := font.TitleSource()
	if err != nil {
		panic(err)
	}
	uiSource, err := text.NewGoTextFaceSource(bytes.NewReader(goregular.TTF))
	if err != nil {
		panic(err)
	}
	titleFace = &text.GoTextFace{
		Source: titleSource,
		Size:   48,
	}
	subtitleFace = &text.GoTextFace{
		Source: uiSource,
		Size:   12,
	}
	labelFace = &text.GoTextFace{
		Source: uiSource,
		Size:   18,
	}
}

const (
	difficultyLabelY    = 310
	difficultyGlassLift = 50
	difficultyTapPadX   = 8
	difficultyTapPadY   = 12
	glassTapPad         = 14
	continueBtnW        = 160
	continueBtnH        = 40
	continueBtnY        = ScreenH/2 + 24
	saveBtnW            = 160
	saveBtnH            = 40
	backBtnW            = 160
	backBtnH            = 40
	backBtnY            = ScreenH - GroundHeight - 56
	nameFieldW          = 200
	nameFieldH          = 36
	enterNameTitleY     = 110
	enterNameScoreY     = 175
	enterNameLabelY     = 250
	nameFieldY          = 285
	saveBtnY            = nameFieldY + nameFieldH + 24
	highScoresLinkY     = ScreenH/2 + 170
	highScoresLinkPadX  = 12
	highScoresLinkPadY  = 8
	highScoreTableY     = 118
	highScoreRowH       = 26
	highScoreColPlace   = 18
	highScoreColName    = 52
	highScoreColLevel   = 168
	highScoreColScore   = 228
	highScoreColDiff    = 318
	loadTimeoutFrames   = 600
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
	StateLoading State = iota
	StateReady
	StatePlaying
	StateBossFight
	StateLevelComplete
	StateContinue
	StateEnterName
	StateHighScores
)

type loadResult struct {
	manager *sound.Manager
	err     error
}

type Game struct {
	state               State
	rawScore            int
	level               int
	levelDadsPassed     int
	difficulty          Difficulty
	lives               int
	livesMax            int
	invincibleFrames    int
	playerName          string
	highScores          *HighScores
	bgScrollX           float64
	decorSeed           int
	bird                *Bird
	pipes               *PipeManager
	bossFight           *BossFight
	frames              int
	touchIDs            []ebiten.TouchID
	nameInputOpen       bool
	music               *sound.Manager
	musicMenu           bool
	loadStarted         bool
	loadTextSet         bool
	loadFrames          int
	loadDone            chan loadResult
	loadingScreenHidden bool
}

func New() *Game {
	return &Game{
		state:      StateLoading,
		difficulty: DifficultyEasy,
		bird:       NewBird(),
		pipes:      NewPipeManager(),
		bossFight:  NewBossFight(),
		highScores: NewHighScores(InitScoreStore()),
		livesMax:   MaxLives,
		loadDone:   make(chan loadResult, 1),
	}
}

func (g *Game) displayScore() int {
	return g.rawScore * g.difficulty.Config().ScoreMultiplier
}

func (g *Game) reset() {
	HideNameInput()
	g.nameInputOpen = false
	g.state = StateReady
	g.rawScore = 0
	g.level = 0
	g.levelDadsPassed = 0
	g.frames = 0
	g.lives = 0
	g.livesMax = MaxLives
	g.invincibleFrames = 0
	g.playerName = ""
	g.bgScrollX = 0
	g.decorSeed = 0
	g.bird.Reset()
	g.pipes.Reset()
}

func (g *Game) startGame() {
	cfg := g.difficulty.Config()
	g.bird.ApplyConfig(cfg)
	g.bird.Reset()
	g.pipes.ApplyConfig(cfg)
	g.pipes.Reset()
	g.level = 1
	g.levelDadsPassed = 0
	g.rawScore = 0
	g.lives = MaxLives
	g.livesMax = MaxLives
	g.invincibleFrames = 0
	g.playerName = ""
	g.bgScrollX = 0
	g.decorSeed = rand.Int()
	g.state = StatePlaying
	g.flap()
}

func (g *Game) enterBossFight() {
	g.pipes.Reset()
	g.bossFight.Reset()
	g.state = StateBossFight
}

func (g *Game) advanceToNextLevel() {
	g.lives, g.livesMax = awardLevelCompleteLife(g.lives, g.livesMax)
	g.level++
	g.levelDadsPassed = 0
	g.bird.Reset()
	g.pipes.Reset()
	g.state = StatePlaying
	g.flap()
}

func awardLevelCompleteLife(lives, livesMax int) (newLives, newLivesMax int) {
	// If the player still has missing hearts, they gain one without changing the max.
	if lives < livesMax {
		return lives + 1, livesMax
	}

	// Otherwise, they're full. Increase max lives and fill to it, until the cap.
	if livesMax < MaxLivesCap {
		return livesMax + 1, livesMax + 1
	}

	return lives, livesMax
}

func (g *Game) bossGameOver() {
	if g.music != nil {
		g.music.PlayGameOver()
	}
	g.playerName = ""
	g.state = StateEnterName
	g.nameInputOpen = true
}

func (g *Game) flap() {
	g.bird.Flap()
	if g.music != nil {
		g.music.PlayJump()
	}
}

func (g *Game) loseLife() {
	if g.music != nil {
		g.music.PlayLifeLost()
	}
	g.lives--
	if g.lives > 0 {
		g.state = StateContinue
	} else {
		if g.music != nil {
			g.music.PlayGameOver()
		}
		g.playerName = ""
		g.state = StateEnterName
		g.nameInputOpen = true
	}
}

func (g *Game) continueGame() {
	g.bird.Reset()
	g.flap()
	g.invincibleFrames = LifeInvincibleTicks
	g.state = StatePlaying
}

func (g *Game) submitHighScore() {
	SyncNameInput(&g.playerName)
	g.highScores.Add(g.playerName, g.displayScore(), g.difficulty, g.level)
	g.reset()
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

func continueButtonBounds() (x, y, w, h float64) {
	return (ScreenW - continueBtnW) / 2, continueBtnY, continueBtnW, continueBtnH
}

func saveButtonBounds() (x, y, w, h float64) {
	return (ScreenW - saveBtnW) / 2, saveBtnY, saveBtnW, saveBtnH
}

func backButtonBounds() (x, y, w, h float64) {
	return (ScreenW - backBtnW) / 2, backBtnY, backBtnW, backBtnH
}

func highScoresLinkBounds() (x, y, w, h float64) {
	w = 180
	h = 18 + highScoresLinkPadY*2
	x = (ScreenW - w) / 2
	y = highScoresLinkY - highScoresLinkPadY
	return x, y, w, h
}

func nameFieldBounds() (x, y, w, h float64) {
	return (ScreenW - nameFieldW) / 2, nameFieldY, nameFieldW, nameFieldH
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

func (g *Game) levelCompleteInput() bool {
	return g.continueInput()
}

func (g *Game) saveHighScoreInput() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return true
	}
	px, py, ok := g.readyPointerJustPressed()
	if !ok {
		return false
	}
	x, y, w, h := saveButtonBounds()
	return pointInRect(px, py, x, y, w, h)
}

func (g *Game) highScoresBackInput() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return true
	}
	px, py, ok := g.readyPointerJustPressed()
	if !ok {
		return false
	}
	x, y, w, h := backButtonBounds()
	return pointInRect(px, py, x, y, w, h)
}

func (g *Game) appendNameInput(r rune) {
	if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
		return
	}
	if len([]rune(g.playerName)) >= MaxPlayerNameLen {
		return
	}
	g.playerName += string(unicode.ToUpper(r))
}

func (g *Game) backspaceNameInput() {
	if g.playerName == "" {
		return
	}
	_, size := utf8.DecodeLastRuneInString(g.playerName)
	g.playerName = g.playerName[:len(g.playerName)-size]
}

func (g *Game) updateNameInput() {
	for _, r := range ebiten.AppendInputChars(nil) {
		g.appendNameInput(r)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		g.backspaceNameInput()
	}
}

func (g *Game) updateLoading() {
	g.loadFrames++
	if !g.loadTextSet {
		SetLoadingText("Loading audio...")
		g.loadTextSet = true
	}
	if !g.loadStarted {
		g.loadStarted = true
		go func() {
			manager, err := sound.NewManager()
			g.loadDone <- loadResult{manager: manager, err: err}
		}()
	}
	select {
	case res := <-g.loadDone:
		if res.err != nil {
			log.Printf("music disabled: %v", res.err)
		} else {
			g.music = res.manager
		}
		g.state = StateReady
	default:
		if g.loadFrames >= loadTimeoutFrames {
			log.Printf("audio load timeout after %d frames; starting without music", g.loadFrames)
			g.state = StateReady
		}
	}
}

func (g *Game) Update() error {
	g.frames++
	g.highScores.PollBridge()
	g.highScores.PollRefresh()

	switch g.state {
	case StateLoading:
		g.updateLoading()

	case StateReady:
		g.bird.X = difficultyLabelX(g.difficulty)
		g.bird.Y = difficultyLabelY - difficultyGlassLift + sinBob(g.frames)*4
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.startGame()
		} else if px, py, ok := g.readyPointerJustPressed(); ok {
			lx, ly, lw, lh := highScoresLinkBounds()
			if pointInRect(px, py, lx, ly, lw, lh) {
				g.state = StateHighScores
				g.highScores.RequestRefresh()
			} else if d, hit := difficultyAtPointer(px, py); hit {
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
		g.bgScrollX += g.pipes.Speed() * BgParallaxFactor
		if delta := g.pipes.CheckScore(g.bird.X); delta > 0 {
			g.rawScore += delta
			g.levelDadsPassed += delta
			if g.music != nil {
				for i := 0; i < delta; i++ {
					g.music.PlayPassDad()
				}
			}
			if g.levelDadsPassed >= DadsRequiredForLevel(g.level) {
				g.enterBossFight()
			}
		}

		if g.invincibleFrames > 0 {
			g.invincibleFrames--
		} else {
			bx, by, bw, bh := g.bird.Bounds()
			if g.pipes.Collides(bx, by, bw, bh) || g.hitBounds(bx, by, bw, bh) {
				g.loseLife()
			}
		}

		if g.flapInput() {
			g.flap()
		}

	case StateBossFight:
		won, lost := g.bossFight.Update()
		if won {
			g.state = StateLevelComplete
		} else if lost {
			g.bossGameOver()
		} else if g.flapInput() {
			if g.bossFight.CanThrow() {
				g.bossFight.Throw()
				if g.music != nil {
					g.music.PlayJump()
				}
			}
		}

	case StateLevelComplete:
		if g.levelCompleteInput() {
			g.advanceToNextLevel()
		}

	case StateContinue:
		if g.continueInput() {
			g.continueGame()
		}

	case StateEnterName:
		if g.nameInputOpen {
			x, y, w, h := nameFieldBounds()
			ShowNameInput(g.playerName, x, y, w, h, MaxPlayerNameLen)
			g.nameInputOpen = false
		}
		SyncNameInput(&g.playerName)
		g.updateNameInput()
		if px, py, ok := g.readyPointerJustPressed(); ok {
			fx, fy, fw, fh := nameFieldBounds()
			if pointInRect(px, py, fx, fy, fw, fh) {
				FocusNameInput()
			}
		}
		if g.saveHighScoreInput() {
			g.submitHighScore()
		}

	case StateHighScores:
		if g.highScoresBackInput() {
			g.state = StateReady
		}
	}

	g.syncMusicForState()
	return nil
}

func (g *Game) syncMusicForState() {
	if g.music == nil {
		return
	}
	menu := g.state == StateReady || g.state == StateEnterName || g.state == StateHighScores
	if menu == g.musicMenu {
		return
	}
	g.music.SetMode(menu)
	g.musicMenu = menu
}

func (g *Game) hitBounds(bx, by, bw, bh float64) bool {
	if by < 0 {
		return true
	}
	groundY := float64(FloorSurfaceY)
	return by+bh > groundY
}

func sinBob(frame int) float64 {
	return math.Sin(float64(frame) * 0.08)
}

func (g *Game) Draw(screen *ebiten.Image) {
	drawPubBackground(screen, g.bgScrollX, g.decorSeed, WallpaperIndex(g.level), g.level)

	showPipes := g.state == StatePlaying || g.state == StateContinue
	if g.state != StateLoading {
		if showPipes {
			g.pipes.DrawLamps(screen)
		}
		drawWoodenFloor(screen)
		drawPubStools(screen, g.bgScrollX, g.decorSeed)
		if showPipes {
			g.pipes.DrawDads(screen)
		}
	}

	switch g.state {
	case StateLoading:
		drawTitle(screen, "Pouring beer...", ScreenW/2, ScreenH/2-20)
		drawLabel(screen, "Loading audio...", ScreenW/2, ScreenH/2+20, ColorText)

	case StateReady:
		drawTitle(screen, "Flappy Pappy", ScreenW/2, 140)
		drawSubtitle(screen, "powered by PappaPuben", ScreenW/2, ScreenH/2-100)
		drawDifficultySelector(screen, g.difficulty)
		g.bird.Draw(screen)
		drawLabel(screen, "Tap the beer to Start", ScreenW/2, ScreenH/2+90, ColorText)
		drawLabel(screen, "Tap a level to choose", ScreenW/2, ScreenH/2+110, ColorText)
		drawHighScoresLink(screen)
		if !g.loadingScreenHidden {
			HideLoadingScreen()
			g.loadingScreenHidden = true
		}

	case StateHighScores:
		drawHighScoresScreen(screen, g.highScores)

	case StateBossFight:
		drawBossPlayer(screen)
		g.bossFight.Draw(screen)
		drawGameplayHUD(screen, g)
		drawBossHUD(screen, g.bossFight)

	case StateLevelComplete:
		g.bird.Draw(screen)
		drawGameplayHUD(screen, g)
		drawLevelCompletePrompt(screen, g)

	default:
		g.bird.Draw(screen)
		switch g.state {
		case StateContinue:
			drawContinuePrompt(screen)
		case StateEnterName:
			drawEnterNameScreen(screen, g)
		}
	}

	if g.state == StatePlaying || g.state == StateContinue {
		drawGameplayHUD(screen, g)
	}
}

func drawTitle(screen *ebiten.Image, str string, centerX, y float64) {
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignStart
	op.GeoM.Translate(centerX, y)
	op.ColorScale.ScaleWithColor(ColorText)
	text.Draw(screen, str, titleFace, op)
}

func drawSubtitle(screen *ebiten.Image, str string, centerX, y float64) {
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignStart
	op.GeoM.Translate(centerX, y)
	op.ColorScale.ScaleWithColor(ColorText)
	text.Draw(screen, str, subtitleFace, op)
}

func drawDifficultySelector(screen *ebiten.Image, selected Difficulty) {
	labels := []string{DifficultyEasy.Name(), DifficultyHard.Name(), DifficultyInsane.Name()}

	for i, label := range labels {
		col := ColorTextMuted
		if Difficulty(i) == selected {
			col = ColorText
		}
		drawLabel(screen, label, difficultyLabelX(Difficulty(i)), difficultyLabelY, col)
	}

	cfg := selected.Config()
	multLabel := fmt.Sprintf("%dx score", cfg.ScoreMultiplier)
	drawLabel(screen, multLabel, ScreenW/2, difficultyLabelY+28, ColorText)
}

func drawLabel(screen *ebiten.Image, str string, centerX, y float64, col color.Color) {
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignCenter
	op.SecondaryAlign = text.AlignStart
	op.GeoM.Translate(centerX, y)
	op.ColorScale.ScaleWithColor(col)
	text.Draw(screen, str, labelFace, op)
}

func drawLabelLeft(screen *ebiten.Image, str string, leftX, y float64, col color.Color) {
	op := &text.DrawOptions{}
	op.PrimaryAlign = text.AlignStart
	op.SecondaryAlign = text.AlignStart
	op.GeoM.Translate(leftX, y)
	op.ColorScale.ScaleWithColor(col)
	text.Draw(screen, str, labelFace, op)
}

func drawLifeHUD(screen *ebiten.Image, lives, livesMax int) {
	full := sprite.LifeFull()
	empty := sprite.LifeEmpty()
	scale := float64(LifeGlassW) / float64(sprite.LifeFrameWidth())

	totalW := LifeGlassW*livesMax + LifeGlassGap*(livesMax-1)
	baseX := ScreenW - LifeHUDMargin - totalW
	scaledH := int(float64(full.Bounds().Dy()) * scale)
	baseY := ScreenH - GroundHeight + (GroundHeight-scaledH)/2

	for i := 0; i < livesMax; i++ {
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
	drawLabel(screen, label, x+w/2, y+h/2-9, ColorText)
}

func drawGameplayHUD(screen *ebiten.Image, g *Game) {
	drawLabel(screen, fmt.Sprintf("%d", g.displayScore()), ScreenW/2, 20, ColorText)
	drawLabelLeft(screen, fmt.Sprintf("Level %d", g.level), LifeHUDMargin, 20, ColorText)
	target := DadsRequiredForLevel(g.level)
	drawLabel(screen, fmt.Sprintf("%d/%d", g.levelDadsPassed, target), ScreenW/2, 44, ColorTextMuted)
	drawLifeHUD(screen, g.lives, g.livesMax)
}

func drawBossHUD(screen *ebiten.Image, bf *BossFight) {
	drawLabel(screen, fmt.Sprintf("Hits: %d/%d", bf.hits, BossHitsRequired), ScreenW/2, 68, ColorText)
	drawLabel(screen, fmt.Sprintf("Throws: %d/%d", bf.throwsUsed, BossThrowsAllowed), ScreenW/2, 92, ColorTextMuted)
}

func drawLevelCompletePrompt(screen *ebiten.Image, g *Game) {
	drawLabel(screen, fmt.Sprintf("Level %d complete!", g.level), ScreenW/2, continueBtnY-28, ColorText)
	x, y, w, h := continueButtonBounds()
	drawPubButton(screen, "CONTINUE", x, y, w, h)
}

func drawContinuePrompt(screen *ebiten.Image) {
	drawLabel(screen, "Ah, no! Your beer is a memory now!", ScreenW/2, continueBtnY-28, ColorText)
	x, y, w, h := continueButtonBounds()
	drawPubButton(screen, "CONTINUE", x, y, w, h)
}

func drawHighScoresLink(screen *ebiten.Image) {
	drawLabel(screen, "<< HIGH SCORES >>", ScreenW/2, highScoresLinkY, ColorText)
}

func drawEnterNameScreen(screen *ebiten.Image, g *Game) {
	drawTitle(screen, "GAME OVER", ScreenW/2, enterNameTitleY)
	drawTitle(screen, fmt.Sprintf("Score: %d", g.displayScore()), ScreenW/2, enterNameScoreY)
	drawLabel(screen, fmt.Sprintf("Level reached: %d", g.level), ScreenW/2, enterNameScoreY+40, ColorTextMuted)
	drawLabel(screen, "Enter your name", ScreenW/2, enterNameLabelY, ColorText)

	x, y, w, h := nameFieldBounds()
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), ColorGlass, true)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 2, ColorGlassEdge, true)

	display := g.playerName
	if g.frames%60 < 30 {
		display += "_"
	}
	drawLabel(screen, display, ScreenW/2, y+9, ColorText)

	sx, sy, sw, sh := saveButtonBounds()
	drawPubButton(screen, "SAVE", sx, sy, sw, sh)
}

func drawHighScoresScreen(screen *ebiten.Image, scores *HighScores) {
	drawTitle(screen, "HIGH SCORES", ScreenW/2, 80)

	if scores.ConnectFailed() {
		drawLabel(screen, "Could not reach leaderboard", ScreenW/2, 180, ColorTextMuted)
	} else if !scores.Active() {
		drawLabel(screen, "Scores not synced", ScreenW/2, 160, ColorTextMuted)
		drawLabel(screen, "Set FLAPPY_SQLITECLOUD_URL", ScreenW/2, 190, ColorTextMuted)
		drawLabel(screen, "or web/config.js", ScreenW/2, 218, ColorTextMuted)
	} else {
		entries := scores.List()
		if scores.Loading() {
			drawLabel(screen, "Loading...", ScreenW/2, 180, ColorTextMuted)
		} else if len(entries) == 0 {
			drawLabel(screen, "No scores yet", ScreenW/2, 180, ColorTextMuted)
		} else {
			headerY := float64(highScoreTableY)
			drawLabelLeft(screen, "#", highScoreColPlace, headerY, ColorTextMuted)
			drawLabelLeft(screen, "NAME", highScoreColName, headerY, ColorTextMuted)
			drawLabelLeft(screen, "LVL", highScoreColLevel, headerY, ColorTextMuted)
			drawLabelLeft(screen, "SCORE", highScoreColScore, headerY, ColorTextMuted)
			drawLabelLeft(screen, "DIFF", highScoreColDiff, headerY, ColorTextMuted)

			for i, e := range entries {
				y := headerY + float64(highScoreRowH) + float64(i)*float64(highScoreRowH)
				drawLabelLeft(screen, fmt.Sprintf("%d", i+1), highScoreColPlace, y, ColorText)
				drawLabelLeft(screen, e.Name, highScoreColName, y, ColorText)
				drawLabelLeft(screen, fmt.Sprintf("%d", e.Level), highScoreColLevel, y, ColorText)
				drawLabelLeft(screen, fmt.Sprintf("%d", e.Score), highScoreColScore, y, ColorText)
				drawLabelLeft(screen, e.Difficulty.Name(), highScoreColDiff, y, ColorText)
			}
		}
	}

	bx, by, bw, bh := backButtonBounds()
	drawPubButton(screen, "BACK", bx, by, bw, bh)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}
