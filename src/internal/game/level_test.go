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

func TestBossConstants(t *testing.T) {
	if BossHitsRequired != 5 {
		t.Errorf("BossHitsRequired = %d, want 5", BossHitsRequired)
	}
	if BossThrowsAllowed != 7 {
		t.Errorf("BossThrowsAllowed = %d, want 7", BossThrowsAllowed)
	}
}
