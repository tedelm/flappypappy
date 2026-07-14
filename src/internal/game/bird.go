package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Bird struct {
	X, Y   float64
	VY     float64
	Width  float64
	Height float64
}

func NewBird() *Bird {
	return &Bird{
		X:      BirdStartX,
		Y:      float64(ScreenH-GroundHeight) / 2,
		Width:  BirdWidth,
		Height: BirdHeight,
	}
}

func (b *Bird) Reset() {
	b.X = BirdStartX
	b.Y = float64(ScreenH-GroundHeight) / 2
	b.VY = 0
}

func (b *Bird) Flap() {
	b.VY = FlapStrength
}

func (b *Bird) Update() {
	b.VY += Gravity
	b.Y += b.VY
}

func (b *Bird) Bounds() (x, y, w, h float64) {
	return b.X - b.Width/2, b.Y - b.Height/2, b.Width, b.Height
}

func (b *Bird) Draw(screen *ebiten.Image) {
	cx := float32(b.X)
	cy := float32(b.Y)
	rx := float32(b.Width / 2)

	angle := math.Atan2(b.VY, 6) * 0.4
	if angle > 0.5 {
		angle = 0.5
	}
	if angle < -0.8 {
		angle = -0.8
	}

	vector.DrawFilledCircle(screen, cx, cy, rx, ColorBird, true)

	beakLen := float32(10)
	beakH := float32(6)
	beakX := cx + rx*float32(math.Cos(angle))
	beakY := cy + rx*float32(math.Sin(angle))
	vector.DrawFilledRect(screen, beakX, beakY-beakH/2, beakLen, beakH, ColorBeak, true)

	eyeX := cx + rx*0.3*float32(math.Cos(angle-0.3))
	eyeY := cy + rx*0.3*float32(math.Sin(angle-0.3))
	vector.DrawFilledCircle(screen, eyeX, eyeY, 4, color.White, true)
	vector.DrawFilledCircle(screen, eyeX+1, eyeY, 2, color.Black, true)
}
