package domain

import (
	"context"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID       primitive.ObjectID
	UserId   string
	Status   string
	RideFare *RideFareModel
	Driver   *pb.TripDriver
}

func (t *TripModel) ToProto() *pb.Trip {
	return &pb.Trip{
		Id: t.ID.Hex(),
		SelectedFare: &pb.RideFare{
			Id:                t.RideFare.ID.Hex(),
			UserID:            t.RideFare.UserId,
			PackageSlug:       t.RideFare.PackageSlug,
			TotalPriceInCents: t.RideFare.TotalPriceInCents,
		},
		Driver: t.Driver,
	}
}

type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	SaveRideFare(ctx context.Context, fare *RideFareModel) error
	GetRideFareById(ctx context.Context, fareID string) (*RideFareModel, error)
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*types.OSRMApiResponse, error)
	EstimatePackagesPriceWithRoute(route *types.OSRMApiResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context, fares []*RideFareModel, userID string, route *types.OSRMApiResponse) ([]*RideFareModel, error)

	GetAndValidateFare(ctx context.Context, fareID string, userID string) (*RideFareModel, error)
}
