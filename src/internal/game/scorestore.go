package game

type ScoreStore interface {
	Active() bool
	EnsureSchema() error
	BeginRun(difficulty Difficulty) error
	Save(name string, score int, difficulty Difficulty, level int) error
	Top(limit int) ([]HighScoreEntry, error)
}

// StoreInit holds a score store plus connection metadata for the UI.
type StoreInit struct {
	Store         ScoreStore
	Configured    bool // URL was found (env or config.js)
	ConnectFailed bool // URL found but connection failed
}
