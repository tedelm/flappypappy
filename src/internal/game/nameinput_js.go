//go:build js && wasm

package game

import "syscall/js"

func callJS(name string, args ...any) {
	fn := js.Global().Get(name)
	if fn.Type() != js.TypeFunction {
		return
	}
	fn.Invoke(args...)
}

func ShowNameInput(value string, x, y, w, h float64, maxLen int) {
	callJS("flappyShowNameInput", x, y, w, h, value, maxLen)
}

func HideNameInput() {
	callJS("flappyHideNameInput")
}

func SyncNameInput(dst *string) {
	fn := js.Global().Get("flappySyncNameInput")
	if fn.Type() != js.TypeFunction {
		return
	}
	*dst = fn.Invoke().String()
}

func FocusNameInput() {
	callJS("flappyFocusNameInput")
}
