package core

import (
	"fmt"
	"log"
	"os"
	"encoding/json"
	"simulator/src/models"

	"github.com/joho/godotenv"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client

// Conexión general al broker MQTT
func ConnectMqtt() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No se pudieron cargar las variables de entorno")
	}

	USER := os.Getenv("USER_RABBIT")
	PASSWORD := os.Getenv("PASSWORD_RABBIT")
	HOST_RABBIT := os.Getenv("HOST_RABBIT")
	CLIENT_ID := os.Getenv("CLIENT_ID")

	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://" + HOST_RABBIT)
	opts.SetClientID(CLIENT_ID)
	opts.SetUsername(USER)
	opts.SetPassword(PASSWORD)

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	fmt.Println("Conectado correctamente al broker MQTT")
}

// Suscripción al tópico device.data
func SubscribeToDeviceData() {
	TOPIC := os.Getenv("TOPICCON")

	if token := client.Subscribe(TOPIC, 1, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Mensaje recibido en %s: %s\n", msg.Topic(), string(msg.Payload()))

		var device models.DeviceData
		if err := json.Unmarshal(msg.Payload(), &device); err != nil {
			fmt.Println("Error al parsear el mensaje:", err)
			return
		}

		simulated := GenerateSensorData(device)
		dataJSON, err := json.Marshal(simulated)
		if err != nil {
			fmt.Println("Error al convertir datos simulados a JSON:", err)
			return
		}

		PublishData(string(dataJSON))
	}); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	} else {
		fmt.Println("Suscrito al tópico:", TOPIC)
	}
}

// Publicar mensajes al tópico esp32.datos
func PublishData(message string) {
	TOPIC := os.Getenv("TOPICPUB")

	if token := client.Publish(TOPIC, 0, false, message); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	} else {
		fmt.Println("Mensaje publicado en", TOPIC, ":", message)
	}
}
