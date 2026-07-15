//go:build !js || !wasm

package game

func HideLoadingScreen() {}

func SetLoadingText(msg string) {}
