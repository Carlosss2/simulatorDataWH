package core

import (
	"math"
	"math/rand"
	"simulator/src/models"
	"github.com/google/uuid"
)

// Crea una lectura de sensor genérica
func GenerateSensorData(sensorType string) *models.SensorData {
	return &models.SensorData{
		Id:          uuid.NewString(),
		Type:        sensorType,
		Measurement: simulateMeasurement(sensorType),
	}
}

// Simula una medición dependiendo del tipo de sensor
func simulateMeasurement(sensorType string) float32 {
	var value float32

	switch sensorType {
	case "Temperature":
		value = 36 + rand.Float32()*2 // 36–38 °C
	case "Oxygen":
		value = 90 + rand.Float32()*10 // 90–100 %
	case "HeartRate":
		value = 60 + rand.Float32()*40 // 60–100 bpm
	default:
		value = rand.Float32() * 100 // valores desconocidos
	}

	// Redondear a dos decimales
	return float32(math.Round(float64(value)*100) / 100)
}
