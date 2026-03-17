package messaging

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {

	// establish connection
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %v", err)
	}

	// create a channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	rmq := &RabbitMQ{Conn: conn, Channel: ch}

	// create exhanges and queues
	if err := rmq.setupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("error setting up exchanges and queues: %v", err)
	}

	return rmq, nil
}

func (r *RabbitMQ) Close() {
	if r.Conn != nil {
		r.Conn.Close()
	}

	if r.Channel != nil {
		r.Channel.Close()
	}
}

func (r *RabbitMQ) setupExchangesAndQueues() error {

	_, err := r.Channel.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, body string) error {

	err := r.Channel.PublishWithContext(ctx,
		"",         // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(body),
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}

	return nil
}
