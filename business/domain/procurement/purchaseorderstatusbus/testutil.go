package purchaseorderstatusbus

import (
	"context"
	"fmt"
)

// TestSeedPurchaseOrderStatuses seeds purchase order statuses for testing.
func TestSeedPurchaseOrderStatuses(ctx context.Context, n int, api *Business) ([]PurchaseOrderStatus, error) {
	statuses := make([]PurchaseOrderStatus, n)
	for i := 0; i < n; i++ {
		status, err := api.Create(ctx, NewPurchaseOrderStatus{
			Name:        fmt.Sprintf("Status%d", i),
			Description: fmt.Sprintf("Status%d Description", i),
			SortOrder:   (i + 1) * 100,
		})
		if err != nil {
			return nil, fmt.Errorf("creating purchase order status: %w", err)
		}

		statuses[i] = status
	}

	return statuses, nil
}