//go:build !js || !wasm

package game

import "log"

func startStoreRefresh(h *HighScores) {
	go func() {
		h.saveWG.Wait()
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

func saveScoreAsync(h *HighScores, store ScoreStore, name string, score int, diff Difficulty) {
	h.saveWG.Add(1)
	go func() {
		defer h.saveWG.Done()
		if err := store.Save(name, score, diff); err != nil {
			log.Printf("highscores: save: %v", err)
			return
		}
		h.RequestRefresh()
	}()
}
