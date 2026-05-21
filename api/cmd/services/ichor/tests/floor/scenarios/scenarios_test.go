package scenarios_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// customRowOverrides supplies the Custom closures for scenarios that don't fit
// a family helper. Profile scenarios are config-only (no entities) and
// e2e-baseline is intentionally empty; all three require Custom handlers rather
// than a family walk. See walkProfileWithReceive, walkProfileWithTransfer, and
// walkE2EBaseline in walk_test.go for rationale.
//
// Note: e2e-pick-strict is NOT listed here. discoverScenarios resolves it to
// familyPick (via familyOverrides in harness_test.go). walkPick handles the
// lever-only / no-sales-order case gracefully: when in.SOID == uuid.Nil,
// step 2 (the per-task loop) is a no-op and only the GB-007 endpoint smoke
// fires against the sentinel uuid.Nil order_id.
func customRowOverrides() map[string]ScenarioRow {
	return map[string]ScenarioRow{
		"profile-strict-regulated": {
			Name:     "profile-strict-regulated",
			Category: "generic",
			Custom: func(t *testing.T, h *apitest.Test, db *sqlx.DB, sid uuid.UUID) {
				walkProfileWithReceive(t, h, db, sid)
			},
		},
		"profile-medical-device-rental": {
			Name:     "profile-medical-device-rental",
			Category: "generic",
			Custom: func(t *testing.T, h *apitest.Test, db *sqlx.DB, sid uuid.UUID) {
				walkProfileWithTransfer(t, h, db, sid)
			},
		},
		"e2e-baseline": {
			Name:     "e2e-baseline",
			Category: "generic",
			Custom: func(t *testing.T, h *apitest.Test, db *sqlx.DB, sid uuid.UUID) {
				walkE2EBaseline(t, h, db, sid)
			},
		},
	}
}

// rowsForTest discovers all 21 scenarios and overlays Custom closures from
// customRowOverrides for scenarios that can't be dispatched to a family walk.
func rowsForTest(t *testing.T) []ScenarioRow {
	t.Helper()
	discovered, err := discoverScenarios(scenarioRoots())
	if err != nil {
		t.Fatalf("discoverScenarios: %v", err)
	}
	overrides := customRowOverrides()
	for i, r := range discovered {
		if over, ok := overrides[r.Name]; ok {
			discovered[i] = over
		}
	}
	return discovered
}

// lotFlowScenarios lists the receive scenarios that should fire GB-012 (RFC3339
// manufacture_date POST) and GB-014 (lot-trackings GET) in walkReceive.
// lotFlow=true is passed to walkReceive for any scenario whose name appears here.
var lotFlowScenarios = map[string]bool{
	"receive-lot-tracking":    true,
	"receive-serial-tracking": true,
}

// TestFloorScenarios is the unified table-driven replacement for the 21
// individual TestFloorScenarios_Xxx canary functions. It discovers all scenarios
// via discoverScenarios, overlays Custom closures for profile + e2e-baseline
// scenarios, then runs each row as a parallel sub-test.
//
// Parallelism: -parallel 2 is the safe ceiling on macOS at 21-scenario scale.
// -parallel 4 saturates ephemeral source ports (49152–65535) and fails with
// "can't assign requested address" on the dbtest postgres connection.
// parallel-2 keeps the connection rate under the port budget; 21/21 sub-tests
// PASS in ~207s.
//
// Dispatch rules:
//
//	Custom != nil  → row.Custom(t, h, h.DB.DB, scenarioID)
//	familyReceive  → walkReceive with lotFlow derived from lotFlowScenarios
//	familyPick     → walkPick (handles uuid.Nil SOID for lever-only scenarios)
//	familyTransfer → walkTransfer, LotTracked=true for *-lot-tracked suffix
//	familyCycleCount → walkCycleCount, VarianceMode set from scenario name suffix
//	familyProfile  → t.Fatalf (profile-* must have Custom handlers)
//	""             → t.Fatalf (add to customRowOverrides)
func TestFloorScenarios(t *testing.T) {
	rows := rowsForTest(t)
	for _, row := range rows {
		row := row
		t.Run(row.Name, func(t *testing.T) {
			t.Parallel()
			h := startScenarioTest(t, row.Name)
			scenarioID := loadScenarioFixtures(t, h, row.Name)

			if row.Custom != nil {
				row.Custom(t, h, h.DB.DB, scenarioID)
				return
			}

			switch row.Family {
			case familyReceive:
				walkReceive(t, h, scenarioID, discoverReceiveInputs(t, h, scenarioID), lotFlowScenarios[row.Name])

			case familyPick:
				walkPick(t, h, scenarioID, discoverPickInputs(t, h, scenarioID))

			case familyTransfer:
				in := discoverTransferInputs(t, h, scenarioID)
				if strings.HasSuffix(row.Name, "-lot-tracked") {
					in.LotTracked = true
				}
				walkTransfer(t, h, scenarioID, in)

			case familyCycleCount:
				in := discoverCycleCountInputs(t, h, scenarioID)
				switch {
				case strings.HasSuffix(row.Name, "-variance-over"):
					in.VarianceMode = "over"
				case strings.HasSuffix(row.Name, "-variance-under"):
					in.VarianceMode = "under"
				}
				walkCycleCount(t, h, scenarioID, in)

			case familyProfile:
				t.Fatalf("profile-* scenario %q reached family dispatch — it must have a Custom handler in customRowOverrides", row.Name)

			case "":
				t.Fatalf("scenario %q has empty family and no Custom handler — add it to customRowOverrides in scenarios_test.go", row.Name)

			default:
				t.Fatalf("unknown family %q for scenario %q — update dispatch switch in TestFloorScenarios", row.Family, row.Name)
			}
		})
	}
}
