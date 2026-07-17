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
	meetBossBtnW        = 280
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
	loadTimeoutFrames      = 600
	levelTransitionFrames  = 180 // ~3s at 60 TPS
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
	StateLevelIntro
	StatePlaying
	StateLevelExit
	StateBossIntro
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
	finalDadAttempt     bool
	bossContinue        bool
	difficulty          Difficulty
	lives               int
	livesMax            int
	invincibleFrames    int
	playerName          string
	highScores          *HighScores
	bgScrollX           float64
	decorSeed           int
	transitionFrames    int
	bird                *Bird
	pipes               *PipeManager
	bossFight           *BossFight
	frames              int
	touchIDs            []ebiten.TouchID
	throwCharging       bool
	throwChargeFrames   int
	throwChargeTouch    ebiten.TouchID
	throwChargeHasTouch bool
	nameInputOpen       bool
	music               *sound.Manager
	musicMode           sound.MusicMode
	musicModeSet        bool
	loadStarted         bool
	loadTextSet         bool
	loadFrames          int
	loadDone            chan loadResult
	loadingScreenHidden bool
	continueText        string
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
	g.finalDadAttempt = false
	g.bossContinue = false
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
	g.finalDadAttempt = false
	g.bossContinue = false
	g.pipes.SetSpawnLimit(DadsRequiredForLevel(g.level))
	g.rawScore = 0
	g.lives = MaxLives
	g.livesMax = MaxLives
	g.invincibleFrames = 0
	g.playerName = ""
	g.bgScrollX = 0
	g.decorSeed = rand.Int()
	g.beginLevelIntro()
}

func (g *Game) applyDebugStart() {
	if DebugStartLevel <= 0 {
		return
	}

	cfg := g.difficulty.Config()
	g.bird.ApplyConfig(cfg)
	g.bird.Reset()
	g.pipes.ApplyConfig(cfg)
	g.pipes.Reset()
	g.level = DebugStartLevel
	g.levelDadsPassed = 0
	g.finalDadAttempt = false
	g.bossContinue = false
	g.rawScore = 0
	g.lives = MaxLives
	g.livesMax = MaxLives
	g.invincibleFrames = 0
	g.playerName = ""
	g.bgScrollX = 0
	g.decorSeed = rand.Int()

	if DebugStartBoss != 0 {
		g.enterBossIntro()
		return
	}

	g.pipes.SetSpawnLimit(DadsRequiredForLevel(g.level))
	g.state = StatePlaying
	g.flap()
}

func (g *Game) beginPlaying() {
	g.bird.Reset()
	g.state = StatePlaying
	g.flap()
}

func (g *Game) beginLevelIntro() {
	g.transitionFrames = 0
	g.bird.X = -g.bird.Width
	g.bird.Y = float64(ScreenH-GroundHeight) / 2
	g.bird.VY = 0
	g.state = StateLevelIntro
}

func (g *Game) updateLevelIntro() {
	g.transitionFrames++
	t := float64(g.transitionFrames) / float64(levelTransitionFrames)
	if t > 1 {
		t = 1
	}
	startX := -g.bird.Width
	g.bird.X = startX + (BirdStartX-startX)*t
	g.bird.Y = float64(ScreenH-GroundHeight)/2 + sinBob(g.transitionFrames)*8
	g.bird.VY = 0
	if g.transitionFrames >= levelTransitionFrames {
		g.beginPlaying()
	}
}

func (g *Game) enterBossIntro() {
	g.bird.Reset()
	g.transitionFrames = 0
	g.state = StateBossIntro
}

func (g *Game) beginLevelExit() {
	g.transitionFrames = 0
	g.bird.VY = 0
	g.state = StateLevelExit
}

func (g *Game) updateLevelExit() {
	g.transitionFrames++
	g.pipes.Update()
	g.bgScrollX += g.pipes.Speed() * BgParallaxFactor
	endX := float64(ScreenW) + g.bird.Width
	t := float64(g.transitionFrames) / float64(levelTransitionFrames)
	if t > 1 {
		t = 1
	}
	g.bird.X = BirdStartX + (endX-BirdStartX)*t
	g.bird.Y = float64(ScreenH-GroundHeight)/2 + sinBob(g.transitionFrames)*8
	g.bird.VY = 0
	if g.transitionFrames >= levelTransitionFrames {
		g.enterBossIntro()
	}
}

func (g *Game) enterBossFight() {
	g.pipes.Reset()
	g.bossFight.Reset(g.level)
	g.clearThrowCharge()
	g.bossContinue = false
	g.state = StateBossFight
}

