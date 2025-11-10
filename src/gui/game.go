package gui

import (
	"fmt"
	"log"
	"time"
	"simulator/src/core"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// Game mantiene imágenes y estado UI
type Game struct {
	raspberry *ebiten.Image
	guante    *ebiten.Image

	// Estado de simulación (solo para UI)
	simulating bool
	// debounce para evitar múltiples toggles por un mismo click
	lastMousePressed bool
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
	// detectar clicks sobre botones
	// Botón Iniciar Simulación: x=20,y=380,w=280,h=50
	// Botón Detener Simulación: x=320,y=380,w=280,h=50

	// usamos inpututil para detectar clicks del mouse
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		// iniciar
		if pointInRect(x, y, 20, 380, 280, 50) {
			// llamar a core.StartSimulation()
			if !core.IsSimulationActive() {
				// lanzamos con parámetros razonables; número dispositivos 5, intervalo 1s
				err := core.StartSimulation(5, 1*time.Second)
				if err != nil {
					// reportar al usuario por consola
					fmt.Println("No se pudo iniciar simulación:", err)
				} else {
					g.simulating = true
				}
			}
		}
		// detener
		if pointInRect(x, y, 320, 380, 280, 50) {
			if core.IsSimulationActive() {
				core.StopSimulation()
				g.simulating = false
			}
		}
	}

	// alternativamente sincronizar estado desde core (por si se detiene desde fuera)
	g.simulating = core.IsSimulationActive()

	return nil
}

func pointInRect(px, py, x, y, w, h int) bool {
	return px >= x && px <= x+w && py >= y && py <= y+h
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
	text.Draw(screen, "Monitoreo de Ritmo, Movimiento, Temperatura y Oxigenacion", basicfont.Face7x13, 40, 240, ColorTextDark)

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Scale(0.15, 0.15)
	op2.GeoM.Translate(400, 195)
	screen.DrawImage(g.guante, op2)

	// === Panel de Controles ===
text.Draw(screen, "Controles", basicfont.Face7x13, 20, 360, ColorSecondary)

drawButton(screen, 20, 380, 280, 50, "Iniciar Simulacion", ColorGreenButton)
drawButton(screen, 320, 380, 280, 50, "Detener Simulacion", ColorRedButton)

// Estado de transmisión
status := "OFF"
statusColor := ColorRedButton

if g.simulating {
    status = "ENVIANDO DATOS"
    statusColor = ColorGreenButton
}

// Panel de estado (usa statusColor para indicar visualmente)
drawPanel(screen, 420, 340, 200, 40, statusColor)
text.Draw(screen, "Estado: "+status, basicfont.Face7x13, 430, 365, ColorWhite)

}

func (g *Game) Layout(w, h int) (int, int) {
	return 640, 480
}
