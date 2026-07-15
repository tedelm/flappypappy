package game

type noopStore struct{}

func (noopStore) Active() bool { return false }

func (noopStore) EnsureSchema() error { return nil }

func (noopStore) Save(name string, score int, difficulty Difficulty) error { return nil }

func (noopStore) Top(limit int) ([]HighScoreEntry, error) { return nil, nil }

func NoopStore() ScoreStore { return noopStore{} }
