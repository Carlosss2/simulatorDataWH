package core

import (
	"math"
	"math/rand"
	"simulator/src/models"
)

// Genera datos simulados de sensores en base a un DeviceData recibido
func GenerateSensorData(device models.DeviceData) *models.Message {

	return &models.Message{
		DeviceId:    device.IdDevice,
		UserID:      device.IdUser,
		Bpm:         int(simulateMeasurement("HeartRate")),
		Spo2:        int(simulateMeasurement("Oxygen")),
		Bpm2:        int(simulateMeasurement("HeartRate")),
		Moving:      rand.Intn(2) == 1, // true o false aleatorio
		Temperature: float64(simulateMeasurement("Temperature")),
	}
}

// Simula una medición según el tipo de sensor
func simulateMeasurement(sensorType string) float64 {
	var value float64

	switch sensorType {
	case "Temperature":
		value = 36.0 + rand.Float64()*2.0 // 36–38 °C
	case "Oxygen":
		value = 90.0 + rand.Float64()*10.0 // 90–100 %
	case "HeartRate":
		value = 60.0 + rand.Float64()*40.0 // 60–100 bpm
	default:
		value = rand.Float64() * 100.0 // Valor genérico si no se conoce el tipo
	}

	// Redondear a dos decimales
	return math.Round(value*100) / 100
}

