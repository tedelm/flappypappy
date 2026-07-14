package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"flappy/internal/game"
)

func main() {
	ebiten.SetWindowSize(game.ScreenW, game.ScreenH)
	ebiten.SetWindowTitle("Flappy Bird")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)

	g := game.New()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
