package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"flappy/internal/game/sprite"
)

const (
	duelTurnPlayer = iota
	duelTurnPlayerResolve
	duelTurnEnemy
	duelTurnEnemyResolve
)

const (
	duelPlayerDisplayH      = 100.0
	duelBossDisplayH        = 100.0
	duelPlayerXOffset       = 10.0
	duelGroundClearance     = 12.0
	duelFootballHitsForLife = 2
	duelInvincibleFrames    = 60
	duelFlapStrength        = -6.8
	duelGravity             = 0.5
	duelPlayerMinY          = 50.0
	footballSize            = 36.0
	footballLaunchSpeed     = 12.0
	thrownGlassW            = 32.0
	thrownGlassH            = 28.0
)

type FootballProjectile struct {
	X, Y, VX, VY float64
	Age          int
}

func duelPlayerX() float64 {
	return float64(BirdStartX) + duelPlayerXOffset
}

func duelGroundFeetY() float64 {
	return float64(FloorSurfaceY) - duelGroundClearance
}

func (bf *BossFight) IsDuel() bool {
	return bf.variant == BossVariantNeighbour && bf.mode == bossModeThrow
}

func (bf *BossFight) DuelTurn() int {
	return bf.duelTurn
}

func (bf *BossFight) CanDodge() bool {
	return bf.IsDuel() && !bf.pendingWin && bf.duelTurn == duelTurnEnemyResolve
}

func (bf *BossFight) DuelFlap() {
	if !bf.CanDodge() {
		return
	}
	bf.playerVY = duelFlapStrength
}

func (bf *BossFight) UpdateDuelPlayer() {
	if !bf.CanDodge() {
		return
	}
	bf.playerVY += duelGravity
	bf.playerY += bf.playerVY

	halfH := duelPlayerDisplayH / 2
	minY := duelPlayerMinY + halfH
	maxY := duelGroundFeetY() - halfH
	if bf.playerY < minY {
		bf.playerY = minY
		bf.playerVY = 0
	}
	if bf.playerY > maxY {
		bf.playerY = maxY
		bf.playerVY = 0
	}
}

func (bf *BossFight) ConsumeLifeLost() bool {
	if !bf.lifeLostPending {
		return false
	}
	bf.lifeLostPending = false
	return true
}

func (bf *BossFight) resetNeighbourBoss(level int) {
	frame := sprite.NeighbourStanding()
	contentH := sprite.NeighbourStandingContentH()
	scale := duelBossDisplayH / contentH
	bossW := float64(frame.Bounds().Dx()) * scale
	homeX := float64(ScreenW) - bossMarginX - bossW
	feetY := duelGroundFeetY()

	bf.variant = BossVariantNeighbour
	bf.mode = bossModeThrow
	bf.duelTurn = duelTurnPlayer
	bf.playerThrowFrame = 0
	bf.playerThrowTick = 0
	bf.playerThrowing = false
	bf.glassSpawned = false
	bf.enemyThrowFrame = 0
	bf.enemyThrowTick = 0
	bf.enemyThrowing = false
	bf.footballSpawned = false
	bf.enemyProjectile = nil
	bf.footballHits = 0
	bf.lifeLostPending = false
	bf.playerInvincible = 0
	bf.playerY = feetY - duelPlayerDisplayH/2
	bf.playerVY = 0

	bf.boss = Boss{
		X:           homeX,
		Y:           feetY - duelBossDisplayH,
		width:       bossW,
		height:      duelBossDisplayH,
		homeX:       homeX,
		hitboxInset: BossHitboxInsetForLevel(level),
	}
}

func (bf *BossFight) updateDuel() (won, lost bool) {
	if bf.playerInvincible > 0 {
		bf.playerInvincible--
	}

	if bf.boss.shakeTimer > 0 {
		bf.boss.shakeTimer--
	}

	switch bf.duelTurn {
	case duelTurnPlayer:
		// idle — waiting for player throw input
	case duelTurnPlayerResolve:
		won, lost = bf.updateDuelPlayerTurn()
	case duelTurnEnemy:
		bf.updateDuelEnemyWindUp()
	case duelTurnEnemyResolve:
		won, lost = bf.updateDuelEnemyResolve()
	}

	if won || lost || bf.pendingWin {
		return won, lost
	}
	if bf.throwsUsed >= bf.throwsAllowed && len(bf.projectiles) == 0 && bf.duelTurn == duelTurnPlayer && bf.hits < bf.hitsRequired {
		return false, true
	}
	return false, false
}

