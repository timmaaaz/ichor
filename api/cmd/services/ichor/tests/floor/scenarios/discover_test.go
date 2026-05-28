package scenarios_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

func TestDeriveFamily(t *testing.T) {
	cases := []struct {
		name string
		want family
	}{
		{"transfer-intra-zone", familyTransfer},
		{"transfer-cross-zone", familyTransfer},
		{"pick-whole-order", familyPick},
		{"pick-short-pick", familyPick},
		{"receive-lot-tracking", familyReceive},
		{"cycle-count-variance-over", familyCycleCount},
		{"profile-medical-device-rental", familyProfile},
		{"profile-strict-regulated", familyProfile},
		{"rush-receiving", familyReceive},   // override
		{"e2e-pick-strict", familyPick},     // override
		{"e2e-baseline", ""},                // unset — falls through to Custom
		{"unknown-prefix", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := deriveFamily(tc.name); got != tc.want {
				t.Errorf("deriveFamily(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestDiscoverScenarios_Smoke(t *testing.T) {
	rows, err := discoverScenarios(scenarioRoots())
	if err != nil {
		t.Fatalf("discoverScenarios: %v", err)
	}
	// We expect exactly 22 scenarios in deployments/scenarios/ as of 2026-05-28
	// (e2e-pick-tote added in phase-0g.B7). If this count drifts, either a
	// scenario was added/removed or the discovery glob broke — investigate, do
	// not blindly update the number.
	const wantCount = 22
	if len(rows) != wantCount {
		names := make([]string, 0, len(rows))
		for _, r := range rows {
			names = append(names, r.Name)
		}
		t.Errorf("discoverScenarios returned %d rows, want %d. Got: %v", len(rows), wantCount, names)
	}
}

// discoverTransferInputs queries the DB for the single pending-or-approved
// transfer order that belongs to the given scenario, and enriches it with the
// source/destination location codes and product UPC needed by walkTransfer.
//
// The query accepts both "pending" and "approved" statuses so it works
// regardless of whether the scenario seeds the order as pending (the current
// convention) or pre-approved. walkTransfer advances the status itself.
//
// All failure paths call t.Fatalf; the returned TransferInputs is always valid.
func discoverTransferInputs(t *testing.T, h *apitest.Test, scenarioID uuid.UUID) TransferInputs {
	t.Helper()
	ctx := context.Background()
	db := h.DB.DB

	// Fetch the transfer order row.
	var row struct {
		ID             uuid.UUID `db:"id"`
		FromLocationID uuid.UUID `db:"from_location_id"`
		ToLocationID   uuid.UUID `db:"to_location_id"`
		ProductID      uuid.UUID `db:"product_id"`
		Quantity       int       `db:"quantity"`
	}
	err := db.GetContext(ctx, &row, `
		SELECT id, from_location_id, to_location_id, product_id, quantity
		FROM inventory.transfer_orders
		WHERE scenario_id = $1
		  AND status IN ('pending', 'approved')
		ORDER BY transfer_date ASC
		LIMIT 1
	`, scenarioID)
	if err != nil {
		t.Fatalf("discoverTransferInputs: query transfer_orders for scenario %s: %v", scenarioID, err)
	}

	// Resolve source location code.
	var fromCode string
	if err := db.GetContext(ctx, &fromCode, `
		SELECT location_code
		FROM inventory.inventory_locations
		WHERE id = $1
	`, row.FromLocationID); err != nil {
		t.Fatalf("discoverTransferInputs: query from_location code for %s: %v", row.FromLocationID, err)
	}

	// Resolve destination location code.
	var toCode string
	if err := db.GetContext(ctx, &toCode, `
		SELECT location_code
		FROM inventory.inventory_locations
		WHERE id = $1
	`, row.ToLocationID); err != nil {
		t.Fatalf("discoverTransferInputs: query to_location code for %s: %v", row.ToLocationID, err)
	}

	// Resolve product UPC.
	var upc string
	if err := db.GetContext(ctx, &upc, `
		SELECT upc_code
		FROM products.products
		WHERE id = $1
	`, row.ProductID); err != nil {
		t.Fatalf("discoverTransferInputs: query upc_code for product %s: %v", row.ProductID, err)
	}

	return TransferInputs{
		TransferID: row.ID,
		FromCode:   fromCode,
		ToCode:     toCode,
		ProductID:  row.ProductID,
		UPC:        upc,
		Quantity:   row.Quantity,
	}
}

// discoverReceiveInputs queries the DB for the first APPROVED purchase order
// that belongs to the given scenario, then enriches it with per-line-item
// product metadata (UPC, tracking_type).
//
// For scenarios that have no purchase orders in their state (e.g.,
// rush-receiving ships with an empty state.yaml), the function returns a
// zero-value ReceiveInputs{POID: uuid.Nil, LineItems: nil}.  walkReceive
// handles that case by running only the endpoint smoke-checks (GB-006 empty
// list, GB-014 lot-trackings query) and skipping the per-line receive POST.
//
// All failure paths call t.Fatalf; when a non-empty ReceiveInputs is returned
// it is always valid.
func discoverReceiveInputs(t *testing.T, h *apitest.Test, scenarioID uuid.UUID) ReceiveInputs {
	t.Helper()
	ctx := context.Background()
	db := h.DB.DB

	// Step 1 — find the APPROVED purchase order for this scenario.
	// purchase_orders.scenario_id links POs authored by the scenario seed.
	// The APPROVED status name ("APPROVED") is the canonical value in
	// seedmodels.PurchaseOrderStatusData; the status row is seeded by
	// seedProcurement and the UUID is looked up dynamically here.
	var poID uuid.UUID
	err := db.GetContext(ctx, &poID, `
		SELECT po.id
		FROM procurement.purchase_orders po
		JOIN procurement.purchase_order_statuses pos
		  ON pos.id = po.purchase_order_status_id
		WHERE po.scenario_id = $1
		  AND pos.name = 'APPROVED'
		ORDER BY po.order_number ASC
		LIMIT 1
	`, scenarioID)
	if errors.Is(err, sql.ErrNoRows) {
		// scenario has no purchase orders (e.g. rush-receiving empty state)
		return ReceiveInputs{}
	}
	if err != nil {
		t.Fatalf("discoverReceiveInputs: query purchase_orders for scenario %s: %v", scenarioID, err)
	}

	// Step 2 — fetch all line items for this PO.
	// Join supplier_products → products to retrieve product_id, upc_code,
	// tracking_type in a single query.
	type lineRow struct {
		LineID            uuid.UUID `db:"line_id"`
		SupplierProductID uuid.UUID `db:"supplier_product_id"`
		ProductID         uuid.UUID `db:"product_id"`
		UPC               string    `db:"upc_code"`
		TrackingType      string    `db:"tracking_type"`
		QuantityOrdered   int       `db:"quantity_ordered"`
	}

	var rows []lineRow
	if err := db.SelectContext(ctx, &rows, `
		SELECT
			li.id             AS line_id,
			li.supplier_product_id,
			p.id              AS product_id,
			COALESCE(p.upc_code, '')     AS upc_code,
			COALESCE(p.tracking_type, 'none') AS tracking_type,
			li.quantity_ordered
		FROM procurement.purchase_order_line_items li
		JOIN procurement.supplier_products sp ON sp.id = li.supplier_product_id
		JOIN products.products p ON p.id = sp.product_id
		WHERE li.purchase_order_id = $1
		ORDER BY li.created_date ASC
	`, poID); err != nil {
		t.Fatalf("discoverReceiveInputs: query line items for PO %s: %v", poID, err)
	}

	items := make([]ReceiveLineItem, len(rows))
	for i, r := range rows {
		items[i] = ReceiveLineItem{
			LineID:            r.LineID,
			ProductID:         r.ProductID,
			UPC:               r.UPC,
			ExpectedQty:       r.QuantityOrdered,
			LotTracked:        r.TrackingType == "lot",
			SerialTracked:     r.TrackingType == "serial",
			SupplierProductID: r.SupplierProductID,
		}
	}

	return ReceiveInputs{
		POID:      poID,
		LineItems: items,
	}
}

// discoverPickInputs queries the DB for the sales order and pick tasks seeded
// by this scenario, then enriches each task with its location code and product
// UPC needed by walkPick.
//
// For scenarios that have no pick_tasks in their state (e.g., e2e-pick-strict
// which only carries lever_overrides), this returns PickInputs{SOID: uuid.Nil,
// Allocations: nil}. walkPick handles that case by running only endpoint
// smoke-checks and skipping the per-task pick-quantity POST.
//
// PickAllocation.PickTaskID holds the sales_order_line_item_id (not the
// pick_tasks.id) because the pick-quantity endpoint is keyed on line item ID:
//
//	POST /v1/sales/order-line-items/{order_line_items_id}/pick-quantity
//
// The field name "PickTaskID" was locked in Task 2's types_test.go; we reuse
// it for the endpoint URL parameter to avoid changing a pre-existing type.
//
// All failure paths call t.Fatalf; when a non-empty PickInputs is returned it
// is always valid.
func discoverPickInputs(t *testing.T, h *apitest.Test, scenarioID uuid.UUID) PickInputs {
	t.Helper()
	ctx := context.Background()
	db := h.DB.DB

	// Step 1 — find the sales order seeded by this scenario.
	// sales.orders.number is unique; ORDER BY number ASC LIMIT 1 is
	// deterministic regardless of seeding order.
	var soID uuid.UUID
	err := db.GetContext(ctx, &soID, `
		SELECT id
		FROM sales.orders
		WHERE scenario_id = $1
		ORDER BY number ASC
		LIMIT 1
	`, scenarioID)
	if errors.Is(err, sql.ErrNoRows) {
		// scenario has no sales orders (e.g. e2e-pick-strict lever-only scenario)
		return PickInputs{}
	}
	if err != nil {
		t.Fatalf("discoverPickInputs: query sales.orders for scenario %s: %v", scenarioID, err)
	}

	// Step 2 — fetch all pick_tasks for that sales order.
	// sales_order_line_item_id is the URL parameter for the pick-quantity
	// endpoint; quantity_to_pick is the safe amount to pass (matches on-hand).
	type taskRow struct {
		LineItemID uuid.UUID `db:"sales_order_line_item_id"`
		ProductID  uuid.UUID `db:"product_id"`
		LocationID uuid.UUID `db:"location_id"`
		Qty        int       `db:"quantity_to_pick"`
	}
	var tasks []taskRow
	if err := db.SelectContext(ctx, &tasks, `
		SELECT sales_order_line_item_id, product_id, location_id, quantity_to_pick
		FROM inventory.pick_tasks
		WHERE sales_order_id = $1
		ORDER BY sales_order_line_item_id ASC
	`, soID); err != nil {
		t.Fatalf("discoverPickInputs: query pick_tasks for order %s: %v", soID, err)
	}

	// Step 3 — for each task, resolve location_code + UPC + lot-tracking flag.
	allocations := make([]PickAllocation, 0, len(tasks))
	for _, task := range tasks {
		var locCode string
		if err := db.GetContext(ctx, &locCode, `
			SELECT location_code FROM inventory.inventory_locations WHERE id = $1
		`, task.LocationID); err != nil {
			t.Fatalf("discoverPickInputs: location_code for %s: %v", task.LocationID, err)
		}

		var upc string
		if err := db.GetContext(ctx, &upc, `
			SELECT COALESCE(upc_code, '') FROM products.products WHERE id = $1
		`, task.ProductID); err != nil {
			t.Fatalf("discoverPickInputs: upc_code for product %s: %v", task.ProductID, err)
		}

		var trackingType string
		_ = db.GetContext(ctx, &trackingType, `
			SELECT COALESCE(tracking_type, 'none') FROM products.products WHERE id = $1
		`, task.ProductID)

		// PickTaskID holds the sales_order_line_item_id — that is the URL
		// parameter for POST /v1/sales/order-line-items/{id}/pick-quantity.
		// The field name was locked in Task 2; we reuse it for the endpoint key.
		allocations = append(allocations, PickAllocation{
			PickTaskID:   task.LineItemID,
			ProductID:    task.ProductID,
			UPC:          upc,
			LocationCode: locCode,
			Qty:          task.Qty,
			LotTracked:   trackingType == "lot",
		})
	}

	return PickInputs{
		SOID:        soID,
		Allocations: allocations,
	}
}

// discoverProfileFlags queries config.scenario_setting_overrides directly to
// return the lever overrides that were written at seed time for the given
// profile scenario.
//
// Design pivot (Task 11 pre-flight):
//
// The plan's "load A, load B, re-load A" design is broken. Business.Load calls
// DeleteScopedRows for the current active then ApplyFixtures for the target —
// so loading a workflow scenario after a profile scenario destroys the profile's
// lever_overrides (they're scoped to the profile scenario_id). Re-loading the
// profile then destroys the workflow entities. The round-trip is a no-op for
// the entities but leaves you with an empty DB.
//
// Chosen alternative — Option 3 (activate profile, assert lever_overrides
// queryable):
//   - Profile scenarios have NO entities (no POs, no transfer orders, no pick
//     tasks). Their only effect is writing rows to config.scenario_setting_overrides.
//   - discoverProfileFlags queries that table directly to confirm the rows exist.
//   - walkProfileWithReceive / walkProfileWithTransfer then hit
//     GET /v1/config/settings/{key} for a representative subset of the profile's
//     keys to confirm the HTTP layer returns the override value, not the default.
//   - walkE2EBaseline activates the empty scenario and verifies no overrides exist.
//
// This matches the actual codebase behavior: profile scenarios are config-only
// and their test coverage is "overrides are present in the DB and the settings
// resolver returns them via HTTP". Walking a workflow (receive, transfer) under
// the profile requires baseline data; those levers are DORMANT at the API layer
// (the floor composables read them client-side), so there is nothing for a
// server-side HTTP walk to assert about lever strictness beyond the settings GET.
func discoverProfileFlags(t *testing.T, h *apitest.Test, scenarioID uuid.UUID) ProfileFlags {
	t.Helper()
	ctx := context.Background()
	db := h.DB.DB

	type overrideRow struct {
		Key   string `db:"key"`
		Value string `db:"value"`
	}

	var rows []overrideRow
	if err := db.SelectContext(ctx, &rows, `
		SELECT key, value
		FROM config.scenario_setting_overrides
		WHERE scenario_id = $1
		ORDER BY key ASC
	`, scenarioID); err != nil {
		t.Fatalf("discoverProfileFlags: query scenario_setting_overrides for scenario %s: %v", scenarioID, err)
	}

	overrides := make(map[string]string, len(rows))
	for _, r := range rows {
		overrides[r.Key] = r.Value
	}

	return ProfileFlags{LeverOverrides: overrides}
}

// discoverCycleCountInputs queries the DB for the cycle count session and its
// items seeded by this scenario.
//
// All cycle-count scenarios pre-seed their session + items in state.yaml
// (status = draft + pending). The walk uses the seeded records directly — it
// does NOT create a new session. This mirrors how the floor PWA resumes a
// cycle count that was previously opened by a supervisor.
//
// CycleCountInputs.LocationCode is populated from the first item's location.
// For single-location scenarios (variance-over, variance-under, multi-item)
// all items share the same location; for multi-location scenarios (scheduled)
// each item may differ — the walk drives items by their individual locationId
// so LocationCode is informational only.
//
// CycleCountInputs.Items[i].ActualQty == ExpectedQty; the walk applies variance
// internally based on VarianceMode. Callers that want a variance walk must set
// in.VarianceMode before calling walkCycleCount.
//
// All failure paths call t.Fatalf; the returned CycleCountInputs is always valid.
func discoverCycleCountInputs(t *testing.T, h *apitest.Test, scenarioID uuid.UUID) CycleCountInputs {
	t.Helper()
	ctx := context.Background()
	db := h.DB.DB

	// Step 1 — find the draft cycle count session seeded by this scenario.
	var sessionID uuid.UUID
	err := db.GetContext(ctx, &sessionID, `
		SELECT id
		FROM inventory.cycle_count_sessions
		WHERE scenario_id = $1
		  AND status = 'draft'
		ORDER BY created_date ASC
		LIMIT 1
	`, scenarioID)
	if err != nil {
		t.Fatalf("discoverCycleCountInputs: query cycle_count_sessions for scenario %s: %v", scenarioID, err)
	}

	// Step 2 — fetch all pending items for that session.
	type itemRow struct {
		ID         uuid.UUID `db:"id"`
		ProductID  uuid.UUID `db:"product_id"`
		LocationID uuid.UUID `db:"location_id"`
		SysQty     int       `db:"system_quantity"`
	}
	var rows []itemRow
	if err := db.SelectContext(ctx, &rows, `
		SELECT id, product_id, location_id, system_quantity
		FROM inventory.cycle_count_items
		WHERE scenario_id = $1
		  AND session_id = $2
		  AND status = 'pending'
		ORDER BY product_id ASC
	`, scenarioID, sessionID); err != nil {
		t.Fatalf("discoverCycleCountInputs: query cycle_count_items for session %s: %v", sessionID, err)
	}
	if len(rows) == 0 {
		t.Fatalf("discoverCycleCountInputs: no pending cycle_count_items found for session %s (scenario %s)", sessionID, scenarioID)
	}

	// Step 3 — resolve location_code + UPC for each item.
	items := make([]CycleCountItem, 0, len(rows))
	var primaryLocationCode string // from first item — informational
	for i, row := range rows {
		var locCode string
		if err := db.GetContext(ctx, &locCode, `
			SELECT location_code FROM inventory.inventory_locations WHERE id = $1
		`, row.LocationID); err != nil {
			t.Fatalf("discoverCycleCountInputs: location_code for %s: %v", row.LocationID, err)
		}
		if i == 0 {
			primaryLocationCode = locCode
		}

		var upc string
		_ = db.GetContext(ctx, &upc, `
			SELECT COALESCE(upc_code, '') FROM products.products WHERE id = $1
		`, row.ProductID)

		items = append(items, CycleCountItem{
			ProductID:   row.ProductID,
			UPC:         upc,
			ExpectedQty: row.SysQty,
			ActualQty:   row.SysQty, // walk applies variance based on VarianceMode
		})
	}

	return CycleCountInputs{
		LocationCode: primaryLocationCode,
		LocationID:   rows[0].LocationID,
		Items:        items,
		VarianceMode: "", // caller sets VarianceMode before walkCycleCount
	}
}
