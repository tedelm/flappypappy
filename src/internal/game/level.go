package game

import "flappy/internal/game/sprite"

const (
	LevelStartDads     = 7
	LevelDadsIncrement = 6

	BossHitsRequired  = 5
	BossThrowsAllowed = 8

	bossBaseMoveSpeed   = 2.5
	bossMoveSpeedPerLvl = 0.35
	bossMaxMoveSpeed    = 5.5

	bossBaseHitboxInset   = 0.15
	bossHitboxInsetPerLvl = 0.03
	bossMaxHitboxInset    = 0.35

	bossBaseYFlipMin      = 40
	bossBaseYFlipMax      = 90
	bossYFlipShrinkPerLvl = 6

	bossBaseDashGapMin      = 70
	bossBaseDashGapMax      = 140
	bossDashGapShrinkPerLvl = 10

	NumWallpapers = 6
)

func DadsRequiredForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	return LevelStartDads + (level-1)*LevelDadsIncrement
}

func BossHitsForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	return BossHitsRequired + (level - 1)
}

func BossThrowsForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	return BossThrowsAllowed + (level - 1)
}

func BossMoveSpeedForLevel(level int) float64 {
	if level < 1 {
		level = 1
	}
	speed := bossBaseMoveSpeed + float64(level-1)*bossMoveSpeedPerLvl
	if speed > bossMaxMoveSpeed {
		return bossMaxMoveSpeed
	}
	return speed
}

func BossHitboxInsetForLevel(level int) float64 {
	if level < 1 {
		level = 1
	}
	inset := bossBaseHitboxInset + float64(level-1)*bossHitboxInsetPerLvl
	if inset > bossMaxHitboxInset {
		return bossMaxHitboxInset
	}
	return inset
}

// BossYFlipIntervalForLevel returns min/max frames between random vertical direction flips.
func BossYFlipIntervalForLevel(level int) (minFrames, maxFrames int) {
	if level < 1 {
		level = 1
	}
	shrink := (level - 1) * bossYFlipShrinkPerLvl
	minFrames = bossBaseYFlipMin - shrink
	maxFrames = bossBaseYFlipMax - shrink
	if minFrames < 18 {
		minFrames = 18
	}
	if maxFrames < minFrames+20 {
		maxFrames = minFrames + 20
	}
	return minFrames, maxFrames
}

// BossDashGapForLevel returns min/max frames between sideways dash attempts.
func BossDashGapForLevel(level int) (minFrames, maxFrames int) {
	if level < 1 {
		level = 1
	}
	shrink := (level - 1) * bossDashGapShrinkPerLvl
	minFrames = bossBaseDashGapMin - shrink
	maxFrames = bossBaseDashGapMax - shrink
	if minFrames < 28 {
		minFrames = 28
	}
	if maxFrames < minFrames+30 {
		maxFrames = minFrames + 30
	}
	return minFrames, maxFrames
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
