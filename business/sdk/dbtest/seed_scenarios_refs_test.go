package dbtest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func fixedLookups(t *testing.T) refLookups {
	t.Helper()
	productID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	locationID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	labelID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	supplierID := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee")
	warehouseID := uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff")
	currencyID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	purchaseOrderStatusID := uuid.MustParse("33333333-3333-4333-8333-333333333333")

	return refLookups{
		productIDBySKU: func(_ context.Context, sku string) (uuid.UUID, error) {
			if sku != "SKU-0001" {
				return uuid.Nil, errors.New("not found")
			}
			return productID, nil
		},
		locationIDByCode: func(_ context.Context, code string) (uuid.UUID, error) {
			if code != "RCV-01" {
				return uuid.Nil, errors.New("not found")
			}
			return locationID, nil
		},
		labelIDByCode: func(_ context.Context, code string) (uuid.UUID, error) {
			if code != "TOTE-001" {
				return uuid.Nil, errors.New("not found")
			}
			return labelID, nil
		},
		supplierIDByCode: func(_ context.Context, code string) (uuid.UUID, error) {
			if code != "SUP-001" {
				return uuid.Nil, errors.New("not found")
			}
			return supplierID, nil
		},
		warehouseIDByCode: func(_ context.Context, code string) (uuid.UUID, error) {
			if code != "WH-MAIN" {
				return uuid.Nil, errors.New("not found")
			}
			return warehouseID, nil
		},
		currencyIDByCode: func(_ context.Context, code string) (uuid.UUID, error) {
			if code != "USD" {
				return uuid.Nil, errors.New("not found")
			}
			return currencyID, nil
		},
		userIDByUsername: func(_ context.Context, username string) (uuid.UUID, error) {
			if username != "jdoe" {
				return uuid.Nil, errors.New("not found")
			}
			return userID, nil
		},
		purchaseOrderStatusIDByName: func(_ context.Context, name string) (uuid.UUID, error) {
			if name != "Pending" {
				return uuid.Nil, errors.New("not found")
			}
			return purchaseOrderStatusID, nil
		},
	}
}

func TestResolveRefs(t *testing.T) {
	t.Parallel()

	scenarioID := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
	lookups := fixedLookups(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		in        map[string]any
		expect    map[string]any
		expectErr string
	}{
		{
			name: "product_ref resolves to product_id, scenario_id injected",
			in: map[string]any{
				"quantity":    50,
				"product_ref": "SKU-0001",
			},
			expect: map[string]any{
				"quantity":    50,
				"product_id":  "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
				"scenario_id": "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
			},
		},
		{
			name: "three refs together",
			in: map[string]any{
				"product_ref":  "SKU-0001",
				"location_ref": "RCV-01",
				"tote_ref":     "TOTE-001",
			},
			expect: map[string]any{
				"product_id":       "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
				"location_id":      "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
				"label_catalog_id": "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
				"scenario_id":      "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
			},
		},
		{
			name: "existing scenario_id preserved (not overwritten)",
			in: map[string]any{
				"product_ref": "SKU-0001",
				"scenario_id": "9999eeee-eeee-4eee-8eee-eeeeeeeeeeee",
			},
			expect: map[string]any{
				"product_id":  "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
				"scenario_id": "9999eeee-eeee-4eee-8eee-eeeeeeeeeeee",
			},
		},
		{
			name: "already-resolved _id pass-through",
			in: map[string]any{
				"product_id": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
				"quantity":   10,
			},
			expect: map[string]any{
				"product_id":  "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
				"quantity":    10,
				"scenario_id": "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
			},
		},
		{
			name: "unknown _ref key errors",
			in: map[string]any{
				"vendor_ref": "Acme Inc",
			},
			expectErr: "unknown ref key",
		},
		{
			name: "non-string ref value errors",
			in: map[string]any{
				"product_ref": 42,
			},
			expectErr: "must be a string",
		},
		{
			name: "resolver failure propagates",
			in: map[string]any{
				"product_ref": "SKU-NONEXISTENT",
			},
			expectErr: "resolve product_ref",
		},
		{
			name: "supplier_ref resolves to supplier_id, scenario_id injected",
			in: map[string]any{
				"supplier_ref": "SUP-001",
			},
			expect: map[string]any{
				"supplier_id": "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee",
				"scenario_id": "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
			},
		},
		{
			name: "warehouse_ref resolves to warehouse_id, scenario_id injected",
			in: map[string]any{
				"warehouse_ref": "WH-MAIN",
			},
			expect: map[string]any{
				"warehouse_id": "ffffffff-ffff-4fff-8fff-ffffffffffff",
				"scenario_id":  "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
			},
		},
		{
			name: "currency_ref resolves to currency_id, scenario_id injected",
			in: map[string]any{
				"currency_ref": "USD",
			},
			expect: map[string]any{
				"currency_id": "11111111-1111-4111-8111-111111111111",
				"scenario_id": "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
			},
		},
		{
			name: "user_ref resolves to user_id, scenario_id injected",
			in: map[string]any{
				"user_ref": "jdoe",
			},
			expect: map[string]any{
				"user_id":     "22222222-2222-4222-8222-222222222222",
				"scenario_id": "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
			},
		},
		{
			name: "purchase_order_status_ref resolves to purchase_order_status_id, scenario_id injected",
			in: map[string]any{
				"purchase_order_status_ref": "Pending",
			},
			expect: map[string]any{
				"purchase_order_status_id": "33333333-3333-4333-8333-333333333333",
				"scenario_id":              "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveRefs(ctx, tc.in, scenarioID, lookups)
			if tc.expectErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (out=%v)", tc.expectErr, got)
				}
				if !strings.Contains(err.Error(), tc.expectErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.expectErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.expect) {
				t.Fatalf("key count mismatch: got %d %v, want %d %v", len(got), got, len(tc.expect), tc.expect)
			}
			for k, v := range tc.expect {
				gv, ok := got[k]
				if !ok {
					t.Errorf("missing key %q in output", k)
					continue
				}
				if gv != v {
					t.Errorf("key %q: got %v (%T), want %v (%T)", k, gv, gv, v, v)
				}
			}
		})
	}
}
