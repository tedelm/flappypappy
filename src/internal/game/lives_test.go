package game

import "testing"

func TestAwardLevelCompleteLife(t *testing.T) {
	tests := []struct {
		name     string
		lives    int
		livesMax int
		wantL    int
		wantMax  int
	}{
		{
			name:     "2/3 -> 3/3",
			lives:    2,
			livesMax: 3,
			wantL:    3,
			wantMax:  3,
		},
		{
			name:     "3/3 -> 4/4",
			lives:    3,
			livesMax: 3,
			wantL:    4,
			wantMax:  4,
		},
		{
			name:     "9/10 -> 10/10",
			lives:    9,
			livesMax: 10,
			wantL:    10,
			wantMax:  10,
		},
		{
			name:     "10/10 -> 10/10",
			lives:    10,
			livesMax: 10,
			wantL:    10,
			wantMax:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLives, gotMax := awardLevelCompleteLife(tt.lives, tt.livesMax)
			if gotLives != tt.wantL || gotMax != tt.wantMax {
				t.Fatalf("awardLevelCompleteLife(%d, %d) = (%d, %d), want (%d, %d)",
					tt.lives, tt.livesMax, gotLives, gotMax, tt.wantL, tt.wantMax)
			}
		})
	}
}

