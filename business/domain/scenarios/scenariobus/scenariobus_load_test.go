package scenariobus_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus/stores/scenariodb"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
)

// Test_Load_AppliesLeverOverrides_AndWorkerZones runs the full Load path
// against a real DB. It writes a temporary scenario directory on disk, seeds
// the DB via SeedScenariosFromRoot, calls Load, and asserts both
// override-visibility (via settingsdb's scenario JOIN) and worker-zone effects.
func Test_Load_AppliesLeverOverrides_AndWorkerZones(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Load_AppliesLeverOverrides_AndWorkerZones")
	ctx := context.Background()

	// Baseline seed — populates config.settings (17 levers) plus all FK
	// dependencies required by the user bus.
	if err := dbtest.InsertSeedDataWithDB(db.Log, db.DB); err != nil {
		t.Fatalf("baseline seed: %v", err)
	}

	// Create 1 user with a known username so the bindings.yaml can reference it.
	users, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.User, db.BusDomain.User)
	if err != nil {
		t.Fatalf("seed worker user: %v", err)
	}
	workerUsername := users[0].Username.String()

	// Build temporary scenario directory: one scenario subdirectory under
	// scenariosRoot containing scenario.yaml + bindings.yaml (no state.yaml).
	scenariosRoot := t.TempDir()
	scenarioDir := filepath.Join(scenariosRoot, "test-levers")
	if err := os.MkdirAll(scenarioDir, 0o755); err != nil {
		t.Fatalf("mkdir scenario dir: %v", err)
	}

	scenarioYAML := "name: test-levers\ndescription: B5 + B6 integration test\nlever_overrides:\n  pick.lotScan: required-if-lot-tracked\n  pick.destinationScan: required\n  cycleCount.locationScan: button-confirm\n"
	if err := os.WriteFile(filepath.Join(scenarioDir, "scenario.yaml"), []byte(scenarioYAML), 0o644); err != nil {
		t.Fatalf("write scenario.yaml: %v", err)
	}

	bindingsYAML := fmt.Sprintf("workers:\n  - username: %s\n    zones: [PCK, STG-Z]\n", workerUsername)
	if err := os.WriteFile(filepath.Join(scenarioDir, "bindings.yaml"), []byte(bindingsYAML), 0o644); err != nil {
		t.Fatalf("write bindings.yaml: %v", err)
	}

	// Seed scenario row + lever_override rows into the DB from the temp dir.
	if err := dbtest.SeedScenariosFromRoot(ctx, db.BusDomain, scenariosRoot); err != nil {
		t.Fatalf("seed temp scenario: %v", err)
	}

	// Look up the scenario ID by name using the default (empty-root) bus.
	target, err := db.BusDomain.Scenario.QueryByName(ctx, "test-levers")
	if err != nil {
		t.Fatalf("query scenario by name: %v", err)
	}

	// Build a fresh scenariobus.Business wired to the temp scenariosRoot so
	// Load can read worker zones from bindings.yaml.  The default
	// db.BusDomain.Scenario was constructed with "" (no-op worker zone path).
	scenarioBus := scenariobus.NewBusiness(
		db.Log,
		db.BusDomain.Delegate,
		scenariodb.NewStore(db.Log, db.DB),
		sqldb.NewBeginner(db.DB),
		scenariosRoot,
	)

	// Load the scenario — applies fixtures (none here), sets scenarios_active,
	// and applies worker zone bindings.
	if err := scenarioBus.Load(ctx, target.ID); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Assert: override visible via settingsdb's scenario JOIN.
	// pick.lotScan default is "disabled"; override must win.
	lotScanSetting, err := db.BusDomain.Settings.QueryByKey(ctx, "pick.lotScan")
	if err != nil {
		t.Fatalf("query pick.lotScan: %v", err)
	}
	var lotScanValue string
	if err := json.Unmarshal(lotScanSetting.Value, &lotScanValue); err != nil {
		t.Fatalf("unmarshal pick.lotScan value: %v", err)
	}
	if lotScanValue != "required-if-lot-tracked" {
		t.Errorf("pick.lotScan = %q, want %q (override should win)", lotScanValue, "required-if-lot-tracked")
	}

	// Assert: B6 new-lever override visible.
	// cycleCount.locationScan default is "required"; override must win.
	cycleCountSetting, err := db.BusDomain.Settings.QueryByKey(ctx, "cycleCount.locationScan")
	if err != nil {
		t.Fatalf("query cycleCount.locationScan: %v", err)
	}
	var cycleCountValue string
	if err := json.Unmarshal(cycleCountSetting.Value, &cycleCountValue); err != nil {
		t.Fatalf("unmarshal cycleCount.locationScan value: %v", err)
	}
	if cycleCountValue != "button-confirm" {
		t.Errorf("cycleCount.locationScan = %q, want %q (override should win)", cycleCountValue, "button-confirm")
	}

	// Assert: non-overridden key returns base default.
	// pick.assignmentGranularity was NOT overridden; base default is "whole-order".
	granSetting, err := db.BusDomain.Settings.QueryByKey(ctx, "pick.assignmentGranularity")
	if err != nil {
		t.Fatalf("query pick.assignmentGranularity: %v", err)
	}
	var granValue string
	if err := json.Unmarshal(granSetting.Value, &granValue); err != nil {
		t.Fatalf("unmarshal pick.assignmentGranularity value: %v", err)
	}
	if granValue != "whole-order" {
		t.Errorf("pick.assignmentGranularity = %q, want %q (base default should hold)", granValue, "whole-order")
	}

	// Assert: worker zones were applied to the user.
	uname := userbus.MustParseName(workerUsername)
	queried, err := db.BusDomain.User.Query(
		ctx,
		userbus.QueryFilter{Username: &uname},
		userbus.DefaultOrderBy,
		page.MustParse("1", "10"),
	)
	if err != nil {
		t.Fatalf("user query: %v", err)
	}
	if len(queried) == 0 {
		t.Fatalf("user %q not found after Load", workerUsername)
	}
	gotZones := queried[0].AssignedZones
	wantZones := []string{"PCK", "STG-Z"}
	if !slices.Equal(gotZones, wantZones) {
		t.Errorf("user %q assigned_zones = %v, want %v", workerUsername, gotZones, wantZones)
	}

	// Reset re-applies the active scenario via Load. The DoD asserts "Scenario
	// reset restores default config state" — for B5, that means override
	// visibility and worker zones survive a Reset round-trip unchanged.
	if err := scenarioBus.Reset(ctx); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	// Re-assert override still wins post-Reset.
	lotScanSetting2, err := db.BusDomain.Settings.QueryByKey(ctx, "pick.lotScan")
	if err != nil {
		t.Fatalf("query pick.lotScan post-Reset: %v", err)
	}
	var lotScanValue2 string
	if err := json.Unmarshal(lotScanSetting2.Value, &lotScanValue2); err != nil {
		t.Fatalf("unmarshal pick.lotScan value post-Reset: %v", err)
	}
	if lotScanValue2 != "required-if-lot-tracked" {
		t.Errorf("post-Reset pick.lotScan = %q, want %q (override should still win)", lotScanValue2, "required-if-lot-tracked")
	}

	// Re-assert B6 new-lever override still wins post-Reset.
	cycleCountSetting2, err := db.BusDomain.Settings.QueryByKey(ctx, "cycleCount.locationScan")
	if err != nil {
		t.Fatalf("query cycleCount.locationScan post-Reset: %v", err)
	}
	var cycleCountValue2 string
	if err := json.Unmarshal(cycleCountSetting2.Value, &cycleCountValue2); err != nil {
		t.Fatalf("unmarshal cycleCount.locationScan value post-Reset: %v", err)
	}
	if cycleCountValue2 != "button-confirm" {
		t.Errorf("post-Reset cycleCount.locationScan = %q, want %q (override should still win)", cycleCountValue2, "button-confirm")
	}

	// Re-assert worker zones still applied post-Reset.
	queried2, err := db.BusDomain.User.Query(
		ctx,
		userbus.QueryFilter{Username: &uname},
		userbus.DefaultOrderBy,
		page.MustParse("1", "10"),
	)
	if err != nil {
		t.Fatalf("user query post-Reset: %v", err)
	}
	if len(queried2) == 0 {
		t.Fatalf("user %q not found post-Reset", workerUsername)
	}
	if !slices.Equal(queried2[0].AssignedZones, wantZones) {
		t.Errorf("post-Reset user %q assigned_zones = %v, want %v", workerUsername, queried2[0].AssignedZones, wantZones)
	}
}

