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
