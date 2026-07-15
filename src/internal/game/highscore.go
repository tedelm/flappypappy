package game

import (
	"log"
	"strings"
	"sync"
	"unicode"
)

type HighScoreEntry struct {
	Name  string
	Score int
}

type HighScores struct {
	mu             sync.Mutex
	entries        []HighScoreEntry
	store          ScoreStore
	refreshPending bool
	loading        bool
}

func NewHighScores(store ScoreStore) *HighScores {
	h := &HighScores{store: store}
	if store != nil {
		go func() {
			if err := store.EnsureSchema(); err != nil {
				log.Printf("highscores: schema: %v", err)
			}
			h.RequestRefresh()
		}()
	}
	return h
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

func (h *HighScores) Add(name string, score int, diff Difficulty) {
	entry := HighScoreEntry{
		Name:  normalizePlayerName(name),
		Score: score,
	}

	h.mu.Lock()
	h.insertEntry(entry)
	h.mu.Unlock()

	if h.store != nil {
		store := h.store
		go func() {
			if err := store.Save(entry.Name, entry.Score, diff); err != nil {
				log.Printf("highscores: save: %v", err)
			}
		}()
	}
}

func (h *HighScores) insertEntry(entry HighScoreEntry) {
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
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]HighScoreEntry, len(h.entries))
	copy(out, h.entries)
	return out
}

func (h *HighScores) Loading() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.loading
}

func (h *HighScores) RequestRefresh() {
	if h.store == nil {
		return
	}
	h.mu.Lock()
	h.loading = true
	h.refreshPending = true
	h.mu.Unlock()
	startStoreRefresh(h)
}

func (h *HighScores) PollRefresh() {
	if h.store == nil {
		return
	}
	h.mu.Lock()
	pending := h.refreshPending
	h.mu.Unlock()
	if !pending {
		return
	}
	if entries, done := pollStoreRefresh(); done {
		h.applyEntries(entries)
	}
}

func (h *HighScores) applyEntries(entries []HighScoreEntry) {
	h.mu.Lock()
	h.entries = entries
	h.refreshPending = false
	h.loading = false
	h.mu.Unlock()
}

func (h *HighScores) clearRefreshState() {
	h.mu.Lock()
	h.refreshPending = false
	h.loading = false
	h.mu.Unlock()
}
