package game

import "testing"

func TestThrowPowerClamp(t *testing.T) {
	if got := ThrowPower(0); got != 0 {
		t.Errorf("ThrowPower(0) = %v, want 0", got)
	}
	if got := ThrowPower(ThrowChargeMaxFrames()); got != 1 {
		t.Errorf("ThrowPower(max) = %v, want 1", got)
	}
	if got := ThrowPower(ThrowChargeMaxFrames() / 2); got != 0.5 {
		t.Errorf("ThrowPower(half) = %v, want 0.5", got)
	}
}

func TestProjectileRangeByPower(t *testing.T) {
	bossX := BossHomeXEstimate()
	shortMax := projectileFlightMaxX(0)
	fullMax := projectileFlightMaxX(1)

	if shortMax >= bossX {
		t.Errorf("short throw maxX=%.1f reaches boss homeX=%.1f; expected to fall short", shortMax, bossX)
	}
	if fullMax < bossX {
		t.Errorf("full throw maxX=%.1f does not reach boss homeX=%.1f", fullMax, bossX)
	}
	if fullMax <= shortMax {
		t.Errorf("full throw maxX=%.1f should exceed short maxX=%.1f", fullMax, shortMax)
	}
}

func TestBossIsOutrun(t *testing.T) {
	for _, level := range []int{3, 6, 9, 12} {
		if !BossIsOutrun(level) {
			t.Errorf("BossIsOutrun(%d) should be true", level)
		}
	}
	for _, level := range []int{1, 2, 4, 5, 7, 8, 0, -1} {
		if BossIsOutrun(level) {
			t.Errorf("BossIsOutrun(%d) should be false", level)
		}
	}
}

func skipOutrunCountdown(bf *BossFight) {
	for bf.InCountdown() {
		bf.Update()
	}
}

func TestOutrunReset(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(3)
	if !bf.IsOutrun() {
		t.Fatal("level 3 should be outrun mode")
	}
	if !bf.InCountdown() {
		t.Fatal("level 3 should start in countdown")
	}
	if bf.CountdownDisplay() != "3" {
		t.Errorf("CountdownDisplay() = %q, want 3", bf.CountdownDisplay())
	}
	if bf.TimerRemaining() != OutrunDurationFramesForLevel(3) {
		t.Errorf("timer = %d, want %d", bf.TimerRemaining(), OutrunDurationFramesForLevel(3))
	}
	if bf.lead != OutrunStartLeadForLevel(3) {
		t.Errorf("lead = %v, want %v", bf.lead, OutrunStartLeadForLevel(3))
	}

	bf.Reset(6)
	if !bf.IsOutrun() {
		t.Fatal("level 6 should be outrun mode")
	}
	if bf.TimerRemaining() != OutrunDurationFramesForLevel(6) {
		t.Errorf("level 6 timer = %d, want %d", bf.TimerRemaining(), OutrunDurationFramesForLevel(6))
	}
	if bf.lead != OutrunStartLeadForLevel(6) {
		t.Errorf("level 6 lead = %v, want %v", bf.lead, OutrunStartLeadForLevel(6))
	}

	bf.Reset(1)
	if bf.IsOutrun() {
		t.Fatal("level 1 should not be outrun mode")
	}
	if bf.CanTap() {
		t.Fatal("throw mode should not allow tap")
	}
}

func TestOutrunCountdownFreezesRace(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(3)
	startLead := bf.lead
	startTimer := bf.TimerRemaining()

	for i := 0; i < OutrunCountdownTotalFrames(); i++ {
		if !bf.InCountdown() {
			t.Fatalf("expected countdown still active at frame %d", i)
		}
		won, lost := bf.Update()
		if won || lost {
			t.Fatalf("countdown frame %d returned won=%v lost=%v", i, won, lost)
		}
		if bf.lead != startLead {
			t.Fatalf("lead changed during countdown: %v -> %v", startLead, bf.lead)
		}
		if bf.TimerRemaining() != startTimer {
			t.Fatalf("timer changed during countdown: %d -> %d", startTimer, bf.TimerRemaining())
		}
	}
	if bf.InCountdown() {
		t.Fatal("countdown should be finished")
	}
}

