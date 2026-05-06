package dbtest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/sdk/seedid"
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

			got, err := resolveRefs(ctx, tc.in, scenarioID, lookups, nil)
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

// TestRowRef covers the _label + _row_ref cross-row resolver contract.
func TestRowRef(t *testing.T) {
	t.Parallel()

	const scenarioName = "test-scenario"
	ctx := context.Background()
	lookups := fixedLookups(t)

	// scenarioID is not used for label derivation (name is); pick a fixed value.
	scenarioID := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")

	// expectedPOID is deterministic from the label key used in buildRowIndex.
	expectedPOID := seedid.Stable("scenario:" + scenarioName + ":label:po1")

	t.Run("happy path — single _label resolves cross-row", func(t *testing.T) {
		t.Parallel()

		state := map[string][]map[string]any{
			"purchase_orders": {
				{
					"_label":       "po1",
					"supplier_ref": "SUP-001",
				},
			},
			"purchase_order_line_items": {
				{
					"purchase_order_row_ref": "po1",
					"product_ref":            "SKU-0001",
					"expected_qty":           50,
				},
				{
					"purchase_order_row_ref": "po1",
					"product_ref":            "SKU-0001",
					"expected_qty":           20,
				},
			},
		}

		rowIndex, err := buildRowIndex(scenarioName, state)
		if err != nil {
			t.Fatalf("buildRowIndex: %v", err)
		}

		// Resolve the PO row.
		poRow := state["purchase_orders"][0]
		resolvedPO, err := resolveRefs(ctx, poRow, scenarioID, lookups, rowIndex)
		if err != nil {
			t.Fatalf("resolveRefs PO: %v", err)
		}
		if resolvedPO["id"] != expectedPOID.String() {
			t.Errorf("PO id: got %v, want %v", resolvedPO["id"], expectedPOID.String())
		}
		if _, hasLabel := resolvedPO["_label"]; hasLabel {
			t.Error("PO output must not contain _label key")
		}

		// Resolve the line item rows.
		for i, liRow := range state["purchase_order_line_items"] {
			resolvedLI, err := resolveRefs(ctx, liRow, scenarioID, lookups, rowIndex)
			if err != nil {
				t.Fatalf("resolveRefs line_item[%d]: %v", i, err)
			}
			if resolvedLI["purchase_order_id"] != expectedPOID.String() {
				t.Errorf("line_item[%d] purchase_order_id: got %v, want %v", i, resolvedLI["purchase_order_id"], expectedPOID.String())
			}
			if _, hasRowRef := resolvedLI["purchase_order_row_ref"]; hasRowRef {
				t.Errorf("line_item[%d] output must not contain purchase_order_row_ref key", i)
			}
		}
	})

	t.Run("unknown row_ref errors", func(t *testing.T) {
		t.Parallel()

		state := map[string][]map[string]any{
			"purchase_order_line_items": {
				{"purchase_order_row_ref": "nonexistent"},
			},
		}
		rowIndex, err := buildRowIndex(scenarioName, state)
		if err != nil {
			t.Fatalf("buildRowIndex: %v", err)
		}

		_, err = resolveRefs(ctx, state["purchase_order_line_items"][0], scenarioID, lookups, rowIndex)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), `row_ref "nonexistent" not found`) {
			t.Errorf("error %q does not contain expected message", err.Error())
		}
	})

	t.Run("duplicate _label errors", func(t *testing.T) {
		t.Parallel()

		state := map[string][]map[string]any{
			"purchase_orders": {
				{"_label": "po1"},
				{"_label": "po1"},
			},
		}
		_, err := buildRowIndex(scenarioName, state)
		if err == nil {
			t.Fatal("expected error for duplicate label, got nil")
		}
		if !strings.Contains(err.Error(), `duplicate _label "po1"`) {
			t.Errorf("error %q does not contain expected message", err.Error())
		}
	})

	t.Run("empty _label errors", func(t *testing.T) {
		t.Parallel()

		state := map[string][]map[string]any{
			"purchase_orders": {
				{"_label": ""},
			},
		}
		_, err := buildRowIndex(scenarioName, state)
		if err == nil {
			t.Fatal("expected error for empty label, got nil")
		}
		if !strings.Contains(err.Error(), "_label must be non-empty string") {
			t.Errorf("error %q does not contain expected message", err.Error())
		}
	})

	t.Run("forward reference works", func(t *testing.T) {
		// purchase_order_line_items sorts before purchase_orders alphabetically;
		// the pre-pass must make this safe.
		t.Parallel()

		state := map[string][]map[string]any{
			// alphabetically first — references label defined in purchase_orders
			"purchase_order_line_items": {
				{
					"purchase_order_row_ref": "po1",
					"product_ref":            "SKU-0001",
				},
			},
			// alphabetically second — defines the label
			"purchase_orders": {
				{"_label": "po1", "supplier_ref": "SUP-001"},
			},
		}

		rowIndex, err := buildRowIndex(scenarioName, state)
		if err != nil {
			t.Fatalf("buildRowIndex: %v", err)
		}

		liRow := state["purchase_order_line_items"][0]
		resolvedLI, err := resolveRefs(ctx, liRow, scenarioID, lookups, rowIndex)
		if err != nil {
			t.Fatalf("resolveRefs (forward ref): %v", err)
		}
		if resolvedLI["purchase_order_id"] != expectedPOID.String() {
			t.Errorf("purchase_order_id: got %v, want %v", resolvedLI["purchase_order_id"], expectedPOID.String())
		}
	})

	t.Run("combined _label with existing _ref types", func(t *testing.T) {
		t.Parallel()

		state := map[string][]map[string]any{
			"purchase_orders": {
				{
					"_label":        "po1",
					"supplier_ref":  "SUP-001",
					"warehouse_ref": "WH-MAIN",
				},
			},
		}

		rowIndex, err := buildRowIndex(scenarioName, state)
		if err != nil {
			t.Fatalf("buildRowIndex: %v", err)
		}

		poRow := state["purchase_orders"][0]
		resolved, err := resolveRefs(ctx, poRow, scenarioID, lookups, rowIndex)
		if err != nil {
			t.Fatalf("resolveRefs: %v", err)
		}
		// _label stripped
		if _, hasLabel := resolved["_label"]; hasLabel {
			t.Error("output must not contain _label key")
		}
		// id auto-injected
		if resolved["id"] != expectedPOID.String() {
			t.Errorf("id: got %v, want %v", resolved["id"], expectedPOID.String())
		}
		// supplier_ref resolved
		if resolved["supplier_id"] != "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee" {
			t.Errorf("supplier_id: got %v", resolved["supplier_id"])
		}
		// warehouse_ref resolved
		if resolved["warehouse_id"] != "ffffffff-ffff-4fff-8fff-ffffffffffff" {
			t.Errorf("warehouse_id: got %v", resolved["warehouse_id"])
		}
	})
}
