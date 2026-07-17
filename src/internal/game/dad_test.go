package game

import (
	"testing"

	"flappy/internal/game/sprite"
)

func TestDadPerFrameFillsZoneHeight(t *testing.T) {
	const zoneH = 200.0
	for variant := 0; variant < sprite.DadVariantCount; variant++ {
		for frame := 0; frame < sprite.FrameCount; frame++ {
			ch := sprite.ContentH(variant, frame)
			if ch <= 0 {
				t.Fatalf("variant %d frame %d: ContentH=%v", variant, frame, ch)
			}
			scale := zoneH / ch
			filled := ch * scale
			if filled < zoneH-0.01 || filled > zoneH+0.01 {
				t.Fatalf("variant %d frame %d: contentH*scale = %v, want %v", variant, frame, filled, zoneH)
			}
		}
	}
}

func TestDadHitboxNarrowerThanPipeWhenSpriteNarrow(t *testing.T) {
	const (
		pipeX     = 100.0
		variant   = 2 // dad_3
		frame     = 0
	)
	// Short bottom zone so scaled opaque width stays under PipeWidth.
	gapBottom := float64(FloorSurfaceY) - 50
	x, y, w, h := dadHitbox(pipeX, gapBottom, variant, frame)
	zoneH := float64(FloorSurfaceY) - gapBottom
	if y != gapBottom || h != zoneH {
		t.Fatalf("dadHitbox vertical = (%v,%v), want (%v,%v)", y, h, gapBottom, zoneH)
	}
	if w <= 0 || w > PipeWidth {
		t.Fatalf("dadHitbox width = %v, want in (0, %v]", w, PipeWidth)
	}
	if w >= PipeWidth {
		t.Fatalf("dadHitbox width = %v, want < %v for short zone", w, PipeWidth)
	}
	if x < pipeX || x+w > pipeX+PipeWidth+0.01 {
		t.Fatalf("dadHitbox x-range [%v,%v] outside pipe [%v,%v]", x, x+w, pipeX, pipeX+PipeWidth)
	}
}

func TestDadHitboxMissesBesideOpaqueColumn(t *testing.T) {
	const (
		pipeX   = 100.0
		variant = 2
		frame   = 0
	)
	gapBottom := float64(FloorSurfaceY) - 50
	hx, hy, hw, hh := dadHitbox(pipeX, gapBottom, variant, frame)
	if hw >= PipeWidth {
		t.Fatalf("need narrow hitbox for this test, got w=%v", hw)
	}

	birdW, birdH := 28.0, 40.0
	birdX := hx - birdW - 1
	birdY := hy + hh/2 - birdH/2
	if aabbOverlap(birdX, birdY, birdW, birdH, hx, hy, hw, hh) {
		t.Fatal("expected miss beside opaque dad hitbox")
	}
	if !aabbOverlap(birdX, birdY, birdW, birdH, pipeX, hy, PipeWidth, hh) {
		t.Fatal("sanity: bird would overlap old full-width pipe column")
	}

	birdX = hx + hw/2 - birdW/2
	if !aabbOverlap(birdX, birdY, birdW, birdH, hx, hy, hw, hh) {
		t.Fatal("expected hit on opaque dad hitbox")
	}
}

func TestLampHitboxCenteredInPipe(t *testing.T) {
	const pipeX = 50.0
	const gapY = 40.0 // short top zone so lamp hitW can be < PipeWidth
	x, y, w, h := lampHitbox(pipeX, gapY)
	if y != 0 || h != gapY {
		t.Fatalf("lampHitbox vertical = (%v,%v), want (0,%v)", y, h, gapY)
	}
	if w <= 0 || w > PipeWidth {
		t.Fatalf("lampHitbox width = %v, want in (0, %v]", w, PipeWidth)
	}
	if x < pipeX || x+w > pipeX+PipeWidth+0.01 {
		t.Fatalf("lampHitbox x-range [%v,%v] outside pipe [%v,%v]", x, x+w, pipeX, pipeX+PipeWidth)
	}
}

func TestPipeCollidesUsesNarrowHitbox(t *testing.T) {
	pm := NewPipeManager()
	zoneH := 50.0
	gapBottom := float64(FloorSurfaceY) - zoneH
	gapH := 150.0
	gapY := gapBottom - gapH
	pm.pipes = []*Pipe{{
		X:    100,
		GapY: gapY,
		GapH: gapH,
	}}

	const (
		variant = 2
		frame   = 0
	)
	hx, hy, hw, hh := dadHitbox(100, gapBottom, variant, frame)
	if hw >= PipeWidth {
		t.Fatalf("need narrow hitbox for this test, got w=%v", hw)
	}

	birdW, birdH := 14.0, 20.0
	besideX := hx - birdW - 1
	besideY := hy + 2
	if pm.collidesAt(besideX, besideY, birdW, birdH, variant, frame) {
		t.Fatal("Collides: expected miss beside dad opaque width")
	}

	hitX := hx + hw/2 - birdW/2
	hitY := hy + hh/2 - birdH/2
	if !pm.collidesAt(hitX, hitY, birdW, birdH, variant, frame) {
		t.Fatal("Collides: expected hit on dad opaque width")
	}
}
