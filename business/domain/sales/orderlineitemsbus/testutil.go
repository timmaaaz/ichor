package orderlineitemsbus

import (
	"context"
	"math/rand"
	"time"

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

// TestNewOrderLineItemsHistorical creates order line items distributed across time based on order dates.
// Line items are created 0-2 hours after their corresponding order.
func TestNewOrderLineItemsHistorical(n int, orderDates map[uuid.UUID]time.Time, orderIDs uuid.UUIDs, productIDs uuid.UUIDs, lineItemStatusIDs uuid.UUIDs, userIDs uuid.UUIDs) []NewOrderLineItem {
	items := make([]NewOrderLineItem, 0, n)

	for i := 0; i < n; i++ {
		orderID := orderIDs[i%len(orderIDs)]
		orderDate := orderDates[orderID]
		// Line item created 0-2 hours after order
		lineItemDate := orderDate.Add(time.Duration(rand.Intn(120)) * time.Minute)

		items = append(items, NewOrderLineItem{
			OrderID:                       orderID,
			ProductID:                     productIDs[i%len(productIDs)],
			Quantity:                      (i % 10) + 1,
			Discount:                      0.0,
			LineItemFulfillmentStatusesID: lineItemStatusIDs[i%len(lineItemStatusIDs)],
			CreatedBy:                     userIDs[i%len(userIDs)],
			CreatedDate:                   &lineItemDate,
		})
	}
	return items
}

// TestSeedOrderLineItemsHistorical seeds order line items with historical dates based on order dates.
func TestSeedOrderLineItemsHistorical(ctx context.Context, n int, orderDates map[uuid.UUID]time.Time, orderIDs uuid.UUIDs, productIDs uuid.UUIDs, lineItemStatusIDs uuid.UUIDs, userIDs uuid.UUIDs, api *Business) ([]OrderLineItem, error) {
	newItems := TestNewOrderLineItemsHistorical(n, orderDates, orderIDs, productIDs, lineItemStatusIDs, userIDs)
	items := make([]OrderLineItem, len(newItems))
	for i, ni := range newItems {
		item, err := api.Create(ctx, ni)
		if err != nil {
			return []OrderLineItem{}, err
		}
		items[i] = item
	}
	return items, nil
}
