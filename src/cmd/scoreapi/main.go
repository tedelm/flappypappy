package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

const (
	createTableSQL = `
CREATE TABLE IF NOT EXISTS highscores (
	id           INTEGER PRIMARY KEY AUTOINCREMENT,
	player_name  TEXT    NOT NULL,
	score        INTEGER NOT NULL,
	level        INTEGER NOT NULL DEFAULT 1,
	difficulty   TEXT    NOT NULL,
	created_at   TEXT    NOT NULL DEFAULT (datetime('now'))
);`
	createIndexSQL   = `CREATE INDEX IF NOT EXISTS idx_highscores_score ON highscores(score DESC);`
	addLevelColumnSQL = `ALTER TABLE highscores ADD COLUMN level INTEGER NOT NULL DEFAULT 1;`
	maxNameLen       = 12
	maxLimit         = 50
	defaultLimit     = 10
	defaultListen    = "127.0.0.1:8088"
	defaultDBPath    = "/var/lib/flappy/flappypappy.sqlite"
)

type scoreRow struct {
	PlayerName string `json:"player_name"`
	Score      int    `json:"score"`
	Level      int    `json:"level"`
	Difficulty string `json:"difficulty"`
}

func main() {
	listen := envOr("LISTEN", defaultListen)
	dbPath := envOr("DB_PATH", defaultDBPath)
	apiKey := strings.TrimSpace(os.Getenv("API_KEY"))
	if apiKey == "" {
		log.Fatal("API_KEY is required")
	}

	db, err := openDB(dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "error"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /scores", requireAPIKey(apiKey, func(w http.ResponseWriter, r *http.Request) {
		limit := defaultLimit
		if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = n
		}
		if limit > maxLimit {
			limit = maxLimit
		}
		rows, err := listScores(db, limit)
		if err != nil {
			log.Printf("list scores: %v", err)
			http.Error(w, "query failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, rows)
	}))
	mux.HandleFunc("POST /scores", requireAPIKey(apiKey, func(w http.ResponseWriter, r *http.Request) {
		var in scoreRow
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := dec.Decode(&in); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		in.PlayerName = normalizePlayerName(in.PlayerName)
		in.Difficulty = normalizeDifficulty(in.Difficulty)
		if in.Score < 0 {
			http.Error(w, "invalid score", http.StatusBadRequest)
			return
		}
		if in.Level < 1 {
			in.Level = 1
		}
		if err := insertScore(db, in); err != nil {
			log.Printf("insert score: %v", err)
			http.Error(w, "insert failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, in)
	}))

	handler := withCORS(mux)
	srv := &http.Server{
		Addr:              listen,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("scoreapi listening on %s (db %s)", listen, dbPath)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode = WAL`); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(addLevelColumnSQL); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(createIndexSQL); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func listScores(db *sql.DB, limit int) ([]scoreRow, error) {
	q, err := db.Query(
		`SELECT player_name, score, level, difficulty FROM highscores ORDER BY score DESC, level DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer q.Close()

	out := make([]scoreRow, 0, limit)
	for q.Next() {
		var row scoreRow
		if err := q.Scan(&row.PlayerName, &row.Score, &row.Level, &row.Difficulty); err != nil {
			return nil, err
		}
		if row.Level < 1 {
			row.Level = 1
		}
		out = append(out, row)
	}
	return out, q.Err()
}

func insertScore(db *sql.DB, row scoreRow) error {
	_, err := db.Exec(
		`INSERT INTO highscores (player_name, score, difficulty, level) VALUES (?, ?, ?, ?)`,
		row.PlayerName, row.Score, row.Difficulty, row.Level,
	)
	return err
}

func requireAPIKey(want string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		got := strings.TrimSpace(r.Header.Get("X-API-Key"))
		if got == "" {
			got = strings.TrimSpace(r.URL.Query().Get("apikey"))
		}
		if got != want {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func normalizePlayerName(name string) string {
	name = strings.TrimSpace(name)
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	name = strings.TrimSpace(b.String())
	if name == "" {
		return "PLAYER"
	}
	runes := []rune(name)
	if len(runes) > maxNameLen {
		name = string(runes[:maxNameLen])
	}
	return name
}

func normalizeDifficulty(s string) string {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "HARD":
		return "HARD"
	case "NORMAL":
		return "NORMAL"
	default:
		return "EASY"
	}
}
