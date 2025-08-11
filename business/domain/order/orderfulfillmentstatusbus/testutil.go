package orderfulfillmentstatusbus

import (
	"context"
	"fmt"
)

func TestNewOrderFulfillmentStatuses(n int) []NewOrderFulfillmentStatus {
	// Use actual order fulfillment statuses
	statusNames := []string{"PENDING", "PROCESSING", "SHIPPED", "DELIVERED", "CANCELLED"}
	statuses := make([]NewOrderFulfillmentStatus, 0, n)
	for i := 0; i < n; i++ {
		name := statusNames[i%len(statusNames)]
		statuses = append(statuses, NewOrderFulfillmentStatus{
			Name:        name,
			Description: fmt.Sprintf("Description for %s", name),
		})
	}
	return statuses
}

func TestSeedOrderFulfillmentStatuses(ctx context.Context, n int, api *Business) ([]OrderFulfillmentStatus, error) {
	newStatuses := TestNewOrderFulfillmentStatuses(n)
	statuses := make([]OrderFulfillmentStatus, len(newStatuses))
	for i, ns := range newStatuses {
		s, err := api.Create(ctx, ns)
		if err != nil {
			return []OrderFulfillmentStatus{}, err
		}
		statuses[i] = s
	}
	return statuses, nil
}
