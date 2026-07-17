package game

import "testing"

func TestApplyDebugStartDisabled(t *testing.T) {
	oldLevel, oldBoss := DebugStartLevel, DebugStartBoss
	DebugStartLevel = 0
	DebugStartBoss = 0
	defer func() {
		DebugStartLevel = oldLevel
		DebugStartBoss = oldBoss
	}()

	g := New()
	g.state = StateReady
	g.applyDebugStart()

	if g.state != StateReady {
		t.Errorf("state = %v, want StateReady when debug disabled", g.state)
	}
}

func TestApplyDebugStartLevel(t *testing.T) {
	oldLevel, oldBoss := DebugStartLevel, DebugStartBoss
	DebugStartLevel = 3
	DebugStartBoss = 0
	defer func() {
		DebugStartLevel = oldLevel
		DebugStartBoss = oldBoss
	}()

	g := New()
	g.applyDebugStart()

	if g.level != 3 {
		t.Errorf("level = %d, want 3", g.level)
	}
	if g.state != StatePlaying {
		t.Errorf("state = %v, want StatePlaying", g.state)
	}
}

func TestApplyDebugStartBossIntro(t *testing.T) {
	oldLevel, oldBoss := DebugStartLevel, DebugStartBoss
	DebugStartLevel = 3
	DebugStartBoss = 1
	defer func() {
		DebugStartLevel = oldLevel
		DebugStartBoss = oldBoss
	}()

	g := New()
	g.applyDebugStart()

	if g.level != 3 {
		t.Errorf("level = %d, want 3", g.level)
	}
	if g.state != StateBossIntro {
		t.Errorf("state = %v, want StateBossIntro", g.state)
	}
}
