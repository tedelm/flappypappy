package game

import "image/color"

const (
	ScreenW = 400
	ScreenH = 600

	// Easy-tier defaults; Hard/Insane values live in difficulty.go
	Gravity           = 0.5
	FlapStrength      = -6.8
	PipeSpeed         = 3.0
	PipeGap           = 150
	PipeWidth         = 60
	SpawnInterval     = 120
	GroundHeight      = 115
	FloorHeight       = 64
	FloorSurfaceY     = ScreenH - FloorHeight
	WainscotingTopY   = ScreenH / 2
	WainscotingHeight = FloorSurfaceY - WainscotingTopY

	MaxLives            = 3
	MaxLivesCap         = 10
	LifeInvincibleTicks = 90
	LifeGlassW          = 24
	LifeGlassGap        = 6
	LifeHUDMargin       = 10

	MaxHighScores    = 10
	MaxPlayerNameLen = 8

	BgParallaxFactor = 0.35

	DecorChunkMin         = 200
	DecorChunkJitter      = 120
	DecorPaintChance      = 45
	DecorStoolBelowGround = 55
	DecorTableBelowStool  = 7
	DecorStoolY           = ScreenH - GroundHeight + DecorStoolBelowGround
	DecorPaintingY        = 100
	DecorStoolScale       = 0.120
	DecorPaintingScale    = 0.80
	DecorPaintingScaleMin = 0.9
	DecorPaintingScaleMax = 1.2
	DecorTableChunkMin    = 280
	DecorTableChunkJitter = 160
	DecorTableChance      = 40
	DecorTableY           = DecorStoolY + DecorTableBelowStool
	DecorTableW           = 56
	DecorTableH           = 14
	DecorStoolsPerTable   = 3
	DecorStoolAroundX     = 32

	BirdStartX = 80
	BirdWidth  = 28
	BirdHeight = 40
)

var (
	ColorSky     = color.RGBA{135, 206, 235, 255}
	ColorPubBase = color.RGBA{42, 28, 24, 255}

	WallpaperBaseColors = []color.RGBA{
		ColorPubBase,      // burgundy
		{28, 36, 24, 255}, // forest green
		{24, 28, 42, 255}, // deep navy blue
		{36, 24, 42, 255}, // royal purple
		{42, 30, 20, 255}, // burnt rust/orange
		{24, 38, 38, 255}, // teal
	}
	ColorText        = color.RGBA{245, 245, 240, 255}
	ColorTextMuted   = color.RGBA{200, 200, 195, 255}
	ColorTextOutline = color.RGBA{20, 12, 8, 220}
	ColorKeg         = color.RGBA{139, 69, 19, 255}
	ColorKegBand     = color.RGBA{184, 134, 11, 255}
	ColorKegRim      = color.RGBA{101, 67, 33, 255}
	ColorKegEdge     = color.RGBA{80, 50, 20, 255}
	ColorGlass       = color.RGBA{200, 200, 210, 255}
	ColorGlassEdge   = color.RGBA{140, 140, 150, 255}
	ColorBeer        = color.RGBA{218, 165, 32, 255}
	ColorFoam        = color.RGBA{255, 250, 240, 255}
	ColorBubble      = color.RGBA{255, 230, 150, 255}
)
