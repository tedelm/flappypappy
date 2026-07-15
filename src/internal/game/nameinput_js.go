//go:build js && wasm

package game

import "syscall/js"

func ShowNameInput(value string, x, y, w, h float64, maxLen int) {
	js.Global().Call("flappyShowNameInput", x, y, w, h, value, maxLen)
}

func HideNameInput() {
	js.Global().Call("flappyHideNameInput")
}

func SyncNameInput(dst *string) {
	v := js.Global().Call("flappySyncNameInput").String()
	*dst = v
}

func FocusNameInput() {
	js.Global().Call("flappyFocusNameInput")
}
