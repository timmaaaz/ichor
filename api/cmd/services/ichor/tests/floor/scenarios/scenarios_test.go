package scenarios_test

import (
	"testing"
)

// TestFloorScenarios_TransferIntraZone is the canary test for the transfer
// family. It provisions a fresh docker-backed Postgres database, seeds all 21
// scenarios, loads the transfer-intra-zone fixture set, discovers the transfer
// order and associated location codes via direct DB query, then walks the
// canonical 5-step floor-worker transfer through the live HTTP mux with
// ScenariosEnabled: true.
//
// GB regressions this test guards:
//   - GB-008: location-code scan (inventory-locations endpoint must return
//     items by location_code_exact for both source and destination codes)
//   - GB-010: status transitions (pending→approved→in_transit) must not return
//     4xx; a broken /approve or /claim handler surfaces here before /execute
//   - GB-015: /execute must succeed when source stock >= transfer quantity;
//     a broken Execute app or a scenario with mis-seeded stock surfaces here
//
// GB-011 (insufficient stock): if the scenario seeds transfer quantity > source
// stock, /execute returns 422. That is a seed correctness issue, not a code
// regression. The skip comment below documents the expected behaviour and
// why the test is skipped rather than failing in that case.
//
// Note: the transfer-intra-zone scenario seeds 50 units at STG-A01 and
// requests a transfer of 50 units → stock exactly meets demand → no GB-011.
func TestFloorScenarios_TransferIntraZone(t *testing.T) {
	t.Parallel()
	h := startScenarioTest(t, "transfer-intra-zone")
	scenarioID := loadScenarioFixtures(t, h, "transfer-intra-zone")
	in := discoverTransferInputs(t, h, scenarioID)
	walkTransfer(t, h, scenarioID, in)
}
