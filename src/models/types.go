package models

type SensorData struct {
	Id string `json:"sensor_id"`
	Type string `json:"type"`
	Measurement float32 `json:"measurement"`
}
