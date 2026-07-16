package game

type Difficulty int

const (
	DifficultyEasy Difficulty = iota
	DifficultyHard
	DifficultyInsane
)

type DifficultyConfig struct {
	ScoreMultiplier int
	Gravity         float64
	FlapStrength    float64
	PipeSpeed       float64
	PipeGap         int
	SpawnInterval   int
}

var difficultyConfigs = []DifficultyConfig{
	{ScoreMultiplier: 1, Gravity: 0.5, FlapStrength: -6.8, PipeSpeed: 3.0, PipeGap: 150, SpawnInterval: 120},
	{ScoreMultiplier: 2, Gravity: 0.58, FlapStrength: -7.2, PipeSpeed: 3.7, PipeGap: 130, SpawnInterval: 100},
	{ScoreMultiplier: 3, Gravity: 0.66, FlapStrength: -7.6, PipeSpeed: 4.4, PipeGap: 110, SpawnInterval: 82},
}

var difficultyNames = []string{"EASY", "HARD", "INSANE"}

func (d Difficulty) Config() DifficultyConfig {
	return difficultyConfigs[d]
}

func (d Difficulty) Name() string {
	if int(d) < 0 || int(d) >= len(difficultyNames) {
		return difficultyNames[DifficultyEasy]
	}
	return difficultyNames[d]
}

func DifficultyFromName(name string) Difficulty {
	for i, n := range difficultyNames {
		if n == name {
			return Difficulty(i)
		}
	}
	return DifficultyEasy
}

func (d Difficulty) Next() Difficulty {
	return Difficulty((int(d) + 1) % len(difficultyConfigs))
}

func (d Difficulty) Prev() Difficulty {
	return Difficulty((int(d) + len(difficultyConfigs) - 1) % len(difficultyConfigs))
}
