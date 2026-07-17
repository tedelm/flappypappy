package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"flappy/internal/game/sprite"
)

const (
	boxingWinKOFrames     = 90
	boxingLoseKOFrames    = 90
	boxingPlayerHitFrames = 18

	boxingPunchJab      = 0
	boxingPunchCross    = 1
	boxingPunchUppercut = 2

	boxingDadStraight = 0
	boxingDadUppercut = 1
)

func (bf *BossFight) IsBoxing() bool {
	return bf.mode == bossModeBoxing
}

func (bf *BossFight) boxingLosePending() bool {
	return bf.IsBoxing() && !bf.pendingWin && bf.deathPhase != deathPhaseNone
}

func (bf *BossFight) BoxingVictoryReady() bool {
	return bf.IsBoxing() && bf.boxingVictoryHold
}

func (bf *BossFight) resetBoxingBoss(level int) {
	frame := sprite.DadFighterFrame(sprite.DadFighterIdle0)
	contentH := sprite.DadFighterContentH()
	if contentH <= 0 {
		contentH = float64(frame.Bounds().Dy())
	}
	scale := boxingDadDisplayH / contentH
	bossW := float64(frame.Bounds().Dx()) * scale
	feetY := float64(FloorSurfaceY) - boxingGroundClearance
	homeX := float64(ScreenW) - bossMarginX - bossW

	bf.variant = BossVariantDad
	bf.mode = bossModeBoxing
	bf.outrunDuration = BoxingDurationFrames
	bf.timerFrames = BoxingDurationFrames
	bf.countdownFrames = OutrunCountdownTotalFrames()
	bf.lead = boxingBossMaxHP
	bf.playerX = boxingPlayerScreenX
	bf.playerY = feetY - boxingPlayerDisplayH/2
	bf.playerSpeed = 0
	bf.scrollX = 0
	bf.runAnimTick = 0
	bf.sprintFrames = 0
	bf.sprintCooldown = boxingBurstCooldownFrames()
	bf.laughCooldown = 0
	bf.playerThrowFrame = 0
	bf.playerThrowTick = 0
	bf.playerThrowing = false
	bf.lastThrowPower = 0
	bf.throwsUsed = 0
	bf.enemyThrowFrame = 0
	bf.enemyThrowTick = 0
	bf.enemyThrowing = false
	bf.playerInvincible = 0
	bf.deathPhase = deathPhaseNone
	bf.deathTimer = 0
	bf.pendingWin = false
	bf.boxingVictoryHold = false
	bf.boxingFramesSinceTap = boxingDadSuppressFrames
	bf.boxingStamina = boxingStaminaMax
	bf.boxingExhausted = false
	bf.footballHits = boxingDadStraight // reuse as dad punch kind

	bf.boss = Boss{
		X:           homeX,
		Y:           feetY - boxingDadDisplayH,
		width:       bossW,
		height:      boxingDadDisplayH,
		homeX:       homeX,
		hitboxInset: BossHitboxInsetForLevel(level),
		faceRight:   false,
	}
}

func (bf *BossFight) updateBoxing() (won, lost bool) {
	if bf.boxingLosePending() {
		return bf.updateBoxingLose()
	}

	if bf.countdownFrames > 0 {
		bf.countdownFrames--
		return false, false
	}

	if bf.boss.shakeTimer > 0 {
		bf.boss.shakeTimer--
	}
	if bf.playerInvincible > 0 {
		bf.playerInvincible--
	}
	bf.boxingFramesSinceTap++

	if bf.boxingFramesSinceTap >= boxingDadSuppressFrames {
		bf.boxingStamina += boxingStaminaRegenPerFrame
		if bf.boxingStamina > boxingStaminaMax {
			bf.boxingStamina = boxingStaminaMax
		}
	}
	if bf.boxingExhausted && bf.boxingStamina >= boxingStaminaRecoverTo {
		bf.boxingExhausted = false
	}

	bf.updateBoxingAnims()
	bf.updateBoxingBurst()

	bf.timerFrames--
	if bf.timerFrames < 0 {
		bf.timerFrames = 0
	}

	if bf.lead <= 0 {
		bf.lead = 0
		bf.startBoxingWin()
		return false, false
	}
	if bf.timerFrames <= 0 {
		bf.startBoxingLose()
		return false, false
	}
	return false, false
}

