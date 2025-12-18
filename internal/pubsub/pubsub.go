package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	QueueTypeDurable   SimpleQueueType = "durable"
	QueueTypeTransient SimpleQueueType = "transient"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body: data,
	}
	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	if err != nil {
		return err
	}

	return nil
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName,	key string, queueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	newChann, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	durable, autoDelete, exclusive := false, false, false
	switch queueType {
		case QueueTypeDurable:
			durable = true
		case QueueTypeTransient:
			autoDelete, exclusive = true, true
		
	}
	

	newQueue, err := newChann.QueueDeclare(queueName, durable, autoDelete, exclusive, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	} 

	err = newChann.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return newChann, newQueue, nil
}
