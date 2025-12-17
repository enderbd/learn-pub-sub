package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitConnStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnStr)
	if err != nil {
		log.Fatalf("could not dial into RabbitMQL %v", err)
	}
	defer conn.Close()
	fmt.Println("Connected to RabbitMQ ...")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Shutting down")
}
