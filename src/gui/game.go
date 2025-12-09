package gui

import (
	"fmt"
	"image/color"

	"math"
	"math/rand"
	"time"

	"simulator/src/core"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

// Colores atractivos y modernos
var (
	colorBg       = color.RGBA{15, 23, 42, 255}    // Azul oscuro elegante
	colorGlove    = color.RGBA{148, 163, 184, 255} // Gris azulado
	colorThumb    = color.RGBA{34, 211, 238, 255}  // Cyan vibrante
	colorIndex    = color.RGBA{52, 211, 153, 255}  // Verde esmeralda
	colorMiddle   = color.RGBA{250, 204, 21, 255}  // Amarillo dorado
	colorRing     = color.RGBA{251, 146, 60, 255}  // Naranja cálido
	colorPinky    = color.RGBA{244, 114, 182, 255} // Rosa suave
	colorCloud    = color.RGBA{226, 232, 240, 255} // Gris claro
	colorDataText = color.RGBA{100, 116, 139, 255} // Gris medio
)

// Finger representa un dedo del guante
type Finger struct {
	Name     string
	X, Y     float64
	Color    color.RGBA
	Active   bool
	DataRate float64
}

// Particle representa una partícula visual
type Particle struct {
	X, Y       float64
	VX, VY     float64
	Life       float64
	MaxLife    float64
	Color      color.RGBA
	Size       float64
	FingerName string
}

// Game contiene todo el estado del UI
type Game struct {
	fingers       []Finger
	particles     []Particle
	time          float64
	lastEmit      float64
	dataCounters  map[string]int
	sendEvents    <-chan int
	simulating    bool
	cloudX        float32
	cloudY        float32
	buttonStartX  float32
	buttonStartY  float32
	buttonStopX   float32
	buttonStopY   float32
	buttonW       float32
	buttonH       float32
}

// NewGame crea e inicializa el Game con posiciones y colores
func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		dataCounters: make(map[string]int),
		cloudX:       float32(screenWidth) / 2,
		cloudY:       80,
		buttonW:      180,
		buttonH:      44,
	}
	// Posiciones de los dedos en el guante (centro de la pantalla)
	centerX := float64(screenWidth) / 2
	centerY := float64(screenHeight)/2 + 40

	// Reubicamos dedos alrededor de una palma más grande
	g.fingers = []Finger{
		{Name: "Pulgar", X: centerX - 120, Y: centerY + 40, Color: colorThumb, DataRate: 1.2},
		{Name: "Índice", X: centerX - 60, Y: centerY - 10, Color: colorIndex, DataRate: 1.0},
		{Name: "Medio", X: centerX, Y: centerY - 30, Color: colorMiddle, DataRate: 0.8},
		{Name: "Anular", X: centerX + 60, Y: centerY - 10, Color: colorRing, DataRate: 1.1},
		{Name: "Meñique", X: centerX + 120, Y: centerY + 10, Color: colorPinky, DataRate: 0.9},
	}

	// Posición de los botones
	g.buttonStartX = 20
	g.buttonStartY = float32(screenHeight) - 80
	g.buttonStopX = g.buttonStartX + g.buttonW + 20
	g.buttonStopY = g.buttonStartY

	// **Todos los dedos activos** para que envíen datos simultáneamente
	for i := range g.fingers {
		g.fingers[i].Active = true
	}

	return g
}

