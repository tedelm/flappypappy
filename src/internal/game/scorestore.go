package game

import "strings"

type ScoreStore interface {
	Active() bool
	EnsureSchema() error
	Save(name string, score int, difficulty Difficulty) error
	Top(limit int) ([]HighScoreEntry, error)
}

// StoreInit holds a score store plus connection metadata for the UI.
type StoreInit struct {
	Store         ScoreStore
	Configured    bool // URL was found (env or config.js)
	ConnectFailed bool // URL found but connection failed (native only)
}

func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
