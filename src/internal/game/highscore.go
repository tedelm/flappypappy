package game

import (
	"log"
	"sort"
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
	active         bool
	configured     bool
	connectFailed  bool
	initStarted    bool
	saveWG         sync.WaitGroup
	refreshPending bool
	loading        bool
}

func NewHighScores(init StoreInit) *HighScores {
	h := &HighScores{
		store:         init.Store,
		active:        init.Store != nil && init.Store.Active(),
		configured:    init.Configured,
		connectFailed: init.ConnectFailed,
	}
	if h.active {
		h.startInit()
	}
	return h
}

func (h *HighScores) Active() bool {
	h.ensureActive()
	return h.active
}

func (h *HighScores) Configured() bool {
	return h.configured
}

func (h *HighScores) ConnectFailed() bool {
	return h.connectFailed
}

func (h *HighScores) ensureActive() {
	if h.active || h.connectFailed || h.store == nil {
		return
	}
	if !h.store.Active() {
		return
	}
	h.active = true
	h.configured = true
	h.startInit()
}

func (h *HighScores) startInit() {
	if h.initStarted {
		return
	}
	h.initStarted = true
	store := h.store
	go func() {
		if err := store.EnsureSchema(); err != nil {
			log.Printf("highscores: schema: %v", err)
		}
		h.RequestRefresh()
	}()
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
	h.ensureActive()

	entry := HighScoreEntry{
		Name:  normalizePlayerName(name),
		Score: score,
	}

	h.mu.Lock()
	h.insertEntry(entry)
	h.mu.Unlock()

	if h.active {
		saveScoreAsync(h, h.store, entry.Name, entry.Score, diff)
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
	h.ensureActive()
	if !h.active {
		return
	}
	h.mu.Lock()
	h.loading = true
	h.refreshPending = true
	h.mu.Unlock()
	startStoreRefresh(h)
}

func (h *HighScores) PollRefresh() {
	if !h.active {
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
	h.entries = mergeEntries(h.entries, entries)
	h.refreshPending = false
	h.loading = false
	h.mu.Unlock()
}

func mergeEntries(local, remote []HighScoreEntry) []HighScoreEntry {
	type entryKey struct {
		name  string
		score int
	}
	seen := make(map[entryKey]bool, len(local)+len(remote))
	out := make([]HighScoreEntry, 0, len(local)+len(remote))
	add := func(e HighScoreEntry) {
		k := entryKey{e.Name, e.Score}
		if seen[k] {
			return
		}
		seen[k] = true
		out = append(out, e)
	}
	for _, e := range local {
		add(e)
	}
	for _, e := range remote {
		add(e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if len(out) > MaxHighScores {
		out = out[:MaxHighScores]
	}
	return out
}

func (h *HighScores) clearRefreshState() {
	h.mu.Lock()
	h.refreshPending = false
	h.loading = false
	h.mu.Unlock()
}
