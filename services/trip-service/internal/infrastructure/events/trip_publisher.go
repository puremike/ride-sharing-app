package events

import (
	"context"
	"encoding/json"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/contracts"
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

func (r *TripEventPublisher) PublishTripCreatedEvent(ctx context.Context, trip *domain.TripModel) error {

	payload := messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.rabbitmq.PublishMessage(ctx, contracts.TripEventCreated, contracts.AmqpMessage{
		OwnerID: trip.UserId,
		Data:    payloadJSON,
	})
}
