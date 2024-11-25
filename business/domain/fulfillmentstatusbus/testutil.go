package fulfillmentstatusbus

import (
	"context"
	"fmt"
	"math/rand"
	"sort"

	"github.com/google/uuid"
)

// TestNewFulfillmentStatus is a helper method for testing.
func TestNewFulfillmentStatus(n int) []NewFulfillmentStatus {
	newFulfillmentStatus := make([]NewFulfillmentStatus, n)

	idx := rand.Intn(10000)
	for i := 0; i < n; i++ {
		idx++

		nas := NewFulfillmentStatus{
			IconID: uuid.New(),
			Name:   fmt.Sprintf("FulfillmentStatus%d", idx),
		}

		newFulfillmentStatus[i] = nas
	}

	return newFulfillmentStatus
}

// TestSeedFulfillmentStatus is a helper method for testing.
func TestSeedFulfillmentStatus(ctx context.Context, n int, api *Business) ([]FulfillmentStatus, error) {
	newFulfillmentStatuses := TestNewFulfillmentStatus(n)

	aprvlStatuses := make([]FulfillmentStatus, len(newFulfillmentStatuses))
	for i, nc := range newFulfillmentStatuses {
		as, err := api.Create(ctx, nc)
		if err != nil {
			return nil, fmt.Errorf("seeding city: idx: %d : %w", i, err)
		}

		aprvlStatuses[i] = as
	}

	// sort cities by name
	sort.Slice(aprvlStatuses, func(i, j int) bool {
		return aprvlStatuses[i].Name <= aprvlStatuses[j].Name
	})

	return aprvlStatuses, nil
}
