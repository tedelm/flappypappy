package game

import "testing"

func TestDadsRequiredForLevel(t *testing.T) {
	tests := []struct {
		level int
		want  int
	}{
		{1, 10},
		{2, 15},
		{3, 20},
		{4, 25},
		{5, 30},
		{0, 10},
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
	if BossHitsRequired != 5 {
		t.Errorf("BossHitsRequired = %d, want 5", BossHitsRequired)
	}
	if BossThrowsAllowed != 7 {
		t.Errorf("BossThrowsAllowed = %d, want 7", BossThrowsAllowed)
	}
}
