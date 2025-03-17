package warehousebus

import (
	"context"
	"strconv"

	"github.com/google/uuid"
)

func TestNewWarehouses(n int, createdBy uuid.UUID, streetIDs uuid.UUIDs) []NewWarehouse {
	newWarehouses := make([]NewWarehouse, n)

	for i := 0; i < n; i++ {
		newWarehouses[i] = NewWarehouse{
			StreetID:  streetIDs[i%len(streetIDs)],
			Name:      "Warehouse " + strconv.Itoa(i),
			CreatedBy: createdBy,
		}
	}
	return newWarehouses
}

func TestSeedWarehouses(ctx context.Context, n int, createdBy uuid.UUID, streetIDs uuid.UUIDs, api *Business) ([]Warehouse, error) {
	newWarehouses := TestNewWarehouses(n, createdBy, streetIDs)
	warehouses := make([]Warehouse, n)

	for i, newWarehouse := range newWarehouses {
		w, err := api.Create(ctx, newWarehouse)
		if err != nil {
			return []Warehouse{}, err
		}
		warehouses[i] = w
	}
	return warehouses, nil
}
