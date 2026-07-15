//go:build !js || !wasm

package game

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	sqlitecloud "github.com/sqlitecloud/sqlitecloud-go"
)

const createTableSQL = `
CREATE TABLE IF NOT EXISTS highscores (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	player_name  TEXT    NOT NULL,
	score        INTEGER NOT NULL,
	difficulty   TEXT    NOT NULL,
	created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);`

const createIndexSQL = `CREATE INDEX IF NOT EXISTS idx_highscores_score ON highscores(score DESC);`

type sqliteCloudStore struct {
	db *sqlitecloud.SQCloud
}

func InitScoreStore() StoreInit {
	url, source := resolveSQLiteCloudURL()
	if url == "" {
		log.Printf("highscores: disabled (no URL configured)")
		return StoreInit{Store: NoopStore()}
	}
	db, err := sqlitecloud.Connect(url)
	if err != nil {
		log.Printf("highscores: connect failed: %v", err)
		return StoreInit{Store: NoopStore(), Configured: true, ConnectFailed: true}
	}
	log.Printf("highscores: connected via %s", source)
	return StoreInit{Store: &sqliteCloudStore{db: db}}
}

func NewScoreStore() ScoreStore {
	return InitScoreStore().Store
}

func (s *sqliteCloudStore) Active() bool { return true }

func resolveSQLiteCloudURL() (url, source string) {
	if u := strings.TrimSpace(os.Getenv("FLAPPY_SQLITECLOUD_URL")); u != "" {
		return u, "env"
	}
	if u := readConfigJS(); u != "" {
		return u, "config.js"
	}
	return "", ""
}

func readConfigJS() string {
	seen := make(map[string]bool)
	for _, start := range configSearchRoots() {
		dir := start
		for i := 0; i < 8; i++ {
			abs, err := filepath.Abs(dir)
			if err != nil {
				break
			}
			if seen[abs] {
				break
			}
			seen[abs] = true

			if u := readConfigAt(filepath.Join(dir, "web", "config.js")); u != "" {
				return u
			}
			if modRoot := findGoModRoot(dir); modRoot != "" {
				parent := filepath.Dir(modRoot)
				if u := readConfigAt(filepath.Join(parent, "web", "config.js")); u != "" {
					return u
				}
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return ""
}

func configSearchRoots() []string {
	var roots []string
	if wd, err := os.Getwd(); err == nil {
		roots = append(roots, wd)
	}
	if exe, err := os.Executable(); err == nil {
		roots = append(roots, filepath.Dir(exe))
	}
	return roots
}

func findGoModRoot(dir string) string {
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func readConfigAt(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return parseConfigURL(string(data))
}

func parseConfigURL(content string) string {
	const key = "FLAPPY_SQLITECLOUD_URL"
	idx := strings.Index(content, key)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(content[idx+len(key):])
	if !strings.HasPrefix(rest, "=") {
		return ""
	}
	rest = strings.TrimSpace(rest[1:])
	if len(rest) == 0 {
		return ""
	}
	quote := rest[0]
	if quote != '"' && quote != '\'' {
		return ""
	}
	end := strings.IndexByte(rest[1:], quote)
	if end < 0 {
		return ""
	}
	url := rest[1 : end+1]
	if strings.TrimSpace(url) == "" {
		return ""
	}
	return url
}

func (s *sqliteCloudStore) EnsureSchema() error {
	if err := s.db.Execute(createTableSQL); err != nil {
		return err
	}
	return s.db.Execute(createIndexSQL)
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
