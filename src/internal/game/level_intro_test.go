package game

import "testing"

func TestStartGameEntersLevelIntro(t *testing.T) {
	g := New()
	g.difficulty = DifficultyEasy
	g.startGame()
	if g.state != StateLevelIntro {
		t.Fatalf("state = %v, want StateLevelIntro", g.state)
	}
}

func TestLevelIntroAutoStartsPlaying(t *testing.T) {
	g := New()
	g.difficulty = DifficultyEasy
	g.startGame()

	for i := 0; i < levelTransitionFrames+5; i++ {
		_ = g.Update()
		if g.state == StatePlaying {
			break
		}
	}
	if g.state != StatePlaying {
		t.Fatalf("state = %v, want StatePlaying after intro", g.state)
	}
	if g.bird.X != BirdStartX {
		t.Fatalf("bird.X = %v, want BirdStartX after beginPlaying", g.bird.X)
	}
}

func TestAdvanceToNextLevelUsesIntro(t *testing.T) {
	g := New()
	g.difficulty = DifficultyEasy
	g.level = 1
	g.lives = MaxLives
	g.livesMax = MaxLives
	g.advanceToNextLevel()
	if g.level != 2 {
		t.Fatalf("level = %d, want 2", g.level)
	}
	if g.state != StateLevelIntro {
		t.Fatalf("state = %v, want StateLevelIntro", g.state)
	}
}
