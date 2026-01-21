package handlers

import (
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/types"

	"github.com/gin-gonic/gin"
)

type httpHander struct {
	svc domain.TripService
}

func NewHttpHander(svc domain.TripService) *httpHander {
	return &httpHander{
		svc: svc,
	}
}

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (s *httpHander) HandleTripPreview(c *gin.Context) {
	var payload previewTripRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	route, err := s.svc.GetRoute(c.Request.Context(), &payload.Pickup, &payload.Destination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, route)
}
