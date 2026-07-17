package game

import (
	"testing"

	"flappy/internal/game/sprite"
)

func TestBossVariantForLevel(t *testing.T) {
	if got := BossVariantForLevel(4); got != BossVariantNeighbour {
		t.Errorf("BossVariantForLevel(4) = %d, want BossVariantNeighbour", got)
	}
	if got := BossVariantForLevel(5); got != BossVariantDad {
		t.Errorf("BossVariantForLevel(5) = %d, want BossVariantDad", got)
	}
	for _, level := range []int{1, 2, 3, 6, 7} {
		if got := BossVariantForLevel(level); got != BossVariantLady {
			t.Errorf("BossVariantForLevel(%d) = %d, want BossVariantLady", level, got)
		}
	}
}

func TestDuelResetStartsPlayerTurn(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(4)
	if !bf.IsDuel() {
		t.Fatal("level 4 should be duel mode")
	}
	if bf.DuelTurn() != duelTurnPlayer {
		t.Errorf("duelTurn = %d, want duelTurnPlayer", bf.DuelTurn())
	}
	if !bf.CanThrow() {
		t.Fatal("should be able to throw on player turn")
	}
	if bf.CanDodge() {
		t.Fatal("should not dodge on player turn")
	}
}

func TestDuelTurnAlternation(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(4)

	bf.duelThrow(1)
	if bf.DuelTurn() != duelTurnPlayerResolve {
		t.Fatalf("after throw duelTurn = %d, want duelTurnPlayerResolve", bf.DuelTurn())
	}
	if bf.CanThrow() {
		t.Fatal("should not throw while glass is resolving")
	}

	// Fast-forward throw anim and clear projectile.
	for i := 0; i < sprite.PlayerThrowFrameCount*sprite.PlayerThrowFrameTicks+5; i++ {
		bf.updateDuelPlayerTurn()
	}
	for len(bf.projectiles) > 0 {
		bf.projectiles = nil
		bf.updateDuelPlayerTurn()
	}

	if bf.DuelTurn() != duelTurnEnemy {
		t.Fatalf("after player resolve duelTurn = %d, want duelTurnEnemy", bf.DuelTurn())
	}

	// Advance enemy wind-up to football resolve.
	for bf.DuelTurn() == duelTurnEnemy {
		bf.updateDuelEnemyWindUp()
	}
	if bf.DuelTurn() != duelTurnEnemyResolve {
		t.Fatalf("after enemy wind-up duelTurn = %d, want duelTurnEnemyResolve", bf.DuelTurn())
	}
	if !bf.CanDodge() {
		t.Fatal("should dodge during enemy resolve")
	}

	// Move football off screen.
	if bf.enemyProjectile != nil {
		bf.enemyProjectile.X = -footballSize * 2
		bf.enemyProjectile.Y = bf.playerY
	}
	bf.updateDuelEnemyResolve()

	if bf.DuelTurn() != duelTurnPlayer {
		t.Fatalf("after enemy resolve duelTurn = %d, want duelTurnPlayer", bf.DuelTurn())
	}
}

func TestDuelFootballHitsTriggerLifeLost(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(4)
	bf.duelTurn = duelTurnEnemyResolve
	bf.enemyProjectile = &FootballProjectile{X: duelPlayerX(), Y: bf.playerY}

	bf.footballHits = 0
	bf.updateDuelEnemyResolve()
	if bf.ConsumeLifeLost() {
		t.Fatal("life lost should not trigger after first hit")
	}
	if bf.footballHits != 1 {
		t.Fatalf("footballHits = %d, want 1", bf.footballHits)
	}

	bf.duelTurn = duelTurnEnemyResolve
	bf.enemyProjectile = &FootballProjectile{X: duelPlayerX(), Y: bf.playerY}
	bf.updateDuelEnemyResolve()
	if !bf.ConsumeLifeLost() {
		t.Fatal("life lost should trigger after second hit")
	}
	if bf.footballHits != 0 {
		t.Errorf("footballHits = %d, want 0 after life lost", bf.footballHits)
	}
}

func TestDuelThrowSpawnsGlassAtReleaseFrame(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(4)
	bf.duelThrow(0.8)
	if len(bf.projectiles) != 0 {
		t.Fatal("glass should not spawn immediately on throw")
	}

	spawnTick := sprite.PlayerThrowReleaseFrame * sprite.PlayerThrowFrameTicks
	for i := 0; i < spawnTick-1; i++ {
		bf.updateDuelPlayerTurn()
		if bf.glassSpawned {
			t.Fatalf("glass spawned too early at tick %d", i+1)
		}
	}
	bf.updateDuelPlayerTurn()
	if !bf.glassSpawned {
		t.Fatal("glass should spawn at release frame")
	}
}
