package game

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"flappy/internal/game/sprite"
)

const (
	bossDisplayH      = 140.0
	bossMarginX       = 20.0
	bossMoveMinY      = 60.0
	bossSideDashAmp   = 50.0
	bossDashSpeed     = 3.2
	bossHomeReturnSpd = 1.4
	bossPlayerY       = float64(FloorSurfaceY+80) / 2
	particleGravity   = 0.18

	throwChargeMaxFrames = 30
	throwLaunchAngle     = -0.55
	throwSpeedMin        = 4.4
	throwSpeedMax        = 11.0
	projectileGravity    = 0.32
	projectileAngleClamp = 1.2

	deathTipFrames    = 28
	deathExplodeHold  = 20
	deathSettleFrames = 36

	bossHitShakeFrames = 20
	bossHitShakeAmp    = 5.0
	bossHitShakeRot    = 0.08
	bossHitFlipEvery   = 3
)

const (
	deathPhaseNone = iota
	deathPhaseTip
	deathPhaseExplode
	deathPhaseSettle
)

const (
	bossModeThrow = iota
	bossModeOutrun
	bossModeBoxing
)

var bossExplodeOrange = color.RGBA{255, 140, 40, 255}

type Projectile struct {
	X, Y, VX, VY float64
	Angle        float64
	Age          int
}

type beerParticle struct {
	X, Y, VX, VY  float64
	Life, MaxLife float64
	R             float64
	Col           color.RGBA
}

type Boss struct {
	X, Y          float64
	width, height float64
	homeX         float64
	dirY          float64
	speedY        float64
	baseSpeedY    float64
	speedX        float64
	hitboxInset   float64
	yFlipTimer    int
	yFlipMin      int
	yFlipMax      int
	dashTimer     int
	dashGapMin    int
	dashGapMax    int
	dashing       bool
	rotation      float64
	hidden        bool
	defeated      bool
	deathFrame    int
	tipStartY     float64
	tipTargetY    float64
	shakeTimer    int
	faceRight     bool
	sprinting     bool
}

type BossFight struct {
	mode          int
	variant       int
	throwsUsed    int
	hits          int
	hitsRequired  int
	throwsAllowed int
	projectiles   []*Projectile
	particles     []beerParticle
	pendingWin    bool
	deathPhase    int
	deathTimer    int
	explodeSFX    bool
	hitSFX        bool
	loseSFX       bool
	boss          Boss

	// Neighbour duel (level 4 alternating turns).
	duelTurn          int
	lastThrowPower    float64
	playerThrowFrame  int
	playerThrowTick   int
	playerThrowing    bool
	glassSpawned      bool
	enemyThrowFrame   int
	enemyThrowTick    int
	enemyThrowing     bool
	footballSpawned   bool
	enemyProjectile   *FootballProjectile
	footballHits      int
	lifeLostPending   bool
	playerInvincible  int
	playerVY          float64

	// Outrun (scrolling race every 3rd level).
	level           int
	outrunDuration  int
	timerFrames     int
	countdownFrames int
	playerX         float64
	playerY         float64
	playerSpeed     float64
	lead            float64
	scrollX         float64
	tapSFX          bool
	laughSFX        bool
	laughCooldown   int
	sprintFrames    int
	sprintCooldown  int
	runAnimTick     int

	// Boxing victory hold (paused win pose until CONTINUE).
	boxingVictoryHold    bool
	boxingFramesSinceTap int
	boxingStamina        float64
	boxingExhausted      bool
}

func NewBossFight() *BossFight {
	bf := &BossFight{}
	bf.Reset(1)
	return bf
}

