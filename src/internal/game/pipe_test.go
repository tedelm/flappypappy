package game

import "testing"

func TestPipeSpawnLimit(t *testing.T) {
	pm := NewPipeManager()
	pm.ApplyConfig(DifficultyEasy.Config())
	pm.Reset()
	pm.SetSpawnLimit(3)
	pm.spawnTimer = 0

	for i := 0; i < 500; i++ {
		pm.Update()
	}

	if pm.Spawned() != 3 {
		t.Fatalf("Spawned() = %d, want 3", pm.Spawned())
	}
	if len(pm.Pipes()) > 3 {
		t.Fatalf("len(pipes) = %d, want <= 3", len(pm.Pipes()))
	}
}

func TestLastDadFullyCleared(t *testing.T) {
	pm := NewPipeManager()
	pm.pipes = []*Pipe{{
		X:    100,
		GapY: 120,
		GapH: 150,
	}}

	if pm.LastDadFullyCleared(100 + PipeWidth) {
		t.Fatal("expected not cleared when birdX == pipe right edge")
	}
	if !pm.LastDadFullyCleared(100 + PipeWidth + 1) {
		t.Fatal("expected cleared when birdX past pipe right edge")
	}

	pm.pipes = nil
	if !pm.LastDadFullyCleared(0) {
		t.Fatal("no pipes should count as fully cleared")
	}
}

func TestResetBeforeLastDad(t *testing.T) {
	pm := NewPipeManager()
	pm.ApplyConfig(DifficultyEasy.Config())
	pm.SetSpawnLimit(10)
	pm.Reset()
	pm.spawn()
	pm.spawn()

	pm.ResetBeforeLastDad()

	if pm.Spawned() != 10 {
		t.Fatalf("Spawned() = %d, want 10", pm.Spawned())
	}
	if len(pm.Pipes()) != 1 {
		t.Fatalf("len(pipes) = %d, want 1", len(pm.Pipes()))
	}
	p := pm.Pipes()[0]
	if p.Scored {
		t.Fatal("last dad should be unscored")
	}
	if p.X <= BirdStartX {
		t.Fatalf("last dad X=%v should be ahead of BirdStartX=%v", p.X, BirdStartX)
	}

	pm.spawnTimer = 0
	for i := 0; i < 200; i++ {
		pm.Update()
	}
	if pm.Spawned() != 10 {
		t.Fatalf("after Update Spawned() = %d, want 10 (no extra spawns)", pm.Spawned())
	}
}

func TestContinueRewindsFinalDad(t *testing.T) {
	g := New()
	g.difficulty = DifficultyEasy
	g.level = 1
	target := DadsRequiredForLevel(g.level)
	g.levelDadsPassed = target
	g.rawScore = target
	g.lives = 2
	g.finalDadAttempt = true
	g.pipes.SetSpawnLimit(target)

	g.continueGame()

	if g.levelDadsPassed != target-1 {
		t.Fatalf("levelDadsPassed = %d, want %d", g.levelDadsPassed, target-1)
	}
	if g.rawScore != target-1 {
		t.Fatalf("rawScore = %d, want %d", g.rawScore, target-1)
	}
	if g.finalDadAttempt {
		t.Fatal("finalDadAttempt should be cleared")
	}
	if len(g.pipes.Pipes()) != 1 {
		t.Fatalf("len(pipes) = %d, want 1", len(g.pipes.Pipes()))
	}
	if g.state != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying", g.state)
	}
}

func TestContinueBeforeScoringLastDadKeepsScore(t *testing.T) {
	g := New()
	g.difficulty = DifficultyEasy
	g.level = 1
	target := DadsRequiredForLevel(g.level)
	g.levelDadsPassed = target - 1
	g.rawScore = target - 1
	g.lives = 2
	g.finalDadAttempt = true
	g.pipes.SetSpawnLimit(target)

	g.continueGame()

	if g.rawScore != target-1 {
		t.Fatalf("rawScore = %d, want %d (unchanged)", g.rawScore, target-1)
	}
	if g.levelDadsPassed != target-1 {
		t.Fatalf("levelDadsPassed = %d, want %d", g.levelDadsPassed, target-1)
	}
}

func TestBossLossContinueRestartsCurrentLevel(t *testing.T) {
	g := New()
	g.difficulty = DifficultyEasy
	g.level = 2
	target := DadsRequiredForLevel(g.level)
	g.levelDadsPassed = target
	g.rawScore = target + 3 // prior levels + this level's dads
	g.lives = 2
	g.state = StateBossFight

	g.bossGameOver()

	if g.lives != 1 {
		t.Fatalf("lives = %d, want 1", g.lives)
	}
	if g.state != StateContinue {
		t.Fatalf("state = %v, want StateContinue", g.state)
	}
	if !g.bossContinue {
		t.Fatal("bossContinue should be set")
	}

	g.continueGame()

	if g.lives != 1 {
		t.Fatalf("after continue lives = %d, want 1", g.lives)
	}
	if g.level != 2 {
		t.Fatalf("level = %d, want 2", g.level)
	}
	if g.levelDadsPassed != 0 {
		t.Fatalf("levelDadsPassed = %d, want 0", g.levelDadsPassed)
	}
	if g.rawScore != 3 {
		t.Fatalf("rawScore = %d, want 3", g.rawScore)
	}
	if g.bossContinue {
		t.Fatal("bossContinue should be cleared")
	}
	if g.state != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying", g.state)
	}
}

func TestBossLossNoLivesGameOver(t *testing.T) {
	g := New()
	g.lives = 1
	g.state = StateBossFight

	g.bossGameOver()

	if g.lives != 0 {
		t.Fatalf("lives = %d, want 0", g.lives)
	}
	if g.state != StateEnterName {
		t.Fatalf("state = %v, want StateEnterName", g.state)
	}
	if g.bossContinue {
		t.Fatal("bossContinue should not be set on final death")
	}
}