func (bf *BossFight) updateBoxingBurst() {
	if bf.sprintFrames > 0 {
		bf.sprintFrames--
		if bf.sprintFrames == 0 {
			bf.sprintCooldown = boxingBurstCooldownFrames()
			bf.enemyThrowing = false
			if bf.footballHits == boxingDadStraight {
				bf.footballHits = boxingDadUppercut
			} else {
				bf.footballHits = boxingDadStraight
			}
		}
		return
	}
	if bf.sprintCooldown > 0 {
		bf.sprintCooldown--
		if bf.sprintCooldown == 0 {
			if bf.playerThrowing || bf.boxingFramesSinceTap < boxingDadSuppressFrames {
				bf.sprintCooldown = 1
				return
			}
			bf.sprintFrames = boxingBurstDurationFrames()
			bf.enemyThrowing = true
			bf.enemyThrowFrame = 0
			bf.enemyThrowTick = 0
			bf.playerInvincible = boxingPlayerHitFrames
			bf.hitSFX = true
			bf.timerFrames -= boxingHitTimePenaltyFrames
			if bf.timerFrames < 0 {
				bf.timerFrames = 0
			}
		}
	}
}

func (bf *BossFight) boxingPlayerPunchKind() int {
	if bf.throwsUsed <= 0 {
		return boxingPunchJab
	}
	return (bf.throwsUsed - 1) % 3
}

func (bf *BossFight) boxingPlayerPunchCount() int {
	switch bf.boxingPlayerPunchKind() {
	case boxingPunchCross:
		return sprite.PlayerPunchCrossCount
	case boxingPunchUppercut:
		return sprite.PlayerUppercutFrameCount
	default:
		return sprite.PlayerPunchJabCount
	}
}

func (bf *BossFight) boxingDadPunchCount() int {
	if bf.footballHits == boxingDadUppercut {
		return sprite.DadFighterUppercutCount
	}
	return sprite.DadFighterStraightCount
}

func (bf *BossFight) updateBoxingAnims() {
	if bf.playerThrowing {
		ticks := sprite.PlayerPunchFrameTicks
		if bf.boxingPlayerPunchKind() == boxingPunchUppercut {
			ticks = sprite.PlayerUppercutFrameTicks
		}
		bf.playerThrowTick++
		if bf.playerThrowTick >= ticks {
			bf.playerThrowTick = 0
			bf.playerThrowFrame++
			if bf.playerThrowFrame >= bf.boxingPlayerPunchCount() {
				bf.playerThrowing = false
				bf.playerThrowFrame = 0
			}
		}
	}

	if bf.enemyThrowing {
		bf.enemyThrowTick++
		if bf.enemyThrowTick >= sprite.DadFighterFrameTicks {
			bf.enemyThrowTick = 0
			bf.enemyThrowFrame++
			max := bf.boxingDadPunchCount()
			if bf.enemyThrowFrame >= max {
				bf.enemyThrowFrame = max - 1
			}
		}
	}

	bf.runAnimTick++
}

func (bf *BossFight) BoxingExhausted() bool {
	return bf.IsBoxing() && bf.boxingExhausted
}

