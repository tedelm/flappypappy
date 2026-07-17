package game

import (
	"testing"

	"flappy/internal/game/sprite"
)

func skipBoxingCountdown(bf *BossFight) {
	for bf.InCountdown() {
		bf.Update()
	}
}

func TestBoxingReset(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(5)
	if !bf.IsBoxing() {
		t.Fatal("level 5 should be boxing mode")
	}
	if bf.IsOutrun() {
		t.Fatal("boxing should not be outrun")
	}
	if bf.IsDuel() {
		t.Fatal("boxing should not be duel")
	}
	if bf.variant != BossVariantDad {
		t.Errorf("variant = %d, want BossVariantDad", bf.variant)
	}
	if bf.TimerRemaining() != BoxingDurationFrames {
		t.Errorf("timer = %d, want %d", bf.TimerRemaining(), BoxingDurationFrames)
	}
	if bf.lead != boxingBossMaxHP {
		t.Errorf("boss HP = %v, want %v", bf.lead, boxingBossMaxHP)
	}
	if !bf.InCountdown() {
		t.Fatal("expected countdown on reset")
	}
	if bf.CanTap() {
		t.Fatal("should not tap during countdown")
	}
}

func TestBossIsBoxing(t *testing.T) {
	if !BossIsBoxing(5) {
		t.Error("BossIsBoxing(5) should be true")
	}
	for _, level := range []int{1, 2, 3, 4, 6, 7} {
		if BossIsBoxing(level) {
			t.Errorf("BossIsBoxing(%d) should be false", level)
		}
	}
}

func TestDadFighterUppercutCount(t *testing.T) {
	if sprite.DadFighterUppercutCount != 2 {
		t.Errorf("DadFighterUppercutCount = %d, want 2", sprite.DadFighterUppercutCount)
	}
}

func TestBoxingTapIgnoredDuringCountdown(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(5)
	before := bf.lead
	bf.Tap()
	if bf.lead != before {
		t.Errorf("tap during countdown changed HP: %v -> %v", before, bf.lead)
	}
	if bf.CanTap() {
		t.Error("CanTap should be false during countdown")
	}
	if bf.ConsumeTapSFX() {
		t.Error("tap during countdown should not set SFX")
	}
}

func TestBoxingTapDamagesBoss(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(5)
	skipBoxingCountdown(bf)
	before := bf.lead
	bf.Tap()
	if bf.lead >= before {
		t.Errorf("tap should reduce boss HP: before=%v after=%v", before, bf.lead)
	}
	if !bf.ConsumeTapSFX() {
		t.Error("tap should set tap SFX flag")
	}
	if !bf.playerThrowing {
		t.Error("tap should start punch anim")
	}
	if bf.BossHPFrac() >= 1 {
		t.Error("BossHPFrac should drop after tap")
	}
}

func TestBoxingLoseWithoutTaps(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(5)
	skipBoxingCountdown(bf)
	lost := false
	for i := 0; i < BoxingDurationFrames+boxingLoseKOFrames+10; i++ {
		_, l := bf.Update()
		if l {
			lost = true
			break
		}
	}
	if !lost {
		t.Fatal("expected lose when timer runs out without tapping")
	}
}

func TestBoxingWinHoldsForContinue(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(5)
	skipBoxingCountdown(bf)
	ready := false
	for i := 0; i < BoxingDurationFrames+boxingWinKOFrames+10; i++ {
		if bf.CanTap() {
			bf.Tap()
		}
		w, l := bf.Update()
		if l {
			t.Fatalf("lost at frame %d with HP=%v timer=%d", i, bf.lead, bf.timerFrames)
		}
		if w {
			t.Fatal("boxing should not auto-return won; expect victory hold")
		}
		if bf.BoxingVictoryReady() {
			ready = true
			break
		}
	}
	if !ready {
		t.Fatal("expected BoxingVictoryReady after depleting boss HP")
	}
	if bf.CanTap() {
		t.Error("should not tap during victory hold")
	}
	w, l := bf.Update()
	if w || l {
		t.Errorf("victory hold Update should stay frozen, got won=%v lost=%v", w, l)
	}
}

