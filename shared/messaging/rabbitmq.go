package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"ride-sharing/shared/contracts"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange = "trips"
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
	if r.Channel != nil {
		r.Channel.Close()
	}

	if r.Conn != nil {
		r.Conn.Close()
	}
}

func (r *RabbitMQ) setupExchangesAndQueues() error {

	err := r.Channel.ExchangeDeclare(
		TripExchange, // name
		"topic",      // type
		false,        // durability
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare an exchange : %s, %v", TripExchange, err)
	}

	err = r.declareQueueAndBind(FindAvailableDriversQueue, TripExchange, []string{contracts.TripEventCreated, contracts.TripEventDriverNotInterested})

	if err != nil {
		return fmt.Errorf("failed to declare queue: %s and bind: %v", FindAvailableDriversQueue, err)
	}

	// q, err := r.Channel.QueueDeclare(
	// 	"hello", // name
	// 	false,   // durable
	// 	false,   // delete when unused
	// 	false,   // exclusive
	// 	false,   // no-wait
	// 	nil,     // arguments
	// )

	// if err != nil {
	// 	return fmt.Errorf("failed to declare a queue: %v", err)
	// }

	// err = r.Channel.QueueBind(
	// 	q.Name,               // queue name
	// 	"trip.event.created", // routing key
	// 	TripExchange,         // exchange
	// 	false,
	// 	nil,
	// )
	// if err != nil {
	// 	return fmt.Errorf("failed to bind queue: %v", err)
	// }

	return nil
}

func (r *RabbitMQ) declareQueueAndBind(queueName string, exchangeName string, MessageType []string) error {

	q, err := r.Channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare a queue: %v", err)
	}

	for _, msgType := range MessageType {
		err = r.Channel.QueueBind(
			q.Name,       // queue name
			msgType,      // routing key
			exchangeName, // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue: %v", err)
		}
	}

	return nil
}

type MessageHandler func(context.Context, amqp.Delivery) error

func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {

	// Qos (Quality of Service) setting. This is best practice for manual acknowledgments; it tells RabbitMQ not to overwhelm your worker with too many messages at once.

	// This tells RabbitMQ: "Don't give me more than 1 unacknowledged message at a time"
	// 1 - Prefetch Count
	// 0 - Prefetch Size (0 means no specific limit)
	// false - Global (false means this QoS setting applies to the current channel only)
	if err := r.Channel.Qos(1, 0, false); err != nil {
		return err
	}

	msgs, err := r.Channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack - we want to manually ack after processing
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	if err != nil {
		return err
	}

	ctx := context.Background()

	go func() {
		for msg := range msgs {

			if err := handler(ctx, msg); err != nil {
				log.Printf("failed to handle messages: %v; message: %s", err, msg.Body)

				// Nack (Negative Acknowledgment) What it says: "I have received this message, but I failed to process it. You can requeue it for another attempt or discard it."
				// msg.Nack(false, true) - The 'false' parameter means we're negatively acknowledging a single message (not multiple), and the 'false' parameter means we want to discard the message (not requeue it - Dead Letter Exchange). If you want to requeue it, set the second parameter to 'true'.

				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("failed to nack message: %v; message: %s", nackErr, msg.Body)
				}

				continue
			}

			// Ack (Acknowledgment) What it says: "I have received this message, I have processed it successfully, and you can now safely delete it from the queue."
			// msg.Ack(false) - The 'false' parameter means we're acknowledging a single message (not multiple)

			if ackErr := msg.Ack(false); ackErr != nil {
				log.Printf("failed to ack message: %v; message: %s", ackErr, msg.Body)
			}

		}
	}()

	return nil
}

func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, body contracts.AmqpMessage) error {

	log.Printf("Publishing message to exchange %s with routing key %s", TripExchange, routingKey)

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal message body: %v", err)
	}

	err = r.Channel.PublishWithContext(ctx,
		TripExchange, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         payload,
			DeliveryMode: amqp.Persistent,
		})

	if err != nil {
		return fmt.Errorf("failed to publish a message: %v", err)
	}

	return nil
}
