package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"ride-sharing/shared/env"
	"ride-sharing/shared/messaging"
	"syscall"
	"time"

	grpcserver "google.golang.org/grpc"
)

var (
	grpcAddr    = ":9092"
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		sig := <-sigChan
		log.Printf("Received shutdown signal: %s. Shutting down....\n", sig.String())
		cancel()
	}()

	listen, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Printf("failed to listen on %s: %v\n", grpcAddr, err)
		return
	}

	instanceID := time.Now().UnixNano()
	svc := NewService(instanceID)

	// In your log
	log.Printf("[%d] Consumer check: Found 0 matches out of 0...", instanceID)

	// RabbitMQ Connection
	rmqConn, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rmqConn.Close()
	log.Print("Connected to RabbitMQ")

	consumer := NewTripConsumer(rmqConn, svc)

	go func() {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("Failed to listen to messages: %v", err)
			return
		}
	}()

	gRPCServer := grpcserver.NewServer()

	NewGRPCHandler(gRPCServer, svc)

	go func() {
		if err := gRPCServer.Serve(listen); err != nil {
			log.Printf("Failed to serve on %s: %v\n", listen.Addr().String(), err)
		}
		cancel()
	}()

	// graceful shutdown
	<-ctx.Done()
	log.Print("Shutting down gRPC Server")
	gRPCServer.GracefulStop()
	log.Print("gRPC server stopped")
}
