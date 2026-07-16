package sprite

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

const NumPubWalls = 6

//go:embed pub_wall.png
var pubWallPNG []byte

//go:embed pub_wall_2.png
var pubWall2PNG []byte

//go:embed pub_wall_3.png
var pubWall3PNG []byte

//go:embed pub_wall_4.png
var pubWall4PNG []byte

//go:embed pub_wall_5.png
var pubWall5PNG []byte

//go:embed pub_wall_6.png
var pubWall6PNG []byte

var pubWallPNGs = [][]byte{
	pubWallPNG,
	pubWall2PNG,
	pubWall3PNG,
	pubWall4PNG,
	pubWall5PNG,
	pubWall6PNG,
}

var (
	pubWallOnce   sync.Once
	pubWallImages []*ebiten.Image
)

func ensurePubWallsLoaded() {
	pubWallOnce.Do(func() {
		pubWallImages = make([]*ebiten.Image, NumPubWalls)
		for i, data := range pubWallPNGs {
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				panic(err)
			}
			pubWallImages[i] = ebiten.NewImageFromImage(ToRGBA(img))
		}
	})
}

func PubWall(index int) *ebiten.Image {
	ensurePubWallsLoaded()
	if index < 0 {
		index = 0
	}
	if index >= NumPubWalls {
		index = NumPubWalls - 1
	}
	return pubWallImages[index]
}
