package sprite

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed player_punching.png
var playerPunchingPNG []byte

//go:embed player_punching_uppercut.png
var playerPunchingUppercutPNG []byte

//go:embed player_punching_win.png
var playerPunchingWinPNG []byte

//go:embed player_punching_lose.png
var playerPunchingLosePNG []byte

const (
	PlayerPunchFrameCols  = 8
	PlayerPunchFrameRows  = 1
	PlayerPunchFrameCount = PlayerPunchFrameCols * PlayerPunchFrameRows
	PlayerPunchFrameTicks = 4

	PlayerPunchIdle0   = 0
	PlayerPunchIdle1   = 4
	PlayerPunchIdle2   = 7
	PlayerPunchJabStart = 1
	PlayerPunchJabCount = 3
	PlayerPunchGuard    = 3
	PlayerPunchCrossStart = 5
	PlayerPunchCrossCount = 3

	PlayerUppercutFrameCols  = 4
	PlayerUppercutFrameRows  = 1
	PlayerUppercutFrameCount = PlayerUppercutFrameCols * PlayerUppercutFrameRows
	PlayerUppercutFrameTicks = 4

	PlayerPunchWinFrameCols  = 3
	PlayerPunchWinFrameRows  = 1
	PlayerPunchWinFrameCount = PlayerPunchWinFrameCols * PlayerPunchWinFrameRows

	PlayerPunchLoseFrameCols  = 5
	PlayerPunchLoseFrameRows  = 1
	PlayerPunchLoseFrameCount = PlayerPunchLoseFrameCols * PlayerPunchLoseFrameRows
)

var (
	playerPunchOnce     sync.Once
	playerPunchFrames   [PlayerPunchFrameCount]*ebiten.Image
	playerPunchContentH float64
	playerPunchFeetPad  float64

	playerUppercutOnce   sync.Once
	playerUppercutFrames [PlayerUppercutFrameCount]*ebiten.Image

	playerPunchWinOnce   sync.Once
	playerPunchWinFrames [PlayerPunchWinFrameCount]*ebiten.Image

	playerPunchLoseOnce   sync.Once
	playerPunchLoseFrames [PlayerPunchLoseFrameCount]*ebiten.Image
)

func loadSheetFrames(pngBytes []byte, cols, rows, count int, out []*ebiten.Image) (contentH, feetPad float64) {
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		panic(err)
	}
	sheet := KeyBlackTransparent(ToRGBA(img))
	b := sheet.Bounds()
	sheetW, sheetH := b.Dx(), b.Dy()

	for i := 0; i < count; i++ {
		col := i % cols
		row := i / cols
		rect := image.Rect(
			col*sheetW/cols,
			row*sheetH/rows,
			(col+1)*sheetW/cols,
			(row+1)*sheetH/rows,
		)
		cell := cropRGBA(sheet, rect)
		pad, ch := MeasureRGBA(cell)
		if pad > feetPad {
			feetPad = pad
		}
		if ch > contentH {
			contentH = ch
		}
		out[i] = ebiten.NewImageFromImage(cell)
	}
	if contentH == 0 {
		contentH = float64(sheetH / rows)
	}
	return contentH, feetPad
}

func ensurePlayerPunchLoaded() {
	playerPunchOnce.Do(func() {
		frames := make([]*ebiten.Image, PlayerPunchFrameCount)
		playerPunchContentH, playerPunchFeetPad = loadSheetFrames(
			playerPunchingPNG, PlayerPunchFrameCols, PlayerPunchFrameRows, PlayerPunchFrameCount, frames,
		)
		for i := range frames {
			playerPunchFrames[i] = frames[i]
		}
	})
}

func ensurePlayerUppercutLoaded() {
	playerUppercutOnce.Do(func() {
		frames := make([]*ebiten.Image, PlayerUppercutFrameCount)
		loadSheetFrames(playerPunchingUppercutPNG, PlayerUppercutFrameCols, PlayerUppercutFrameRows, PlayerUppercutFrameCount, frames)
		for i := range frames {
			playerUppercutFrames[i] = frames[i]
		}
	})
}

func ensurePlayerPunchWinLoaded() {
	playerPunchWinOnce.Do(func() {
		frames := make([]*ebiten.Image, PlayerPunchWinFrameCount)
		loadSheetFrames(playerPunchingWinPNG, PlayerPunchWinFrameCols, PlayerPunchWinFrameRows, PlayerPunchWinFrameCount, frames)
		for i := range frames {
			playerPunchWinFrames[i] = frames[i]
		}
	})
}

func ensurePlayerPunchLoseLoaded() {
	playerPunchLoseOnce.Do(func() {
		frames := make([]*ebiten.Image, PlayerPunchLoseFrameCount)
		loadSheetFrames(playerPunchingLosePNG, PlayerPunchLoseFrameCols, PlayerPunchLoseFrameRows, PlayerPunchLoseFrameCount, frames)
		for i := range frames {
			playerPunchLoseFrames[i] = frames[i]
		}
	})
}

func PlayerPunchFrame(i int) *ebiten.Image {
	ensurePlayerPunchLoaded()
	if i < 0 {
		i = 0
	}
	return playerPunchFrames[i%PlayerPunchFrameCount]
}

func PlayerUppercutFrame(i int) *ebiten.Image {
	ensurePlayerUppercutLoaded()
	if i < 0 {
		i = 0
	}
	return playerUppercutFrames[i%PlayerUppercutFrameCount]
}

func PlayerPunchWinFrame(i int) *ebiten.Image {
	ensurePlayerPunchWinLoaded()
	if i < 0 {
		i = 0
	}
	return playerPunchWinFrames[i%PlayerPunchWinFrameCount]
}

func PlayerPunchLoseFrame(i int) *ebiten.Image {
	ensurePlayerPunchLoseLoaded()
	if i < 0 {
		i = 0
	}
	return playerPunchLoseFrames[i%PlayerPunchLoseFrameCount]
}

func PlayerPunchContentH() float64 {
	ensurePlayerPunchLoaded()
	return playerPunchContentH
}

func PlayerPunchFeetPad() float64 {
	ensurePlayerPunchLoaded()
	return playerPunchFeetPad
}

// PlayerPunchIdleFrameIndex returns a looping idle frame from the punch sheet.
func PlayerPunchIdleFrameIndex(tick int) int {
	idle := [...]int{PlayerPunchIdle0, PlayerPunchIdle1, PlayerPunchIdle2}
	return idle[(tick/10)%len(idle)]
}