func (bf *BossFight) updateDuelPlayerTurn() (won, lost bool) {
	if bf.playerThrowing {
		bf.playerThrowTick++
		if bf.playerThrowTick >= sprite.PlayerThrowFrameTicks {
			bf.playerThrowTick = 0
			bf.playerThrowFrame++
			if bf.playerThrowFrame == sprite.PlayerThrowReleaseFrame && !bf.glassSpawned {
				bf.spawnDuelGlass()
			}
			if bf.playerThrowFrame >= sprite.PlayerThrowFrameCount {
				bf.playerThrowing = false
			}
		}
	}

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
		if p.X > float64(ScreenW)+thrownGlassW || p.Y > float64(FloorSurfaceY) {
			continue
		}
		px, py, pw, ph := p.duelBounds()
		bx, by, bw, bh := bf.boss.Bounds()
		if aabbOverlap(px, py, pw, ph, bx, by, bw, bh) {
			bf.hits++
			bf.hitSFX = true
			bf.boss.shakeTimer = bossHitShakeFrames
			bf.spawnSplash(p.X, p.Y)
			if bf.hits >= bf.hitsRequired {
				bf.startDeath()
				return false, false
			}
			continue
		}
		remaining = append(remaining, p)
	}
	bf.projectiles = remaining

	if !bf.playerThrowing && len(bf.projectiles) == 0 {
		bf.duelTurn = duelTurnEnemy
		bf.enemyThrowFrame = 0
		bf.enemyThrowTick = 0
		bf.enemyThrowing = true
		bf.footballSpawned = false
	}
	return false, false
}

func (bf *BossFight) spawnDuelGlass() {
	bf.glassSpawned = true
	power := bf.lastThrowPower
	speed := throwSpeedMin + (throwSpeedMax-throwSpeedMin)*power
	p := &Projectile{
		X:     duelPlayerX(),
		Y:     bf.duelPlayerThrowY(),
		VX:    math.Cos(throwLaunchAngle) * speed,
		VY:    math.Sin(throwLaunchAngle) * speed,
		Angle: throwLaunchAngle,
	}
	bf.projectiles = append(bf.projectiles, p)
	bf.spawnFoamBurst(p)
}

func (bf *BossFight) duelPlayerThrowY() float64 {
	if bf.duelTurn == duelTurnPlayer || bf.duelTurn == duelTurnPlayerResolve {
		return duelGroundFeetY() - duelPlayerDisplayH/2
	}
	return bf.playerY
}

func (bf *BossFight) updateDuelEnemyWindUp() {
	if !bf.enemyThrowing {
		return
	}
	bf.enemyThrowTick++
	if bf.enemyThrowTick >= sprite.NeighbourThrowFrameTicks {
		bf.enemyThrowTick = 0
		bf.enemyThrowFrame++
		if bf.enemyThrowFrame == sprite.NeighbourThrowReleaseFrame && !bf.footballSpawned {
			bf.spawnFootball()
		}
		if bf.enemyThrowFrame >= sprite.NeighbourThrowFrameCount {
			bf.enemyThrowing = false
			if !bf.footballSpawned {
				bf.spawnFootball()
			}
			bf.duelTurn = duelTurnEnemyResolve
			bf.playerY = duelGroundFeetY() - duelPlayerDisplayH/2
			bf.playerVY = 0
		}
	}
}

func (bf *BossFight) spawnFootball() {
	bf.footballSpawned = true
	cx, cy := bf.boss.center()
	startX := cx - bf.boss.width*0.25
	startY := cy - bf.boss.height*0.15
	targetX := duelPlayerX()
	targetY := bf.playerY
	dx := targetX - startX
	dy := targetY - startY
	dist := math.Hypot(dx, dy)
	if dist < 1 {
		dist = 1
	}
	bf.enemyProjectile = &FootballProjectile{
		X:  startX,
		Y:  startY,
		VX: dx / dist * footballLaunchSpeed,
		VY: dy / dist * footballLaunchSpeed,
	}
}

func (bf *BossFight) updateDuelEnemyResolve() (won, lost bool) {
	bf.UpdateDuelPlayer()

	fp := bf.enemyProjectile
	if fp == nil {
		bf.duelTurn = duelTurnPlayer
		return false, false
	}

	fp.VY += projectileGravity * 0.35
	fp.X += fp.VX
	fp.Y += fp.VY
	fp.Age++

	offScreen := fp.X < -footballSize || fp.X > float64(ScreenW)+footballSize ||
		fp.Y > float64(FloorSurfaceY)+footballSize || fp.Y < -footballSize

	if !offScreen && bf.playerInvincible <= 0 {
		px := duelPlayerX() - thrownGlassW/2
		py := bf.playerY - duelPlayerDisplayH/2
		pw := thrownGlassW
		ph := duelPlayerDisplayH * 0.85
		fx := fp.X - footballSize/2
		fy := fp.Y - footballSize/2
		if aabbOverlap(px, py, pw, ph, fx, fy, footballSize, footballSize) {
			bf.footballHits++
			bf.enemyProjectile = nil
			if bf.footballHits >= duelFootballHitsForLife {
				bf.footballHits = 0
				bf.lifeLostPending = true
				bf.playerInvincible = duelInvincibleFrames
			}
			bf.duelTurn = duelTurnPlayer
			return false, false
		}
	}

	if offScreen {
		bf.enemyProjectile = nil
		bf.duelTurn = duelTurnPlayer
	}
	return false, false
}

