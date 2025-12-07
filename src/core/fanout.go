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
	
	GlobalDeviceCount = 100

	// control del ciclo de simulación
	simCancel context.CancelFunc
	simMu     sync.Mutex
	simActive bool

	// canal público que notifica a la GUI cuando un mensaje se publicó OK (envía DeviceID)
	SendOK chan int
)

// simulateDevice produce measurements periodicamente para un device y las envia al canal jobs.
func simulateDevice(ctx context.Context, deviceID, userID int, interval time.Duration, jobs chan<- models.DeviceData, prodWG *sync.WaitGroup) {
	prodWG.Add(1)
	ticker := time.NewTicker(interval)
	go func() {}()
	namedLoopDevice(ctx, deviceID, userID, ticker, jobs, prodWG)
}

func namedLoopDevice(ctx context.Context, deviceID, userID int, ticker *time.Ticker, jobs chan<- models.DeviceData, prodWG *sync.WaitGroup) {
	defer prodWG.Done()
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dev := models.DeviceData{
				IdDevice: deviceID,
				IdUser:   userID,
			}
			select {
			case jobs <- dev:
			case <-ctx.Done():
				return
			}
		}
	}
}

// worker consume jobs y genera mensajes simulados
func worker(ctx context.Context, jobs <-chan models.DeviceData, results chan<- *models.Message, workerWG *sync.WaitGroup) {
	defer workerWG.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case dev, ok := <-jobs:
			if !ok {
				return
			}
			msg := GenerateSensorData(dev)
			sleepRandomDelay()
			select {
			case results <- msg:
			case <-ctx.Done():
				return
			}
		}
	}
}

func sleepRandomDelay() {
	ms := rand.Intn(200)
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// publisher publica y notifica a UI
func publisher(ctx context.Context, results <-chan *models.Message, pubWG *sync.WaitGroup) {
	pubWG.Add(1)
	defer pubWG.Done()

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

			// Notificar UI
			if SendOK != nil {
				select {
				case SendOK <- msg.DeviceId:
				default:
				}
			}
		}
	}
}

// startWorkers lanza el pool de workers
func startWorkers(ctx context.Context, workerCount int, jobs <-chan models.DeviceData, results chan<- *models.Message, workerWG *sync.WaitGroup) {
	for i := 0; i < workerCount; i++ {
		go worker(ctx, jobs, results, workerWG)
	}
}

// cleanupHandler gestiona el shutdown ordenado
func cleanupHandler(ctx context.Context, jobs chan models.DeviceData, results chan *models.Message, prodWG *sync.WaitGroup, workerWG *sync.WaitGroup, pubWG *sync.WaitGroup) {
	<-ctx.Done()

	prodWG.Wait()
	close(jobs)

	workerWG.Wait()
	close(results)

	pubWG.Wait()

	if SendOK != nil {
		close(SendOK)
		SendOK = nil
	}
}

// GetSendOK expone el canal a la UI
func GetSendOK() <-chan int {
	return SendOK
}

// StartSimulation inicia toda la simulación
func StartSimulation(interval time.Duration) error {
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

	jobs := make(chan models.DeviceData, 1000)
	results := make(chan *models.Message, 1000)

	SendOK = make(chan int, 1024)

	var prodWG sync.WaitGroup
	var workerWG sync.WaitGroup
	var pubWG sync.WaitGroup

	// 
	numDevices := GlobalDeviceCount

	workerCount := numDevices
	if workerCount < 4 {
		workerCount = 4
	}
	if workerCount > 500 {
		workerCount = 500
	}

	startWorkers(ctx, workerCount, jobs, results, &workerWG)
	go publisher(ctx, results, &pubWG)

	startDeviceProducers(ctx, numDevices, interval, jobs, &prodWG)
	go cleanupHandler(ctx, jobs, results, &prodWG, &workerWG, &pubWG)

	return nil
}

// startDeviceProducers crea los dispositivos
func startDeviceProducers(ctx context.Context, numDevices int, interval time.Duration, jobs chan<- models.DeviceData, prodWG *sync.WaitGroup) {
	for i := 2; i <= numDevices+1; i++ {
		deviceID := i
		userID := i
		go simulateDevice(ctx, deviceID, userID, interval, jobs, prodWG)
	}
}

// StopSimulation detiene todo
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

// IsSimulationActive indica si está activa
func IsSimulationActive() bool {
	simMu.Lock()
	defer simMu.Unlock()
	return simActive
}
