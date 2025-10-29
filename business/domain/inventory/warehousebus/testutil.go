package warehousebus

import (
	"context"
	"strconv"

	"github.com/google/uuid"
)

func TestNewWarehouses(n int, createdBy uuid.UUID, streetIDs uuid.UUIDs) []NewWarehouse {
	newWarehouses := make([]NewWarehouse, n)

	for i := 0; i < n; i++ {
		nw := NewWarehouse{
			StreetID:  streetIDs[i%len(streetIDs)],
			Name:      "Warehouse " + strconv.Itoa(i),
			CreatedBy: createdBy,
		}

		// Even-indexed warehouses get manual codes to test preservation
		// Odd-indexed warehouses get empty codes to test auto-generation
		if i%2 == 0 {
			nw.Code = "WH-TEST-" + strconv.Itoa(i/2+1)
		}

		newWarehouses[i] = nw
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
