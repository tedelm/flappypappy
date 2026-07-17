package game

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestDrawBossResultBanner(t *testing.T) {
	screen := ebiten.NewImage(ScreenW, ScreenH)
	drawBossResultBanner(screen, true, 0)
	drawBossResultBanner(screen, false, 30)
	drawBossResultBanner(screen, true, 90)
}
