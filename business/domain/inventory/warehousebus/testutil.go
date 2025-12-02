package warehousebus

import (
	"context"
	"strconv"
	"time"

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

// TestNewWarehousesHistorical creates warehouses distributed far back in time (1+ years).
// Warehouses are infrastructure that should exist long before operations start.
func TestNewWarehousesHistorical(n int, daysBack int, createdBy uuid.UUID, streetIDs uuid.UUIDs) []NewWarehouse {
	newWarehouses := make([]NewWarehouse, n)
	now := time.Now()

	for i := 0; i < n; i++ {
		// Distribute warehouses far back in time (they were built long ago)
		daysAgo := (i * daysBack) / n
		createdDate := now.AddDate(0, 0, -daysAgo)

		nw := NewWarehouse{
			StreetID:    streetIDs[i%len(streetIDs)],
			Name:        "Warehouse " + strconv.Itoa(i),
			CreatedBy:   createdBy,
			CreatedDate: &createdDate,
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

// TestSeedWarehousesHistorical seeds warehouses with historical date distribution.
func TestSeedWarehousesHistorical(ctx context.Context, n int, daysBack int, createdBy uuid.UUID, streetIDs uuid.UUIDs, api *Business) ([]Warehouse, error) {
	newWarehouses := TestNewWarehousesHistorical(n, daysBack, createdBy, streetIDs)
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
