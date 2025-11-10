package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"simulator/src/models"
)

var (
	// control del ciclo de simulación
	simCancel context.CancelFunc
	simMu     sync.Mutex
	simActive bool
)

//Funciones go 

func worker(ctx context.Context, jobs <-chan models.DeviceData, results chan<- *models.Message, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case dev, ok := <-jobs:
			if !ok {
				return
			}
			msg := GenerateSensorData(dev)
			time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
			select {
			case results <- msg:
			case <-ctx.Done():
				return
			}
		}
	}
}

func publisher(ctx context.Context, results <-chan *models.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-results:
			if !ok {
				return
			}
			dataJSON, err := json.Marshal(msg)
			if err != nil {
				log.Println("Error al serializar mensaje simulado:", err)
				continue
			}
			PublishData(string(dataJSON))
		}
	}
}

func producer(ctx context.Context, interval time.Duration, jobs chan<- models.DeviceData, results chan<- *models.Message, wg *sync.WaitGroup) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			close(results)
			return

		case <-ticker.C:
			
			dev := models.DeviceData{
				IdDevice: 1,
				IdUser:   1,
			}

			select {
			case jobs <- dev:
			case <-ctx.Done():
				return
			}
		}
	}
}


func StartSimulation(numDevices int, interval time.Duration) error {
	simMu.Lock()
	defer simMu.Unlock()

	if simActive {
		return fmt.Errorf("simulación ya está activada")
	}

	if client == nil || !client.IsConnected() {
		return fmt.Errorf("MQTT no conectado - conecta primero")
	}

	ctx, cancel := context.WithCancel(context.Background())
	simCancel = cancel
	simActive = true

	jobs := make(chan models.DeviceData)
	results := make(chan *models.Message)

	workerCount := numDevices

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, jobs, results, &wg)
	}

	go publisher(ctx, results)
	go producer(ctx, interval, jobs, results, &wg)


	return nil
}

// StopSimulation detiene la simulación si está corriendo
func StopSimulation() {
	simMu.Lock()
	defer simMu.Unlock()
	if !simActive {
		return
	}
	if simCancel != nil {
		simCancel()
	}
	simActive = false
}

// IsSimulationActive devuelve si la simulación está activa
func IsSimulationActive() bool {
	simMu.Lock()
	defer simMu.Unlock()
	return simActive
}
