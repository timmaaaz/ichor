package scenarios_test

// HTTP-issuing pattern in this file (established by labels/labelapi tests):
//
//	body := mustJSON(t, payload)
//	req := httptest.NewRequest(http.MethodPost, "/v1/...", bytes.NewReader(body))
//	req.Header.Set("Authorization", "Bearer "+token)
//	w := httptest.NewRecorder()
//	h.ServeHTTP(w, req)
//
// apitest.Test.ServeHTTP is public and exposes the inner mux directly.
// For GET requests with no body, pass nil as the body argument.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// mustDecodeJSON unmarshals the recorder body into dst. It calls t.Fatalf on
// any decode error and is used by all walk helpers in Tasks 6-11.
func mustDecodeJSON(t *testing.T, w *httptest.ResponseRecorder, dst any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), dst); err != nil {
		t.Fatalf("mustDecodeJSON: %v — body: %s", err, w.Body.String())
	}
}

// mustJSON serialises v to JSON. Calls t.Fatalf if marshalling fails.
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("mustJSON: %v", err)
	}
	return b
}

// adminToken returns a token for admin@example.com, the ZZZADMIN user who
// has full transfer_orders table-access permissions in seed.sql.
// walkTransfer requires this because the FLOOR_WORKER role does not include
// inventory.transfer_orders in its table access grants.
func adminToken(t *testing.T, h *apitest.Test) string {
	t.Helper()
	tok := apitest.Token(h.DB.BusDomain.User, h.Auth, "admin@example.com")
	if tok == "" {
		t.Fatal("adminToken: apitest.Token returned empty string — admin@example.com not found in DB")
	}
	return tok
}

// doRequest issues a single authenticated HTTP request against the test mux and
// returns the response recorder. It calls t.Fatalf if the status does not match
// wantStatus.
func doRequest(t *testing.T, h *apitest.Test, method, path string, token string, body []byte, wantStatus int, gbNote string) *httptest.ResponseRecorder {
	t.Helper()
	var bodyReader *bytes.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	} else {
		bodyReader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != wantStatus {
		t.Fatalf("%s %s: want status %d, got %d — body: %s\n%s", method, path, wantStatus, w.Code, w.Body.String(), gbNote)
	}
	return w
}

