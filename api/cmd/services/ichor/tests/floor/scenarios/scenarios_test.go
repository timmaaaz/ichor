package scenarios_test

import (
	"testing"
)

// TestFloorScenarios_TransferIntraZone is the canary test for the transfer
// family. It provisions a fresh docker-backed Postgres database, seeds all 21
// scenarios, loads the transfer-intra-zone fixture set, discovers the transfer
// order and associated location codes via direct DB query, then walks the
// canonical 6-step floor-worker transfer through the live HTTP mux with
// ScenariosEnabled: true.
//
// GB regressions this test guards:
//   - GB-008: location-code scan (inventory-locations endpoint must return
//     items by location_code_exact for both source and destination codes)
//   - GB-010: status transitions (pending→approved→in_transit) must not return
//     4xx; a broken /approve or /claim handler surfaces here before /execute
//
// Note: GB-015 (FEFO subquery alias scope) lives in pickingapp, not
// transferorderapp. Transfer /execute calls DecrementQuantity directly.
// GB-015 coverage is in walkPick (Task 9), not here.
//
// GB-011 (insufficient stock): if the scenario seeds transfer quantity > source
// stock, /execute returns 422. That is a seed-correctness issue, not a code
// regression. The harness signals this by failing the canary loudly — the
// Playwright walks (Phase B) PATCH the quantity to work around it, but the
// harness intentionally surfaces the gap. transfer-intra-zone seeds matched
// stock (50 units at STG-A01, transfer qty 50) → no GB-011 surface here.
func TestFloorScenarios_TransferIntraZone(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "transfer-intra-zone")
	scenarioID := loadScenarioFixtures(t, h, "transfer-intra-zone")
	in := discoverTransferInputs(t, h, scenarioID)
	walkTransfer(t, h, scenarioID, in)
}

func TestFloorScenarios_TransferCrossZone(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "transfer-cross-zone")
	scenarioID := loadScenarioFixtures(t, h, "transfer-cross-zone")
	in := discoverTransferInputs(t, h, scenarioID)
	walkTransfer(t, h, scenarioID, in)
}

func TestFloorScenarios_TransferLotTracked(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "transfer-lot-tracked")
	scenarioID := loadScenarioFixtures(t, h, "transfer-lot-tracked")
	in := discoverTransferInputs(t, h, scenarioID)
	in.LotTracked = true
	walkTransfer(t, h, scenarioID, in)
}

func TestFloorScenarios_TransferMultiLine(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "transfer-multi-line")
	scenarioID := loadScenarioFixtures(t, h, "transfer-multi-line")
	// Multi-line: discoverTransferInputs returns the first approved transfer;
	// additional transfers are not covered by this single walk. If we need
	// full coverage of multi-line, file as a follow-up — out of Phase A scope.
	in := discoverTransferInputs(t, h, scenarioID)
	walkTransfer(t, h, scenarioID, in)
}

// TestFloorScenarios_ReceiveRushMultiLine is the canary for the receive family.
// It exercises a 4-line PO (2 non-tracked + 2 lot-tracked SKUs) under the
// receive-rush-multi-line scenario.
//
// GB regressions this test guards:
//   - GB-006: purchase_order_line_items scenario filter (step 1 GET)
//   - GB-014: lot_trackings scenario filter / JOIN ambiguity (step 4 GET)
//
// lotFlow=false: this scenario's lot-tracked lines test the receive-quantity
// path but we skip the explicit lot-tracking POST (GB-012) — those lines
// have pre-seeded lot_trackings rows in state.yaml and do not exercise the
// RFC3339 parse path. GB-012 coverage is in the lot/serial-tracking canaries.
func TestFloorScenarios_ReceiveRushMultiLine(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "receive-rush-multi-line")
	scenarioID := loadScenarioFixtures(t, h, "receive-rush-multi-line")
	walkReceive(t, h, scenarioID, discoverReceiveInputs(t, h, scenarioID), false)
}

// TestFloorScenarios_ReceiveLotTracking exercises the lot-tracking receive
// path with lotFlow=true, which fires the RFC3339 manufacture_date POST
// (GB-012) and the lot-trackings GET (GB-014).
//
// GB regressions this test guards:
//   - GB-006: purchase_order_line_items scenario filter (step 1 GET)
//   - GB-012: parseTimeFlexible accepts RFC3339 manufacture_date (step 3 POST)
//   - GB-014: lot_trackings scenario filter / JOIN ambiguity (step 4 GET)
func TestFloorScenarios_ReceiveLotTracking(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "receive-lot-tracking")
	scenarioID := loadScenarioFixtures(t, h, "receive-lot-tracking")
	walkReceive(t, h, scenarioID, discoverReceiveInputs(t, h, scenarioID), true)
}

