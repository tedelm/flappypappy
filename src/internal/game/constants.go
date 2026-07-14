package game

import "image/color"

const (
	ScreenW = 400
	ScreenH = 600

	Gravity       = 0.5
	FlapStrength  = -6.0
	PipeSpeed     = 3.0
	PipeGap       = 150
	PipeWidth     = 60
	SpawnInterval = 120
	GroundHeight  = 80
	KegSegmentH   = 32

	BirdStartX = 80
	BirdWidth  = 28
	BirdHeight = 40
)

var (
	ColorSky       = color.RGBA{135, 206, 235, 255}
	ColorKeg       = color.RGBA{139, 69, 19, 255}
	ColorKegBand   = color.RGBA{184, 134, 11, 255}
	ColorKegRim    = color.RGBA{101, 67, 33, 255}
	ColorKegEdge   = color.RGBA{80, 50, 20, 255}
	ColorGlass     = color.RGBA{200, 200, 210, 255}
	ColorGlassEdge = color.RGBA{140, 140, 150, 255}
	ColorBeer      = color.RGBA{218, 165, 32, 255}
	ColorFoam      = color.RGBA{255, 250, 240, 255}
	ColorBubble    = color.RGBA{255, 230, 150, 255}
	ColorFoamLight = color.RGBA{255, 255, 255, 200}
)
