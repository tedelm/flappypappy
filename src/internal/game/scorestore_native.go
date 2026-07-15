//go:build !js || !wasm

package game

import (
	"fmt"
	"log"
	"os"

	sqlitecloud "github.com/sqlitecloud/sqlitecloud-go"
)

const scoreSchemaSQL = `
CREATE TABLE IF NOT EXISTS highscores (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	player_name  TEXT    NOT NULL,
	score        INTEGER NOT NULL,
	difficulty   TEXT    NOT NULL,
	created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_highscores_score ON highscores(score DESC);
`

type sqliteCloudStore struct {
	db *sqlitecloud.SQCloud
}

func NewScoreStore() ScoreStore {
	url := os.Getenv("FLAPPY_SQLITECLOUD_URL")
	if url == "" {
		return NoopStore()
	}
	db, err := sqlitecloud.Connect(url)
	if err != nil {
		log.Printf("highscores: connect: %v", err)
		return NoopStore()
	}
	return &sqliteCloudStore{db: db}
}

func (s *sqliteCloudStore) EnsureSchema() error {
	return s.db.Execute(scoreSchemaSQL)
}

func (s *sqliteCloudStore) Save(name string, score int, difficulty Difficulty) error {
	name = escapeSQLString(normalizePlayerName(name))
	diff := escapeSQLString(difficulty.Name())
	sql := fmt.Sprintf(
		"INSERT INTO highscores (player_name, score, difficulty) VALUES ('%s', %d, '%s');",
		name, score, diff,
	)
	return s.db.Execute(sql)
}

func (s *sqliteCloudStore) Top(limit int) ([]HighScoreEntry, error) {
	result, err := s.db.Select(fmt.Sprintf(
		"SELECT player_name, score FROM highscores ORDER BY score DESC LIMIT %d;",
		limit,
	))
	if err != nil {
		return nil, err
	}
	rows := result.GetNumberOfRows()
	entries := make([]HighScoreEntry, 0, rows)
	for r := uint64(0); r < rows; r++ {
		name, err := result.GetStringValue(r, 0)
		if err != nil {
			return nil, err
		}
		score, err := result.GetInt64Value(r, 1)
		if err != nil {
			return nil, err
		}
		entries = append(entries, HighScoreEntry{Name: name, Score: int(score)})
	}
	return entries, nil
}
