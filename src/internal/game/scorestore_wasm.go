//go:build js && wasm

package game

import "syscall/js"

type jsStore struct{}

func InitScoreStore() StoreInit {
	configured := jsHasScoreURL()
	return StoreInit{
		Store:         jsStore{},
		Configured:    configured,
		ConnectFailed: configured && jsScoreDBFailed(),
	}
}

func NewScoreStore() ScoreStore {
	return InitScoreStore().Store
}

func (jsStore) Active() bool {
	return jsHasScoreURL() && !jsScoreDBFailed()
}

func jsHasScoreURL() bool {
	fn := js.Global().Get("flappyHasScoreURL")
	if fn.Type() == js.TypeFunction {
		return fn.Invoke().Bool()
	}
	v := js.Global().Get("FLAPPY_SQLITECLOUD_URL")
	if v.Type() != js.TypeString {
		return false
	}
	return v.String() != ""
}

func jsScoreBridgeLive() bool {
	fn := js.Global().Get("flappyScoreBridgeLive")
	if fn.Type() != js.TypeFunction {
		return false
	}
	return fn.Invoke().Bool()
}

func jsScoreDBFailed() bool {
	fn := js.Global().Get("flappyScoreDBFailed")
	if fn.Type() != js.TypeFunction {
		return false
	}
	return fn.Invoke().Bool()
}

func (jsStore) EnsureSchema() error {
	callJS("flappyInitScoreDB")
	return nil
}

func (jsStore) Save(name string, score int, difficulty Difficulty, level int) error {
	callJS("flappySaveScore", normalizePlayerName(name), score, difficulty.Name(), level)
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
		return false
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
		level := row.Get("level").Int()
		if level <= 0 {
			level = 1
		}
		difficulty := row.Get("difficulty").String()
		entries = append(entries, HighScoreEntry{
			Name:       name,
			Score:      score,
			Level:      level,
			Difficulty: DifficultyFromName(difficulty),
		})
	}
	return entries
}