// walkTransfer drives a floor-worker transfer through the canonical 6-step
// wizard using real HTTP against the test mux:
//
//  1. GET  /v1/inventory/transfer-orders/{id}            → 200, status=pending    (sanity)
//  2. POST /v1/inventory/transfer-orders/{id}/approve    → 200, status=approved   (GB-010)
//  3. POST /v1/inventory/transfer-orders/{id}/claim      → 200, status=in_transit (GB-010)
//  4. GET  /v1/inventory/inventory-locations?location_code_exact={fromCode} → 200, ≥1 item (GB-008)
//  5. GET  /v1/inventory/inventory-locations?location_code_exact={toCode}   → 200, ≥1 item (GB-008)
//  6. POST /v1/inventory/transfer-orders/{id}/execute    → 200, status=completed
//
// Steps 4-5 validate that the location codes the scenario wired up are actually
// queryable via the inventory-locations API (GB-008: location-code scan).
//
// Step 6 guards transfer-execute integrity: transferorderapp.Execute calls
// DecrementQuantity directly and does NOT route through
// pickingapp.QueryAvailableForAllocation (where GB-015's FEFO subquery lives).
// GB-015 coverage belongs to walkPick (Task 9). Step 6 here guards
// DecrementQuantity correctness + state transition to completed.
//
// If the scenario seeds insufficient stock, the app returns
// 422/FailedPrecondition (GB-011). The harness surfaces this as a loud failure
// rather than a skip — the Playwright walks (Phase B) PATCH around GB-011, but
// the harness intentionally signals the gap.
//
// The scenarioID parameter is currently unused by walkTransfer (the transfer
// state is queried by transfer_id alone), but is retained for signature
// consistency with walkReceive, walkPick, walkCycleCount — Task 12's
// table-driven dispatch calls all family walks with the same 4-arg shape.
func walkTransfer(t *testing.T, h *apitest.Test, scenarioID uuid.UUID, in TransferInputs) {
	_ = scenarioID // signature-consistent; see doc comment
	t.Helper()
	token := adminToken(t, h)
	idStr := in.TransferID.String()

	// Step 1 — sanity: GET the transfer order, expect status=pending.
	{
		w := doRequest(t, h, http.MethodGet,
			"/v1/inventory/transfer-orders/"+idStr,
			token, nil, http.StatusOK,
			"Step 1 GET transfer order failed — scenario fixtures may not have loaded correctly")

		var to struct {
			Status string `json:"status"`
		}
		mustDecodeJSON(t, w, &to)
		if to.Status != "pending" {
			t.Fatalf("walkTransfer step 1: want status=pending, got %q (transfer_id=%s)", to.Status, idStr)
		}
	}

	// Step 2 — GB-010: advance pending → approved via POST /approve.
	// GB-010 regression: this endpoint must accept a pending transfer and
	// return it in approved state. A 422/400 here means the status machine
	// or the Approve handler is broken.
	{
		w := doRequest(t, h, http.MethodPost,
			"/v1/inventory/transfer-orders/"+idStr+"/approve",
			token, nil, http.StatusOK,
			"GB-010: POST /approve returned non-200 — status machine may be broken (approve must accept pending)")

		var to struct {
			Status string `json:"status"`
		}
		mustDecodeJSON(t, w, &to)
		if to.Status != "approved" {
			t.Fatalf("walkTransfer step 2 (GB-010): want status=approved after /approve, got %q", to.Status)
		}
	}

	// Step 3 — GB-010: advance approved → in_transit via POST /claim.
	// GB-010 regression: /claim must accept an approved transfer and return
	// it with status=in_transit. Failing here means the Claim handler or the
	// status transition is broken before we even reach the execute step.
	{
		w := doRequest(t, h, http.MethodPost,
			"/v1/inventory/transfer-orders/"+idStr+"/claim",
			token, nil, http.StatusOK,
			"GB-010: POST /claim returned non-200 — status machine may be broken (claim must accept approved)")

		var to struct {
			Status string `json:"status"`
		}
		mustDecodeJSON(t, w, &to)
		if to.Status != "in_transit" {
			t.Fatalf("walkTransfer step 3 (GB-010): want status=in_transit after /claim, got %q", to.Status)
		}
	}

	// Step 4 — GB-008: from-location must be queryable by location_code_exact.
	// GB-008 regression: the inventory-locations query filter must resolve the
	// location code the scenario wired as the source. A 200 with empty items
	// means the location code didn't survive the scenario fixture load.
	{
		url := fmt.Sprintf("/v1/inventory/inventory-locations?location_code_exact=%s", in.FromCode)
		w := doRequest(t, h, http.MethodGet, url, token, nil, http.StatusOK,
			"GB-008: GET inventory-locations by from_location_code returned non-200")

		var result struct {
			Items []struct {
				LocationID string `json:"location_id"`
			} `json:"items"`
			Total int `json:"total"`
		}
		mustDecodeJSON(t, w, &result)
		if result.Total == 0 || len(result.Items) == 0 {
			t.Fatalf("GB-008: from_location %q not found via location_code_exact filter (total=%d)", in.FromCode, result.Total)
		}
		if result.Items[0].LocationID == "" {
			t.Fatalf("GB-008: from_location %q item missing location_id in response", in.FromCode)
		}
	}

	// Step 5 — GB-008: to-location must be queryable by location_code_exact.
	{
		url := fmt.Sprintf("/v1/inventory/inventory-locations?location_code_exact=%s", in.ToCode)
		w := doRequest(t, h, http.MethodGet, url, token, nil, http.StatusOK,
			"GB-008: GET inventory-locations by to_location_code returned non-200")

		var result struct {
			Items []struct {
				LocationID string `json:"location_id"`
			} `json:"items"`
			Total int `json:"total"`
		}
		mustDecodeJSON(t, w, &result)
		if result.Total == 0 || len(result.Items) == 0 {
			t.Fatalf("GB-008: to_location %q not found via location_code_exact filter (total=%d)", in.ToCode, result.Total)
		}
		if result.Items[0].LocationID == "" {
			t.Fatalf("GB-008: to_location %q item missing location_id in response", in.ToCode)
		}
	}

	// Step 6 — execute the in_transit transfer (completes status + decrements
	// source stock via DecrementQuantity; does NOT exercise FEFO — GB-015's
	// FEFO subquery lives in pickingapp and is covered by walkPick).
	// If seed quantity < transfer quantity this returns 422 (GB-011); the
	// harness fails loudly to signal the seed gap.
	{
		w := doRequest(t, h, http.MethodPost,
			"/v1/inventory/transfer-orders/"+idStr+"/execute",
			token, nil, http.StatusOK,
			"POST /execute returned non-200 — may be GB-011 (insufficient stock) or a transfer execute / DecrementQuantity regression. GB-015 (FEFO alias scope) is exercised by walkPick, not this step.")

		var to struct {
			Status string `json:"status"`
		}
		mustDecodeJSON(t, w, &to)
		if to.Status != "completed" {
			t.Fatalf("walkTransfer step 6: want status=completed after /execute, got %q", to.Status)
		}
	}
}

