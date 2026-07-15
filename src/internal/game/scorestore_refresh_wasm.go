//go:build js && wasm

package game

func startStoreRefresh(h *HighScores) {
	jsBeginRefresh(MaxHighScores)
}

func pollStoreRefresh() ([]HighScoreEntry, bool) {
	if !jsScoresReady() {
		return nil, false
	}
	return jsGetScores(), true
}
