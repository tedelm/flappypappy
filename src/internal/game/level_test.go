package game

import "testing"

func TestDadsRequiredForLevel(t *testing.T) {
	tests := []struct {
		level int
		want  int
	}{
		{1, 7},
		{2, 13},
		{3, 19},
		{4, 25},
		{5, 31},
		{0, 7},
	}
	for _, tt := range tests {
		if got := DadsRequiredForLevel(tt.level); got != tt.want {
			t.Errorf("DadsRequiredForLevel(%d) = %d, want %d", tt.level, got, tt.want)
		}
	}
}

func TestWallpaperIndex(t *testing.T) {
	tests := []struct {
		level int
		want  int
	}{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 2},
		{4, 3},
		{5, 4},
		{6, 5},
		{7, 0},
	}
	for _, tt := range tests {
		if got := WallpaperIndex(tt.level); got != tt.want {
			t.Errorf("WallpaperIndex(%d) = %d, want %d", tt.level, got, tt.want)
		}
	}
}

func TestLevelPaintingWindow(t *testing.T) {
	tests := []struct {
		level      int
		wantStart  int
		wantCount  int
	}{
		{1, 0, 3},
		{2, 2, 3},
		{5, 8, 3},
		{6, 10, 2},
	}
	for _, tt := range tests {
		gotStart, gotCount := LevelPaintingWindow(tt.level)
		if gotStart != tt.wantStart || gotCount != tt.wantCount {
			t.Fatalf("LevelPaintingWindow(%d) = (%d, %d), want (%d, %d)", tt.level, gotStart, gotCount, tt.wantStart, tt.wantCount)
		}
	}
}

func TestLevelPaintingPoolDeterministicAndUnique(t *testing.T) {
	a := LevelPaintingPool(7)
	b := LevelPaintingPool(7)
	if len(a) != 3 {
		t.Fatalf("LevelPaintingPool(7) len = %d, want 3", len(a))
	}
	if len(b) != 3 {
		t.Fatalf("LevelPaintingPool(7) second len = %d, want 3", len(b))
	}
	seen := make(map[int]bool, 3)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("LevelPaintingPool(7) not deterministic: %v vs %v", a, b)
		}
		if a[i] < 0 || a[i] >= 12 {
			t.Fatalf("LevelPaintingPool(7) index out of range: %d", a[i])
		}
		if seen[a[i]] {
			t.Fatalf("LevelPaintingPool(7) contains duplicate index: %v", a)
		}
		seen[a[i]] = true
	}
}

func TestLevelPaintingPoolVariesByLevel(t *testing.T) {
	a := LevelPaintingPool(7)
	b := LevelPaintingPool(8)
	same := len(a) == len(b)
	if same {
		for i := range a {
			if a[i] != b[i] {
				same = false
				break
			}
		}
	}
	if same {
		t.Fatalf("expected different pools for levels 7 and 8, got %v and %v", a, b)
	}
}

func TestBossConstants(t *testing.T) {
	if BossHitsRequired != 4 {
		t.Errorf("BossHitsRequired = %d, want 4", BossHitsRequired)
	}
	if BossThrowsAllowed != 10 {
		t.Errorf("BossThrowsAllowed = %d, want 10", BossThrowsAllowed)
	}
}

func TestBossHitsAndThrowsForLevel(t *testing.T) {
	cases := []struct {
		level, wantHits, wantThrows int
	}{
		{1, 4, 10},
		{2, 5, 11},
		{5, 8, 14},
		{0, 4, 10}, // clamped to level 1
	}
	for _, tt := range cases {
		hits := BossHitsForLevel(tt.level)
		throws := BossThrowsForLevel(tt.level)
		if hits != tt.wantHits || throws != tt.wantThrows {
			t.Errorf("level %d: hits/throws = %d/%d, want %d/%d",
				tt.level, hits, throws, tt.wantHits, tt.wantThrows)
		}
		if hits > throws {
			t.Errorf("level %d: hits %d exceeds throws %d", tt.level, hits, throws)
		}
	}
}

func TestBossMoveSpeedForLevel(t *testing.T) {
	if got := BossMoveSpeedForLevel(1); got != bossBaseMoveSpeed {
		t.Errorf("BossMoveSpeedForLevel(1) = %v, want %v", got, bossBaseMoveSpeed)
	}
	l5 := BossMoveSpeedForLevel(5)
	want5 := bossBaseMoveSpeed + 4*bossMoveSpeedPerLvl
	if l5 != want5 {
		t.Errorf("BossMoveSpeedForLevel(5) = %v, want %v", l5, want5)
	}
	if BossMoveSpeedForLevel(1) >= BossMoveSpeedForLevel(5) {
		t.Error("expected higher level to move faster")
	}
	if got := BossMoveSpeedForLevel(100); got != bossMaxMoveSpeed {
		t.Errorf("BossMoveSpeedForLevel(100) = %v, want cap %v", got, bossMaxMoveSpeed)
	}
}

func TestBossHitboxInsetForLevel(t *testing.T) {
	if got := BossHitboxInsetForLevel(1); got != bossBaseHitboxInset {
		t.Errorf("BossHitboxInsetForLevel(1) = %v, want %v", got, bossBaseHitboxInset)
	}
	l5 := BossHitboxInsetForLevel(5)
	want5 := bossBaseHitboxInset + 4*bossHitboxInsetPerLvl
	if l5 != want5 {
		t.Errorf("BossHitboxInsetForLevel(5) = %v, want %v", l5, want5)
	}
	if BossHitboxInsetForLevel(1) >= BossHitboxInsetForLevel(5) {
		t.Error("expected higher level to have larger hitbox inset")
	}
	if got := BossHitboxInsetForLevel(100); got != bossMaxHitboxInset {
		t.Errorf("BossHitboxInsetForLevel(100) = %v, want cap %v", got, bossMaxHitboxInset)
	}
}

func TestBossFightResetScalesWithLevel(t *testing.T) {
	bf := NewBossFight()
	bf.Reset(1)
	speed1 := bf.boss.baseSpeedY
	inset1 := bf.boss.hitboxInset

	bf.Reset(5)
	if bf.boss.baseSpeedY <= speed1 {
		t.Errorf("level 5 speed %v should exceed level 1 speed %v", bf.boss.baseSpeedY, speed1)
	}
	if bf.boss.hitboxInset <= inset1 {
		t.Errorf("level 5 inset %v should exceed level 1 inset %v", bf.boss.hitboxInset, inset1)
	}
	wantHits := BossHitsForLevel(5)
	wantThrows := BossThrowsForLevel(5)
	if bf.hitsRequired != wantHits || bf.throwsAllowed != wantThrows {
		t.Errorf("hits/throws = %d/%d, want %d/%d", bf.hitsRequired, bf.throwsAllowed, wantHits, wantThrows)
	}
}
