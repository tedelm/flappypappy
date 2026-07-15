package game

import "strings"

type ScoreStore interface {
	EnsureSchema() error
	Save(name string, score int, difficulty Difficulty) error
	Top(limit int) ([]HighScoreEntry, error)
}

func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
