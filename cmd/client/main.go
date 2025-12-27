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
	
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}
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
	err = pubsub.SubscribeJSON(
		conn,routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gameState.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.QueueTypeTransient,
		handlerMove(gameState, publishCh),
	)
	if err != nil {
		log.Fatalf("could not subscribe to army moves: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.QueueTypeTransient,
		handlerPause(gameState),
		)
	if err != nil {
		log.Fatalf("Could not subscribe %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"war",
		routing.WarRecognitionsPrefix + ".*",
		pubsub.QueueTypeDurable,
		handlerWar(gameState, publishCh),
		)

	if err != nil {
		log.Fatalf("Could not subscribe the hanleWar %v", err)
	}

	for  {
		userInput := gamelogic.GetInput()
		if len(userInput) < 1 {
			continue
		}
		switch userInput[0] {
		case "spawn":
			err = gameState.CommandSpawn(userInput)	
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			mv, err := gameState.CommandMove(userInput)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+mv.Player.Username,
				mv,
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)

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
