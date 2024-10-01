package citydb

import (
	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/location/citybus"
)

type city struct {
	ID       uuid.UUID `db:"city_id"`
	RegionID uuid.UUID `db:"region_id"`
	Name     string    `db:"name"`
}

func toDBCity(bus citybus.City) city {
	return city{
		ID:       bus.ID,
		RegionID: bus.RegionID,
		Name:     bus.Name,
	}
}

func toBusCity(dbCity city) citybus.City {
	return citybus.City{
		ID:       dbCity.ID,
		RegionID: dbCity.RegionID,
		Name:     dbCity.Name,
	}
}

func toBusCities(dbCities []city) []citybus.City {
	cities := make([]citybus.City, len(dbCities))
	for i, cty := range dbCities {
		cities[i] = toBusCity(cty)
	}
	return cities
}
