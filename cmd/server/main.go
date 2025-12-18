package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitConnStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnStr)
	if err != nil {
		log.Fatalf("Server could not dial into RabbitMQL %v", err)
	}
	defer conn.Close()
	fmt.Println("Server connected to RabbitMQ ...")

	pauseResumeChannel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Server could not open a channel: %v", err)
	}


	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, "game_logs.*", pubsub.QueueTypeDurable )
	if err != nil {
		log.Fatalf("Server: could not declare and bind a queue %v", err)
	}

	gamelogic.PrintServerHelp()
	run := true
	for run {
		userInput := gamelogic.GetInput()
		if len(userInput) < 1 {
			continue
		}
		switch userInput[0] {
		case "pause":
			err = pubsub.PublishJSON(
				pauseResumeChannel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				})
			if err != nil {
				log.Fatalf("Could not send the stop message %v", err)
			}
			fmt.Println("Pause message sent!")
		case "resume":
			err = pubsub.PublishJSON(
				pauseResumeChannel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				})
			if err != nil {
				log.Fatalf("Could not send the resume message %v", err)
			}
			fmt.Println("Resume message sent!")
		case "quit":
			fmt.Println("Exiting the server!")
			run = false
		default:
			fmt.Printf("Unrecognized command %s\n", userInput[0])
		}

	}
}