func TestOutrunTapIgnoredDuringCountdown(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(3)
	before := bf.lead
	bf.Tap()
	if bf.lead != before {
		t.Errorf("tap during countdown changed lead: %v -> %v", before, bf.lead)
	}
	if bf.CanTap() {
		t.Error("CanTap should be false during countdown")
	}
	if bf.ConsumeTapSFX() {
		t.Error("tap during countdown should not set SFX")
	}
}

func TestOutrunCountdownDisplaySequence(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(3)
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

func TestOutrunTapIncreasesLead(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(3)
	skipOutrunCountdown(bf)
	before := bf.lead
	bf.Tap()
	if bf.lead <= before {
		t.Errorf("tap should increase lead: before=%v after=%v", before, bf.lead)
	}
	if !bf.ConsumeTapSFX() {
		t.Error("tap should set tap SFX flag")
	}
}

func TestOutrunCatchWithoutTaps(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(3)
	skipOutrunCountdown(bf)
	lost := false
	for i := 0; i < OutrunDurationFramesForLevel(3); i++ {
		_, l := bf.Update()
		if l {
			lost = true
			break
		}
	}
	if !lost {
		t.Fatal("expected lose when never tapping")
	}
}

func TestOutrunSurviveWithTaps(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(3)
	skipOutrunCountdown(bf)
	won := false
	dur := OutrunDurationFramesForLevel(3)
	for i := 0; i < dur+deathTipFrames+deathExplodeHold+deathSettleFrames+10; i++ {
		// ~4 taps/sec keeps the gap.
		if i%15 == 0 {
			bf.Tap()
		}
		w, l := bf.Update()
		if l {
			t.Fatalf("lost at frame %d with lead=%v timer=%d", i, bf.lead, bf.timerFrames)
		}
		if w {
			won = true
			break
		}
	}
	if !won {
		t.Fatal("expected win after surviving timer with steady taps")
	}
}

func TestOutrunHarderEachChase(t *testing.T) {
	if OutrunDurationFramesForLevel(6) <= OutrunDurationFramesForLevel(3) {
		t.Error("level 6 duration should exceed level 3")
	}
	if OutrunStartLeadForLevel(6) >= OutrunStartLeadForLevel(3) {
		t.Error("level 6 start lead should be shorter than level 3")
	}
	if OutrunWifeCloseBaseForLevel(6) <= OutrunWifeCloseBaseForLevel(3) {
		t.Error("level 6 wife close base should exceed level 3")
	}
	if OutrunWifeCloseRiseForLevel(6) <= OutrunWifeCloseRiseForLevel(3) {
		t.Error("level 6 wife close rise should exceed level 3")
	}
	if OutrunTapLeadBoostForLevel(6) >= OutrunTapLeadBoostForLevel(3) {
		t.Error("level 6 tap boost should be weaker than level 3")
	}
}

func TestOutrunLaughSFX(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(3)
	if bf.laughCooldown < outrunLaughMinFrames || bf.laughCooldown > outrunLaughMaxFrames {
		t.Fatalf("initial laughCooldown %d outside [%d,%d]", bf.laughCooldown, outrunLaughMinFrames, outrunLaughMaxFrames)
	}
	bf.laughCooldown = 60

	for bf.InCountdown() {
		bf.Update()
		if bf.ConsumeLaughSFX() {
			t.Fatal("laugh SFX should not fire during countdown")
		}
	}

	fired := false
	for i := 0; i < outrunLaughMaxFrames+10; i++ {
		if i%15 == 0 {
			bf.Tap()
		}
		bf.Update()
		if bf.ConsumeLaughSFX() {
			fired = true
			break
		}
	}
	if !fired {
		t.Fatal("expected laugh SFX after cooldown elapses")
	}
	if bf.laughCooldown < outrunLaughMinFrames || bf.laughCooldown > outrunLaughMaxFrames {
		t.Errorf("rescheduled laughCooldown %d outside [%d,%d]", bf.laughCooldown, outrunLaughMinFrames, outrunLaughMaxFrames)
	}
	if bf.ConsumeLaughSFX() {
		t.Error("ConsumeLaughSFX should clear the flag")
	}
}