func (bf *BossFight) StaminaFrac() float64 {
	if boxingStaminaMax <= 0 {
		return 0
	}
	f := bf.boxingStamina / boxingStaminaMax
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

func (bf *BossFight) boxingTap() {
	if bf.boxingExhausted || bf.boxingStamina < boxingTapStaminaCost {
		return
	}
	bf.boxingStamina -= boxingTapStaminaCost
	if bf.boxingStamina <= 0 {
		bf.boxingStamina = 0
		bf.boxingExhausted = true
	}
	bf.throwsUsed++
	dmg := boxingTapDamage
	if (bf.throwsUsed-1)%3 == boxingPunchUppercut {
		dmg = boxingUppercutDamage
	}
	bf.lead -= dmg
	if bf.lead < 0 {
		bf.lead = 0
	}
	bf.tapSFX = true
	bf.boss.shakeTimer = bossHitShakeFrames
	bf.playerThrowing = true
	bf.playerThrowTick = 0
	bf.playerThrowFrame = 0
	bf.boxingFramesSinceTap = 0
}

func (bf *BossFight) startBoxingWin() {
	bf.pendingWin = true
	bf.deathPhase = deathPhaseSettle
	bf.deathTimer = boxingWinKOFrames
	bf.boxingVictoryHold = false
	bf.boss.defeated = true
	bf.boss.shakeTimer = 0
	bf.enemyThrowing = false
	bf.playerThrowing = false
	bf.explodeSFX = true
}

func (bf *BossFight) startBoxingLose() {
	bf.pendingWin = false
	bf.deathPhase = deathPhaseSettle
	bf.deathTimer = boxingLoseKOFrames
	bf.boxingVictoryHold = false
	bf.timerFrames = 0
	bf.boss.shakeTimer = 0
	bf.enemyThrowing = false
	bf.playerThrowing = false
	bf.playerInvincible = 0
	bf.sprintFrames = 0
	bf.loseSFX = true
}

func (bf *BossFight) updateBoxingWin() (won, lost bool) {
	bf.runAnimTick++
	if bf.boxingVictoryHold {
		return false, false
	}
	if bf.deathTimer > 0 {
		bf.deathTimer--
		if bf.deathTimer <= 0 {
			bf.boxingVictoryHold = true
		}
	}
	return false, false
}

func (bf *BossFight) updateBoxingLose() (won, lost bool) {
	bf.deathTimer--
	bf.runAnimTick++
	if bf.deathTimer <= 0 {
		bf.deathPhase = deathPhaseNone
		return false, true
	}
	return false, false
}

func (bf *BossFight) drawDadFighter(screen *ebiten.Image) {
	b := &bf.boss
	if b.hidden {
		return
	}

	var frame *ebiten.Image
	switch {
	case b.defeated:
		if bf.boxingVictoryHold {
			frame = sprite.DadFighterBeatenFrame(sprite.DadFighterBeatenFrameCount - 1)
		} else {
			elapsed := boxingWinKOFrames - bf.deathTimer
			idx := elapsed * sprite.DadFighterBeatenFrameCount / boxingWinKOFrames
			if idx >= sprite.DadFighterBeatenFrameCount {
				idx = sprite.DadFighterBeatenFrameCount - 1
			}
			if idx < 0 {
				idx = 0
			}
			frame = sprite.DadFighterBeatenFrame(idx)
		}
	case bf.boxingLosePending():
		frame = sprite.DadFighterFrame(sprite.DadFighterIdleFrameIndex(bf.runAnimTick))
	case b.shakeTimer > 0:
		frame = sprite.DadFighterFrame(sprite.DadFighterHitFrame)
	case bf.enemyThrowing || bf.sprintFrames > 0:
		start := sprite.DadFighterStraightStart
		count := sprite.DadFighterStraightCount
		if bf.footballHits == boxingDadUppercut {
			start = sprite.DadFighterUppercutStart
			count = sprite.DadFighterUppercutCount
		}
		idx := bf.enemyThrowFrame
		if idx < 0 {
			idx = 0
		}
		if idx >= count {
			idx = count - 1
		}
		frame = sprite.DadFighterFrame(start + idx)
	default:
		frame = sprite.DadFighterFrame(sprite.DadFighterIdleFrameIndex(bf.runAnimTick))
	}

	contentH := sprite.DadFighterContentH()
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

func drawBoxingPlayer(screen *ebiten.Image, bf *BossFight) {
	cx := bf.playerX
	feetY := float64(FloorSurfaceY) - boxingGroundClearance
	cy := feetY - boxingPlayerDisplayH/2

	var frame *ebiten.Image
	switch {
	case bf.pendingWin && bf.boss.defeated:
		idx := (bf.runAnimTick / 8) % sprite.PlayerPunchWinFrameCount
		frame = sprite.PlayerPunchWinFrame(idx)
	case bf.boxingLosePending():
		elapsed := boxingLoseKOFrames - bf.deathTimer
		idx := elapsed * sprite.PlayerPunchLoseFrameCount / boxingLoseKOFrames
		if idx >= sprite.PlayerPunchLoseFrameCount {
			idx = sprite.PlayerPunchLoseFrameCount - 1
		}
		if idx < 0 {
			idx = 0
		}
		frame = sprite.PlayerPunchLoseFrame(idx)
	case bf.boxingExhausted:
		frame = sprite.PlayerPunchFrame(sprite.PlayerPunchGuard)
	case bf.playerInvincible > 0:
		frame = sprite.PlayerPunchFrame(sprite.PlayerPunchGuard)
	case bf.playerThrowing:
		idx := bf.playerThrowFrame
		count := bf.boxingPlayerPunchCount()
		if idx >= count {
			idx = count - 1
		}
		switch bf.boxingPlayerPunchKind() {
		case boxingPunchCross:
			frame = sprite.PlayerPunchFrame(sprite.PlayerPunchCrossStart + idx)
		case boxingPunchUppercut:
			frame = sprite.PlayerUppercutFrame(idx)
		default:
			frame = sprite.PlayerPunchFrame(sprite.PlayerPunchJabStart + idx)
		}
	default:
		frame = sprite.StandingReadyFrame(sprite.StandingReadyFrameIndex(int64(bf.runAnimTick)))
	}

	drawSpritePlayer(screen, frame, cx, cy, 0, boxingPlayerDisplayH, 0)
}
