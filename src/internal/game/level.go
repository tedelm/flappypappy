package game

import "flappy/internal/game/sprite"

const (
	LevelStartDads     = 10
	LevelDadsIncrement = 5

	BossHitsRequired  = 5
	BossThrowsAllowed = 7

	NumWallpapers = 6
)

func DadsRequiredForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	return LevelStartDads + (level-1)*LevelDadsIncrement
}

func WallpaperIndex(level int) int {
	if level < 1 {
		level = 1
	}
	return (level - 1) % NumWallpapers
}

func LevelPaintingWindow(level int) (start, count int) {
	if level < 1 {
		level = 1
	}
	start = 2 * (level - 1)
	if start >= sprite.NumPubPaintings {
		return start, 0
	}
	count = 3
	if remaining := sprite.NumPubPaintings - start; remaining < count {
		count = remaining
	}
	return start, count
}

func LevelPaintingPool(level int) []int {
	if level < 1 {
		level = 1
	}
	if level <= 6 {
		start, count := LevelPaintingWindow(level)
		pool := make([]int, 0, count)
		for i := 0; i < count; i++ {
			pool = append(pool, start+i)
		}
		return pool
	}

	pool := make([]int, 0, 3)
	used := make(map[int]bool, 3)
	for salt := 0; len(pool) < 3 && len(used) < sprite.NumPubPaintings; salt++ {
		idx := int(decorHash(level, salt) % uint32(sprite.NumPubPaintings))
		if used[idx] {
			continue
		}
		used[idx] = true
		pool = append(pool, idx)
	}
	return pool
}
