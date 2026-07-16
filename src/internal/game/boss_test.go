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
