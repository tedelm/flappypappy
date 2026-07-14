package game

import "image/color"

const (
	ScreenW = 400
	ScreenH = 600

	Gravity       = 0.5
	FlapStrength  = -8.0
	PipeSpeed     = 3.0
	PipeGap       = 150
	PipeWidth     = 60
	SpawnInterval = 120
	GroundHeight  = 80

	BirdStartX = 80
	BirdWidth  = 34
	BirdHeight = 24
)

var (
	ColorSky      = color.RGBA{135, 206, 235, 255}
	ColorGround   = color.RGBA{222, 184, 135, 255}
	ColorPipe     = color.RGBA{34, 139, 34, 255}
	ColorPipeEdge = color.RGBA{0, 100, 0, 255}
	ColorBird     = color.RGBA{255, 220, 0, 255}
	ColorBeak     = color.RGBA{255, 140, 0, 255}
)
