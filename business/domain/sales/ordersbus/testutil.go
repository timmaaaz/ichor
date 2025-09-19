package ordersbus

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func TestNewOrders(n int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs) []NewOrder {
	// Use actual Orders
	orders := make([]NewOrder, 0, n)
	for i := 0; i < n; i++ {
		orders = append(orders, NewOrder{
			Number:              fmt.Sprintf("TST-%d", i+1),
			CustomerID:          customerIDs[i%len(customerIDs)],
			DueDate:             time.Now().AddDate(0, 0, i+1),
			FulfillmentStatusID: ofIDs[i%len(ofIDs)],
			CreatedBy:           userIDs[i%len(userIDs)],
		})
	}
	return orders
}

func TestSeedOrders(ctx context.Context, n int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs, api *Business) ([]Order, error) {
	newOrders := TestNewOrders(n, userIDs, customerIDs, ofIDs)
	orders := make([]Order, len(newOrders))
	for i, ns := range newOrders {
		s, err := api.Create(ctx, ns)
		if err != nil {
			return []Order{}, err
		}
		orders[i] = s
	}
	return orders, nil
}