func TestBoxingCountdownDisplay(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(5)
	seen := map[string]bool{}
	for bf.InCountdown() {
		seen[bf.CountdownDisplay()] = true
		bf.Update()
	}
	for _, want := range []string{"3", "2", "1", "GO!"} {
		if !seen[want] {
			t.Errorf("never saw countdown display %q", want)
		}
	}
}

func TestBoxingDadHitRemovesFiveSeconds(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(5)
	skipBoxingCountdown(bf)
	bf.sprintCooldown = 1
	bf.boxingFramesSinceTap = boxingDadSuppressFrames
	bf.playerThrowing = false
	before := bf.timerFrames
	bf.Update()
	if !bf.enemyThrowing {
		t.Fatal("expected dad punch to start")
	}
	// One normal tick + 5s penalty.
	want := before - 1 - boxingHitTimePenaltyFrames
	if bf.timerFrames != want {
		t.Errorf("timer = %d, want %d (before=%d penalty=%d)", bf.timerFrames, want, before, boxingHitTimePenaltyFrames)
	}
	if !bf.ConsumeHitSFX() {
		t.Fatal("expected hitSFX after dad punch")
	}
	if bf.ConsumeHitSFX() {
		t.Fatal("ConsumeHitSFX should clear the flag")
	}
}

func TestBoxingMashSuppressesDadPunch(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(5)
	skipBoxingCountdown(bf)
	bf.sprintCooldown = 1
	bf.boxingFramesSinceTap = 0
	bf.playerThrowing = false
	beforeTimer := bf.timerFrames
	bf.Update()
	if bf.enemyThrowing || bf.sprintFrames > 0 {
		t.Fatal("dad should not punch while player is mashing")
	}
	if bf.sprintCooldown != 1 {
		t.Errorf("sprintCooldown = %d, want 1 (deferred)", bf.sprintCooldown)
	}
	if beforeTimer-bf.timerFrames != 1 {
		t.Errorf("expected only normal 1-frame drain while suppressed, got delta %d", beforeTimer-bf.timerFrames)
	}

	// After silence long enough, punch can start.
	bf.boxingFramesSinceTap = boxingDadSuppressFrames
	bf.sprintCooldown = 1
	bf.Update()
	if !bf.enemyThrowing {
		t.Fatal("expected dad punch after player stopped mashing")
	}
}

func TestBoxingExhaustFromMashing(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(5)
	skipBoxingCountdown(bf)
	taps := int(boxingStaminaMax/boxingTapStaminaCost) + 1
	for i := 0; i < taps; i++ {
		bf.Tap()
	}
	if !bf.BoxingExhausted() {
		t.Fatal("expected exhausted after mashing taps")
	}
	if bf.CanTap() {
		t.Fatal("CanTap should be false while exhausted")
	}
	hpBefore := bf.lead
	bf.Tap()
	if bf.lead != hpBefore {
		t.Errorf("tap while exhausted should not damage boss: %v -> %v", hpBefore, bf.lead)
	}
}

func TestBoxingRecoverFromExhaust(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(5)
	skipBoxingCountdown(bf)
	bf.boxingStamina = 0
	bf.boxingExhausted = true
	bf.boxingFramesSinceTap = boxingDadSuppressFrames
	for i := 0; i < 200; i++ {
		bf.Update()
		if !bf.BoxingExhausted() {
			break
		}
	}
	if bf.BoxingExhausted() {
		t.Fatalf("expected recovery, stamina=%v", bf.boxingStamina)
	}
	if bf.boxingStamina < boxingStaminaRecoverTo {
		t.Errorf("stamina = %v, want >= %v", bf.boxingStamina, boxingStaminaRecoverTo)
	}
	if !bf.CanTap() {
		t.Fatal("CanTap should be true after recovery")
	}
}
