package game

import "testing"

func TestHighScoresOrderByScoreThenLevel(t *testing.T) {
	h := &HighScores{}

	h.insertEntry(HighScoreEntry{Name: "A", Score: 100, Level: 2})
	h.insertEntry(HighScoreEntry{Name: "B", Score: 100, Level: 4})
	h.insertEntry(HighScoreEntry{Name: "C", Score: 120, Level: 1})

	got := h.List()
	if len(got) != 3 {
		t.Fatalf("got %d entries, want 3", len(got))
	}

	want := []HighScoreEntry{
		{Name: "C", Score: 120, Level: 1},
		{Name: "B", Score: 100, Level: 4},
		{Name: "A", Score: 100, Level: 2},
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestMergeEntriesDedupesByLevelAndSorts(t *testing.T) {
	local := []HighScoreEntry{
		{Name: "A", Score: 100, Level: 2},
	}
	remote := []HighScoreEntry{
		{Name: "A", Score: 100, Level: 2},
		{Name: "A", Score: 100, Level: 5},
	}

	got := mergeEntries(local, remote)
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2", len(got))
	}
	if got[0].Level != 5 || got[1].Level != 2 {
		t.Fatalf("unexpected order: %+v", got)
	}
}

