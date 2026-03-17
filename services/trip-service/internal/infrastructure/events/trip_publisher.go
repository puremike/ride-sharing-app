package events

import (
	"context"
	"ride-sharing/shared/messaging"
)

type TripEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventPublisher(rabbitmq *messaging.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (r *TripEventPublisher) PublishTripCreatedEvent(ctx context.Context) error {
	if err := r.rabbitmq.PublishMessage(ctx, "hello", "Trip Created Event"); err != nil {
		return err
	}

	return nil
}
