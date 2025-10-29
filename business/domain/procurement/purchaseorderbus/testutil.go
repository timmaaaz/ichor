package purchaseorderbus

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func TestNewPurchaseOrders(n int, supplierIDs uuid.UUIDs, statusIDs uuid.UUIDs, warehouseIDs uuid.UUIDs, streetIDs uuid.UUIDs, userIDs uuid.UUIDs) []NewPurchaseOrder {
	orders := make([]NewPurchaseOrder, 0, n)
	for i := 0; i < n; i++ {
		subtotal := 1000.00 + float64(i*100)
		tax := subtotal * 0.08
		shipping := 50.00
		total := subtotal + tax + shipping

		orders = append(orders, NewPurchaseOrder{
			OrderNumber:              fmt.Sprintf("PO-%d", i+1),
			SupplierID:               supplierIDs[i%len(supplierIDs)],
			PurchaseOrderStatusID:    statusIDs[i%len(statusIDs)],
			DeliveryWarehouseID:      warehouseIDs[i%len(warehouseIDs)],
			DeliveryLocationID:       uuid.Nil, // Using street delivery
			DeliveryStreetID:         streetIDs[i%len(streetIDs)],
			OrderDate:                time.Now().UTC(),
			ExpectedDeliveryDate:     time.Now().UTC().Add(time.Hour * 24 * 14), // 2 weeks
			Subtotal:                 subtotal,
			TaxAmount:                tax,
			ShippingCost:             shipping,
			TotalAmount:              total,
			Currency:                 "USD",
			RequestedBy:              userIDs[i%len(userIDs)],
			Notes:                    fmt.Sprintf("Test purchase order %d", i+1),
			SupplierReferenceNumber:  fmt.Sprintf("SUP-REF-%d", i+1),
			CreatedBy:                userIDs[i%len(userIDs)],
		})
	}
	return orders
}

func TestSeedPurchaseOrders(ctx context.Context, n int, supplierIDs uuid.UUIDs, statusIDs uuid.UUIDs, warehouseIDs uuid.UUIDs, streetIDs uuid.UUIDs, userIDs uuid.UUIDs, api *Business) ([]PurchaseOrder, error) {
	newOrders := TestNewPurchaseOrders(n, supplierIDs, statusIDs, warehouseIDs, streetIDs, userIDs)
	orders := make([]PurchaseOrder, len(newOrders))
	for i, no := range newOrders {
		order, err := api.Create(ctx, no)
		if err != nil {
			return []PurchaseOrder{}, fmt.Errorf("creating purchase order: %w", err)
		}
		orders[i] = order
	}
	return orders, nil
}
