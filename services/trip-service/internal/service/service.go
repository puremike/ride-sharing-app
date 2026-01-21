package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	repo domain.TripRepository
}

func NewService(repo domain.TripRepository) *service {
	return &service{
		repo: repo,
	}
}

const QueryDefaultContext = 5 * time.Second

func (s *service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {

	ctx, cancel := context.WithTimeout(ctx, QueryDefaultContext)
	defer cancel()

	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserId:   fare.UserId,
		Status:   "pending",
		RideFare: fare,
		Driver:   &pb.TripDriver{},
	}

	return s.repo.CreateTrip(ctx, t)
}

func (s *service) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*types.OSRMApiResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, QueryDefaultContext)
	defer cancel()

	// baseURL := "https://osrm.selfmadeengineer.com"

	originalOSRMURL := "http://router.project-osrm.org"

	url := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson", originalOSRMURL, pickup.Longitude, pickup.Latitude, destination.Longitude, destination.Latitude)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get route from OSRM API: %v", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	routeResp := &types.OSRMApiResponse{}
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return routeResp, nil
}

func (s *service) EstimatePackagesPriceWithRoute(route *types.OSRMApiResponse) []*domain.RideFareModel {
	baseFares := getBaseFares()
	estimatedFares := make([]*domain.RideFareModel, len(baseFares))

	for i, f := range baseFares {
		estimatedFares[i] = EstimateFareRoute(f, route)
	}

	return estimatedFares
}

func EstimateFareRoute(fare *domain.RideFareModel, route *types.OSRMApiResponse) *domain.RideFareModel {
	pricingCfg := types.PricingConfig{}
	carPackagePrice := fare.TotalPriceInCents

	distanceInKm := route.Routes[0].Distance
	durationInMin := route.Routes[0].Duration

	distanceFare := pricingCfg.PricePerUnitOfDistance * distanceInKm
	durationFare := pricingCfg.PricePerMinute * durationInMin
	totalFare := distanceFare + durationFare + carPackagePrice

	return &domain.RideFareModel{
		PackageSlug:       fare.PackageSlug,
		TotalPriceInCents: totalFare,
	}
}

func (s *service) GenerateTripFares(ctx context.Context, fares []*domain.RideFareModel, userID string, route *types.OSRMApiResponse) ([]*domain.RideFareModel, error) {

	ctx, cancel := context.WithTimeout(ctx, QueryDefaultContext)
	defer cancel()

	rideFares := make([]*domain.RideFareModel, len(fares))

	for i, f := range fares {

		fare := &domain.RideFareModel{
			UserId:            userID,
			ID:                primitive.NewObjectID(),
			TotalPriceInCents: f.TotalPriceInCents,
			PackageSlug:       f.PackageSlug,
			Route:             route,
		}

		if err := s.repo.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save fare: %v", err)
		}

		rideFares[i] = fare

	}

	return rideFares, nil
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 200,
		},
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 350,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 400,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1000,
		},
	}
}

func (s *service) GetAndValidateFare(ctx context.Context, fareID string, userID string) (*domain.RideFareModel, error) {

	fare, err := s.repo.GetRideFareById(ctx, fareID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip fare: %v", err)
	}

	if fare == nil {
		return nil, fmt.Errorf("fare not found")
	}

	// User validation
	if fare.UserId != userID {
		return nil, fmt.Errorf("unauthorized: fare does not belong to user")
	}

	return nil, nil
}
