package formdataapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// Test_FormData_Atomicity is the fast-follow #1 trip-wire for the formdataapp non-atomic
// multi-entity submit (a pre-existing gap F9 flagged and excluded).
//
// UpsertFormData opens a tx (BeginTxx), defers Rollback, and Commits — but the registry-driven
// entity writes route through POOL-bound app instances captured at startup, so the writes
// autocommit on the base pool and the form's tx brackets NOTHING. A bad FK partway through a
// multi-entity submit therefore leaves the EARLIER entity writes durably committed.
//
// The test posts a sales.orders (order 1) + sales.order_line_items (order 2) submit where the
// SECOND line item carries a well-formed but non-existent product_id. That payload passes app
// validation (unlike buildOrderWithInvalidLineItem, which omits product_id and dies at
// validation), so it reaches the DB and fails on the real Postgres products FK — AFTER the order
// and the first line item have already been written. It then asserts ZERO rows for the order.
//
// NOTE: both ordersbus.Create and orderlineitemsbus.Create mint their own ID (uuid.New()) and
// ignore the client-supplied "id", so we assert on the order's "number" — a business field we
// control that the bus honors — rather than on a client id the bus discards.
//
//	RED  (pre-fix):  order + first line item autocommit on the pool -> COUNT == 1 -> fails.
//	GREEN (post-fix): the writes ride the form's tx and roll back   -> COUNT == 0.
func Test_FormData_Atomicity(t *testing.T) {
	t.Parallel()

	test := apitest.StartTest(t, "Test_FormData_Atomicity")

	sd, err := insertSeedData(test.DB, test.Auth)
	require.NoError(t, err, "seeding")

	// Build a multi-entity submit whose LATER write fails at the DB (not validation):
	// order (valid) -> line item #1 (valid) -> line item #2 (non-existent product_id -> FK 23503).
	payload, orderNumber := buildOrderWithBadFKLineItem(sd)

	body, err := json.Marshal(payload)
	require.NoError(t, err, "marshal payload")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/v1/formdata/%s/upsert", sd.Forms[salesOrderForm].ID),
		bytes.NewReader(body))
	r.Header.Set("Authorization", "Bearer "+sd.Admins[0].Token)
	r.Header.Set("Content-Type", "application/json")
	test.ServeHTTP(w, r)

	require.GreaterOrEqual(t, w.Code, http.StatusBadRequest,
		"submit must fail on the second line item's FK violation; body=%s", w.Body.String())

	// The mandate: a failed multi-entity submit must leave NONE of its earlier writes committed.
	ctx := context.Background()

	var orderCount int
	require.NoError(t, test.DB.DB.GetContext(ctx, &orderCount,
		"SELECT COUNT(*) FROM sales.orders WHERE number = $1", orderNumber))
	require.Equal(t, 0, orderCount,
		"atomicity violated: the sales.orders row committed even though a later line-item write "+
			"failed. The form's tx wraps nothing — registry writes must ride it via tx-bound buses.")

	// Corroborate the array write: no line item may reference an order carrying our number.
	var lineItemCount int
	require.NoError(t, test.DB.DB.GetContext(ctx, &lineItemCount, `
		SELECT COUNT(*) FROM sales.order_line_items oli
		JOIN sales.orders o ON oli.order_id = o.id
		WHERE o.number = $1`, orderNumber))
	require.Equal(t, 0, lineItemCount,
		"atomicity violated: a sales.order_line_items row committed even though the submit failed.")
}

// buildOrderWithBadFKLineItem mirrors buildOrderWithLineItemsPayload but returns the unique order
// "number" (our durable marker) and sets the SECOND line item's product_id to a syntactically
// valid but non-existent UUID. The order and the first line item are fully valid, so they are
// written first; the second item fails the products FK at the DB layer.
func buildOrderWithBadFKLineItem(sd apitest.SeedData) (map[string]any, string) {
	orderNumber := fmt.Sprintf("TEST-ORDER-ATOMIC-%d", time.Now().UnixNano())

	lineItems := []map[string]any{
		{
			"id":                                uuid.New().String(),
			"order_id":                          "{{sales.orders.id}}",
			"product_id":                        sd.Products[0].ProductID, // real
			"quantity":                          "5",
			"discount":                          "0",
			"line_item_fulfillment_statuses_id": sd.LineItemFulfillmentStatuses[0].ID,
		},
		{
			"id":                                uuid.New().String(),
			"order_id":                          "{{sales.orders.id}}",
			"product_id":                        uuid.New().String(), // non-existent -> FK violation
			"quantity":                          "10",
			"discount":                          "0",
			"line_item_fulfillment_statuses_id": sd.LineItemFulfillmentStatuses[0].ID,
		},
	}

	payload := map[string]any{
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
				"id":                          uuid.New().String(),
				"number":                      orderNumber,
				"customer_id":                 sd.Customers[0].ID,
				"order_date":                  time.Now().Format("2006-01-02"),
				"due_date":                    time.Now().AddDate(0, 0, 30).Format("2006-01-02"),
				"order_fulfillment_status_id": sd.OrderFulfillmentStatuses[0].ID,
				"created_by":                  sd.Admins[0].ID,
			},
			"sales.order_line_items": lineItems,
		},
	}

	return payload, orderNumber
}
