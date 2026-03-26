package cyclecountitembus

import (
	"context"
	"sort"

	"github.com/google/uuid"
)

// TestNewCycleCountItems generates n new cycle count items for testing.
func TestNewCycleCountItems(n int, sessionIDs, productIDs, locationIDs []uuid.UUID) []NewCycleCountItem {
	items := make([]NewCycleCountItem, n)

	for i := range n {
		items[i] = NewCycleCountItem{
			SessionID:      sessionIDs[i%len(sessionIDs)],
			ProductID:      productIDs[i%len(productIDs)],
			LocationID:     locationIDs[i%len(locationIDs)],
			SystemQuantity: (i + 1) * 10,
		}
	}

	return items
}

// TestSeedCycleCountItems creates n cycle count items in the database for testing.
func TestSeedCycleCountItems(ctx context.Context, n int, sessionIDs, productIDs, locationIDs []uuid.UUID, api *Business) ([]CycleCountItem, error) {
	newItems := TestNewCycleCountItems(n, sessionIDs, productIDs, locationIDs)

	items := make([]CycleCountItem, len(newItems))
	for i, nci := range newItems {
		item, err := api.Create(ctx, nci)
		if err != nil {
			return nil, err
		}
		items[i] = item
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ID.String() < items[j].ID.String()
	})

	return items, nil
}
