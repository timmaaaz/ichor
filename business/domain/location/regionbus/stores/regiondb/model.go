package regiondb

import (
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/regionbus"
	"github.com/google/uuid"
)

type region struct {
	ID        uuid.UUID `db:"region_id"`
	CountryID uuid.UUID `db:"country_id"`
	Name      string    `db:"name"`
	Code      string    `db:"code"`
}

func toBusRegion(dbRegion region) regionbus.Region {
	return regionbus.Region{
		ID:        dbRegion.ID,
		CountryID: dbRegion.CountryID,
		Name:      dbRegion.Name,
		Code:      dbRegion.Code,
	}
}

func toBusRegions(dbRegions []region) []regionbus.Region {
	regions := make([]regionbus.Region, len(dbRegions))
	for i, reg := range dbRegions {
		regions[i] = toBusRegion(reg)
	}
	return regions
}