// TestFloorScenarios_ReceiveSerialTracking exercises the serial-tracking
// receive path. The scenario seeds an umbrella lot row for the serial numbers;
// lotFlow=true fires the lot-tracking POST with RFC3339 dates (GB-012).
//
// GB regressions this test guards:
//   - GB-006: purchase_order_line_items scenario filter (step 1 GET)
//   - GB-012: parseTimeFlexible accepts RFC3339 manufacture_date (step 3 POST)
//   - GB-014: lot_trackings scenario filter / JOIN ambiguity (step 4 GET)
func TestFloorScenarios_ReceiveSerialTracking(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "receive-serial-tracking")
	scenarioID := loadScenarioFixtures(t, h, "receive-serial-tracking")
	walkReceive(t, h, scenarioID, discoverReceiveInputs(t, h, scenarioID), true)
}

// TestFloorScenarios_ReceiveDiscrepancy exercises the discrepancy receive path
// where the line item quantity_ordered (50) exceeds on-hand inventory (30).
// The receive-quantity endpoint does not enforce stock availability — it
// increments quantity_received on the line item and returns 200. The
// discrepancy is intentional and does not cause a 4xx/5xx.
//
// GB regressions this test guards:
//   - GB-006: purchase_order_line_items scenario filter (step 1 GET)
//   - GB-014: lot_trackings scenario filter / JOIN ambiguity (step 4 GET)
func TestFloorScenarios_ReceiveDiscrepancy(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "receive-discrepancy")
	scenarioID := loadScenarioFixtures(t, h, "receive-discrepancy")
	walkReceive(t, h, scenarioID, discoverReceiveInputs(t, h, scenarioID), false)
}

// TestFloorScenarios_RushReceiving exercises the rush-receiving scenario, which
// ships with an empty state.yaml (no purchase orders). discoverReceiveInputs
// returns ReceiveInputs{POID: uuid.Nil, LineItems: nil}.
//
// walkReceive uses the sentinel PO ID to verify that the purchase-order-line-items
// endpoint returns 200 with an empty list for an unknown PO (not 404 or 500),
// then runs the lot-trackings smoke GET (GB-014).
//
// GB regressions this test guards:
//   - GB-006: purchase_order_line_items endpoint returns 200 for unknown PO
//   - GB-014: lot_trackings scenario filter / JOIN ambiguity (step 4 GET)
func TestFloorScenarios_RushReceiving(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "rush-receiving")
	scenarioID := loadScenarioFixtures(t, h, "rush-receiving")
	walkReceive(t, h, scenarioID, discoverReceiveInputs(t, h, scenarioID), false)
}

// TestFloorScenarios_PickWholeOrder is the canary for a full, no-exception pick.
// The scenario seeds 9 pick tasks across 6 products and 2 staging locations.
// walkPick step 1 (GB-007) verifies order-line-items returns 200 with scenario's
// NULL discount_type COALESCEd. Step 2 (GB-015) verifies the FEFO subquery
// uses sub.* qualifiers so QueryAvailableForAllocation returns inventory rows.
//
// GB regressions this test guards:
//   - GB-007: COALESCE(discount_type,'flat') in order-line-items SELECT (step 1)
//   - GB-015: FEFO sub.* column qualifiers in queryFEFO inner subquery (step 2)
func TestFloorScenarios_PickWholeOrder(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "pick-whole-order")
	scenarioID := loadScenarioFixtures(t, h, "pick-whole-order")
	walkPick(t, h, scenarioID, discoverPickInputs(t, h, scenarioID))
}

// TestFloorScenarios_PickShortPick exercises the short-pick scenario where
// quantity_to_pick (50) < quantity ordered (100). walkPick uses quantity_to_pick
// as the pick quantity so the POST succeeds (50 units available in inventory).
// The scenario validates that walkPick handles below-order-quantity picks
// without 422 from the "quantity exceeds remaining" guard.
//
// GB regressions this test guards:
//   - GB-007: COALESCE(discount_type,'flat') in order-line-items SELECT (step 1)
//   - GB-015: FEFO sub.* column qualifiers in queryFEFO inner subquery (step 2)
func TestFloorScenarios_PickShortPick(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "pick-short-pick")
	scenarioID := loadScenarioFixtures(t, h, "pick-short-pick")
	walkPick(t, h, scenarioID, discoverPickInputs(t, h, scenarioID))
}

// TestFloorScenarios_PickLotTracked exercises picking a lot-tracked product
// (SKU-0029). The FEFO strategy is critical here: lot-tracked products must
// be allocated from the earliest-expiry lot. GB-015 is most impactful for
// lot-tracked products because the subquery joins against lot_trackings to
// find expiry dates.
//
// GB regressions this test guards:
//   - GB-007: COALESCE(discount_type,'flat') in order-line-items SELECT (step 1)
//   - GB-015: FEFO sub.* column qualifiers in queryFEFO inner subquery (step 2)
func TestFloorScenarios_PickLotTracked(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "pick-lot-tracked")
	scenarioID := loadScenarioFixtures(t, h, "pick-lot-tracked")
	walkPick(t, h, scenarioID, discoverPickInputs(t, h, scenarioID))
}

