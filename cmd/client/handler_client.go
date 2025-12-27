package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)



func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype{
	return func(state routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(state)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype{
		defer fmt.Print("> ")
		mvOutcome := gs.HandleMove(move)
		switch mvOutcome {
		case gamelogic.MoveOutComeSafe :
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			key := fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername())
			err := pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				key,
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
				)
			if err != nil {
				fmt.Println("Could not publish the MoveOutcomeMakeWar")
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
		
	}
}

func handlerWar(gs *gamelogic.GameState, publishCh *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype{
		defer fmt.Print("> ")
		warOutcome, winner, loser := gs.HandleWar(rw)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			msg := fmt.Sprintf("%s won a was agains %s", winner, loser)
			err := publishGameLog(publishCh, gs.GetUsername(), msg)
			if err != nil {
				fmt.Printf("error %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			msg := fmt.Sprintf("%s won a was agains %s", winner, loser)
			err := publishGameLog(publishCh, gs.GetUsername(), msg)
			if err != nil {
				fmt.Printf("error %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			msg := fmt.Sprintf("%s had a draw against %s", winner, loser)
			err := publishGameLog(publishCh, gs.GetUsername(), msg)
			if err != nil {
				fmt.Printf("error %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("No suitable war outcome!")
			return pubsub.NackDiscard
		}
		
	}
}

func publishGameLog(publishCh *amqp.Channel, username, msg string) error {
	return pubsub.PublishGob(
		publishCh,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),
			Message:     msg,
		},
	)
}
