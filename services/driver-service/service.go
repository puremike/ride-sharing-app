package main

import (
	"log"
	math "math/rand/v2"
	pb "ride-sharing/shared/proto/driver"
	"ride-sharing/shared/util"
	"sync"

	"github.com/mmcloughlin/geohash"
)

type driverInMap struct {
	Driver *pb.Driver
}

type Service struct {
	drivers    []*driverInMap
	mu         sync.RWMutex
	instanceID int64
}

func NewService(instanceID int64) *Service {
	return &Service{
		instanceID: instanceID,
		drivers:    make([]*driverInMap, 0),
	}
}

func (s *Service) RegisterDriver(driverId string, packageSlug string) (*pb.Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	randomIndex := math.IntN(len(PredefinedRoutes))
	randomRoute := PredefinedRoutes[randomIndex]

	randomPlate := GenerateRandomPlate()
	randomAvatar := util.GetRandomAvatar(randomIndex)

	// we can ignore this property for now, but it must be sent to the frontend.
	geohash := geohash.Encode(randomRoute[0][0], randomRoute[0][1])

	driver := &pb.Driver{
		Geohash:        geohash,
		Location:       &pb.Location{Latitude: randomRoute[0][0], Longitude: randomRoute[0][1]},
		Name:           "Lando Norris",
		PackageSlug:    packageSlug,
		ProfilePicture: randomAvatar,
		CarPlate:       randomPlate,
		Id:             driverId,
	}

	s.drivers = append(s.drivers, &driverInMap{
		Driver: driver,
	})
	log.Printf("DRIVER REGISTERED: %s (%s). Current Registry Size: %d", driverId, packageSlug, len(s.drivers))

	return driver, nil
}

// func (s *Service) UnregisterDriver(driverId string) {
// 	s.mu.Lock()
// 	defer s.mu.Unlock()

// 	for i, driver := range s.drivers {
// 		if driver.Driver.Id == driverId {
// 			s.drivers = append(s.drivers[:i], s.drivers[i+1:]...)
// 		}
// 	}
// }

func (s *Service) UnregisterDriver(driverId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < len(s.drivers); i++ {
		if s.drivers[i].Driver.Id == driverId {
			s.drivers = append(s.drivers[:i], s.drivers[i+1:]...)
			log.Printf("DRIVER REMOVED: %s. Current Registry Size: %d", driverId, len(s.drivers))
			return
		}
	}
}

func (s *Service) FindAvailableDrivers(packageSlug string) ([]string, int) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	var matchingDrivers []string

	for _, driver := range s.drivers {
		if driver.Driver.PackageSlug == packageSlug {
			matchingDrivers = append(matchingDrivers, driver.Driver.Id)
		}
	}

	if len(matchingDrivers) == 0 {
		return []string{}, 0
	}

	return matchingDrivers, len(matchingDrivers)
}
