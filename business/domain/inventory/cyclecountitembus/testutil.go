package cyclecountitembus

import (
	"context"
	"fmt"
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
// If assigneeIDs is non-empty, each item is round-robin-assigned to a user via
// Business.Update after creation (setting CountedBy). Passing nil preserves the
// original unassigned behavior.
func TestSeedCycleCountItems(ctx context.Context, n int, sessionIDs, productIDs, locationIDs, assigneeIDs []uuid.UUID, api *Business) ([]CycleCountItem, error) {
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

	if len(assigneeIDs) > 0 {
		for i := range items {
			assignee := assigneeIDs[i%len(assigneeIDs)]
			updated, err := api.Update(ctx, items[i], UpdateCycleCountItem{CountedBy: &assignee})
			if err != nil {
				return nil, fmt.Errorf("assign cycle count item %d: %w", i, err)
			}
			items[i] = updated
		}
	}

	return items, nil
}