// walkReceive drives the floor-worker receive flow against the test mux:
//
//  1. GET  /v1/procurement/purchase-order-line-items/purchase-order/{po_id} → 200 (GB-006)
//  2. POST /v1/procurement/purchase-order-line-items/{id}/receive-quantity   → 200 per line
//  3. POST /v1/inventory/lot-trackings (lotFlow only)                        → 201 (GB-012)
//  4. GET  /v1/inventory/lot-trackings                                       → 200 (GB-014)
//
// GB-006: Step 1 exercises the purchase_order_line_items scenario filter.
// The ApplyScenarioFilter appends a WHERE clause on scenario_id; if the JOIN
// between purchase_order_line_items and an aliased table introduces an
// ambiguous column reference, this step 500s.
//
// GB-012: Step 3 exercises parseTimeFlexible in lottrackingsapp — the floor
// PWA sends manufacture_date as RFC3339 (time.RFC3339Nano). The regression
// caused a 400 InvalidArgument when only timeutil.FORMAT was accepted.
// A non-2xx response on the lot-tracking POST is a GB-012 regression suspect.
//
// GB-014: Step 4 exercises the lot_trackings Query with ApplyScenarioFilter
// active. The regression was an ambiguous column reference on scenario_id in
// the lot_trackings JOIN query (lt.scenario_id vs sp.scenario_id). A 500
// response on GET /v1/inventory/lot-trackings is the GB-014 regression signal.
//
// When in.POID == uuid.Nil (rush-receiving empty state.yaml), steps 1 and 2
// use a sentinel path that verifies the endpoint returns 200 for an unknown
// PO ID without panicking; the lot-tracking steps still fire regardless.
//
// The lotFlow parameter controls whether lot/serial-tracking steps fire:
//   - true:  receive-lot-tracking, receive-serial-tracking (GB-012 + GB-014)
//   - false: receive-rush-multi-line, receive-discrepancy, rush-receiving
func walkReceive(t *testing.T, h *apitest.Test, scenarioID uuid.UUID, in ReceiveInputs, lotFlow bool) {
	_ = scenarioID // retained for signature consistency with walkPick/walkCycleCount
	t.Helper()
	token := adminToken(t, h)

	// Fixed admin user UUID from seed.sql — used as received_by in receive-quantity
	// POST. The ZZZADMIN user (5cf37266-...) is the admin_gopher seeded in
	// seed.sql and is referenced as created_by/updated_by across all scenario
	// state.yaml files.
	const adminUserID = "5cf37266-3473-4006-984f-9325122678b7"

	// Step 1 — GB-006: GET line items by PO ID under an active scenario.
	// The endpoint exercises ApplyScenarioFilter on purchase_order_line_items.
	// When scenario_id is ambiguous in a JOIN, this returns 500. Assert 200
	// and that the item count matches the number of line items we discovered.
	//
	// For rush-receiving (no PO), we use uuid.Nil; the endpoint returns an
	// empty JSON array but must still return 200.
	{
		poIDStr := in.POID.String()
		url := "/v1/procurement/purchase-order-line-items/purchase-order/" + poIDStr
		w := doRequest(t, h, http.MethodGet, url, token, nil, http.StatusOK,
			"GB-006: GET purchase-order-line-items by PO returned non-200 — scenario filter on purchase_order_line_items may be broken")

		// The endpoint returns a JSON array directly (PurchaseOrderLineItems wraps []PurchaseOrderLineItem).
		var items []struct {
			ID string `json:"id"`
		}
		mustDecodeJSON(t, w, &items)

		if in.POID != uuid.Nil && len(items) != len(in.LineItems) {
			t.Fatalf("GB-006: GET purchase-order-line-items for PO %s returned %d items, want %d (scenario_id filter may be wrong)",
				poIDStr, len(items), len(in.LineItems))
		}
	}

	// Step 2 — per-line POST receive-quantity.
	// Each line is received for its full quantity_ordered. This verifies that
	// the receive-quantity endpoint works end-to-end and updates the line item
	// state in the DB.
	//
	// For receive-discrepancy the scenario seeds a line with quantity_ordered=50
	// and inventory on-hand=30 — the discrepancy is intentional (see
	// deployments/scenarios/receive-discrepancy/state.yaml). The receive-quantity
	// endpoint does not enforce stock availability; it only increments
	// quantity_received on the line item. A 200 here does NOT mean the physical
	// stock was decremented — that happens via a separate inventory-items update.
	// receive-discrepancy tests that the endpoint accepts a quantity larger than
	// on-hand without rejecting it (the discrepancy is surfaced to the user, not
	// blocked).
	for _, line := range in.LineItems {
		body := mustJSON(t, map[string]any{
			"quantity":    strconv.Itoa(line.ExpectedQty),
			"received_by": adminUserID,
		})
		url := "/v1/procurement/purchase-order-line-items/" + line.LineID.String() + "/receive-quantity"
		doRequest(t, h, http.MethodPost, url, token, body, http.StatusOK,
			fmt.Sprintf("POST receive-quantity for line %s (product %s qty %d) returned non-200",
				line.LineID, line.ProductID, line.ExpectedQty))
	}

	// Step 3 — GB-012: POST a new lot tracking with manufacture_date in RFC3339
	// format. Only fires when lotFlow == true (lot-tracking and serial-tracking
	// scenarios). The regression caused a 400 when the floor PWA sent RFC3339
	// instead of timeutil.FORMAT; parseTimeFlexible was the fix.
	//
	// We pick the first lot-tracked (or serial-tracked, which also has an
	// umbrella lot) line item's supplier_product_id. If no lot/serial-tracked
	// line exists, we use the first line's supplier_product_id as a fallback —
	// the POST validates supplier_product_id exists regardless of tracking type.
	if lotFlow {
		// Identify the supplier_product_id to use: prefer a lot/serial-tracked
		// line; fall back to the first line.
		spID := ""
		for _, li := range in.LineItems {
			if li.LotTracked || li.SerialTracked {
				spID = li.SupplierProductID.String()
				break
			}
		}
		if spID == "" && len(in.LineItems) > 0 {
			spID = in.LineItems[0].SupplierProductID.String()
		}

		if spID != "" {
			// GB-012: manufacture_date MUST be RFC3339 (time.RFC3339Nano) to
			// reproduce the regression trigger path. If parseTimeFlexible regresses
			// back to timeutil.FORMAT-only, this POST returns 400.
			manufactureDate := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339Nano)
			expirationDate := time.Now().Add(365 * 24 * time.Hour).UTC().Format(time.RFC3339Nano)
			receivedDate := time.Now().UTC().Format(time.RFC3339Nano)

			body := mustJSON(t, map[string]any{
				"supplier_product_id": spID,
				"lot_number":          "LOT-HARNESS-" + time.Now().Format("20060102150405"),
				"manufacture_date":    manufactureDate,
				"expiration_date":     expirationDate,
				"received_date":       receivedDate,
				"quantity":            "1",
				"quality_status":      "good",
			})
			// The lot-trackings create handler returns 200 (the web framework
			// defaults to StatusOK; only nil returns trigger 204). A 400 here
			// is the GB-012 regression signal — parseTimeFlexible rejected RFC3339.
			doRequest(t, h, http.MethodPost, "/v1/inventory/lot-trackings", token, body, http.StatusOK,
				"GB-012: POST /v1/inventory/lot-trackings with RFC3339 manufacture_date returned non-200 — parseTimeFlexible may have regressed to timeutil.FORMAT-only")
		}
	}

	// Step 4 — GB-014: GET lot-trackings while a scenario is active.
	// The regression was an ambiguous column reference on scenario_id in the
	// lot_trackings query JOIN (lt.scenario_id vs sp.scenario_id) that caused
	// 500 when ScenariosEnabled == true. A 200 here confirms the fix holds.
	//
	// We do NOT assert a specific lot count because rush-receiving and
	// receive-discrepancy don't seed any lot_trackings rows; the important
	// invariant is that the endpoint does NOT 500 — it returns 200 with items
	// or total = 0.
	{
		doRequest(t, h, http.MethodGet, "/v1/inventory/lot-trackings?page=1&rows=10", token, nil, http.StatusOK,
			"GB-014: GET /v1/inventory/lot-trackings returned non-200 — ambiguous scenario_id column reference may have reappeared in lot_trackings JOIN")
	}
}

