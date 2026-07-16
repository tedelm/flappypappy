package game

import (
	"testing"
)

func TestDecorPaintingIndexStable(t *testing.T) {
	seed := 42
	worldX := 1200
	pool := LevelPaintingPool(2)
	a := pool[int(decorHash(seed, worldX))%len(pool)]
	b := pool[int(decorHash(seed, worldX))%len(pool)]
	if a != b {
		t.Fatalf("painting index not stable: %d vs %d", a, b)
	}
	start, count := LevelPaintingWindow(2)
	if a < start || a >= start+count {
		t.Fatalf("painting index %d outside level 2 window [%d, %d)", a, start, start+count)
	}
}

func TestDecorPaintingIndexRespectsFixedWindow(t *testing.T) {
	seed := 99
	start, count := LevelPaintingWindow(5)
	pool := LevelPaintingPool(5)
	seen := make(map[int]bool)
	for worldX := 0; worldX < 2000; worldX += DecorChunkMin {
		idx := pool[int(decorHash(seed, worldX))%len(pool)]
		if idx < start || idx >= start+count {
			t.Fatalf("painting index %d outside level 5 window [%d, %d)", idx, start, start+count)
		}
		seen[idx] = true
	}
	if len(seen) < 2 {
		t.Fatalf("expected varied painting indices inside fixed window, got %d unique", len(seen))
	}
}

func TestDecorPaintingIndexRespectsRandomPool(t *testing.T) {
	seed := 99
	pool := LevelPaintingPool(7)
	allowed := make(map[int]bool, len(pool))
	for _, idx := range pool {
		allowed[idx] = true
	}
	seen := make(map[int]bool)
	for worldX := 0; worldX < 2000; worldX += DecorChunkMin {
		idx := pool[int(decorHash(seed, worldX))%len(pool)]
		if !allowed[idx] {
			t.Fatalf("painting index %d not present in level 7 pool %v", idx, pool)
		}
		seen[idx] = true
	}
	if len(seen) < 2 {
		t.Fatalf("expected varied painting indices inside random pool, got %d unique", len(seen))
	}
}

func TestDecorPaintingScaleStable(t *testing.T) {
	seed := 42
	worldX := 1200
	a := decorPaintingScale(seed, worldX)
	b := decorPaintingScale(seed, worldX)
	if a != b {
		t.Fatalf("painting scale not stable: %v vs %v", a, b)
	}
}

func TestDecorPaintingScaleInRange(t *testing.T) {
	min := DecorPaintingScale * DecorPaintingScaleMin
	max := DecorPaintingScale * DecorPaintingScaleMax
	for seed := 0; seed < 20; seed++ {
		for worldX := 0; worldX < 2000; worldX += DecorChunkMin {
			scale := decorPaintingScale(seed, worldX)
			if scale < min || scale > max {
				t.Fatalf("scale %v out of range [%v, %v] for seed=%d worldX=%d", scale, min, max, seed, worldX)
			}
		}
	}
}

func TestDecorPaintingScaleVaries(t *testing.T) {
	seed := 99
	seen := make(map[float64]bool)
	for worldX := 0; worldX < 2000; worldX += DecorChunkMin {
		seen[decorPaintingScale(seed, worldX)] = true
	}
	if len(seen) < 2 {
		t.Fatalf("expected varied painting scales across positions, got %d unique", len(seen))
	}
}
