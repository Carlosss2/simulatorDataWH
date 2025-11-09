package core


import (
    "fmt"
    "log"
    "github.com/streadway/amqp"
	"github.com/joho/godotenv"
	"os"
)

func InitMqtt() *amqp.Connection{

	err := godotenv.Load()

	if err != nil {
		println("Las variables de entorno no se cargaron correctamente")
	}
    //Cargar credenciales
	USER := os.Getenv("USER_RABBIT");
	PASSWORD := os.Getenv("PASSWORD_RABBIT");
	HOST_RABBIT := os.Getenv("HOST_RABBIT");
	
	url := fmt.Sprintf("amqp://%s:%s@%s:5672/",USER, PASSWORD,HOST_RABBIT);

    conn, err := amqp.Dial(url);
    if err != nil {
        log.Fatal("No se pudo conectar:", err)
    }
    fmt.Println("Conectado exitosamente")
    return  conn
}
