package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

var (
	grpcAddr = ":9092"
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
		log.Printf("Failed to listen on %s: %v\n", grpcAddr, err)
		return
	}

	gRPCServer := grpcserver.NewServer()
	svc := NewService()
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
