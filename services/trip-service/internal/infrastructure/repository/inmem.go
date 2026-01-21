package repository

import (
	"context"
	"fmt"
	"ride-sharing/services/trip-service/internal/domain"
	"sync"
)

type inMemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
	mutex     sync.RWMutex
}

func NewInMemRepository() *inMemRepository {
	return &inMemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inMemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.trips[trip.ID.Hex()] = trip

	return trip, nil
}

func (r *inMemRepository) SaveRideFare(ctx context.Context, fare *domain.RideFareModel) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.rideFares[fare.ID.Hex()] = fare

	return nil
}

func (r *inMemRepository) GetRideFareById(ctx context.Context, fareID string) (*domain.RideFareModel, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	fare, exits := r.rideFares[fareID]
	if !exits {
		return nil, fmt.Errorf("fare with id, %s does not exist", fareID)
	}
	return fare, nil
}
