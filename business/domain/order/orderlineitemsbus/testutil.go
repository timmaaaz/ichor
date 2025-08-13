package orderlineitemsbus

import (
	"context"

	"github.com/google/uuid"
)

func TestNewOrderLineItems(n int, orderIDs uuid.UUIDs, productIDs uuid.UUIDs, lineItemStatusIDs uuid.UUIDs, userIDs uuid.UUIDs) []NewOrderLineItem {
	// Use actual order line itemes
	statuses := make([]NewOrderLineItem, 0, n)
	for i := 0; i < n; i++ {
		statuses = append(statuses, NewOrderLineItem{
			OrderID:                       orderIDs[i%len(orderIDs)],
			ProductID:                     productIDs[i%len(productIDs)],
			Quantity:                      1,
			LineItemFulfillmentStatusesID: lineItemStatusIDs[i%len(lineItemStatusIDs)],
			CreatedBy:                     userIDs[i%len(userIDs)],
		})
	}
	return statuses
}

func TestSeedOrderLineItems(ctx context.Context, n int, orderIDs uuid.UUIDs, productIDs uuid.UUIDs, lineItemStatusIDs uuid.UUIDs, userIDs uuid.UUIDs, api *Business) ([]OrderLineItem, error) {
	newStatuses := TestNewOrderLineItems(n, orderIDs, productIDs, lineItemStatusIDs, userIDs)
	statuses := make([]OrderLineItem, len(newStatuses))
	for i, ns := range newStatuses {
		s, err := api.Create(ctx, ns)
		if err != nil {
			return []OrderLineItem{}, err
		}
		statuses[i] = s
	}
	return statuses, nil
}
