//go:build !js || !wasm

package game

import "log"

func startStoreRefresh(h *HighScores) {
	go func() {
		entries, err := h.store.Top(MaxHighScores)
		if err != nil {
			log.Printf("highscores: refresh: %v", err)
			h.clearRefreshState()
			return
		}
		h.applyEntries(entries)
	}()
}

func pollStoreRefresh() ([]HighScoreEntry, bool) {
	return nil, false
}
