package formdataapi_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// =============================================================================
// Helper Functions for Phase 4 Testing
// =============================================================================

// buildOrderWithLineItemsPayload creates a FormDataRequest for testing orders with line items.
// lineItemCount: number of line items to include (1-100)
func buildOrderWithLineItemsPayload(sd apitest.SeedData, lineItemCount int) map[string]any {
	orderID := uuid.New()

	lineItems := make([]map[string]any, lineItemCount)
	for i := 0; i < lineItemCount; i++ {
		lineItems[i] = map[string]any{
			"id":                                 uuid.New().String(),
			"order_id":                           "{{sales.orders.id}}", // Template variable
			"product_id":                         sd.Products[i%len(sd.Products)].ProductID,
			"quantity":                           fmt.Sprintf("%d", (i+1)*5),
			"discount":                           "0",
			"line_item_fulfillment_statuses_id": sd.LineItemFulfillmentStatuses[0].ID,
			"created_by":                         sd.Admins[0].ID,
		}
	}

	return map[string]any{
		"operations": map[string]any{
			"sales.orders": map[string]any{
				"operation": "create",
				"order":     1,
			},
			"sales.order_line_items": map[string]any{
				"operation": "create",
				"order":     2,
			},
		},
		"data": map[string]any{
			"sales.orders": map[string]any{
				"id":                   orderID.String(),
				"number":               fmt.Sprintf("TEST-ORDER-%d", time.Now().Unix()),
				"customer_id":          sd.Customers[0].ID,
				"due_date":             time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
				"fulfillment_status_id": sd.OrderFulfillmentStatuses[0].ID,
				"created_by":           sd.Admins[0].ID,
			},
			"sales.order_line_items": lineItems,
		},
	}
}

// buildOrderWithInvalidLineItem creates a payload where the second line item has missing required field.
func buildOrderWithInvalidLineItem(sd apitest.SeedData) map[string]any {
	orderID := uuid.New()

	lineItems := []map[string]any{
		{
			"id":                                 uuid.New().String(),
			"order_id":                           "{{sales.orders.id}}",
			"product_id":                         sd.Products[0].ProductID,
			"quantity":                           "5",
			"discount":                           "0",
			"line_item_fulfillment_statuses_id": sd.LineItemFulfillmentStatuses[0].ID,
			"created_by":                         sd.Admins[0].ID,
		},
		{
			"id":       uuid.New().String(),
			"order_id": "{{sales.orders.id}}",
			// Missing required "product_id" field
			"quantity":                           "10",
			"discount":                           "0",
			"line_item_fulfillment_statuses_id": sd.LineItemFulfillmentStatuses[0].ID,
			"created_by":                         sd.Admins[0].ID,
		},
	}

	return map[string]any{
		"operations": map[string]any{
			"sales.orders": map[string]any{
				"operation": "create",
				"order":     1,
			},
			"sales.order_line_items": map[string]any{
				"operation": "create",
				"order":     2,
			},
		},
		"data": map[string]any{
			"sales.orders": map[string]any{
				"id":                   orderID.String(),
				"number":               fmt.Sprintf("TEST-ORDER-INVALID-%d", time.Now().Unix()),
				"customer_id":          sd.Customers[0].ID,
				"due_date":             time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
				"fulfillment_status_id": sd.OrderFulfillmentStatuses[0].ID,
				"created_by":           sd.Admins[0].ID,
			},
			"sales.order_line_items": lineItems,
		},
	}
}

// buildOrderWithEmptyLineItems creates a payload with an empty line items array.
func buildOrderWithEmptyLineItems(sd apitest.SeedData) map[string]any {
	orderID := uuid.New()

	return map[string]any{
		"operations": map[string]any{
			"sales.orders": map[string]any{
				"operation": "create",
				"order":     1,
			},
			"sales.order_line_items": map[string]any{
				"operation": "create",
				"order":     2,
			},
		},
		"data": map[string]any{
			"sales.orders": map[string]any{
				"id":                   orderID.String(),
				"number":               fmt.Sprintf("TEST-ORDER-EMPTY-%d", time.Now().Unix()),
				"customer_id":          sd.Customers[0].ID,
				"due_date":             time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
				"fulfillment_status_id": sd.OrderFulfillmentStatuses[0].ID,
				"created_by":           sd.Admins[0].ID,
			},
			"sales.order_line_items": []map[string]any{},
		},
	}
}

// verifyOrderLineItems checks that an order exists with the expected number of line items.
func verifyOrderLineItems(t *testing.T, db *dbtest.Database, orderID uuid.UUID, expectedCount int) {
	t.Helper()

	// 1. Verify order exists
	var order struct {
		ID uuid.UUID `db:"id"`
	}
	err := db.DB.Get(&order, "SELECT id FROM sales.orders WHERE id = $1", orderID)
	if err != nil {
		t.Fatalf("Order should exist in database: %v", err)
	}

	// 2. Count line items
	var count int
	err = db.DB.Get(&count, "SELECT COUNT(*) FROM sales.order_line_items WHERE order_id = $1", orderID)
	if err != nil {
		t.Fatalf("Should be able to query line items: %v", err)
	}
	if count != expectedCount {
		t.Fatalf("Expected %d line items, got %d", expectedCount, count)
	}

	// 3. Verify FK integrity
	var orphanedCount int
	err = db.DB.Get(&orphanedCount, `
		SELECT COUNT(*) FROM sales.order_line_items oli
		LEFT JOIN sales.orders o ON oli.order_id = o.id
		WHERE o.id IS NULL
	`)
	if err != nil {
		t.Fatalf("Should be able to check FK integrity: %v", err)
	}
	if orphanedCount != 0 {
		t.Fatalf("No orphaned line items should exist, found %d", orphanedCount)
	}
}

// verifyNoOrphanedLineItems ensures transaction rollback worked correctly.
func verifyNoOrphanedLineItems(t *testing.T, db *dbtest.Database) {
	t.Helper()

	var orphanedCount int
	err := db.DB.Get(&orphanedCount, `
		SELECT COUNT(*) FROM sales.order_line_items oli
		LEFT JOIN sales.orders o ON oli.order_id = o.id
		WHERE o.id IS NULL
	`)
	if err != nil {
		t.Fatalf("Should be able to check for orphaned line items: %v", err)
	}
	if orphanedCount != 0 {
		t.Fatalf("No orphaned line items should exist after rollback, found %d", orphanedCount)
	}
}