// Update corre ~60 FPS y maneja inputs, eventos del core y actualización de partículas
func (g *Game) Update() error {
	g.time += 1.0 / 60.0

	// Leer clicks de mouse (Start / Stop)
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if pointInRectInt(x, y, int(g.buttonStartX), int(g.buttonStartY), int(g.buttonW), int(g.buttonH)) {
			if !core.IsSimulationActive() {
				// Iniciar simulación con valor por defecto (puedes cambiarlo)
				err := core.StartSimulation(1*time.Second)
				if err != nil {
					fmt.Println("No se pudo iniciar simulación:", err)
				} else {
					g.simulating = true
					// obtener canal de eventos de core
					g.sendEvents = core.GetSendOK()
				}
			}
		}
		if pointInRectInt(x, y, int(g.buttonStopX), int(g.buttonStopY), int(g.buttonW), int(g.buttonH)) {
			if core.IsSimulationActive() {
				core.StopSimulation()
				g.simulating = false
				// limpiar partículas al detener la simulación
				g.particles = nil
				// sendEvents será cerrado por core; dejamos referencia para que Update lo detecte
			}
		}
	}

	// actualizar estado de simulación (por si se detiene desde otro sitio)
	g.simulating = core.IsSimulationActive()

	// Procesar eventos del core (envío exitoso de mensajes)
	g.readSendEvents()

	// Emitir partículas SOLO si la simulación está activa
	if g.simulating && g.time-g.lastEmit > 0.05 {
		for _, finger := range g.fingers {
			// todos los dedos envían al mismo tiempo
			g.emitParticle(finger)
			g.dataCounters[finger.Name]++
		}
		g.lastEmit = g.time
	}

	// Actualizar partículas (movimiento y vida)
	g.updateParticles()

	return nil
}

// readSendEvents consume eventos SendOK desde core sin bloquear el frame loop
func (g *Game) readSendEvents() {
	if g.sendEvents == nil {
		// intentar recuperar canal si la simulación está activa
		if core.IsSimulationActive() {
			g.sendEvents = core.GetSendOK()
		}
		return
	}
	for {
		select {
		case id, ok := <-g.sendEvents:
			if !ok {
				// canal cerrado -> dejarlo nil
				g.sendEvents = nil
				return
			}
			// Mapear deviceID a finger index cíclicamente y dar un burst visual
			g.handleDeviceSent(id)
		default:
			return
		}
	}
}

// handleDeviceSent mapea deviceID a dedo y emite partículas visuales para reforzar el envío
func (g *Game) handleDeviceSent(deviceID int) {
	if len(g.fingers) == 0 {
		return
	}
	index := (deviceID - 2) % len(g.fingers) // deviceIDs empiezan en 2
	if index < 0 {
		index = 0
	}
	// emitir burst visual
	g.emitMultipleFromFinger(g.fingers[index], 10)
	// aumentar contador visual
	g.dataCounters[g.fingers[index].Name]++
}

// emitMultipleFromFinger emite n partículas inmediatamente desde un dedo
func (g *Game) emitMultipleFromFinger(finger Finger, n int) {
	for i := 0; i < n; i++ {
		g.emitParticle(finger)
	}
}

// emitParticle crea y añade una partícula con velocidad hacia arriba (hacia la nube)
func (g *Game) emitParticle(finger Finger) {
	angle := -math.Pi/2 + (rand.Float64()-0.5)*0.4
	speed := 2.0 + rand.Float64()*2.0

	p := Particle{
		X:          finger.X,
		Y:          finger.Y,
		VX:         math.Cos(angle) * speed,
		VY:         math.Sin(angle) * speed,
		Life:       2.0 + rand.Float64(),
		MaxLife:    2.0 + rand.Float64(),
		Color:      finger.Color,
		Size:       3 + rand.Float64()*2,
		FingerName: finger.Name,
	}

	g.particles = append(g.particles, p)
}

// updateParticles actualiza posición y vida de partículas y limpia las muertas
func (g *Game) updateParticles() {
	if len(g.particles) == 0 {
		return
	}
	newParticles := make([]Particle, 0, len(g.particles))
	for i := range g.particles {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY -= 0.05 // gravedad invertida (suben)
		p.Life -= 1.0 / 60.0
		if p.Life > 0 {
			newParticles = append(newParticles, *p)
		}
	}
	g.particles = newParticles
}

