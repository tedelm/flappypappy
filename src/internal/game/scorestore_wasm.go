//go:build js && wasm

package game

import "syscall/js"

type jsStore struct{}

func NewScoreStore() ScoreStore {
	return jsStore{}
}

func (jsStore) EnsureSchema() error {
	callJS("flappyInitScoreDB")
	return nil
}

func (jsStore) Save(name string, score int, difficulty Difficulty) error {
	callJS("flappySaveScore", normalizePlayerName(name), score, difficulty.Name())
	return nil
}

func (jsStore) Top(limit int) ([]HighScoreEntry, error) {
	return jsGetScores(), nil
}

func jsBeginRefresh(limit int) {
	callJS("flappyRefreshScores", limit)
}

func jsScoresReady() bool {
	fn := js.Global().Get("flappyScoresReady")
	if fn.Type() != js.TypeFunction {
		return true
	}
	return fn.Invoke().Bool()
}

func jsGetScores() []HighScoreEntry {
	fn := js.Global().Get("flappyGetScores")
	if fn.Type() != js.TypeFunction {
		return nil
	}
	v := fn.Invoke()
	if v.Type() != js.TypeObject {
		return nil
	}
	length := v.Length()
	entries := make([]HighScoreEntry, 0, length)
	for i := 0; i < length; i++ {
		row := v.Index(i)
		name := row.Get("name").String()
		score := row.Get("score").Int()
		entries = append(entries, HighScoreEntry{Name: name, Score: score})
	}
	return entries
}
