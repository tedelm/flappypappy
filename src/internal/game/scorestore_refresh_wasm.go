//go:build js && wasm

package game

import "log"

func startStoreRefresh(h *HighScores) {
	jsBeginRefresh(MaxHighScores)
}

func pollStoreRefresh() ([]HighScoreEntry, bool) {
	if !jsScoresReady() {
		return nil, false
	}
	return jsGetScores(), true
}

func saveScoreAsync(h *HighScores, store ScoreStore, name string, score int, diff Difficulty) {
	if err := store.Save(name, score, diff); err != nil {
		log.Printf("highscores: save: %v", err)
		return
	}
	h.RequestRefresh()
}

func pollBridgeReady(h *HighScores) {
	if h.connectFailed {
		return
	}
	if jsScoreDBFailed() {
		h.mu.Lock()
		h.connectFailed = true
		h.active = false
		h.refreshPending = false
		h.loading = false
		h.mu.Unlock()
		return
	}
	if !jsScoreBridgeLive() {
		return
	}
	h.mu.Lock()
	bridgeSeen := h.bridgeReadySeen
	if !bridgeSeen {
		h.bridgeReadySeen = true
	}
	h.mu.Unlock()
	if !bridgeSeen {
		h.RequestRefresh()
	}
}
