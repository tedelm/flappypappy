//go:build js && wasm

package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"flappy/internal/game"
)

func main() {
	ebiten.SetRunnableOnUnfocused(true)

	g := game.New()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