func (bf *BossFight) Reset(level int) {
	frame := sprite.AngryWomanFrame(0)
	contentH := sprite.AngryWomanContentH()
	scale := bossDisplayH / contentH
	bossW := float64(frame.Bounds().Dx()) * scale
	homeX := float64(ScreenW) - bossMarginX - bossW
	yFlipMin, yFlipMax := BossYFlipIntervalForLevel(level)
	dashGapMin, dashGapMax := BossDashGapForLevel(level)
	baseSpeed := BossMoveSpeedForLevel(level)

	bf.throwsUsed = 0
	bf.hits = 0
	bf.hitsRequired = BossHitsForLevel(level)
	bf.throwsAllowed = BossThrowsForLevel(level)
	bf.projectiles = nil
	bf.particles = nil
	bf.pendingWin = false
	bf.deathPhase = deathPhaseNone
	bf.deathTimer = 0
	bf.explodeSFX = false
	bf.hitSFX = false
	bf.loseSFX = false
	bf.variant = BossVariantLady
	bf.duelTurn = duelTurnPlayer
	bf.enemyProjectile = nil
	bf.lifeLostPending = false
	bf.footballHits = 0
	bf.playerInvincible = 0
	bf.playerVY = 0
	bf.tapSFX = false
	bf.laughSFX = false
	bf.laughCooldown = 0
	bf.sprintFrames = 0
	bf.sprintCooldown = 0
	bf.level = level
	bf.outrunDuration = 0
	bf.timerFrames = 0
	bf.countdownFrames = 0
	bf.playerX = float64(BirdStartX)
	bf.playerY = bossPlayerY
	bf.playerSpeed = 0
	bf.lead = 0
	bf.scrollX = 0
	bf.runAnimTick = 0
	bf.boxingVictoryHold = false
	bf.boxingFramesSinceTap = boxingDadSuppressFrames
	bf.boxingStamina = boxingStaminaMax
	bf.boxingExhausted = false

	if BossIsOutrun(level) {
		bf.mode = bossModeOutrun
		bf.outrunDuration = OutrunDurationFramesForLevel(level)
		bf.timerFrames = bf.outrunDuration
		bf.countdownFrames = OutrunCountdownTotalFrames()
		bf.laughCooldown = outrunLaughIntervalFrames()
		bf.sprintCooldown = outrunSprintCooldownFrames()
		bf.playerX = outrunPlayerScreenX
		bf.playerY = float64(FloorSurfaceY) - BirdHeight/2 - outrunGroundClearance
		bf.playerSpeed = outrunCoastSpeed
		bf.lead = OutrunStartLeadForLevel(level)
		bf.boss = Boss{
			X:           bf.playerX - bf.lead - bossW,
			Y:           float64(FloorSurfaceY) - bossDisplayH - outrunGroundClearance,
			width:       bossW,
			height:      bossDisplayH,
			homeX:       homeX,
			hitboxInset: BossHitboxInsetForLevel(level),
			faceRight:   true,
		}
		return
	}

	if BossIsBoxing(level) {
		bf.resetBoxingBoss(level)
		return
	}

	bf.mode = bossModeThrow
	if BossVariantForLevel(level) == BossVariantNeighbour {
		bf.resetNeighbourBoss(level)
		return
	}
	bf.boss = Boss{
		X:           homeX,
		Y:           bossPlayerY,
		width:       bossW,
		height:      bossDisplayH,
		homeX:       homeX,
		dirY:        1,
		speedY:      baseSpeed,
		baseSpeedY:  baseSpeed,
		hitboxInset: BossHitboxInsetForLevel(level),
		yFlipMin:    yFlipMin,
		yFlipMax:    yFlipMax,
		yFlipTimer:  randRange(yFlipMin, yFlipMax),
		dashGapMin:  dashGapMin,
		dashGapMax:  dashGapMax,
		dashTimer:   randRange(dashGapMin, dashGapMax),
	}
}

func (bf *BossFight) IsOutrun() bool {
	return bf.mode == bossModeOutrun
}

func (bf *BossFight) IsSprinting() bool {
	return bf.mode == bossModeOutrun && bf.sprintFrames > 0
}

func (bf *BossFight) InCountdown() bool {
	return (bf.mode == bossModeOutrun || bf.mode == bossModeBoxing) && bf.countdownFrames > 0
}

