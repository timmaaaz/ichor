package orderlineitemsbus

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus/types"
)

func TestNewOrderLineItems(n int, orderIDs uuid.UUIDs, productIDs uuid.UUIDs, lineItemStatusIDs uuid.UUIDs, userIDs uuid.UUIDs) []NewOrderLineItem {
	statuses := make([]NewOrderLineItem, 0, n)
	descriptions := []string{
		"Standard item",
		"Premium quality",
		"Bulk order",
		"Special request",
		"Regular stock",
	}
	discountTypes := []string{"flat", "percent"}

	for i := 0; i < n; i++ {
		// Generate realistic prices
		unitPrice := float64(rand.Intn(200)+10) + float64(rand.Intn(100))/100 // $10.00 - $209.99
		quantity := (i % 10) + 1
		discountType := discountTypes[i%2]
		var discount float64
		if discountType == "flat" {
			discount = float64(rand.Intn(10)) // $0 - $9 flat discount
		} else {
			discount = float64(rand.Intn(20)) // 0-19% discount
		}

		// Calculate line total
		var lineTotal float64
		if discountType == "flat" {
			lineTotal = float64(quantity)*unitPrice - discount
		} else {
			lineTotal = float64(quantity) * unitPrice * (1 - discount/100)
		}
		if lineTotal < 0 {
			lineTotal = 0
		}

		up, _ := types.ParseMoney(fmt.Sprintf("%.2f", unitPrice))
		disc, _ := types.ParseMoney(fmt.Sprintf("%.2f", discount))
		lt, _ := types.ParseMoney(fmt.Sprintf("%.2f", lineTotal))

		statuses = append(statuses, NewOrderLineItem{
			OrderID:                       orderIDs[i%len(orderIDs)],
			ProductID:                     productIDs[i%len(productIDs)],
			Description:                   descriptions[i%len(descriptions)],
			Quantity:                      quantity,
			UnitPrice:                     up,
			Discount:                      disc,
			DiscountType:                  discountType,
			LineTotal:                     lt,
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

	descriptions := []string{
		"Widget A - Blue",
		"Widget A - Red",
		"Premium Service Package",
		"Consulting Hours",
		"Software License",
		"Hardware Component",
		"Maintenance Contract",
		"Training Session",
	}
	discountTypes := []string{"flat", "percent"}

	for i := 0; i < n; i++ {
		orderID := orderIDs[i%len(orderIDs)]
		orderDate := orderDates[orderID]
		// Line item created 0-2 hours after order
		lineItemDate := orderDate.Add(time.Duration(rand.Intn(120)) * time.Minute)

		// Generate realistic prices - vary by product position
		basePrice := float64((i%len(productIDs)+1)*25) + float64(rand.Intn(50))
		unitPrice := basePrice + float64(rand.Intn(100))/100 // Add cents

		quantity := (i % 10) + 1
		discountType := discountTypes[i%2]
		var discount float64
		if discountType == "flat" {
			discount = float64(rand.Intn(int(unitPrice/10))) // Up to 10% of unit price as flat discount
		} else {
			discount = float64(rand.Intn(15)) // 0-14% discount
		}

		// Calculate line total
		var lineTotal float64
		if discountType == "flat" {
			lineTotal = float64(quantity)*unitPrice - discount
		} else {
			lineTotal = float64(quantity) * unitPrice * (1 - discount/100)
		}
		if lineTotal < 0 {
			lineTotal = 0
		}

		up, _ := types.ParseMoney(fmt.Sprintf("%.2f", unitPrice))
		disc, _ := types.ParseMoney(fmt.Sprintf("%.2f", discount))
		lt, _ := types.ParseMoney(fmt.Sprintf("%.2f", lineTotal))

		items = append(items, NewOrderLineItem{
			OrderID:                       orderID,
			ProductID:                     productIDs[i%len(productIDs)],
			Description:                   descriptions[i%len(descriptions)],
			Quantity:                      quantity,
			UnitPrice:                     up,
			Discount:                      disc,
			DiscountType:                  discountType,
			LineTotal:                     lt,
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
