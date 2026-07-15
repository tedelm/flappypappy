//go:build !js || !wasm

package game

func ShowNameInput(value string, x, y, w, h float64, maxLen int) {}

func HideNameInput() {}

func SyncNameInput(dst *string) {}

func FocusNameInput() {}
