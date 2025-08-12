package orderfulfillmentstatusbus

import (
	"context"
	"fmt"
)

func TestNewOrderFulfillmentStatuses() []NewOrderFulfillmentStatus {
	// Use actual order fulfillment statuses
	statusNames := []string{"PENDING", "PROCESSING", "SHIPPED", "DELIVERED", "CANCELLED"}
	statuses := make([]NewOrderFulfillmentStatus, 0, len(statusNames))
	for i := 0; i < len(statusNames); i++ {
		name := statusNames[i%len(statusNames)]
		statuses = append(statuses, NewOrderFulfillmentStatus{
			Name:        name,
			Description: fmt.Sprintf("Description for %s", name),
		})
	}
	return statuses
}

func TestSeedOrderFulfillmentStatuses(ctx context.Context, api *Business) ([]OrderFulfillmentStatus, error) {
	newStatuses := TestNewOrderFulfillmentStatuses()
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
