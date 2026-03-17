package main

import (
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
)

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (p *previewTripRequest) toProto() *pb.PreviewTripRequest {
	return &pb.PreviewTripRequest{
		UserID: p.UserID,
		Pickup: &pb.Coordinate{
			Latitude:  p.Pickup.Latitude,
			Longitude: p.Pickup.Longitude,
		},
		Destination: &pb.Coordinate{
			Latitude:  p.Destination.Latitude,
			Longitude: p.Destination.Longitude,
		},
	}
}

type createTripRequest struct {
	UserID     string `json:"userID"`
	RideFareID string `json:"rideFareID"`
}

func (c *createTripRequest) toProto() *pb.CreateTripRequest {
	return &pb.CreateTripRequest{
		UserID:     c.UserID,
		RideFareID: c.RideFareID,
	}
}
