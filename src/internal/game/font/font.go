package font

import (
	"bytes"
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed title.ttf
var titleTTF []byte

func TitleSource() (*text.GoTextFaceSource, error) {
	return text.NewGoTextFaceSource(bytes.NewReader(titleTTF))
}