// TestFloorScenarios_PickZoneSliced exercises picking across 4 tasks spread
// across 3 staging zones (STG-A01, STG-A02, STG-B01, STG-B02). Verifies that
// the FEFO query handles multiple location_ids correctly across calls — the
// sub.* fix must hold for each location independently.
//
// GB regressions this test guards:
//   - GB-007: COALESCE(discount_type,'flat') in order-line-items SELECT (step 1)
//   - GB-015: FEFO sub.* column qualifiers in queryFEFO inner subquery (step 2)
func TestFloorScenarios_PickZoneSliced(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "pick-zone-sliced")
	scenarioID := loadScenarioFixtures(t, h, "pick-zone-sliced")
	walkPick(t, h, scenarioID, discoverPickInputs(t, h, scenarioID))
}

// TestFloorScenarios_E2EPickStrict exercises the e2e-pick-strict lever-variant
// scenario. This scenario carries only lever overrides (pick.lotScan=required-
// if-lot-tracked) and no state.yaml, so discoverPickInputs returns
// PickInputs{SOID: uuid.Nil, Allocations: nil}. walkPick step 1 fires with a
// sentinel order_id to verify the endpoint returns 200+empty-list rather than
// 404 or 500. Step 2 is skipped (no tasks). The scenario validates that
// walkPick gracefully handles lever-only scenarios.
//
// GB regressions this test guards:
//   - GB-007: order-line-items endpoint returns 200 for unknown order_id
func TestFloorScenarios_E2EPickStrict(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "e2e-pick-strict")
	scenarioID := loadScenarioFixtures(t, h, "e2e-pick-strict")
	walkPick(t, h, scenarioID, discoverPickInputs(t, h, scenarioID))
}

// =============================================================================
// Cycle-count family canaries (Task 10)
//
// These 4 tests lock in the cycle-count baseline. Cycle-count was Track-E-clean
// on 2026-05-19 so no GB regression labels are asserted. The purpose is to catch
// future regressions: any breakage in the cycle-count session lifecycle (draft →
// in_progress → items counted → completed) will surface here before reaching
// the Playwright walks.
//
// walkCycleCount applies variance inside itself based on in.VarianceMode:
//   "over"  → actualQty = expectedQty + 10 for every item
//   "under" → actualQty = max(expectedQty - 10, 0) for every item
//   ""      → actualQty = expectedQty (no variance, zero-delta adjustments)
// =============================================================================

// TestFloorScenarios_CycleCountVarianceOver exercises an over-count: the worker
// enters 10 more than the system quantity for each item. The complete() handler
// creates positive inventory adjustments for the variance and marks the session
// completed.
func TestFloorScenarios_CycleCountVarianceOver(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "cycle-count-variance-over")
	scenarioID := loadScenarioFixtures(t, h, "cycle-count-variance-over")
	in := discoverCycleCountInputs(t, h, scenarioID)
	in.VarianceMode = "over"
	walkCycleCount(t, h, scenarioID, in)
}

// TestFloorScenarios_CycleCountVarianceUnder exercises an under-count: the
// worker enters 10 less than the system quantity. The complete() handler creates
// negative inventory adjustments and marks the session completed.
func TestFloorScenarios_CycleCountVarianceUnder(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "cycle-count-variance-under")
	scenarioID := loadScenarioFixtures(t, h, "cycle-count-variance-under")
	in := discoverCycleCountInputs(t, h, scenarioID)
	in.VarianceMode = "under"
	walkCycleCount(t, h, scenarioID, in)
}

// TestFloorScenarios_CycleCountMultiItem exercises a 5-item count at a single
// location (PCK-01, SKUs 0001–0005). No variance — the walk submits exact counts
// to verify that the complete() path handles multi-item sessions cleanly.
func TestFloorScenarios_CycleCountMultiItem(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "cycle-count-multi-item")
	scenarioID := loadScenarioFixtures(t, h, "cycle-count-multi-item")
	walkCycleCount(t, h, scenarioID, discoverCycleCountInputs(t, h, scenarioID))
}

// TestFloorScenarios_CycleCountScheduled exercises a pre-scheduled session
// (status=draft, 3 items across 3 different locations). No variance — tests that
// the walk correctly handles multi-location sessions where each item has a
// distinct locationId.
func TestFloorScenarios_CycleCountScheduled(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "cycle-count-scheduled")
	scenarioID := loadScenarioFixtures(t, h, "cycle-count-scheduled")
	walkCycleCount(t, h, scenarioID, discoverCycleCountInputs(t, h, scenarioID))
}
