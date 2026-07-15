//go:build js && wasm

package game

import (
	"log"
	"sync"

	"syscall/js"
)

var missingJSWarn sync.Map

func callJS(name string, args ...any) {
	fn := js.Global().Get(name)
	if fn.Type() != js.TypeFunction {
		if _, loaded := missingJSWarn.LoadOrStore(name, true); !loaded {
			log.Printf("js bridge: function %q not available", name)
		}
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
