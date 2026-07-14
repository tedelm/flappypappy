package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Bird struct {
	X, Y          float64
	VY            float64
	Width         float64
	Height        float64
	gravity       float64
	flapStrength  float64
}

func NewBird() *Bird {
	b := &Bird{
		X:      BirdStartX,
		Y:      float64(ScreenH-GroundHeight) / 2,
		Width:  BirdWidth,
		Height: BirdHeight,
	}
	b.ApplyConfig(DifficultyEasy.Config())
	return b
}

func (b *Bird) ApplyConfig(cfg DifficultyConfig) {
	b.gravity = cfg.Gravity
	b.flapStrength = cfg.FlapStrength
}

func (b *Bird) Reset() {
	b.X = BirdStartX
	b.Y = float64(ScreenH-GroundHeight) / 2
	b.VY = 0
}

func (b *Bird) Flap() {
	b.VY = b.flapStrength
}

func (b *Bird) Update() {
	b.VY += b.gravity
	b.Y += b.VY
}

func (b *Bird) Bounds() (x, y, w, h float64) {
	return b.X - b.Width/2, b.Y - b.Height/2, b.Width, b.Height
}

func (b *Bird) tiltAngle() float64 {
	angle := math.Atan2(b.VY, 6) * 0.4
	if angle > 0.5 {
		return 0.5
	}
	if angle < -0.8 {
		return -0.8
	}
	return angle
}

func (b *Bird) Draw(screen *ebiten.Image) {
	cx := b.X
	cy := b.Y
	angle := b.tiltAngle()

	halfH := b.Height / 2
	topHW := b.Width/2 - 2
	botHW := b.Width/2 - 6
	inset := 3.0

	glass := toWorldQuad([4][2]float64{
		{-topHW, -halfH},
		{topHW, -halfH},
		{botHW, halfH},
		{-botHW, halfH},
	}, cx, cy, angle)

	foamY := -halfH + 10
	beerTopY := foamY + 6
	beerBottomY := halfH - inset

	beer := toWorldQuad([4][2]float64{
		{localEdgeX(topHW, botHW, halfH, beerTopY, true), beerTopY},
		{localEdgeX(topHW, botHW, halfH, beerTopY, false), beerTopY},
		{localEdgeX(topHW, botHW, halfH, beerBottomY, false), beerBottomY},
		{localEdgeX(topHW, botHW, halfH, beerBottomY, true), beerBottomY},
	}, cx, cy, angle)

	drawFilledQuad(screen, glass, ColorGlass)
	drawFilledQuad(screen, beer, ColorBeer)

	foamCX, foamCY := toWorld(0, foamY, cx, cy, angle)
	foamW := float32((topHW - inset) * 2)
	vector.DrawFilledRect(screen, float32(foamCX)-foamW/2, float32(foamCY)-3, foamW, 7, ColorFoam, true)
	vector.DrawFilledCircle(screen, float32(foamCX-topHW+6), float32(foamCY), 5, ColorFoam, true)
	vector.DrawFilledCircle(screen, float32(foamCX), float32(foamCY-2), 6, ColorFoam, true)
	vector.DrawFilledCircle(screen, float32(foamCX+topHW-6), float32(foamCY+1), 5, ColorFoam, true)

	bubbleOffsets := [][2]float64{{-4, 8}, {3, 14}, {-1, 20}}
	for _, off := range bubbleOffsets {
		bx, by := toWorld(off[0], -halfH+off[1], cx, cy, angle)
		vector.DrawFilledCircle(screen, float32(bx), float32(by), 2, ColorBubble, true)
	}

	drawStrokedQuad(screen, glass, 2, ColorGlassEdge)

	hlTopX := localEdgeX(topHW, botHW, halfH, -halfH+6, true) + 3
	hlBotX := localEdgeX(topHW, botHW, halfH, halfH-6, true) + 3
	hlTopX, hlTopY := toWorld(hlTopX, -halfH+6, cx, cy, angle)
	hlBotX, hlBotY := toWorld(hlBotX, halfH-6, cx, cy, angle)
	vector.StrokeLine(screen, float32(hlTopX), float32(hlTopY),
		float32(hlBotX), float32(hlBotY), 1.5, color.RGBA{230, 230, 240, 180}, true)
}

func localEdgeX(topHW, botHW, halfH, y float64, left bool) float64 {
	t := (y + halfH) / (2 * halfH)
	hw := topHW + (botHW-topHW)*t
	if left {
		return -hw + 3
	}
	return hw - 3
}

func toWorld(lx, ly, cx, cy, angle float64) (float64, float64) {
	x, y := rotPoint(cx, cy, cx+lx, cy+ly, angle)
	return float64(x), float64(y)
}

func toWorldQuad(local [4][2]float64, cx, cy, angle float64) [4][2]float64 {
	var out [4][2]float64
	for i, p := range local {
		out[i][0], out[i][1] = toWorld(p[0], p[1], cx, cy, angle)
	}
	return out
}

func rotPoint(cx, cy, px, py, angle float64) (float32, float32) {
	dx := px - cx
	dy := py - cy
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	return float32(cx + dx*cos - dy*sin), float32(cy + dx*sin + dy*cos)
}

func drawFilledQuad(screen *ebiten.Image, pts [4][2]float64, col color.Color) {
	var path vector.Path
	path.MoveTo(float32(pts[0][0]), float32(pts[0][1]))
	for i := 1; i < 4; i++ {
		path.LineTo(float32(pts[i][0]), float32(pts[i][1]))
	}
	path.Close()

	var drawOpts vector.DrawPathOptions
	drawOpts.ColorScale.ScaleWithColor(col)
	drawOpts.AntiAlias = true
	vector.FillPath(screen, &path, &vector.FillOptions{}, &drawOpts)
}

func drawStrokedQuad(screen *ebiten.Image, pts [4][2]float64, width float32, col color.Color) {
	var path vector.Path
	path.MoveTo(float32(pts[0][0]), float32(pts[0][1]))
	for i := 1; i < 4; i++ {
		path.LineTo(float32(pts[i][0]), float32(pts[i][1]))
	}
	path.Close()

	var drawOpts vector.DrawPathOptions
	drawOpts.ColorScale.ScaleWithColor(col)
	drawOpts.AntiAlias = true
	vector.StrokePath(screen, &path, &vector.StrokeOptions{Width: width}, &drawOpts)
}
