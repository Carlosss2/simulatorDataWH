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

func worker(ctx context.Context, workerId int, jobs <-chan models.DeviceData, results chan<- *models.Message, wg *sync.WaitGroup) {
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

func producer(ctx context.Context, numDevices int, interval time.Duration, jobs chan<- models.DeviceData, results chan<- *models.Message, wg *sync.WaitGroup) {
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
			for i := 1; i <= numDevices; i++ {
				dev := models.DeviceData{
					IdDevice: i,
					IdUser:   i,
				}
				select {
				case jobs <- dev:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}


// StartSimulation inicia el patrón fan-out / fan-in que genera mensajes simulados y los publica.
// numDevices: cuántos dispositivos simular (por ejemplo 5)
//interval: intervalo entre "requests" generadas por el productor (por ejemplo 500*time.Millisecond)
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
	if workerCount > 8 {
		workerCount = 8
	}
	if workerCount <= 0 {
		workerCount = 1
	}

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(ctx, i, jobs, results, &wg)
	}

	go publisher(ctx, results)
	go producer(ctx, numDevices, interval, jobs, results, &wg)

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
