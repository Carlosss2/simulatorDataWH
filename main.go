package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"simulator/src/gui"
)

func main() {
	game := gui.StartUI()

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Simulador WarmHeart IoT")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