// Test_Load_AppliesFixtures_MultiTableFKChain pins the FK-safe INSERT
// ordering in scenariodb.ApplyFixtures by loading a custom scenario whose
// state.yaml authors a 2-table FK chain in which the child table
// (inventory.cycle_count_items) sorts alphabetically BEFORE the parent
// (inventory.cycle_count_sessions). Under the pre-fix `SELECT DISTINCT
// target_table` with no ORDER BY, Postgres could return child first and
// Load would FK-violate. Under the fix, scopedTables is reversed so
// parents always insert first.
//
// A custom temp scenario is used (rather than a shipped scenario) because
// shipped scenarios under deployments/scenarios/ author unlabelled rows
// on tables with NOT NULL id and no default — those rows currently fail
// INSERT for unrelated reasons (resolveRefs only auto-injects id for
// _label rows). The custom state.yaml below labels every row to sidestep
// that pre-existing seed-pipeline gap.
func Test_Load_AppliesFixtures_MultiTableFKChain(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_Load_AppliesFixtures_MultiTableFKChain")
	ctx := context.Background()

	if err := dbtest.InsertSeedDataWithDB(db.Log, db.DB); err != nil {
		t.Fatalf("baseline seed: %v", err)
	}

	scenariosRoot := t.TempDir()
	scenarioDir := filepath.Join(scenariosRoot, "fk-chain-test")
	if err := os.MkdirAll(scenarioDir, 0o755); err != nil {
		t.Fatalf("mkdir scenario dir: %v", err)
	}

	scenarioYAML := "name: fk-chain-test\ndescription: regression test for follow-up B FK INSERT ordering\n"
	if err := os.WriteFile(filepath.Join(scenarioDir, "scenario.yaml"), []byte(scenarioYAML), 0o644); err != nil {
		t.Fatalf("write scenario.yaml: %v", err)
	}

	// Both rows labelled so resolveRefs auto-injects id. created_by hardcodes
	// the floor_worker1 deterministic UUID (see YAML comment below).
	stateYAML := `cycle_count_sessions:
  - _label: sess1
    name: "FK-Chain-Regression"
    status: draft
    # floor_worker1 deterministic UUID — created by dbtest.InsertSeedDataWithDB
    # via the user seeder's position-based stable IDs (not random). Same UUID
    # is reused by the shipped cycle-count-multi-item state.yaml.
    created_by: "c0000000-0000-4000-8000-000000000001"
    created_date: "2026-05-07T00:00:00Z"
    updated_date: "2026-05-07T00:00:00Z"

cycle_count_items:
  - _label: item1
    session_row_ref: sess1
    product_ref: SKU-0001
    location_ref: PCK-01
    system_quantity: 5
    status: pending
    created_date: "2026-05-07T00:00:00Z"
    updated_date: "2026-05-07T00:00:00Z"
`
	if err := os.WriteFile(filepath.Join(scenarioDir, "state.yaml"), []byte(stateYAML), 0o644); err != nil {
		t.Fatalf("write state.yaml: %v", err)
	}

	if err := dbtest.SeedScenariosFromRoot(ctx, db.BusDomain, scenariosRoot); err != nil {
		t.Fatalf("seed temp scenario: %v", err)
	}

	target, err := db.BusDomain.Scenario.QueryByName(ctx, "fk-chain-test")
	if err != nil {
		t.Fatalf("query scenario by name: %v", err)
	}

	if err := db.BusDomain.Scenario.Load(ctx, target.ID); err != nil {
		t.Fatalf("Load fk-chain-test: %v", err)
	}

	var sessCount, itemCount int
	if err := db.DB.GetContext(ctx, &sessCount,
		`SELECT COUNT(*) FROM inventory.cycle_count_sessions WHERE scenario_id = $1`,
		target.ID); err != nil {
		t.Fatalf("count cycle_count_sessions: %v", err)
	}
	if sessCount != 1 {
		t.Errorf("inventory.cycle_count_sessions count = %d, want 1", sessCount)
	}
	if err := db.DB.GetContext(ctx, &itemCount,
		`SELECT COUNT(*) FROM inventory.cycle_count_items WHERE scenario_id = $1`,
		target.ID); err != nil {
		t.Fatalf("count cycle_count_items: %v", err)
	}
	if itemCount != 1 {
		t.Errorf("inventory.cycle_count_items count = %d, want 1 (FK-safe INSERT ordering broken?)", itemCount)
	}
}
