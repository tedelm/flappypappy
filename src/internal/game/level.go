package game

import (
	"math/rand"

	"flappy/internal/game/sprite"
)

const (
	LevelStartDads     = 1 // 7
	LevelDadsIncrement = 0 // 6

	BossHitsRequired  = 4
	BossThrowsAllowed = 10

	// Outrun chase: every 3rd level; base duration 20s at 60 TPS.
	OutrunDurationFrames = 1200
	OutrunCountdownSecs  = 3
	outrunGoHoldFrames   = 40
	outrunSecFrames      = 60

	outrunStartLead       = 110.0
	outrunCatchLead       = 28.0
	outrunMaxLead         = 320.0
	outrunTapLeadBoost    = 40.0
	outrunWifeCloseBase   = 1.5
	outrunWifeCloseRise   = 0.4
	outrunTapSpeedBoost   = 2.2
	outrunCoastSpeed      = 3.0
	outrunMaxPlayerSpeed  = 9.0
	outrunSpeedDecay      = 0.14
	outrunPlayerScreenX   = 160.0
	outrunGroundClearance = 10.0

	outrunDurationPerChase   = 300 // +5s each subsequent chase
	outrunStartLeadPerChase  = 18.0
	outrunStartLeadMin       = 70.0
	outrunWifeCloseBaseStep  = 0.35
	outrunWifeCloseRiseStep  = 0.2
	outrunTapLeadBoostStep   = 5.0
	outrunTapLeadBoostMin    = 22.0

	// Chase laugh SFX: random interval 3–6s at 60 TPS.
	outrunLaughMinFrames = 180
	outrunLaughMaxFrames = 360

	bossBaseMoveSpeed   = 2.2
	bossMoveSpeedPerLvl = 0.25
	bossMaxMoveSpeed    = 4.5

	bossBaseHitboxInset   = 0.12
	bossHitboxInsetPerLvl = 0.02
	bossMaxHitboxInset    = 0.28

	bossBaseYFlipMin      = 40
	bossBaseYFlipMax      = 90
	bossYFlipShrinkPerLvl = 4

	bossBaseDashGapMin      = 70
	bossBaseDashGapMax      = 140
	bossDashGapShrinkPerLvl = 6

	NumWallpapers = 6
)

func DadsRequiredForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	return LevelStartDads + (level-1)*LevelDadsIncrement
}

// BossIsOutrun is true for every 3rd level's scrolling-race boss fight (3, 6, 9, …).
func BossIsOutrun(level int) bool {
	return level > 0 && level%3 == 0
}

// OutrunChaseIndex is 1 for the first chase (level 3), 2 for level 6, etc.
func OutrunChaseIndex(level int) int {
	if level < 1 {
		return 0
	}
	return level / 3
}

func OutrunDurationFramesForLevel(level int) int {
	n := OutrunChaseIndex(level)
	if n < 1 {
		n = 1
	}
	return OutrunDurationFrames + (n-1)*outrunDurationPerChase
}

func OutrunStartLeadForLevel(level int) float64 {
	n := OutrunChaseIndex(level)
	if n < 1 {
		n = 1
	}
	lead := outrunStartLead - float64(n-1)*outrunStartLeadPerChase
	if lead < outrunStartLeadMin {
		return outrunStartLeadMin
	}
	return lead
}

func OutrunWifeCloseBaseForLevel(level int) float64 {
	n := OutrunChaseIndex(level)
	if n < 1 {
		n = 1
	}
	return outrunWifeCloseBase + float64(n-1)*outrunWifeCloseBaseStep
}

func OutrunWifeCloseRiseForLevel(level int) float64 {
	n := OutrunChaseIndex(level)
	if n < 1 {
		n = 1
	}
	return outrunWifeCloseRise + float64(n-1)*outrunWifeCloseRiseStep
}

func OutrunTapLeadBoostForLevel(level int) float64 {
	n := OutrunChaseIndex(level)
	if n < 1 {
		n = 1
	}
	boost := outrunTapLeadBoost - float64(n-1)*outrunTapLeadBoostStep
	if boost < outrunTapLeadBoostMin {
		return outrunTapLeadBoostMin
	}
	return boost
}

func OutrunCountdownTotalFrames() int {
	return OutrunCountdownSecs*outrunSecFrames + outrunGoHoldFrames
}

func outrunLaughIntervalFrames() int {
	span := outrunLaughMaxFrames - outrunLaughMinFrames + 1
	return outrunLaughMinFrames + rand.Intn(span)
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
