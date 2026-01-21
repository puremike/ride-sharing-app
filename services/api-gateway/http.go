package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"ride-sharing/services/api-gateway/grpc_clients"
	"ride-sharing/shared/contracts"

	"github.com/gin-gonic/gin"
)

// func handleTripPreview(c *gin.Context)

func handleTripPreview(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payload previewTripRequest

	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		log.Printf("failed to unmarshal request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// basic validation
	if payload.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	tripService, err := grpc_clients.NewTripServiceClient()
	if err != nil {
		log.Printf("failed to create trip-service client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tripService.Close()

	tripPreview, err := tripService.Client.PreviewTrip(c.Request.Context(), payload.toProto())

	if err != nil {
		log.Printf("failed to preview trip: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"failed to preview trip": err.Error()})
		return
	}

	// url := fmt.Sprintf("http://trip-service:8083/%s/preview", APIVersion)

	// resp, err := http.Post(url, "application/json", bytes.NewReader(bodyBytes))
	// if err != nil {
	// 	log.Printf("failed to make request to trip-service: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// defer resp.Body.Close()

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	log.Printf("failed to read response body: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	// var parsedBody any
	// if err := json.Unmarshal(body, &parsedBody); err != nil {
	// 	log.Printf("failed to unmarshal response body: %v", err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }

	response := contracts.APIResponse{Data: tripPreview}

	c.JSON(http.StatusOK, response)
}
