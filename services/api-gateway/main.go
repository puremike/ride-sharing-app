package main

import (
	"log"
	"ride-sharing/shared/env"
)

var (
	httpAddr   = env.GetString("HTTP_ADDR", ":8081")
	APIVersion = "v1"
)

func main() {

	mux := route()
	if err := server(mux); err != nil {
		log.Fatal(err)
	}
}
