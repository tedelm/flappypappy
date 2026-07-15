package game

const (
	LevelStartDads     = 10
	LevelDadsIncrement = 5

	BossHitsRequired  = 5
	BossThrowsAllowed = 7
)

func DadsRequiredForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	return LevelStartDads + (level-1)*LevelDadsIncrement
}
