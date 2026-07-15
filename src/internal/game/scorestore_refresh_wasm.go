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
	}
}
