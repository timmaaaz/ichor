package purchaseorderbus

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func TestNewPurchaseOrders(n int, supplierIDs uuid.UUIDs, statusIDs uuid.UUIDs, warehouseIDs uuid.UUIDs, streetIDs uuid.UUIDs, userIDs uuid.UUIDs, currencyIDs uuid.UUIDs) []NewPurchaseOrder {
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
			CurrencyID:               currencyIDs[i%len(currencyIDs)],
			RequestedBy:              userIDs[i%len(userIDs)],
			Notes:                    fmt.Sprintf("Test purchase order %d", i+1),
			SupplierReferenceNumber:  fmt.Sprintf("SUP-REF-%d", i+1),
			CreatedBy:                userIDs[i%len(userIDs)],
		})
	}
	return orders
}

func TestSeedPurchaseOrders(ctx context.Context, n int, supplierIDs uuid.UUIDs, statusIDs uuid.UUIDs, warehouseIDs uuid.UUIDs, streetIDs uuid.UUIDs, userIDs uuid.UUIDs, currencyIDs uuid.UUIDs, api *Business) ([]PurchaseOrder, error) {
	newOrders := TestNewPurchaseOrders(n, supplierIDs, statusIDs, warehouseIDs, streetIDs, userIDs, currencyIDs)
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

// TestNewPurchaseOrdersHistorical creates purchase orders distributed across a time range for seeding.
// daysBack specifies how many days of history to generate (e.g., 90, 180).
// Order dates are evenly distributed across the time range. The first few orders
// have varied delivery dates (today, +3 days, overdue) for dashboard window testing.
func TestNewPurchaseOrdersHistorical(n int, daysBack int, supplierIDs uuid.UUIDs, statusIDs uuid.UUIDs, warehouseIDs uuid.UUIDs, streetIDs uuid.UUIDs, userIDs uuid.UUIDs, currencyIDs uuid.UUIDs) []NewPurchaseOrder {
	orders := make([]NewPurchaseOrder, 0, n)
	now := time.Now()

	for i := 0; i < n; i++ {
		// Distribute evenly across the time range
		daysAgo := (i * daysBack) / n
		orderDate := now.AddDate(0, 0, -daysAgo)
		// Varied delivery dates for dashboard windows
		var expectedDelivery time.Time
		switch i {
		case 0:
			expectedDelivery = now // Today window
		case 1:
			expectedDelivery = now.AddDate(0, 0, 3) // 7 Days window
		case 2:
			expectedDelivery = now.AddDate(0, 0, -7) // Overdue (past)
		default:
			expectedDelivery = orderDate.AddDate(0, 0, 14) // Standard 2 weeks
		}

		subtotal := 1000.00 + float64(i*100)
		tax := subtotal * 0.08
		shipping := 50.00
		total := subtotal + tax + shipping

		orders = append(orders, NewPurchaseOrder{
			OrderNumber:              fmt.Sprintf("PO-HIST-%d", i+1),
			SupplierID:               supplierIDs[i%len(supplierIDs)],
			PurchaseOrderStatusID:    statusIDs[i%len(statusIDs)],
			DeliveryWarehouseID:      warehouseIDs[i%len(warehouseIDs)],
			DeliveryLocationID:       uuid.Nil,
			DeliveryStreetID:         streetIDs[i%len(streetIDs)],
			OrderDate:                orderDate, // Still regular field
			ExpectedDeliveryDate:     expectedDelivery,
			Subtotal:                 subtotal,
			TaxAmount:                tax,
			ShippingCost:             shipping,
			TotalAmount:              total,
			CurrencyID:               currencyIDs[i%len(currencyIDs)],
			RequestedBy:              userIDs[i%len(userIDs)],
			Notes:                    fmt.Sprintf("Historical PO %d", i+1),
			SupplierReferenceNumber:  fmt.Sprintf("SUP-HIST-%d", i+1),
			CreatedBy:                userIDs[i%len(userIDs)],
			CreatedDate:              &orderDate, // Use same as order date
		})
	}
	return orders
}

// TestSeedPurchaseOrdersHistorical seeds purchase orders with historical date distribution.
func TestSeedPurchaseOrdersHistorical(ctx context.Context, n int, daysBack int, supplierIDs uuid.UUIDs, statusIDs uuid.UUIDs, warehouseIDs uuid.UUIDs, streetIDs uuid.UUIDs, userIDs uuid.UUIDs, currencyIDs uuid.UUIDs, api *Business) ([]PurchaseOrder, error) {
	newOrders := TestNewPurchaseOrdersHistorical(n, daysBack, supplierIDs, statusIDs, warehouseIDs, streetIDs, userIDs, currencyIDs)
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
