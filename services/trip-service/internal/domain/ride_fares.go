package domain

import (
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserId            string
	PackageSlug       string // ex: van, sedan, suv
	TotalPriceInCents float64
	Route             *types.OSRMApiResponse
}

func (r *RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
		Id:                r.ID.Hex(),
		UserID:            r.UserId,
	}
}

func ToRideFareProtoSlice(fares []*RideFareModel) []*pb.RideFare {

	var protoFares []*pb.RideFare

	for _, f := range fares {
		protoFares = append(protoFares, f.ToProto())
	}

	return protoFares
}
