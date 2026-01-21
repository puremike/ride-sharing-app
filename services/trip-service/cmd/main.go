package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"ride-sharing/services/trip-service/internal/infrastructure/grpc"
	"ride-sharing/services/trip-service/internal/infrastructure/repository"
	"ride-sharing/services/trip-service/internal/service"
	"ride-sharing/shared/env"

	grpcserver "google.golang.org/grpc"
)

var (
	// httpAddr = env.GetString("TRIP_HTTP_ADDR", ":8083")
	grpcAddr = env.GetString("TRIP_GRPC_ADDR", ":9093")
)

func main() {

	inMem := repository.NewInMemRepository()
	svc := service.NewService(inMem)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// go routine to listen for interrupt signal
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		sig := <-sigChan
		log.Printf("Received signal: %v. Shutting down...\n", sig.String())
		cancel()
	}()

	// a listener to annouce local tcp address
	listen, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Printf("Failed to listen on %s: %v\n", grpcAddr, err)
		return
	}

	// Starting the GRPC Server
	grpcServer := grpcserver.NewServer()
	grpc.NewGRPCHandler(grpcServer, svc)

	go func() {
		// start the gRPC server
		if err := grpcServer.Serve(listen); err != nil {
			log.Printf("Failed to serve on %s: %v\n", listen.Addr().String(), err)
		}
		cancel()
	}()

	// graceful shutdown
	<-ctx.Done()
	log.Println("Shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped.")

	// mux := route()
	// server := &http.Server{
	// 	Addr:    httpAddr,
	// 	Handler: mux,
	// }

	// if err := server.ListenAndServe(); err != nil {
	// 	log.Fatal(err)
	// }

	// ctx := context.Background()

	// fare := &domain.RideFareModel{
	// 	UserId:            "123",
	// 	PackageSlug:       "standard",
	// 	TotalPriceInCents: 100,
	// }
	// t, err := svc.CreateTrip(ctx, fare)
	// if err != nil {
	// 	log.Println(err)
	// }

	// log.Println(t)

	// for {
	// 	time.Sleep(time.Second)
	// }
}

// func route() http.Handler {

// 	g := gin.Default()
// 	v1 := g.Group("/v1")

// 	inMem := repository.NewInMemRepository()
// 	svc := service.NewService(inMem)

// 	handler := handlers.NewHttpHander(svc)

// 	v1.POST("/preview", handler.HandleTripPreview)

// 	return g
// }
