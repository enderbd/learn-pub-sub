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
		log.Fatalf("Client could not dial into RabbitMQL %v", err)
	}
	defer conn.Close()
	fmt.Println("Client connected to RabbitMQ ...")
	
	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Clound could not be welcomed %v", err)
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, userName)
	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, queueName, routing.PauseKey, pubsub.QueueTypeTransient )
	if err != nil {
		log.Fatalf("Could not declare and bind a queue %v", err)
	}

	gameState := gamelogic.NewGameState(userName)


	for  {
		userInput := gamelogic.GetInput()
		if len(userInput) < 1 {
			continue
		}
		switch userInput[0] {
		case "spawn":
			err = gameState.CommandSpawn(userInput)	
			if err != nil {
				log.Fatalf("Could not execute spawn command %v", err)
				continue
			}
		case "move":
			_, err = gameState.CommandMove(userInput)
			if err != nil {
				log.Fatalf("Client move failed %v", err)
			}
			fmt.Println("user move successful")
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("Unrecognized command %s\n", userInput[0])
		}

	}

}