// CountdownDisplay returns "3"/"2"/"1"/"GO!" while counting down, else "".
func (bf *BossFight) CountdownDisplay() string {
	if !bf.InCountdown() {
		return ""
	}
	if bf.countdownFrames <= outrunGoHoldFrames {
		return "GO!"
	}
	afterGo := bf.countdownFrames - outrunGoHoldFrames
	sec := (afterGo + outrunSecFrames - 1) / outrunSecFrames
	switch sec {
	case 3:
		return "3"
	case 2:
		return "2"
	default:
		return "1"
	}
}

func (bf *BossFight) TimerRemaining() int {
	if bf.timerFrames < 0 {
		return 0
	}
	return bf.timerFrames
}

// LeadFrac is 0 when caught and 1 at max lead (for HUD).
func (bf *BossFight) LeadFrac() float64 {
	span := outrunMaxLead - outrunCatchLead
	if span <= 0 {
		return 0
	}
	f := (bf.lead - outrunCatchLead) / span
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

// BossHPFrac is remaining boss HP 0..1 (boxing uses lead as HP).
func (bf *BossFight) BossHPFrac() float64 {
	if boxingBossMaxHP <= 0 {
		return 0
	}
	f := bf.lead / boxingBossMaxHP
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

// TimerFrac is remaining fight deadline 0..1.
func (bf *BossFight) TimerFrac() float64 {
	dur := bf.outrunDuration
	if dur <= 0 {
		if bf.IsBoxing() {
			dur = BoxingDurationFrames
		} else {
			dur = OutrunDurationFramesForLevel(bf.level)
		}
	}
	if dur <= 0 {
		return 0
	}
	f := float64(bf.TimerRemaining()) / float64(dur)
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

func (bf *BossFight) ScrollX() float64 {
	return bf.scrollX
}

func (bf *BossFight) PlayerX() float64 {
	return bf.playerX
}

func (bf *BossFight) PlayerY() float64 {
	return bf.playerY
}

func randRange(min, max int) int {
	if max <= min {
		return min
	}
	return min + rand.Intn(max-min+1)
}

func (b *Boss) Bounds() (x, y, w, h float64) {
	insetX := b.width * b.hitboxInset
	insetY := b.height * b.hitboxInset
	return b.X + insetX, b.Y + insetY, b.width - insetX*2, b.height - insetY*2
}

func (b *Boss) center() (cx, cy float64) {
	return b.X + b.width/2, b.Y + b.height/2
}

func (b *Boss) randomizeSpeedY() {
	// Keep vertical speed within ~70–130% of the level base.
	b.speedY = b.baseSpeedY * (0.7 + rand.Float64()*0.6)
}

func (b *Boss) scheduleYFlip() {
	b.yFlipTimer = randRange(b.yFlipMin, b.yFlipMax)
}

func (b *Boss) scheduleDash() {
	b.dashTimer = randRange(b.dashGapMin, b.dashGapMax)
}

func (b *Boss) Update() {
	if b.defeated {
		return
	}
	if b.shakeTimer > 0 {
		b.shakeTimer--
	}

	maxY := float64(FloorSurfaceY) - b.height - 20
	minX := b.homeX - bossSideDashAmp
	maxX := b.homeX + 8

	b.yFlipTimer--
	if b.yFlipTimer <= 0 {
		b.dirY = -b.dirY
		b.randomizeSpeedY()
		b.scheduleYFlip()
	}

	b.Y += b.dirY * b.speedY
	if b.Y <= bossMoveMinY {
		b.Y = bossMoveMinY
		b.dirY = 1
		b.randomizeSpeedY()
		b.scheduleYFlip()
	} else if b.Y >= maxY {
		b.Y = maxY
		b.dirY = -1
		b.randomizeSpeedY()
		b.scheduleYFlip()
	}

	if b.dashing {
		b.X += b.speedX
		if b.X <= minX {
			b.X = minX
			b.speedX = math.Abs(b.speedX)
		} else if b.X >= maxX {
			b.X = maxX
			b.dashing = false
			b.speedX = 0
			b.scheduleDash()
		}
		// End dash after drifting back near home.
		if b.speedX > 0 && b.X >= b.homeX-2 {
			b.X = b.homeX
			b.dashing = false
			b.speedX = 0
			b.scheduleDash()
		}
	} else {
		// Ease back toward home when idle.
		if b.X < b.homeX {
			b.X += bossHomeReturnSpd
			if b.X > b.homeX {
				b.X = b.homeX
			}
		}
		b.dashTimer--
		if b.dashTimer <= 0 {
			b.dashing = true
			// Dash left first, then return right.
			b.speedX = -bossDashSpeed * (0.85 + rand.Float64()*0.4)
		}
	}
}

func (b *Boss) Draw(screen *ebiten.Image) {
	if b.hidden {
		return
	}

	frameIdx := sprite.AngryWomanFrameIndex(ebiten.Tick())
	if b.sprinting {
		frameIdx = sprite.AngryWomanFrameIndex(ebiten.Tick() * 2)
	}
	if b.defeated {
		frameIdx = b.deathFrame
	}
	frame := sprite.AngryWomanFrame(frameIdx)
	contentH := sprite.AngryWomanContentH()
	scale := b.height / contentH
	sw := float64(frame.Bounds().Dx()) * scale
	sh := float64(frame.Bounds().Dy()) * scale
	cx, cy := b.center()

	ox, oy, rot := 0.0, 0.0, b.rotation
	flip := b.faceRight
	if b.shakeTimer > 0 && !b.defeated {
		t := float64(b.shakeTimer)
		amp := bossHitShakeAmp * (t / bossHitShakeFrames)
		ox = math.Sin(t*1.8) * amp
		oy = math.Cos(t*2.3) * amp * 0.4
		rot += math.Sin(t*2.1) * bossHitShakeRot * (t / bossHitShakeFrames)
		flip = flip != ((b.shakeTimer/bossHitFlipEvery)%2 == 1)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(-sw/2, -sh/2)
	if flip {
		op.GeoM.Scale(-1, 1)
	}
	op.GeoM.Rotate(rot)
	op.GeoM.Translate(cx+ox, cy+oy)
	screen.DrawImage(frame, op)
}

func (p *Projectile) Bounds() (x, y, w, h float64) {
	return p.X - BirdWidth/2, p.Y - BirdHeight/2, BirdWidth, BirdHeight
}

func drawProjectile(screen *ebiten.Image, p *Projectile) {
	drawBeerGlass(screen, p.X, p.Y, BirdWidth, BirdHeight, p.Angle)
}

func (bf *BossFight) spawnParticles(ox, oy float64, n int, speedMin, speedMax float64, lifeMin, lifeMax float64, rMin, rMax float64, foamBias float64, angleMin, angleMax float64) {
	for i := 0; i < n; i++ {
		ang := angleMin + rand.Float64()*(angleMax-angleMin)
		spd := speedMin + rand.Float64()*(speedMax-speedMin)
		life := lifeMin + rand.Float64()*(lifeMax-lifeMin)
		r := rMin + rand.Float64()*(rMax-rMin)
		col := ColorFoam
		if rand.Float64() > foamBias {
			col = ColorBeer
		}
		bf.particles = append(bf.particles, beerParticle{
			X:       ox + (rand.Float64()-0.5)*4,
			Y:       oy + (rand.Float64()-0.5)*4,
			VX:      math.Cos(ang) * spd,
			VY:      math.Sin(ang) * spd,
			Life:    life,
			MaxLife: life,
			R:       r,
			Col:     col,
		})
	}
}

func (bf *BossFight) spawnBossExplosion(cx, cy float64) {
	n := 45 + rand.Intn(11)
	for i := 0; i < n; i++ {
		ang := rand.Float64() * 2 * math.Pi
		spd := 2.0 + rand.Float64()*6.0
		life := 25.0 + rand.Float64()*20.0
		r := 2.5 + rand.Float64()*5.5
		var col color.RGBA
		switch rand.Intn(3) {
		case 0:
			col = ColorFoam
		case 1:
			col = ColorBeer
		default:
			col = bossExplodeOrange
		}
		bf.particles = append(bf.particles, beerParticle{
			X:       cx + (rand.Float64()-0.5)*10,
			Y:       cy + (rand.Float64()-0.5)*10,
			VX:      math.Cos(ang) * spd,
			VY:      math.Sin(ang) * spd,
			Life:    life,
			MaxLife: life,
			R:       r,
			Col:     col,
		})
	}
}

func (bf *BossFight) spawnFoamBurst(p *Projectile) {
	rimX, rimY := toWorld(0, -BirdHeight/2+4, p.X, p.Y, p.Angle)
	// Burst upward and forward from the tilted rim.
	bf.spawnParticles(rimX, rimY, 10, 1.5, 4.5, 14, 22, 2.5, 5.5, 0.85, -math.Pi*0.85, -math.Pi*0.15)
}

func (bf *BossFight) spawnFlightDrip(p *Projectile) {
	rimX, rimY := toWorld(0, -BirdHeight/2+4, p.X, p.Y, p.Angle)
	bf.spawnParticles(rimX, rimY, 1+rand.Intn(2), 0.4, 1.8, 8, 14, 1.5, 3.5, 0.9, -math.Pi*0.6, math.Pi*0.1)
}

func (bf *BossFight) spawnSplash(x, y float64) {
	bf.spawnParticles(x, y, 18, 1.5, 5.5, 18, 28, 2, 6, 0.55, 0, 2*math.Pi)
}

func (bf *BossFight) updateParticles() {
	alive := bf.particles[:0]
	for i := range bf.particles {
		p := &bf.particles[i]
		p.VY += particleGravity
		p.X += p.VX
		p.Y += p.VY
		p.Life--
		if p.Life > 0 {
			alive = append(alive, *p)
		}
	}
	bf.particles = alive
}

func (bf *BossFight) drawParticles(screen *ebiten.Image) {
	for i := range bf.particles {
		p := &bf.particles[i]
		t := p.Life / p.MaxLife
		if t < 0 {
			t = 0
		}
		col := p.Col
		col.A = uint8(float64(col.A) * t)
		r := float32(p.R * (0.6 + 0.4*t))
		vector.DrawFilledCircle(screen, float32(p.X), float32(p.Y), r, col, true)
	}
}

func (bf *BossFight) CanThrow() bool {
	if bf.pendingWin {
		return false
	}
	if bf.throwsUsed >= bf.throwsAllowed {
		return false
	}
	if len(bf.projectiles) > 0 {
		return false
	}
	if bf.IsDuel() {
		return bf.duelTurn == duelTurnPlayer && !bf.playerThrowing
	}
	return bf.mode == bossModeThrow
}

func (bf *BossFight) CanTap() bool {
	tapMode := bf.mode == bossModeOutrun || bf.mode == bossModeBoxing
	if bf.boxingLosePending() || bf.BoxingVictoryReady() {
		return false
	}
	if bf.IsBoxing() && bf.boxingExhausted {
		return false
	}
	return tapMode && !bf.pendingWin && !bf.InCountdown() && bf.timerFrames > 0
}

func (bf *BossFight) Tap() {
	if !bf.CanTap() {
		return
	}
	if bf.IsBoxing() {
		bf.boxingTap()
		return
	}
	bf.lead += OutrunTapLeadBoostForLevel(bf.level)
	if bf.lead > outrunMaxLead {
		bf.lead = outrunMaxLead
	}
	bf.playerSpeed += outrunTapSpeedBoost
	if bf.playerSpeed > outrunMaxPlayerSpeed {
		bf.playerSpeed = outrunMaxPlayerSpeed
	}
	bf.tapSFX = true
}

func (bf *BossFight) ConsumeTapSFX() bool {
	if !bf.tapSFX {
		return false
	}
	bf.tapSFX = false
	return true
}

func (bf *BossFight) ConsumeLaughSFX() bool {
	if !bf.laughSFX {
		return false
	}
	bf.laughSFX = false
	return true
}

// ThrowPower maps hold frames (0..throwChargeMaxFrames) to 0..1 charge.
func ThrowPower(chargeFrames int) float64 {
	if chargeFrames <= 0 {
		return 0
	}
	if chargeFrames >= throwChargeMaxFrames {
		return 1
	}
	return float64(chargeFrames) / float64(throwChargeMaxFrames)
}

func ThrowChargeMaxFrames() int {
	return throwChargeMaxFrames
}

func (bf *BossFight) Throw(power float64) {
	if !bf.CanThrow() {
		return
	}
	if bf.IsDuel() {
		bf.duelThrow(power)
		return
	}
	if power < 0 {
		power = 0
	} else if power > 1 {
		power = 1
	}
	speed := throwSpeedMin + (throwSpeedMax-throwSpeedMin)*power
	p := &Projectile{
		X:     float64(BirdStartX) + BirdWidth/2,
		Y:     bossPlayerY,
		VX:    math.Cos(throwLaunchAngle) * speed,
		VY:    math.Sin(throwLaunchAngle) * speed,
		Angle: throwLaunchAngle,
	}
	bf.projectiles = append(bf.projectiles, p)
	bf.throwsUsed++
	bf.spawnFoamBurst(p)
}

// projectileFlightMaxX returns how far a throw of the given power travels
// before hitting the floor or leaving the screen (no boss collision).
func projectileFlightMaxX(power float64) float64 {
	if power < 0 {
		power = 0
	} else if power > 1 {
		power = 1
	}
	speed := throwSpeedMin + (throwSpeedMax-throwSpeedMin)*power
	p := Projectile{
		X:  float64(BirdStartX) + BirdWidth/2,
		Y:  bossPlayerY,
		VX: math.Cos(throwLaunchAngle) * speed,
		VY: math.Sin(throwLaunchAngle) * speed,
	}
	maxX := p.X
	for i := 0; i < 600; i++ {
		p.VY += projectileGravity
		p.X += p.VX
		p.Y += p.VY
		if p.X > maxX {
			maxX = p.X
		}
		if p.X > float64(ScreenW)+BirdWidth || p.Y > float64(FloorSurfaceY) {
			break
		}
	}
	return maxX
}

// BossHomeXEstimate is a conservative right-side X used for range tests
// (boss sits near ScreenW - margin).
func BossHomeXEstimate() float64 {
	return float64(ScreenW) - bossMarginX - 80
}

func (bf *BossFight) ConsumeExplodeSFX() bool {
	if !bf.explodeSFX {
		return false
	}
	bf.explodeSFX = false
	return true
}

func (bf *BossFight) ConsumeHitSFX() bool {
	if !bf.hitSFX {
		return false
	}
	bf.hitSFX = false
	return true
}

func (bf *BossFight) ConsumeLoseSFX() bool {
	if !bf.loseSFX {
		return false
	}
	bf.loseSFX = false
	return true
}

func (bf *BossFight) startDeath() {
	bf.pendingWin = true
	bf.deathPhase = deathPhaseTip
	bf.deathTimer = deathTipFrames
	bf.boss.defeated = true
	bf.boss.deathFrame = sprite.AngryWomanFrameIndex(ebiten.Tick())
	bf.boss.rotation = 0
	bf.boss.shakeTimer = 0
	bf.boss.dashing = false
	bf.boss.speedX = 0
	bf.boss.tipStartY = bf.boss.Y
	targetY := float64(FloorSurfaceY) - bf.boss.width - 10
	if targetY < bossMoveMinY {
		targetY = bossMoveMinY
	}
	bf.boss.tipTargetY = targetY
}

func (bf *BossFight) updateDeath() (won, lost bool) {
	switch bf.deathPhase {
	case deathPhaseTip:
		bf.deathTimer--
		progress := 1 - float64(bf.deathTimer)/float64(deathTipFrames)
		if progress < 0 {
			progress = 0
		}
		if progress > 1 {
			progress = 1
		}
		ease := progress * progress * (3 - 2*progress)
		bf.boss.rotation = ease * (math.Pi / 2)
		bf.boss.Y = bf.boss.tipStartY + (bf.boss.tipTargetY-bf.boss.tipStartY)*ease
		if bf.deathTimer <= 0 {
			bf.deathPhase = deathPhaseExplode
			bf.deathTimer = deathExplodeHold
			cx, cy := bf.boss.center()
			bf.spawnBossExplosion(cx, cy)
			bf.boss.hidden = true
			bf.explodeSFX = true
		}
	case deathPhaseExplode:
		bf.deathTimer--
		if bf.deathTimer <= 0 {
			bf.deathPhase = deathPhaseSettle
			bf.deathTimer = deathSettleFrames
		}
	case deathPhaseSettle:
		bf.deathTimer--
		if bf.deathTimer <= 0 {
			return true, false
		}
	}
	return false, false
}

func (bf *BossFight) Update() (won, lost bool) {
	bf.updateParticles()

	if bf.pendingWin {
		if bf.IsBoxing() {
			return bf.updateBoxingWin()
		}
		return bf.updateDeath()
	}

	if bf.mode == bossModeOutrun {
		return bf.updateOutrun()
	}

	if bf.mode == bossModeBoxing {
		return bf.updateBoxing()
	}

	if bf.IsDuel() {
		return bf.updateDuel()
	}

	bf.boss.Update()

	remaining := bf.projectiles[:0]
	for _, p := range bf.projectiles {
		p.VY += projectileGravity
		p.X += p.VX
		p.Y += p.VY
		p.Age++
		p.Angle = math.Atan2(p.VY, p.VX)
		if p.Angle > projectileAngleClamp {
			p.Angle = projectileAngleClamp
		} else if p.Angle < -projectileAngleClamp {
			p.Angle = -projectileAngleClamp
		}
		if p.Age%3 == 0 {
			bf.spawnFlightDrip(p)
		}
		if p.X > float64(ScreenW)+BirdWidth || p.Y > float64(FloorSurfaceY) {
			continue
		}
		px, py, pw, ph := p.Bounds()
		bx, by, bw, bh := bf.boss.Bounds()
		if aabbOverlap(px, py, pw, ph, bx, by, bw, bh) {
			bf.hits++
			bf.hitSFX = true
			bf.boss.shakeTimer = bossHitShakeFrames
			bf.spawnSplash(p.X, p.Y)
			if bf.hits >= bf.hitsRequired {
				bf.startDeath()
			}
			continue
		}
		remaining = append(remaining, p)
	}
	bf.projectiles = remaining

	if bf.pendingWin {
		return false, false
	}
	if bf.throwsUsed >= bf.throwsAllowed && len(bf.projectiles) == 0 && bf.hits < bf.hitsRequired {
		return false, true
	}
	return false, false
}

func (bf *BossFight) updateOutrun() (won, lost bool) {
	if bf.countdownFrames > 0 {
		bf.countdownFrames--
		return false, false
	}

	if bf.boss.shakeTimer > 0 {
		bf.boss.shakeTimer--
	}

	if bf.playerSpeed > outrunCoastSpeed {
		bf.playerSpeed -= outrunSpeedDecay
		if bf.playerSpeed < outrunCoastSpeed {
			bf.playerSpeed = outrunCoastSpeed
		}
	} else if bf.playerSpeed < outrunCoastSpeed {
		bf.playerSpeed = outrunCoastSpeed
	}
	if bf.playerSpeed > outrunCoastSpeed {
		bf.runAnimTick++
	}

	duration := bf.outrunDuration
	if duration <= 0 {
		duration = OutrunDurationFramesForLevel(bf.level)
	}
	progress := 1 - float64(bf.timerFrames)/float64(duration)
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	wifeClose := OutrunWifeCloseBaseForLevel(bf.level) + OutrunWifeCloseRiseForLevel(bf.level)*progress
	wifeClose *= OutrunDifficultyMult(OutrunChaseIndex(bf.level))

	bf.updateOutrunSprint(&wifeClose)

	bf.lead -= wifeClose
	bf.scrollX += bf.playerSpeed
	bf.syncOutrunBossPos()
	bf.boss.sprinting = bf.sprintFrames > 0

	bf.timerFrames--
	if bf.lead <= outrunCatchLead {
		return false, true
	}
	if bf.timerFrames <= 0 {
		bf.startDeath()
		return false, false
	}

	if bf.laughCooldown > 0 {
		bf.laughCooldown--
		if bf.laughCooldown == 0 {
			bf.laughSFX = true
			bf.laughCooldown = outrunLaughIntervalFrames()
		}
	}
	return false, false
}

func (bf *BossFight) updateOutrunSprint(wifeClose *float64) {
	if bf.sprintFrames > 0 {
		*wifeClose *= outrunSprintCloseMult
		bf.sprintFrames--
		if bf.sprintFrames == 0 {
			bf.sprintCooldown = outrunSprintCooldownFrames()
		}
		return
	}
	if bf.sprintCooldown > 0 {
		bf.sprintCooldown--
		if bf.sprintCooldown == 0 {
			bf.sprintFrames = outrunSprintDurationFrames()
			bf.laughSFX = true
		}
	}
}

func (bf *BossFight) syncOutrunBossPos() {
	bf.boss.X = bf.playerX - bf.lead - bf.boss.width*0.35
	bf.boss.Y = float64(FloorSurfaceY) - bf.boss.height - outrunGroundClearance
}

func (bf *BossFight) Draw(screen *ebiten.Image) {
	switch {
	case bf.IsBoxing():
		bf.drawDadFighter(screen)
	case bf.IsDuel():
		bf.drawNeighbourBoss(screen)
		for _, p := range bf.projectiles {
			drawThrownGlassProjectile(screen, p)
		}
		if bf.enemyProjectile != nil {
			drawFootballProjectile(screen, bf.enemyProjectile)
		}
	default:
		bf.boss.Draw(screen)
		if bf.mode == bossModeThrow {
			for _, p := range bf.projectiles {
				drawProjectile(screen, p)
			}
		}
	}
	bf.drawParticles(screen)
}

func (bf *BossFight) runningFrameIndex() int {
	if bf.playerSpeed <= outrunCoastSpeed {
		return 0
	}
	return sprite.RunningFrameIndexFromTick(int64(bf.runAnimTick))
}

func drawBossPlayer(screen *ebiten.Image, chargePower float64) {
	if chargePower < 0 {
		chargePower = 0
	} else if chargePower > 1 {
		chargePower = 1
	}
	tilt := throwLaunchAngle * chargePower
	drawBeerGlass(screen, float64(BirdStartX), bossPlayerY, BirdWidth, BirdHeight, tilt)
}

func drawOutrunPlayer(screen *ebiten.Image, bf *BossFight) {
	const runW, runH = 201.6, 172.8 // 20% larger than 168×144
	cy := float64(FloorSurfaceY) - outrunGroundClearance - runH/2

	if bf.InCountdown() {
		drawStandingPlayer(screen, bf.playerX+28, cy, runW, runH, 0)
		return
	}

	// Slight bob and forward lean based on run speed.
	bob := math.Sin(bf.scrollX * 0.18) * 3
	speedFrac := (bf.playerSpeed - outrunCoastSpeed) / (outrunMaxPlayerSpeed - outrunCoastSpeed)
	if speedFrac < 0 {
		speedFrac = 0
	}
	if speedFrac > 1 {
		speedFrac = 1
	}
	tilt := -0.15 - 0.25*speedFrac
	drawRunningPlayer(screen, bf.playerX+28, cy+bob, runW, runH, tilt, bf.runningFrameIndex())
}