func (bf *BossFight) duelThrow(power float64) {
	if power < 0 {
		power = 0
	} else if power > 1 {
		power = 1
	}
	bf.lastThrowPower = power
	bf.throwsUsed++
	bf.duelTurn = duelTurnPlayerResolve
	bf.playerThrowing = true
	bf.playerThrowFrame = 0
	bf.playerThrowTick = 0
	bf.glassSpawned = false
}

func (p *Projectile) duelBounds() (x, y, w, h float64) {
	return p.X - thrownGlassW/2, p.Y - thrownGlassH/2, thrownGlassW, thrownGlassH
}

func (bf *BossFight) drawNeighbourBoss(screen *ebiten.Image) {
	b := &bf.boss
	if b.hidden {
		return
	}

	var frame *ebiten.Image
	if b.shakeTimer > 0 && !b.defeated {
		frame = sprite.NeighbourHit()
	} else if bf.enemyThrowing || bf.duelTurn == duelTurnEnemy {
		idx := bf.enemyThrowFrame
		if idx >= sprite.NeighbourThrowFrameCount {
			idx = sprite.NeighbourThrowFrameCount - 1
		}
		frame = sprite.NeighbourThrowFrame(idx)
	} else if b.defeated {
		frame = sprite.NeighbourStanding()
	} else {
		frame = sprite.NeighbourStanding()
	}

	contentH := sprite.NeighbourStandingContentH()
	if contentH <= 0 {
		contentH = float64(frame.Bounds().Dy())
	}
	scale := b.height / contentH
	sw := float64(frame.Bounds().Dx()) * scale
	sh := float64(frame.Bounds().Dy()) * scale
	cx, cy := b.center()

	ox, oy, rot := 0.0, 0.0, b.rotation
	if b.shakeTimer > 0 && !b.defeated {
		t := float64(b.shakeTimer)
		amp := bossHitShakeAmp * (t / bossHitShakeFrames)
		ox = math.Sin(t*1.8) * amp
		oy = math.Cos(t*2.3) * amp * 0.4
		rot += math.Sin(t*2.1) * bossHitShakeRot * (t / bossHitShakeFrames)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(-sw/2, -sh/2)
	op.GeoM.Scale(-1, 1) // face left toward player
	op.GeoM.Rotate(rot)
	op.GeoM.Translate(cx+ox, cy+oy)
	screen.DrawImage(frame, op)
}

func drawThrownGlassProjectile(screen *ebiten.Image, p *Projectile) {
	frameIdx := 0
	if p.Age > 8 {
		frameIdx = 1
	}
	frame := sprite.ThrownGlassFrame(frameIdx)
	fw := float64(frame.Bounds().Dx())
	fh := float64(frame.Bounds().Dy())
	if fw <= 0 || fh <= 0 {
		return
	}
	scale := thrownGlassH / fh
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-fw/2, -fh/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Rotate(p.Angle)
	op.GeoM.Translate(p.X, p.Y)
	screen.DrawImage(frame, op)
}

func drawFootballProjectile(screen *ebiten.Image, fp *FootballProjectile) {
	img := sprite.Football()
	fw := float64(img.Bounds().Dx())
	fh := float64(img.Bounds().Dy())
	if fw <= 0 || fh <= 0 {
		return
	}
	scale := footballSize / math.Max(fw, fh)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-fw/2, -fh/2)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(fp.X, fp.Y)
	screen.DrawImage(img, op)
}

func drawDuelPlayer(screen *ebiten.Image, bf *BossFight, chargePower float64) {
	cx := duelPlayerX()
	cy := bf.duelPlayerThrowY()
	if bf.CanDodge() {
		cy = bf.playerY
	}

	var frame *ebiten.Image
	if bf.playerThrowing {
		idx := bf.playerThrowFrame
		if idx >= sprite.PlayerThrowFrameCount {
			idx = sprite.PlayerThrowFrameCount - 1
		}
		frame = sprite.PlayerThrowFrame(idx)
	} else {
		frame = sprite.PlayerStanding()
	}

	angle := 0.0
	if !bf.playerThrowing && chargePower > 0 {
		angle = throwLaunchAngle * chargePower * 0.3
	}
	// Size by height only — sheets are single frames after slicing.
	drawSpritePlayer(screen, frame, cx, cy, 0, duelPlayerDisplayH, angle)
}
