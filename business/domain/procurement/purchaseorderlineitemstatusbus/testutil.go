package purchaseorderlineitemstatusbus

import (
	"context"
	"fmt"
)

// TestSeedPurchaseOrderLineItemStatuses seeds purchase order line item statuses for testing.
func TestSeedPurchaseOrderLineItemStatuses(ctx context.Context, n int, api *Business) ([]PurchaseOrderLineItemStatus, error) {
	statuses := make([]PurchaseOrderLineItemStatus, n)
	for i := 0; i < n; i++ {
		status, err := api.Create(ctx, NewPurchaseOrderLineItemStatus{
			Name:        fmt.Sprintf("LineItemStatus%d", i),
			Description: fmt.Sprintf("LineItemStatus%d Description", i),
			SortOrder:   (i + 1) * 100,
		})
		if err != nil {
			return nil, fmt.Errorf("creating purchase order line item status: %w", err)
		}

		statuses[i] = status
	}

	return statuses, nil
}
