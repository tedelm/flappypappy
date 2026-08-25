package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	maxScore       = 5000
	maxLevel       = 50
	maxRunAge      = 2 * time.Hour
	ticksPerSecond = 60.0
	minTimeFactor  = 0.6
	runsPerMinute  = 30
)

// Difficulty pacing aligned with src/internal/game/difficulty.go
var difficultyPace = map[string]struct {
	multiplier    int
	spawnInterval int
}{
	"EASY":   {multiplier: 1, spawnInterval: 120},
	"HARD":   {multiplier: 2, spawnInterval: 100},
	"INSANE": {multiplier: 3, spawnInterval: 82},
}

type runLimiter struct {
	mu     sync.Mutex
	times  []time.Time
	window time.Duration
	limit  int
}

func newRunLimiter(limit int, window time.Duration) *runLimiter {
	return &runLimiter{limit: limit, window: window}
}

func (l *runLimiter) allow(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-l.window)
	kept := l.times[:0]
	for _, t := range l.times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	l.times = kept
	if len(l.times) >= l.limit {
		return false
	}
	l.times = append(l.times, now)
	return true
}

func signRunToken(secret, runID, issuedAt, difficulty string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(runID + "|" + issuedAt + "|" + difficulty))
	return hex.EncodeToString(mac.Sum(nil))
}

func verifyRunToken(secret, runID, issuedAt, difficulty, token string) bool {
	want := signRunToken(secret, runID, issuedAt, difficulty)
	return subtle.ConstantTimeCompare([]byte(want), []byte(token)) == 1
}

func minElapsed(score int, difficulty string) time.Duration {
	pace, ok := difficultyPace[difficulty]
	if !ok {
		pace = difficultyPace["EASY"]
	}
	if score <= 0 || pace.multiplier < 1 {
		return 0
	}
	raw := float64(score) / float64(pace.multiplier)
	secs := raw * (float64(pace.spawnInterval) / ticksPerSecond) * minTimeFactor
	return time.Duration(secs * float64(time.Second))
}

func newRunID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func createRun(db *sql.DB, secret, difficulty string, now time.Time) (runID, token, issuedAt string, err error) {
	runID, err = newRunID()
	if err != nil {
		return "", "", "", err
	}
	issuedAt = now.UTC().Format(time.RFC3339)
	token = signRunToken(secret, runID, issuedAt, difficulty)
	_, err = db.Exec(
		`INSERT INTO runs (id, difficulty, issued_at) VALUES (?, ?, ?)`,
		runID, difficulty, issuedAt,
	)
	if err != nil {
		return "", "", "", err
	}
	_, _ = db.Exec(
		`DELETE FROM runs WHERE used_at IS NULL AND issued_at < ?`,
		now.UTC().Add(-maxRunAge).Format(time.RFC3339),
	)
	return runID, token, issuedAt, nil
}

var (
	errBadToken      = errors.New("invalid token")
	errRunNotFound   = errors.New("run not found")
	errRunUsed       = errors.New("run already used")
	errTooFast       = errors.New("score too high for elapsed time")
	errRunExpired    = errors.New("run expired")
	errScoreTooHigh  = errors.New("score too high")
	errLevelInvalid  = errors.New("invalid level")
	errDiffMismatch  = errors.New("difficulty mismatch")
)

func consumeRunAndInsert(db *sql.DB, secret string, in scoreSubmit, now time.Time) error {
	if in.Score < 0 || in.Score > maxScore {
		return errScoreTooHigh
	}
	if in.Level < 1 || in.Level > maxLevel {
		return errLevelInvalid
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var difficulty, issuedAt string
	var usedAt sql.NullString
	err = tx.QueryRow(
		`SELECT difficulty, issued_at, used_at FROM runs WHERE id = ?`,
		in.RunID,
	).Scan(&difficulty, &issuedAt, &usedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return errRunNotFound
	}
	if err != nil {
		return err
	}
	if usedAt.Valid {
		return errRunUsed
	}
	if difficulty != in.Difficulty {
		return errDiffMismatch
	}
	if !verifyRunToken(secret, in.RunID, issuedAt, difficulty, in.Token) {
		return errBadToken
	}

	issued, err := time.Parse(time.RFC3339, issuedAt)
	if err != nil {
		return fmt.Errorf("parse issued_at: %w", err)
	}
	elapsed := now.UTC().Sub(issued.UTC())
	if elapsed > maxRunAge {
		return errRunExpired
	}
	if elapsed < minElapsed(in.Score, difficulty) {
		return errTooFast
	}

	res, err := tx.Exec(
		`UPDATE runs SET used_at = ? WHERE id = ? AND used_at IS NULL`,
		now.UTC().Format(time.RFC3339), in.RunID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errRunUsed
	}

	_, err = tx.Exec(
		`INSERT INTO highscores (player_name, score, difficulty, level) VALUES (?, ?, ?, ?)`,
		in.PlayerName, in.Score, in.Difficulty, in.Level,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}
