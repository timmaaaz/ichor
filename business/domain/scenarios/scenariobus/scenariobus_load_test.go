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

	// Baseline seed — populates config.settings (11 levers) plus all FK
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

	scenarioYAML := "name: test-levers\ndescription: B5 integration test\nlever_overrides:\n  pick.lotScan: required-if-lot-tracked\n  pick.destinationScan: required\n"
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
}
