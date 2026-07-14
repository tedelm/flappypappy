package game

import (
	"strings"
	"unicode"
)

type HighScoreEntry struct {
	Name  string
	Score int
}

type HighScores struct {
	entries []HighScoreEntry
}

func NewHighScores() *HighScores {
	return &HighScores{}
}

func normalizePlayerName(name string) string {
	name = strings.TrimSpace(name)
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	name = strings.TrimSpace(b.String())
	if name == "" {
		return "PLAYER"
	}
	runes := []rune(name)
	if len(runes) > MaxPlayerNameLen {
		name = string(runes[:MaxPlayerNameLen])
	}
	return name
}

func (h *HighScores) Add(name string, score int) {
	entry := HighScoreEntry{
		Name:  normalizePlayerName(name),
		Score: score,
	}

	inserted := false
	for i, e := range h.entries {
		if entry.Score > e.Score {
			h.entries = append(h.entries[:i], append([]HighScoreEntry{entry}, h.entries[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		h.entries = append(h.entries, entry)
	}

	if len(h.entries) > MaxHighScores {
		h.entries = h.entries[:MaxHighScores]
	}
}

func (h *HighScores) List() []HighScoreEntry {
	out := make([]HighScoreEntry, len(h.entries))
	copy(out, h.entries)
	return out
}
