//go:build js && wasm

package game

func HideLoadingScreen() {
	callJS("flappyHideLoading")
}

func SetLoadingText(msg string) {
	callJS("flappySetLoadingText", msg)
}
