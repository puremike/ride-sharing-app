package main

import (
	"context"
	"encoding/json"
	"log"
	"ride-sharing/shared/contracts"
	"ride-sharing/shared/messaging"

	"github.com/rabbitmq/amqp091-go"
)

type tripConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *Service
}

func NewTripConsumer(rmq *messaging.RabbitMQ, service *Service) *tripConsumer {
	return &tripConsumer{rabbitmq: rmq, service: service}
}

func (c *tripConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.FindAvailableDriversQueue, func(ctx context.Context, msg amqp091.Delivery) error {

		var tripEventData contracts.AmqpMessage
		if err := json.Unmarshal(msg.Body, &tripEventData); err != nil {
			log.Printf("Failed to unmarshal message body: %v", err)
			return err
		}

		var payload messaging.TripEventData
		if err := json.Unmarshal(tripEventData.Data, &payload); err != nil {
			log.Printf("Failed to unmarshal trip event data: %v", err)
			return err
		}

		// log.Printf("Driver received a message: %+v", payload)

		switch msg.RoutingKey {
		case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
			return c.findAndNotifyAvailableDrivers(ctx, payload)
		}

		return nil
	})

}

func (c *tripConsumer) findAndNotifyAvailableDrivers(ctx context.Context, payload messaging.TripEventData) error {
	suitableIDs, totalInRegistry := c.service.FindAvailableDrivers(payload.Trip.SelectedFare.PackageSlug)

	log.Printf("Registry size: %d", len(suitableIDs))
	log.Printf("Consumer check: Found %d matches out of %d total drivers in registry", len(suitableIDs), totalInRegistry)

	// log.Printf("packageSlug: %s", payload.Trip.SelectedFare.PackageSlug)

	// notify that no suitable drivers found for the trip
	if len(suitableIDs) == 0 {
		log.Print("No suitable drivers found for trip")

		if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{OwnerID: payload.Trip.UserID}); err != nil {
			log.Printf("Failed to publish no drivers found message to exchange: %v", err)
			return err
		}
		return nil
	}

	marshalledEvent, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	suitableDriverID := suitableIDs[0]

	// notify the most suitable driver about the trip request
	if err := c.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
		OwnerID: suitableDriverID,
		Data:    marshalledEvent,
	}); err != nil {
		log.Printf("Failed to publish trip request message to exchange: %v", err)
		return err
	}

	return nil
}
