package main

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func route() http.Handler {

	g := gin.Default()

	tripPreviewCorsConfig := cors.Config{
		AllowOrigins: []string{"http://localhost:3000", "*"},
		AllowMethods: []string{"PUT", "GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:       12 * time.Hour,
	}

	g.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  http.StatusOK,
		})
	})

	time.Sleep(time.Second * 7)

	// g.POST("/trip/preview", middleware.EnableCors(), handleTripPreview)
	g.POST("/trip/preview", cors.New(tripPreviewCorsConfig), handleTripPreview)
	g.GET("/ws/drivers", handleDriversWebSocket)
	g.GET("/ws/riders", handleRidersWebSocket)

	return g
}