// walkPick drives the floor-worker pick flow against the test mux:
//
//  1. GET  /v1/sales/order-line-items?order_id={soID}           → 200 (GB-007)
//  2. POST /v1/sales/order-line-items/{lineItemID}/pick-quantity → 200 per task (GB-015)
//
// GB-007 (NULL discount_type COALESCE): Step 1 exercises the order-line-items
// Query with a scenario filter active. Scenario state.yaml files do not set
// discount_type, leaving the column NULL. Before fix 19f7cb80, orderlineitemsdb
// lacked COALESCE on that column; the NULL scan caused sqlx to return
// "converting NULL to string is unsupported" → 500. Step 1 returning 200
// confirms the COALESCE is in place.
//
// GB-015 (FEFO subquery alias scope): Step 2 exercises
// pickingapp.PickQuantity, which calls
// inventoryitembus.QueryAvailableForAllocation with strategy="fefo". Before fix
// f985029a, queryFEFO used ii.* column qualifiers inside the inner subquery
// (aliased as `sub`), making Postgres resolve them against the outer ii row —
// a correlated subquery that never matched the target product+location. The
// result was 0 inventory items found → 422 "insufficient stock". Step 2
// returning 200 confirms the sub.* qualification is in place.
//
// When in.SOID == uuid.Nil (scenario has no sales orders, e.g. e2e-pick-strict
// which carries only lever overrides), step 2 is skipped entirely and only step
// 1 fires with a sentinel order_id. The endpoint must return 200 with an empty
// list — not 404 or 500.
//
// The scenarioID parameter is unused but retained for signature consistency with
// walkTransfer, walkReceive, walkCycleCount — Task 12's table-driven dispatch
// calls all family walks with the same 4-arg shape.
func walkPick(t *testing.T, h *apitest.Test, scenarioID uuid.UUID, in PickInputs) {
	_ = scenarioID // signature-consistent; see doc comment
	t.Helper()
	token := adminToken(t, h)

	// Fixed admin user UUID from seed.sql — used as picked_by in pick-quantity
	// POST. The ZZZADMIN user (5cf37266-...) is the admin_gopher seeded in
	// seed.sql and referenced as created_by/updated_by across all scenario
	// state.yaml files.
	const adminUserID = "5cf37266-3473-4006-984f-9325122678b7"

	// Step 1 — GB-007: GET order line items for this sales order while an active
	// scenario is set. Scenario rows have discount_type = NULL; the orderlineitemsdb
	// SELECT must COALESCE(discount_type, 'flat') to avoid a NULL→string scan
	// error. A 500 response is the GB-007 regression signal.
	//
	// For e2e-pick-strict (no sales order), use uuid.Nil as the sentinel order_id.
	// The endpoint must return 200 with an empty list for an unknown order ID.
	{
		orderIDStr := in.SOID.String()
		if in.SOID == uuid.Nil {
			orderIDStr = uuid.Nil.String()
		}
		url := fmt.Sprintf("/v1/sales/order-line-items?order_id=%s&page=1&rows=100", orderIDStr)
		w := doRequest(t, h, http.MethodGet, url, token, nil, http.StatusOK,
			"GB-007: GET /v1/sales/order-line-items returned non-200 — COALESCE(discount_type,'flat') may have been removed from the orderlineitemsdb SELECT")

		// The order-line-items endpoint returns a paginated envelope:
		//   {"items": [...], "total": N, "page": N, "rows_per_page": N}
		var result struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			Total int `json:"total"`
		}
		mustDecodeJSON(t, w, &result)

		// When a real sales order is present, the line item count must be > 0.
		// For the sentinel uuid.Nil case we accept an empty list.
		if in.SOID != uuid.Nil && len(result.Items) == 0 {
			t.Fatalf("GB-007: GET order-line-items for order %s returned 0 items — scenario fixtures may not have loaded or order_id filter is broken", in.SOID)
		}
	}

	// Step 2 — GB-015: per-task, POST pick-quantity for each pick task.
	// pickingapp.PickQuantity calls inventoryitembus.QueryAvailableForAllocation
	// with strategy="fefo" and the location_id from the task. The FEFO subquery
	// must use sub.* column qualifiers (not ii.*); an ii.* qualifier makes the
	// filter a correlated reference against the outer row, which almost never
	// matches the targeted product+location. A 422 with
	// "insufficient stock at specified location" is the GB-015 regression signal
	// even when inventory clearly exists.
	//
	// Skipped entirely when in.SOID == uuid.Nil (lever-only scenario, no tasks).
	for i, alloc := range in.Allocations {
		lineItemIDStr := alloc.PickTaskID.String() // PickTaskID holds the line-item UUID

		// Resolve the inventory location UUID so we can pass it in the request body.
		// The walk does this via discoverPickInputs already storing LocationCode;
		// we need the UUID — stored as LocationID in the alloc.ProductID's inventory
		// record. We look it up inline by location_code + scenario filter-compatible
		// path via the inventory-locations endpoint.
		url := fmt.Sprintf("/v1/inventory/inventory-locations?location_code_exact=%s", alloc.LocationCode)
		locW := doRequest(t, h, http.MethodGet, url, token, nil, http.StatusOK,
			fmt.Sprintf("walkPick step 2 pre-check: GET inventory-locations for %q returned non-200", alloc.LocationCode))

		var locResult struct {
			Items []struct {
				LocationID string `json:"location_id"`
			} `json:"items"`
			Total int `json:"total"`
		}
		mustDecodeJSON(t, locW, &locResult)
		if locResult.Total == 0 || len(locResult.Items) == 0 {
			t.Fatalf("walkPick step 2: location %q not found via location_code_exact filter (task index %d)", alloc.LocationCode, i)
		}
		locationID := locResult.Items[0].LocationID

		body := mustJSON(t, map[string]any{
			"quantity":    strconv.Itoa(alloc.Qty),
			"picked_by":   adminUserID,
			"location_id": locationID,
		})
		pickURL := "/v1/sales/order-line-items/" + lineItemIDStr + "/pick-quantity"
		w := doRequest(t, h, http.MethodPost, pickURL, token, body, http.StatusOK,
			fmt.Sprintf("GB-015: POST /v1/sales/order-line-items/%s/pick-quantity returned non-200 — FEFO subquery may have reverted to ii.* qualifier (sub.* required in queryFEFO inner subquery). Check: product=%s location=%s qty=%d",
				lineItemIDStr, alloc.ProductID, alloc.LocationCode, alloc.Qty))

		var result struct {
			PickedQuantity string `json:"picked_quantity"`
		}
		mustDecodeJSON(t, w, &result)
		pickedQty, _ := strconv.Atoi(result.PickedQuantity)
		if pickedQty <= 0 {
			t.Fatalf("GB-015: pick-quantity for line item %s returned picked_quantity=%q, want > 0", lineItemIDStr, result.PickedQuantity)
		}
	}
}
