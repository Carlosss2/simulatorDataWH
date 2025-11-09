package gui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// Paleta de colores
var (
	ColorBackground   = color.RGBA{0xF4, 0xF7, 0xFB, 0xFF} // Fondo general
	ColorPrimary      = color.RGBA{0xD3, 0x2F, 0x2F, 0xFF} // Rojo WarmHeart
	ColorSecondary    = color.RGBA{0x06, 0x42, 0x70, 0xFF} // Azul profundo
	ColorPanel        = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF} // Panel blanco
	ColorTextDark     = color.RGBA{0x22, 0x22, 0x22, 0xFF} // Texto oscuro
	ColorGreenButton  = color.RGBA{0x28, 0xA7, 0x45, 0xFF} // Verde suave
	ColorRedButton    = ColorPrimary                        // Rojo WarmHeart
	ColorWhite        = color.RGBA{255, 255, 255, 255}
)


// Panel rectangular
func drawPanel(screen *ebiten.Image, x, y, w, h int, clr color.Color) {
	panel := ebiten.NewImage(w, h)
	panel.Fill(clr)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))

	// Sombra sutil WarmHeart
	shadow := ebiten.NewImage(w, h)
	shadow.Fill(ColorSecondary)
	shadowOp := &ebiten.DrawImageOptions{}
	shadowOp.GeoM.Translate(float64(x)+3, float64(y)+3)
	shadowOp.ColorScale.ScaleAlpha(0.20)

	screen.DrawImage(shadow, shadowOp)
	screen.DrawImage(panel, op)
}

// Botón UI
func drawButton(screen *ebiten.Image, x, y, w, h int, label string, col color.Color) {
	btn := ebiten.NewImage(w, h)
	btn.Fill(col)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)

	text.Draw(screen, label, basicfont.Face7x13, x+15, y+30, ColorWhite)
}
