package purchaseorderlineitembus

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func TestNewPurchaseOrderLineItems(n int, purchaseOrderIDs uuid.UUIDs, supplierProductIDs uuid.UUIDs, lineItemStatusIDs uuid.UUIDs, userIDs uuid.UUIDs) []NewPurchaseOrderLineItem {
	items := make([]NewPurchaseOrderLineItem, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, NewPurchaseOrderLineItem{
			PurchaseOrderID:      purchaseOrderIDs[i%len(purchaseOrderIDs)],
			SupplierProductID:    supplierProductIDs[i%len(supplierProductIDs)],
			QuantityOrdered:      (i + 1) * 10,
			UnitCost:             100.00 + float64(i*10),
			Discount:             5.00,
			LineTotal:            (100.00 + float64(i*10)) * float64((i+1)*10) - 5.00,
			LineItemStatusID:     lineItemStatusIDs[i%len(lineItemStatusIDs)],
			ExpectedDeliveryDate: time.Now().UTC().Add(time.Hour * 24 * 7), // 7 days from now
			Notes:                "Test line item notes",
			CreatedBy:            userIDs[i%len(userIDs)],
		})
	}
	return items
}

func TestSeedPurchaseOrderLineItems(ctx context.Context, n int, purchaseOrderIDs uuid.UUIDs, supplierProductIDs uuid.UUIDs, lineItemStatusIDs uuid.UUIDs, userIDs uuid.UUIDs, api *Business) ([]PurchaseOrderLineItem, error) {
	newItems := TestNewPurchaseOrderLineItems(n, purchaseOrderIDs, supplierProductIDs, lineItemStatusIDs, userIDs)
	items := make([]PurchaseOrderLineItem, len(newItems))
	for i, ni := range newItems {
		item, err := api.Create(ctx, ni)
		if err != nil {
			return []PurchaseOrderLineItem{}, err
		}
		items[i] = item
	}
	return items, nil
}
