package game

import "testing"

func TestLevelExitStartsAfterLastDad(t *testing.T) {
	g := New()
	g.difficulty = DifficultyEasy
	g.level = 1
	target := DadsRequiredForLevel(g.level)
	g.levelDadsPassed = target
	g.rawScore = target
	g.state = StatePlaying
	g.pipes.SetSpawnLimit(target)
	g.pipes.pipes = nil
	g.pipes.spawned = target
	g.bird.X = BirdStartX
	g.bird.Y = 120
	g.bird.VY = 0

	_ = g.Update()

	if g.state != StateLevelExit {
		t.Fatalf("state = %v, want StateLevelExit", g.state)
	}
}

func TestLevelExitLeadsToBossIntro(t *testing.T) {
	g := New()
	g.difficulty = DifficultyEasy
	g.level = 1
	g.state = StateLevelExit
	g.transitionFrames = 0
	g.bird.X = BirdStartX
	g.bird.Y = 120
	g.bird.VY = 0

	for i := 0; i < levelTransitionFrames+5; i++ {
		_ = g.Update()
		if g.state == StateBossIntro {
			break
		}
	}
	if g.state != StateBossIntro {
		t.Fatalf("state = %v, want StateBossIntro after exit", g.state)
	}
	if g.bird.X != BirdStartX {
		t.Fatalf("bird.X = %v, want reset to BirdStartX for intro", g.bird.X)
	}
}