func (g *Game) clearThrowCharge() {
	g.throwCharging = false
	g.throwChargeFrames = 0
	g.throwChargeHasTouch = false
}

func (g *Game) throwChargePower() float64 {
	if !g.throwCharging {
		return 0
	}
	return ThrowPower(g.throwChargeFrames)
}

// updateThrowCharge handles hold-to-charge for boss throws.
// Returns true and power when the player releases a throw.
func (g *Game) updateThrowCharge() (released bool, power float64) {
	if !g.bossFight.CanThrow() {
		g.clearThrowCharge()
		return false, 0
	}

	if !g.throwCharging {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.throwCharging = true
			g.throwChargeFrames = 0
			g.throwChargeHasTouch = false
			return false, 0
		}
		g.touchIDs = inpututil.AppendJustPressedTouchIDs(g.touchIDs[:0])
		if len(g.touchIDs) > 0 {
			g.throwCharging = true
			g.throwChargeFrames = 0
			g.throwChargeTouch = g.touchIDs[0]
			g.throwChargeHasTouch = true
		}
		return false, 0
	}

	held := false
	if g.throwChargeHasTouch {
		g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
		for _, id := range g.touchIDs {
			if id == g.throwChargeTouch {
				held = true
				break
			}
		}
	} else {
		held = ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	}

	if held {
		if g.throwChargeFrames < ThrowChargeMaxFrames() {
			g.throwChargeFrames++
		}
		return false, 0
	}

	power = ThrowPower(g.throwChargeFrames)
	g.clearThrowCharge()
	return true, power
}

