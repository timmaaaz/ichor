package lineitemfulfillmentstatusbus

import (
	"context"
	"fmt"
)

func TestNewLineItemFulfillmentStatuses() []NewLineItemFulfillmentStatus {
	// Use actual line item fulfillment statuses
	statusNames := []string{"ALLOCATED", "CANCELLED", "PACKED", "PENDING", "PICKED", "SHIPPED"}
	statuses := make([]NewLineItemFulfillmentStatus, 0, len(statusNames))
	for i := 0; i < len(statusNames); i++ {
		name := statusNames[i]
		statuses = append(statuses, NewLineItemFulfillmentStatus{
			Name:        name,
			Description: fmt.Sprintf("Description for %s", name),
		})
	}
	return statuses
}

func TestSeedLineItemFulfillmentStatuses(ctx context.Context, api *Business) ([]LineItemFulfillmentStatus, error) {
	newStatuses := TestNewLineItemFulfillmentStatuses()
	statuses := make([]LineItemFulfillmentStatus, len(newStatuses))
	for i, ns := range newStatuses {
		s, err := api.Create(ctx, ns)
		if err != nil {
			return []LineItemFulfillmentStatus{}, err
		}
		statuses[i] = s
	}
	return statuses, nil
}
