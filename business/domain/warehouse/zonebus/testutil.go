package zonebus

import (
	"context"
	"strconv"

	"github.com/google/uuid"
)

func TestNewZone(n int, createdBy uuid.UUID, warehouseID uuid.UUID) []NewZone {
	newZones := make([]NewZone, n)

	for i := 0; i < n; i++ {
		newZones[i] = NewZone{
			WarehouseID: warehouseID,
			Name:        "Zone " + strconv.Itoa(i),
			Description: "Description " + strconv.Itoa(i),
			CreatedBy:   createdBy,
		}
	}
	return newZones
}

func TestSeedZones(ctx context.Context, n int, createdBy uuid.UUID, warehouseID uuid.UUID, api *Business) ([]Zone, error) {
	newZones := TestNewZone(n, createdBy, warehouseID)
	zones := make([]Zone, n)

	for i, newZone := range newZones {
		z, err := api.Create(ctx, newZone)
		if err != nil {
			return []Zone{}, err
		}
		zones[i] = z
	}
	return zones, nil
}