func (g *Game) advanceToNextLevel() {
	oldLives, oldLivesMax := g.lives, g.livesMax
	g.lives, g.livesMax = awardLevelCompleteLife(g.lives, g.livesMax)
	if g.music != nil && (g.lives > oldLives || g.livesMax > oldLivesMax) {
		g.music.PlayPowerUp()
	}
	g.level++
	g.levelDadsPassed = 0
	g.finalDadAttempt = false
	g.bird.Reset()
	g.pipes.Reset()
	g.pipes.SetSpawnLimit(DadsRequiredForLevel(g.level))
	g.beginLevelIntro()
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
	g.clearThrowCharge()
	if g.music != nil {
		g.music.PlayLifeLost()
		g.music.PlayLosingHorn()
	}
	g.lives--
	if g.lives > 0 {
		g.bossContinue = true
		g.continueText = randomText()
		g.state = StateContinue
		return
	}
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
	target := DadsRequiredForLevel(g.level)
	g.finalDadAttempt = g.levelDadsPassed >= target-1
	g.lives--
	if g.lives > 0 {
		g.continueText = randomText()
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
	if g.bossContinue {
		if g.rawScore >= g.levelDadsPassed {
			g.rawScore -= g.levelDadsPassed
		} else {
			g.rawScore = 0
		}
		g.levelDadsPassed = 0
		g.finalDadAttempt = false
		g.bossContinue = false
		g.pipes.Reset()
		g.pipes.SetSpawnLimit(DadsRequiredForLevel(g.level))
	} else if g.finalDadAttempt {
		target := DadsRequiredForLevel(g.level)
		if g.levelDadsPassed >= target && g.rawScore > 0 {
			g.rawScore--
		}
		g.levelDadsPassed = target - 1
		g.pipes.SetSpawnLimit(target)
		g.pipes.ResetBeforeLastDad()
		g.finalDadAttempt = false
	}
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

func (g *Game) inBossScene() bool {
	return g.state == StateBossIntro || g.state == StateBossFight
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

func meetBossButtonBounds() (x, y, w, h float64) {
	return (ScreenW - meetBossBtnW) / 2, continueBtnY, meetBossBtnW, continueBtnH
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

func (g *Game) meetBossInput() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	px, py, ok := g.readyPointerJustPressed()
	if !ok {
		return false
	}
	x, y, w, h := meetBossButtonBounds()
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
		g.applyDebugStart()
	default:
		if g.loadFrames >= loadTimeoutFrames {
			log.Printf("audio load timeout after %d frames; starting without music", g.loadFrames)
			g.state = StateReady
			g.applyDebugStart()
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

	case StateLevelIntro:
		g.updateLevelIntro()

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
		}

		if g.invincibleFrames > 0 {
			g.invincibleFrames--
		} else {
			bx, by, bw, bh := g.bird.Bounds()
			if g.pipes.Collides(bx, by, bw, bh, DadVariantForLevel(g.level, g.decorSeed)) || g.hitBounds(bx, by, bw, bh) {
				g.loseLife()
			}
		}

		if g.state == StatePlaying {
			target := DadsRequiredForLevel(g.level)
			if g.levelDadsPassed >= target && g.pipes.LastDadFullyCleared(g.bird.X) {
				g.beginLevelExit()
			}
		}

		if g.flapInput() {
			g.flap()
		}

	case StateLevelExit:
		g.updateLevelExit()

	case StateBossIntro:
		if g.meetBossInput() {
			g.enterBossFight()
		}

	case StateBossFight:
		if g.bossFight.BoxingVictoryReady() {
			g.bossFight.Update() // keep win pose anim looping
			if g.levelCompleteInput() {
				g.clearThrowCharge()
				g.advanceToNextLevel()
			}
			break
		}
		if g.bossFight.CanDodge() && g.flapInput() {
			g.bossFight.DuelFlap()
			if g.music != nil {
				g.music.PlayJump()
			}
		}
		won, lost := g.bossFight.Update()
		if g.bossFight.ConsumeHitSFX() && g.music != nil {
			if g.bossFight.IsBoxing() {
				g.music.PlayNeighbourPunch()
			} else {
				g.music.PlayGlassBreak()
				g.music.PlayWrongWithYou()
			}
		}
		if g.bossFight.ConsumeExplodeSFX() && g.music != nil {
			g.music.PlayPlayerWin()
			g.music.PlayYouWin()
		}
		if g.bossFight.ConsumeLoseSFX() && g.music != nil {
			g.music.PlayPlayerPunchLose()
		}
		if g.bossFight.ConsumeLaughSFX() && g.music != nil {
			g.music.PlayLaughingRun()
		}
		if g.bossFight.ConsumeLifeLost() {
			if g.music != nil {
				g.music.PlayLifeLost()
			}
			g.lives--
			if g.lives <= 0 {
				g.clearThrowCharge()
				if g.music != nil {
					g.music.PlayGameOver()
				}
				g.playerName = ""
				g.state = StateEnterName
				g.nameInputOpen = true
			}
		}
		if won {
			g.clearThrowCharge()
			g.state = StateLevelComplete
		} else if lost {
			g.clearThrowCharge()
			g.bossGameOver()
		} else if g.bossFight.IsOutrun() || g.bossFight.IsBoxing() {
			if !g.bossFight.InCountdown() && g.flapInput() {
				g.bossFight.Tap()
				if g.music != nil {
					g.music.PlayFootstep()
				}
			}
		} else if released, power := g.updateThrowCharge(); released {
			if g.bossFight.CanThrow() {
				g.bossFight.Throw(power)
				if g.music != nil {
					g.music.PlayOuch()
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
	var mode sound.MusicMode
	switch {
	case g.state == StateReady || g.state == StateEnterName || g.state == StateHighScores:
		mode = sound.MusicMenu
	case g.inBossScene():
		mode = sound.MusicBoss
	default:
		mode = sound.MusicGame
	}
	if g.musicModeSet && mode == g.musicMode {
		return
	}
	g.music.SetMode(mode)
	g.musicMode = mode
	g.musicModeSet = true
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
	if g.inBossScene() {
		scroll := 0.0
		if g.state == StateBossFight && g.bossFight.IsOutrun() {
			scroll = g.bossFight.ScrollX()
		}
		drawBossBackground(screen, scroll, g.level)
	} else {
		drawPubBackground(screen, g.bgScrollX, g.decorSeed, WallpaperIndex(g.level), g.level)
	}

	showPipes := g.state == StatePlaying || g.state == StateContinue || g.state == StateLevelExit
	if g.state != StateLoading && !g.inBossScene() {
		if showPipes {
			g.pipes.DrawLamps(screen)
		}
		drawWoodenFloor(screen)
		drawPubStools(screen, g.bgScrollX, g.decorSeed)
		if showPipes {
			g.pipes.DrawDads(screen, DadVariantForLevel(g.level, g.decorSeed))
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

	case StateLevelIntro:
		g.bird.Draw(screen)

	case StateBossIntro:
		g.bird.Draw(screen)
		drawGameplayHUD(screen, g)
		drawBossIntroPrompt(screen, g.level)

	case StateBossFight:
		if g.bossFight.IsOutrun() {
			drawOutrunPlayer(screen, g.bossFight)
		} else if g.bossFight.IsBoxing() {
			drawBoxingPlayer(screen, g.bossFight)
		} else if g.bossFight.IsDuel() {
			drawDuelPlayer(screen, g.bossFight, g.throwChargePower())
		} else {
			drawBossPlayer(screen, g.throwChargePower())
		}
		g.bossFight.Draw(screen)
		drawGameplayHUD(screen, g)
		drawBossHUD(screen, g.bossFight, g.throwChargePower(), g.throwCharging)
		if g.bossFight.InCountdown() {
			drawTitle(screen, g.bossFight.CountdownDisplay(), ScreenW/2, float64(ScreenH)/2-40)
		}
		if g.bossFight.pendingWin || g.bossFight.BoxingVictoryReady() {
			drawBossResultBanner(screen, true, g.frames)
		} else if g.bossFight.boxingLosePending() {
			drawBossResultBanner(screen, false, g.frames)
		}
		if g.bossFight.BoxingVictoryReady() {
			x, y, w, h := continueButtonBounds()
			drawPubButton(screen, "CONTINUE", x, y, w, h)
		}

	case StateLevelComplete:
		g.bird.Draw(screen)
		drawGameplayHUD(screen, g)
		drawLevelCompletePrompt(screen, g)

	default:
		g.bird.Draw(screen)

		switch g.state {
		case StateContinue:
			drawContinuePrompt(screen, g.continueText, g.bossContinue, g.frames)
		case StateEnterName:
			drawEnterNameScreen(screen, g)
		}
	}

	if g.state == StatePlaying || g.state == StateContinue || g.state == StateLevelExit {
		drawGameplayHUD(screen, g)
	}
}

func drawTitle(screen *ebiten.Image, str string, centerX, y float64) {
	drawOutlinedTextAt(screen, str, titleFace, centerX, y, text.AlignCenter, ColorText, 3)
}

func drawSubtitle(screen *ebiten.Image, str string, centerX, y float64) {
	drawOutlinedText(screen, str, subtitleFace, centerX, y, text.AlignCenter, ColorText)
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
	drawOutlinedText(screen, str, labelFace, centerX, y, text.AlignCenter, col)
}

func drawLabelLeft(screen *ebiten.Image, str string, leftX, y float64, col color.Color) {
	drawOutlinedText(screen, str, labelFace, leftX, y, text.AlignStart, col)
}

func drawOutlinedText(screen *ebiten.Image, str string, face *text.GoTextFace, x, y float64, align text.Align, col color.Color) {
	drawOutlinedTextAt(screen, str, face, x, y, align, col, 2)
}

func drawOutlinedTextAt(screen *ebiten.Image, str string, face *text.GoTextFace, x, y float64, align text.Align, col color.Color, outline int) {
	for dy := -outline; dy <= outline; dy++ {
		for dx := -outline; dx <= outline; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			op := &text.DrawOptions{}
			op.PrimaryAlign = align
			op.SecondaryAlign = text.AlignStart
			op.GeoM.Translate(x+float64(dx), y+float64(dy))
			op.ColorScale.ScaleWithColor(ColorTextOutline)
			text.Draw(screen, str, face, op)
		}
	}
	op := &text.DrawOptions{}
	op.PrimaryAlign = align
	op.SecondaryAlign = text.AlignStart
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(col)
	text.Draw(screen, str, face, op)
}

func drawLifeHUD(screen *ebiten.Image, lives, livesMax int) {
	drawGlassRow(screen, lives, livesMax, false)
}

func drawGlassRow(screen *ebiten.Image, filled, total int, centered bool) {
	if total <= 0 {
		return
	}
	full := sprite.LifeFull()
	scale := float64(LifeGlassW) / float64(sprite.LifeFrameWidth())
	scaledH := float64(full.Bounds().Dy()) * scale
	baseY := float64(ScreenH - GroundHeight) + (float64(GroundHeight)-scaledH)/2
	drawGlassRowAt(screen, filled, total, centered, baseY)
}

func drawGlassRowAt(screen *ebiten.Image, filled, total int, centered bool, baseY float64) {
	if total <= 0 {
		return
	}
	full := sprite.LifeFull()
	empty := sprite.LifeEmpty()
	scale := float64(LifeGlassW) / float64(sprite.LifeFrameWidth())

	totalW := LifeGlassW*total + LifeGlassGap*(total-1)
	baseX := ScreenW - LifeHUDMargin - totalW
	if centered {
		baseX = (ScreenW - totalW) / 2
	}

	for i := 0; i < total; i++ {
		img := full
		if i >= filled {
			img = empty
		}
		x := baseX + i*(LifeGlassW+LifeGlassGap)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(x), baseY)
		screen.DrawImage(img, op)
	}
}

func drawPubButton(screen *ebiten.Image, label string, x, y, w, h float64) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), ColorKegBand, true)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 2, ColorKegEdge, true)
	drawLabel(screen, label, x+w/2, y+h/2-9, ColorText)
}

func drawGameplayHUD(screen *ebiten.Image, g *Game) {
	if !g.inBossScene() {
		drawLabel(screen, fmt.Sprintf("%d", g.displayScore()), ScreenW/2, 20, ColorText)
		target := DadsRequiredForLevel(g.level)
		drawLabel(screen, fmt.Sprintf("%d/%d", g.levelDadsPassed, target), ScreenW/2, 44, ColorTextMuted)
		drawLifeHUD(screen, g.lives, g.livesMax)
	}
	drawLabelLeft(screen, fmt.Sprintf("Level %d", g.level), LifeHUDMargin, 20, ColorText)
}

func drawBossHUD(screen *ebiten.Image, bf *BossFight, chargePower float64, charging bool) {
	const (
		bossHudBarW = 220.0
		bossHudBarH = 12.0
		lifeBarY    = 72.0
		labelY      = 52.0
		hintY       = float64(FloorSurfaceY) - 130
		chargeBarY  = float64(FloorSurfaceY) - 80
	)

	barX := float64(ScreenW)/2 - bossHudBarW/2

	if bf.IsBoxing() {
		const (
			boxingBossBarY    = 66.0
			boxingTimeBarY    = 92.0
			boxingStaminaBarY = 118.0
			boxingLabelGap    = 8.0
		)
		labelRightX := barX - boxingLabelGap

		drawOutlinedText(screen, "BOSS", labelFace, labelRightX, boxingBossBarY+bossHudBarH/2-6, text.AlignEnd, ColorTextMuted)
		vector.DrawFilledRect(screen, float32(barX), float32(boxingBossBarY), float32(bossHudBarW), float32(bossHudBarH), ColorGlassEdge, true)
		vector.StrokeRect(screen, float32(barX), float32(boxingBossBarY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)
		hpFrac := bf.BossHPFrac()
		if hpFrac > 0 {
			fillW := bossHudBarW * hpFrac
			col := ColorBeer
			if hpFrac <= 0.35 {
				col = color.RGBA{180, 90, 40, 255}
			}
			vector.DrawFilledRect(screen, float32(barX), float32(boxingBossBarY), float32(fillW), float32(bossHudBarH), col, true)
			vector.StrokeRect(screen, float32(barX), float32(boxingBossBarY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)
		}

		drawOutlinedText(screen, "TIME", labelFace, labelRightX, boxingTimeBarY+bossHudBarH/2-6, text.AlignEnd, ColorTextMuted)
		vector.DrawFilledRect(screen, float32(barX), float32(boxingTimeBarY), float32(bossHudBarW), float32(bossHudBarH), ColorGlassEdge, true)
		vector.StrokeRect(screen, float32(barX), float32(boxingTimeBarY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)
		timeFrac := bf.TimerFrac()
		if bf.InCountdown() {
			timeFrac = 1
		}
		if timeFrac > 0 {
			fillW := bossHudBarW * timeFrac
			col := ColorFoam
			if timeFrac <= 0.35 {
				col = color.RGBA{180, 90, 40, 255}
			}
			vector.DrawFilledRect(screen, float32(barX), float32(boxingTimeBarY), float32(fillW), float32(bossHudBarH), col, true)
			vector.StrokeRect(screen, float32(barX), float32(boxingTimeBarY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)
		}
		secs := (bf.TimerRemaining() + 59) / 60
		if bf.InCountdown() {
			secs = (BoxingDurationFrames + 59) / 60
		}
		drawLabelLeft(screen, fmt.Sprintf("%ds", secs), barX+bossHudBarW+10, boxingTimeBarY+bossHudBarH/2-6, ColorText)

		drawOutlinedText(screen, "STAMINA", labelFace, labelRightX, boxingStaminaBarY+bossHudBarH/2-6, text.AlignEnd, ColorTextMuted)
		vector.DrawFilledRect(screen, float32(barX), float32(boxingStaminaBarY), float32(bossHudBarW), float32(bossHudBarH), ColorGlassEdge, true)
		vector.StrokeRect(screen, float32(barX), float32(boxingStaminaBarY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)
		stamFrac := bf.StaminaFrac()
		if bf.InCountdown() {
			stamFrac = 1
		}
		if stamFrac > 0 {
			fillW := bossHudBarW * stamFrac
			col := color.RGBA{80, 180, 120, 255}
			if stamFrac <= 0.35 || bf.BoxingExhausted() {
				col = color.RGBA{180, 90, 40, 255}
			}
			vector.DrawFilledRect(screen, float32(barX), float32(boxingStaminaBarY), float32(fillW), float32(bossHudBarH), col, true)
			vector.StrokeRect(screen, float32(barX), float32(boxingStaminaBarY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)
		}

		if !bf.InCountdown() && !bf.BoxingVictoryReady() && !bf.boxingLosePending() {
			tapFace := &text.GoTextFace{Source: labelFace.Source, Size: 26}
			msg := "TAP!"
			if bf.BoxingExhausted() {
				msg = "TIRED!"
			}
			drawOutlinedText(screen, msg, tapFace, ScreenW/2, float64(ScreenH)/2, text.AlignCenter, ColorTextMuted)
		}
		return
	}

	drawLabel(screen, "Boss", ScreenW/2, labelY, ColorTextMuted)
	vector.DrawFilledRect(screen, float32(barX), float32(lifeBarY), float32(bossHudBarW), float32(bossHudBarH), ColorGlassEdge, true)
	vector.StrokeRect(screen, float32(barX), float32(lifeBarY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)

	if bf.IsOutrun() {
		secs := (bf.TimerRemaining() + 59) / 60
		if bf.InCountdown() {
			secs = (bf.outrunDuration + 59) / 60
			if secs <= 0 {
				secs = (OutrunDurationFramesForLevel(bf.level) + 59) / 60
			}
		}
		frac := bf.LeadFrac()
		if frac > 0 {
			fillW := bossHudBarW * frac
			col := ColorBeer
			if frac <= 0.35 {
				col = color.RGBA{180, 90, 40, 255}
			}
			vector.DrawFilledRect(screen, float32(barX), float32(lifeBarY), float32(fillW), float32(bossHudBarH), col, true)
			vector.StrokeRect(screen, float32(barX), float32(lifeBarY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)
		}
		timerY := lifeBarY + bossHudBarH + 22
		drawLabel(screen, fmt.Sprintf("%ds", secs), ScreenW/2, timerY, ColorText)
		if !bf.InCountdown() {
			tapFace := &text.GoTextFace{Source: labelFace.Source, Size: 26}
			drawOutlinedText(screen, "TAP!", tapFace, ScreenW/2, float64(ScreenH)/2, text.AlignCenter, ColorTextMuted)
		}
		return
	}

	remain := float64(bf.hitsRequired - bf.hits)
	if remain < 0 {
		remain = 0
	}
	frac := 0.0
	if bf.hitsRequired > 0 {
		frac = remain / float64(bf.hitsRequired)
	}
	if frac > 0 {
		fillW := bossHudBarW * frac
		col := ColorBeer
		if frac <= 0.35 {
			col = color.RGBA{180, 90, 40, 255}
		}
		vector.DrawFilledRect(screen, float32(barX), float32(lifeBarY), float32(fillW), float32(bossHudBarH), col, true)
		vector.StrokeRect(screen, float32(barX), float32(lifeBarY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)
	}

	hudHintY := hintY
	hudChargeY := chargeBarY
	if bf.IsDuel() {
		hudHintY = lifeBarY + 28
		hudChargeY = hudHintY + 22
	}

	if bf.IsDuel() {
		switch bf.DuelTurn() {
		case duelTurnEnemy, duelTurnEnemyResolve:
			drawLabel(screen, "Dodge!", ScreenW/2, hudHintY, ColorTextMuted)
		default:
			if !charging {
				drawLabel(screen, "Your turn — hold to throw", ScreenW/2, hudHintY, ColorTextMuted)
			}
		}
	} else if !charging {
		drawLabel(screen, "Hold to throw", ScreenW/2, hudHintY, ColorTextMuted)
	}
	remaining := bf.throwsAllowed - bf.throwsUsed
	if remaining < 0 {
		remaining = 0
	}
	if bf.IsDuel() {
		drawGlassRowAt(screen, remaining, bf.throwsAllowed, true, hudChargeY+bossHudBarH+10)
	} else {
		drawGlassRow(screen, remaining, bf.throwsAllowed, true)
	}

	if chargePower < 0 {
		chargePower = 0
	} else if chargePower > 1 {
		chargePower = 1
	}
	vector.DrawFilledRect(screen, float32(barX), float32(hudChargeY), float32(bossHudBarW), float32(bossHudBarH), ColorGlassEdge, true)
	vector.StrokeRect(screen, float32(barX), float32(hudChargeY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)
	if chargePower > 0 {
		fillW := bossHudBarW * chargePower
		col := ColorBeer
		if chargePower >= 1 {
			col = ColorFoam
		}
		vector.DrawFilledRect(screen, float32(barX), float32(hudChargeY), float32(fillW), float32(bossHudBarH), col, true)
		if chargePower >= 0.4 {
			drawChargeFoamBubbles(screen, barX, hudChargeY, fillW, bossHudBarH, chargePower)
		}
		vector.StrokeRect(screen, float32(barX), float32(hudChargeY), float32(bossHudBarW), float32(bossHudBarH), 2, ColorKegEdge, true)
	}
}

// drawChargeFoamBubbles draws a dense cloudy foam mound over the charged portion of the bar.
// Height and density scale with power (thickest at 1.0).
func drawChargeFoamBubbles(screen *ebiten.Image, barX, barY, fillW, barH, power float64) {
	if fillW < 8 {
		return
	}
	// Ease power from the 0.4 threshold up to 1.0.
	tPower := (power - 0.4) / 0.6
	if tPower < 0 {
		tPower = 0
	} else if tPower > 1 {
		tPower = 1
	}
	ease := tPower * tPower * (3 - 2*tPower)

	height := 8 + 18*ease        // ~8px at start → ~26px at full
	bands := 3 + int(2*ease+0.5) // 3..5 vertical bands
	cols := 8 + int(10*ease+0.5) // 8..18 across

	cx := barX + fillW/2
	tick := float64(ebiten.Tick())

	// Pass 1: large overlapping foam blobs (cloud body).
	for band := 0; band < bands; band++ {
		bandT := float64(band) / float64(bands-1)
		if bands == 1 {
			bandT = 0
		}
		// Half-ellipse: narrower and higher toward the top.
		rowHalfW := (fillW * 0.5) * (1 - 0.35*bandT)
		rowY := barY + barH*0.35 - bandT*height
		for i := 0; i < cols; i++ {
			iT := float64(i) / float64(cols-1)
			if cols == 1 {
				iT = 0.5
			}
			phase := tick*0.11 + float64(band)*1.3 + float64(i)*0.9
			ox := (iT - 0.5) * 2 * rowHalfW
			ox += math.Sin(phase)*3.5 + math.Cos(phase*0.7+float64(band))*2
			oy := math.Sin(phase*1.4+float64(band)) * 2.2
			bx := cx + ox
			by := rowY + oy
			// Keep blobs over the filled span.
			if bx < barX-2 || bx > barX+fillW+2 {
				continue
			}
			r := 3.2 + 3.8*(1-bandT)*ease + 1.2*math.Sin(phase+float64(i)*0.4)
			if r < 2.5 {
				r = 2.5
			}
			vector.DrawFilledCircle(screen, float32(bx), float32(by), float32(r), ColorFoam, true)
			// Soft secondary blob for cloudiness.
			vector.DrawFilledCircle(screen, float32(bx+r*0.35), float32(by-r*0.2), float32(r*0.65), ColorFoam, true)
		}
	}

	// Pass 2: smaller bubble accents nested in the foam.
	bubbleCount := 10 + int(16*ease)
	for i := 0; i < bubbleCount; i++ {
		phase := tick*0.15 + float64(i)*1.1
		iT := float64(i) / float64(bubbleCount)
		ox := (iT - 0.5) * fillW * (0.85 + 0.1*math.Sin(phase))
		oy := -2 - (0.3+0.7*((math.Sin(phase*0.8)+1)*0.5))*height*ease
		bx := cx + ox + math.Cos(phase)*4
		by := barY + barH*0.3 + oy
		if bx < barX || bx > barX+fillW {
			continue
		}
		r := float32(1.4 + 1.6*math.Sin(phase*1.3+float64(i)))
		if r < 1.2 {
			r = 1.2
		}
		vector.DrawFilledCircle(screen, float32(bx), float32(by), r, ColorBubble, true)
	}

	// Pass 3: fill the bar body with soft foam so the head connects to the track.
	bodyCols := 6 + int(8*ease)
	for i := 0; i < bodyCols; i++ {
		phase := tick*0.09 + float64(i)*1.4
		iT := float64(i) / float64(bodyCols-1)
		if bodyCols == 1 {
			iT = 0.5
		}
		bx := barX + 4 + iT*(fillW-8) + math.Sin(phase)*2
		by := barY + barH*0.45 + math.Cos(phase*1.2)*2
		r := float32(3.5 + 2*ease + 0.8*math.Sin(phase))
		vector.DrawFilledCircle(screen, float32(bx), float32(by), r, ColorFoam, true)
	}
}

func drawLevelCompletePrompt(screen *ebiten.Image, g *Game) {
	drawBossResultBanner(screen, true, g.frames)
	x, y, w, h := continueButtonBounds()
	drawPubButton(screen, "CONTINUE", x, y, w, h)
}

func randomText() string {
	randomTexts := []string{
		"Ah, no! You dropped your beer!",
		"You're drunk! You can't play this game!",
		"Ah, no! You knocked over your beer!",
		"Go home and sleep it off!",
	}
	return randomTexts[rand.Intn(len(randomTexts))]
}

func drawContinuePrompt(screen *ebiten.Image, msg string, bossLoss bool, frames int) {
	if bossLoss {
		drawBossResultBanner(screen, false, frames)
	} else {
		drawLabel(screen, msg, ScreenW/2, continueBtnY-28, ColorText)
	}
	x, y, w, h := continueButtonBounds()
	drawPubButton(screen, "CONTINUE", x, y, w, h)
}

const (
	bossResultBannerW  = 280.0
	bossResultBannerH  = 100.0
	bossResultFaceSize = 56.0
)

var bossResultPlaque *ebiten.Image

func ensureBossResultPlaque() *ebiten.Image {
	if bossResultPlaque != nil {
		return bossResultPlaque
	}
	plaque := ebiten.NewImage(int(bossResultBannerW), int(bossResultBannerH))
	tile := sprite.WoodFloor()
	tw := float64(tile.Bounds().Dx())
	th := float64(tile.Bounds().Dy())
	if tw > 0 && th > 0 {
		for ty := 0.0; ty < bossResultBannerH; ty += th {
			for tx := 0.0; tx < bossResultBannerW; tx += tw {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(tx, ty)
				plaque.DrawImage(tile, op)
			}
		}
	} else {
		vector.DrawFilledRect(plaque, 0, 0, float32(bossResultBannerW), float32(bossResultBannerH), ColorKeg, true)
	}
	bossResultPlaque = plaque
	return bossResultPlaque
}

func drawBossResultBanner(screen *ebiten.Image, won bool, frames int) {
	cx := float64(ScreenW) / 2
	cy := float64(continueBtnY) - 12 - bossResultBannerH/2
	x := cx - bossResultBannerW/2
	y := cy - bossResultBannerH/2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(ensureBossResultPlaque(), op)
	vector.StrokeRect(screen, float32(x), float32(y), float32(bossResultBannerW), float32(bossResultBannerH), 3, ColorKegEdge, true)

	label := "Loser"
	if won {
		label = "Winner"
	}
	size := bossResultFaceSize * (1.0 + 0.1*sinBob(frames))
	face := &text.GoTextFace{
		Source: titleFace.Source,
		Size:   size,
	}
	textY := cy - size*0.45
	drawOutlinedTextAt(screen, label, face, cx, textY, text.AlignCenter, ColorText, 3)
}

func drawBossIntroPrompt(screen *ebiten.Image, level int) {
	if BossVariantForLevel(level) == BossVariantNeighbour {
		drawLabel(screen, "The PappaPub neighbour wants", ScreenW/2, continueBtnY-72, ColorText)
		drawLabel(screen, "a kickabout at the stadium —", ScreenW/2, continueBtnY-50, ColorText)
		drawLabel(screen, "hit him with beer glasses!", ScreenW/2, continueBtnY-28, ColorText)
		x, y, w, h := meetBossButtonBounds()
		drawPubButton(screen, "KICK OFF!", x, y, w, h)
		return
	}
	if BossIsBoxing(level) {
		secs := (BoxingDurationFrames + 59) / 60
		drawLabel(screen, "Dad wants a boxing match!", ScreenW/2, continueBtnY-72, ColorText)
		drawLabel(screen, fmt.Sprintf("Knock him out before %d seconds —", secs), ScreenW/2, continueBtnY-50, ColorText)
		drawLabel(screen, "tap fast!", ScreenW/2, continueBtnY-28, ColorText)
		x, y, w, h := meetBossButtonBounds()
		drawPubButton(screen, "BOX!", x, y, w, h)
		return
	}
	if BossIsOutrun(level) {
		secs := (OutrunDurationFramesForLevel(level) + 59) / 60
		drawLabel(screen, "Your best pal's wife is chasing you!", ScreenW/2, continueBtnY-72, ColorText)
		drawLabel(screen, fmt.Sprintf("Outrun her for %d seconds —", secs), ScreenW/2, continueBtnY-50, ColorText)
		drawLabel(screen, "tap fast!", ScreenW/2, continueBtnY-28, ColorText)
	} else {
		drawLabel(screen, "Your best pal's wife is on her way", ScreenW/2, continueBtnY-72, ColorText)
		drawLabel(screen, "to drag him home from the pub,", ScreenW/2, continueBtnY-50, ColorText)
		drawLabel(screen, "you got to stop her!", ScreenW/2, continueBtnY-28, ColorText)
	}
	x, y, w, h := meetBossButtonBounds()
	drawPubButton(screen, "MEET THE LADY BOSS!", x, y, w, h)
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
