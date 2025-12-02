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

// TestNewOrdersHistorical creates orders distributed across a time range for seeding.
// daysBack specifies how many days of history to generate (e.g., 30, 90, 365).
// Orders are evenly distributed across the time range.
func TestNewOrdersHistorical(n int, daysBack int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs) []NewOrder {
	orders := make([]NewOrder, 0, n)
	now := time.Now()

	for i := 0; i < n; i++ {
		// Distribute evenly across the time range
		daysAgo := (i * daysBack) / n
		createdDate := now.AddDate(0, 0, -daysAgo)

		orders = append(orders, NewOrder{
			Number:              fmt.Sprintf("SEED-%d", i+1),
			CustomerID:          customerIDs[i%len(customerIDs)],
			DueDate:             createdDate.AddDate(0, 0, 7), // Due 7 days after creation
			FulfillmentStatusID: ofIDs[i%len(ofIDs)],
			CreatedBy:           userIDs[i%len(userIDs)],
			CreatedDate:         &createdDate, // Explicit historical date
		})
	}
	return orders
}

// TestSeedOrdersHistorical seeds orders with historical date distribution.
func TestSeedOrdersHistorical(ctx context.Context, n int, daysBack int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs, api *Business) ([]Order, error) {
	newOrders := TestNewOrdersHistorical(n, daysBack, userIDs, customerIDs, ofIDs)
	orders := make([]Order, len(newOrders))
	for i, no := range newOrders {
		order, err := api.Create(ctx, no)
		if err != nil {
			return []Order{}, err
		}
		orders[i] = order
	}
	return orders, nil
}