// Draw renderiza toda la UI
func (g *Game) Draw(screen *ebiten.Image) {
	// Fondo
	screen.Fill(colorBg)

	// Dibujar nube
	g.drawCloud(screen)

	// Dibujar palma del guante (círculo grande) - ahora más grande
	centerX := float32(screenWidth) / 2
	centerY := float32(screenHeight)/2 + 50
	vector.DrawFilledCircle(screen, centerX, centerY, 90, colorGlove, false)

	// Dibujar partículas
	for _, p := range g.particles {
		alpha := uint8(255 * (p.Life / p.MaxLife))
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		vector.DrawFilledCircle(screen, float32(p.X), float32(p.Y), float32(p.Size), c, false)
	}

	// Dibujar dedos (más cercanos a palma grande)
	for _, finger := range g.fingers {
		c := finger.Color
		// todos están activos por diseño
		vector.DrawFilledCircle(screen, float32(finger.X), float32(finger.Y), 18, c, false)

		// Efecto de brillo cuando está activo
		pulseAlpha := uint8(100 + 100*math.Sin(g.time*5))
		glowColor := color.RGBA{c.R, c.G, c.B, pulseAlpha}
		vector.DrawFilledCircle(screen, float32(finger.X), float32(finger.Y), 26, glowColor, false)

		// Línea conectando al centro
		vector.StrokeLine(screen, float32(finger.X), float32(finger.Y), centerX, centerY, 2, c, false)
	}

	// Panel de información (lado izquierdo) — simplificado: sin lista de dedos
	g.drawInfoPanel(screen)

	// Botones Start / Stop (abajo a la izquierda)
	g.drawButton(screen, g.buttonStartX, g.buttonStartY, g.buttonW, g.buttonH, "Iniciar Simulacion", color.RGBA{34, 197, 94, 220})
	g.drawButton(screen, g.buttonStopX, g.buttonStopY, g.buttonW, g.buttonH, "Detener Simulacion", color.RGBA{239, 68, 68, 220})

	// Título superior izquierdo
	ebitenutil.DebugPrintAt(screen, "SIMULADOR DE DATOS - GUANTE IoT", 10, 10)
}

// drawCloud dibuja la nube compuesta de círculos
func (g *Game) drawCloud(screen *ebiten.Image) {
	cloudY := g.cloudY
	cloudX := g.cloudX

	circles := []struct{ x, y, r float32 }{
		{cloudX - 40, cloudY, 25},
		{cloudX - 15, cloudY - 10, 30},
		{cloudX + 15, cloudY - 10, 30},
		{cloudX + 40, cloudY, 25},
		{cloudX, cloudY + 10, 28},
	}

	for _, c := range circles {
		vector.DrawFilledCircle(screen, c.x, c.y, c.r, colorCloud, false)
	}

	ebitenutil.DebugPrintAt(screen, "RabbitMQ", int(cloudX)-20, int(cloudY)-5)
}

// drawInfoPanel dibuja un panel simplificado con solo TOTAL y mensaje
func (g *Game) drawInfoPanel(screen *ebiten.Image) {
	panelX := float32(20)
	panelY := float32(110)
	panelW := float32(220)
	panelH := float32(140)

	// fondo y borde
	vector.DrawFilledRect(screen, panelX, panelY, panelW, panelH,
		color.RGBA{30, 41, 59, 200}, false)
	vector.StrokeRect(screen, panelX, panelY, panelW, panelH, 2,
		color.RGBA{71, 85, 105, 255}, false)

	ebitenutil.DebugPrintAt(screen, "DATOS TRANSMITIDOS", int(panelX)+10, int(panelY)+10)

	// total
	total := 0
	for _, count := range g.dataCounters {
		total += count
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("TOTAL: %d pkts", total),
		int(panelX)+10, int(panelY+panelH)-24)
}

// drawButton dibuja un rectángulo y texto para un botón
func (g *Game) drawButton(screen *ebiten.Image, x, y, w, h float32, label string, fill color.RGBA) {
	vector.DrawFilledRect(screen, x, y, w, h, fill, false)
	vector.StrokeRect(screen, x, y, w, h, 2, color.RGBA{0, 0, 0, 80}, false)
	ebitenutil.DebugPrintAt(screen, label, int(x)+12, int(y)+14)
}

// pointInRectInt auxiliar (int)
func pointInRectInt(px, py, x, y, w, h int) bool {
	return px >= x && px <= x+w && py >= y && py <= y+h
}

// Layout fija las dimensiones
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
