package gui

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

type Game struct {
	raspberry *ebiten.Image
	guante    *ebiten.Image
}

func NewGame() *Game {
	raspberry, _, err := ebitenutil.NewImageFromFile("src/gui/assets/raspberry.png")
	if err != nil {
		log.Fatal(err)
	}

	guante, _, err := ebitenutil.NewImageFromFile("src/gui/assets/guanteWH.png")
	if err != nil {
		log.Fatal(err)
	}

	return &Game{
		raspberry: raspberry,
		guante:    guante,
	}
}

func (g *Game) Update() error {
	// Aquí luego agregaremos interacción y simulación.
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Fondo cálido
	screen.Fill(ColorBackground)

	// Título principal (azul para confianza)
	text.Draw(screen, "Area de Simulacion", basicfont.Face7x13, 20, 30, ColorSecondary)

	// === Tarjeta Raspberry ===
	drawPanel(screen, 20, 50, 600, 120, ColorPanel)
	text.Draw(screen, "Raspberry Pi 4", basicfont.Face7x13, 40, 80, ColorPrimary)
	text.Draw(screen, "Controlador Principal", basicfont.Face7x13, 40, 100, ColorTextDark)

	// Imagen Raspberry con tamaño balanceado
	op1 := &ebiten.DrawImageOptions{}
	op1.GeoM.Scale(0.15, 0.15)
	op1.GeoM.Translate(200, 45)
	screen.DrawImage(g.raspberry, op1)


	// === Tarjeta Guante Inteligente ===
	drawPanel(screen, 20, 190, 600, 140, ColorPanel)
	text.Draw(screen, "Guante Inteligente", basicfont.Face7x13, 40, 220, ColorPrimary)
	text.Draw(screen, "Monitoreo de Ritmo, Temperatura y Oxigenacion", basicfont.Face7x13, 40, 240, ColorTextDark)

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Scale(0.15, 0.15)
	op2.GeoM.Translate(320, 195)
	screen.DrawImage(g.guante, op2)

	// === Panel de Controles ===
	text.Draw(screen, "Controles", basicfont.Face7x13, 20, 360, ColorSecondary)

	drawButton(screen, 20, 380, 280, 50, "Iniciar Simulacion", ColorGreenButton)
	drawButton(screen, 320, 380, 280, 50, "Detener Simulacion", ColorRedButton)
}
func (g *Game) Layout(w, h int) (int, int) {
	return 640, 480
}
