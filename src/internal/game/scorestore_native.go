//go:build !js || !wasm

package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type httpScoreStore struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type scoreAPIRow struct {
	PlayerName string `json:"player_name"`
	Score      int    `json:"score"`
	Level      int    `json:"level"`
	Difficulty string `json:"difficulty"`
}

func InitScoreStore() StoreInit {
	url, key, source := resolveScoreAPI()
	if url == "" {
		log.Printf("highscores: disabled (no URL configured)")
		return StoreInit{Store: NoopStore()}
	}
	store := &httpScoreStore{
		baseURL: strings.TrimRight(url, "/"),
		apiKey:  key,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
	if err := store.ping(); err != nil {
		log.Printf("highscores: connect failed: %v", err)
		return StoreInit{Store: NoopStore(), Configured: true, ConnectFailed: true}
	}
	log.Printf("highscores: connected via %s", source)
	return StoreInit{Store: store, Configured: true}
}

func NewScoreStore() ScoreStore {
	return InitScoreStore().Store
}

func (s *httpScoreStore) Active() bool { return true }

func (s *httpScoreStore) ping() error {
	req, err := http.NewRequest(http.MethodGet, s.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health: HTTP %d", resp.StatusCode)
	}
	return nil
}

func resolveScoreAPI() (url, key, source string) {
	url = strings.TrimSpace(os.Getenv("FLAPPY_SCORE_API_URL"))
	key = strings.TrimSpace(os.Getenv("FLAPPY_SCORE_API_KEY"))
	if url != "" {
		return url, key, "env"
	}
	url, key = readConfigJS()
	if url != "" {
		return url, key, "config.js"
	}
	return "", "", ""
}

func readConfigJS() (url, key string) {
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

			if u, k := readConfigAt(filepath.Join(dir, "web", "config.js")); u != "" {
				return u, k
			}
			if modRoot := findGoModRoot(dir); modRoot != "" {
				parent := filepath.Dir(modRoot)
				if u, k := readConfigAt(filepath.Join(parent, "web", "config.js")); u != "" {
					return u, k
				}
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", ""
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

func readConfigAt(path string) (url, key string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ""
	}
	content := string(data)
	return parseConfigJSValue(content, "FLAPPY_SCORE_API_URL"), parseConfigJSValue(content, "FLAPPY_SCORE_API_KEY")
}

func parseConfigJSValue(content, key string) string {
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
	return strings.TrimSpace(rest[1 : end+1])
}

func (s *httpScoreStore) EnsureSchema() error {
	return nil
}

func (s *httpScoreStore) Save(name string, score int, difficulty Difficulty, level int) error {
	body, err := json.Marshal(scoreAPIRow{
		PlayerName: normalizePlayerName(name),
		Score:      score,
		Level:      level,
		Difficulty: difficulty.Name(),
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, s.baseURL+"/scores", bytes.NewReader(body))
	if err != nil {
		return err
	}
	s.setAuth(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("save: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (s *httpScoreStore) Top(limit int) ([]HighScoreEntry, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/scores?limit=%d", s.baseURL, limit), nil)
	if err != nil {
		return nil, err
	}
	s.setAuth(req)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return nil, fmt.Errorf("top: HTTP %d", resp.StatusCode)
	}
	var rows []scoreAPIRow
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	entries := make([]HighScoreEntry, 0, len(rows))
	for _, row := range rows {
		level := row.Level
		if level <= 0 {
			level = 1
		}
		entries = append(entries, HighScoreEntry{
			Name:       row.PlayerName,
			Score:      row.Score,
			Level:      level,
			Difficulty: DifficultyFromName(row.Difficulty),
		})
	}
	return entries, nil
}

func (s *httpScoreStore) setAuth(req *http.Request) {
	if s.apiKey != "" {
		req.Header.Set("X-API-Key", s.apiKey)
	}
}
