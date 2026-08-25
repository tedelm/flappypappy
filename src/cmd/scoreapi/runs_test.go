package main

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestSignAndVerifyRunToken(t *testing.T) {
	secret := "test-secret"
	token := signRunToken(secret, "abc", "2026-01-01T00:00:00Z", "EASY")
	if !verifyRunToken(secret, "abc", "2026-01-01T00:00:00Z", "EASY", token) {
		t.Fatal("expected valid token")
	}
	if verifyRunToken(secret, "abc", "2026-01-01T00:00:00Z", "EASY", token+"ff") {
		t.Fatal("expected invalid token")
	}
	if verifyRunToken(secret, "abc", "2026-01-01T00:00:00Z", "HARD", token) {
		t.Fatal("expected difficulty mismatch to fail")
	}
}

func TestMinElapsed(t *testing.T) {
	if minElapsed(0, "EASY") != 0 {
		t.Fatalf("score 0 should be 0")
	}
	// score 10 on EASY: raw=10, interval=120 → 10*(120/60)*0.6 = 12s
	got := minElapsed(10, "EASY")
	want := 12 * time.Second
	if got != want {
		t.Fatalf("minElapsed(10,EASY)=%v, want %v", got, want)
	}
	// score 10 on HARD: raw=5, interval=100 → 5*(100/60)*0.6 = 5s
	got = minElapsed(10, "HARD")
	want = 5 * time.Second
	if got != want {
		t.Fatalf("minElapsed(10,HARD)=%v, want %v", got, want)
	}
}

func TestConsumeRunAndInsert(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")
	db, err := openDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	secret := "hmac-secret"
	issued := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	runID, token, _, err := createRun(db, secret, "EASY", issued)
	if err != nil {
		t.Fatal(err)
	}

	submit := scoreSubmit{
		RunID: runID, Token: token,
		PlayerName: "TED", Score: 2, Level: 1, Difficulty: "EASY",
	}
	// Too fast: minElapsed(2,EASY)=2.4s
	if err := consumeRunAndInsert(db, secret, submit, issued.Add(time.Second)); !errors.Is(err, errTooFast) {
		t.Fatalf("want too fast, got %v", err)
	}
	// OK after enough time
	now := issued.Add(3 * time.Second)
	if err := consumeRunAndInsert(db, secret, submit, now); err != nil {
		t.Fatalf("insert: %v", err)
	}
	// Reuse
	if err := consumeRunAndInsert(db, secret, submit, now.Add(time.Second)); !errors.Is(err, errRunUsed) {
		t.Fatalf("want reuse, got %v", err)
	}
}

func TestConsumeBadToken(t *testing.T) {
	dir := t.TempDir()
	db, err := openDB(filepath.Join(dir, "t.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	secret := "s"
	issued := time.Now().UTC().Add(-time.Minute)
	runID, _, _, err := createRun(db, secret, "EASY", issued)
	if err != nil {
		t.Fatal(err)
	}
	err = consumeRunAndInsert(db, secret, scoreSubmit{
		RunID: runID, Token: "deadbeef",
		PlayerName: "A", Score: 0, Level: 1, Difficulty: "EASY",
	}, time.Now().UTC())
	if !errors.Is(err, errBadToken) {
		t.Fatalf("want bad token, got %v", err)
	}
}

func TestConsumeScoreTooHigh(t *testing.T) {
	dir := t.TempDir()
	db, err := openDB(filepath.Join(dir, "t.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	secret := "s"
	issued := time.Now().UTC().Add(-time.Hour)
	runID, token, _, err := createRun(db, secret, "EASY", issued)
	if err != nil {
		t.Fatal(err)
	}
	err = consumeRunAndInsert(db, secret, scoreSubmit{
		RunID: runID, Token: token,
		PlayerName: "A", Score: maxScore + 1, Level: 1, Difficulty: "EASY",
	}, time.Now().UTC())
	if !errors.Is(err, errScoreTooHigh) {
		t.Fatalf("want score too high, got %v", err)
	}
}
