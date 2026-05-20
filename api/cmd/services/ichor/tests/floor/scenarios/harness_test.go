package scenarios_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	authbuild "github.com/timmaaaz/ichor/api/cmd/services/auth/build/all"
	ichorbuild "github.com/timmaaaz/ichor/api/cmd/services/ichor/build/all"
	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/api/sdk/http/mux"
	"github.com/timmaaaz/ichor/app/sdk/auth"
	"github.com/timmaaaz/ichor/app/sdk/authclient"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
)

// startScenarioTest constructs an apitest.Test with ScenariosEnabled: true.
// This is the load-bearing fix — the existing apitest.StartTest defaults to
// ScenariosEnabled: false, making ApplyScenarioFilter a no-op throughout
// existing integration tests. That gap hid GB-006, GB-014, GB-015 from CI
// until manual smoke-verify (see session-log 2026-05-19 line 651).
//
// ⚠ Do not regress this — if a future refactor loses ScenariosEnabled,
// the harness becomes silent theater and the bugs hide again.
func startScenarioTest(t *testing.T, testName string) *apitest.Test {
	t.Helper()

	db := dbtest.NewDatabase(t, testName)

	// Seed the full Phase-0g baseline: products, locations, inventory, scenarios.
	// Required before loadScenarioFixtures can call QueryByName + Load because:
	//   - seedScenarios (inside InsertSeedDataWithDB) reads all 21 YAMLs from
	//     deployments/scenarios/ and populates scenarios + scenario_fixtures.
	//   - seedScenarioCustomer needs contact_infos + streets from the baseline.
	// Calling once per *apitest.Test (not per scenario) keeps it O(1).
	if err := dbtest.InsertSeedDataWithDB(db.Log, db.DB); err != nil {
		t.Fatalf("baseline seed: %v", err)
	}

	// -------------------------------------------------------------------------

	auth, err := auth.New(auth.Config{
		Log:       db.Log,
		DB:        db.DB,
		KeyLookup: &apitest.KeyStore{},
	})
	if err != nil {
		t.Fatal(err)
	}

	// -------------------------------------------------------------------------

	server := httptest.NewServer(mux.WebAPI(mux.Config{
		Log:  db.Log,
		Auth: auth,
		DB:   db.DB,
	}, authbuild.Routes()))
	t.Cleanup(server.Close) // mirrors start_ws.go pattern; apitest.StartTest omits this

	authClient := authclient.New(db.Log, server.URL)

	// -------------------------------------------------------------------------

	appMux := mux.WebAPI(mux.Config{
		Log:              db.Log,
		AuthClient:       authClient,
		DB:               db.DB,
		ScenariosEnabled: true, // ⚠ load-bearing — see comment above
	}, ichorbuild.Routes())

	return apitest.New(db, auth, appMux)
}

// loadScenarioFixtures activates a named scenario for the given test:
//
//  1. resolveScenarioPath verifies the scenario YAML exists on disk and
//     enforces D3 (no silent shadowing across roots).
//  2. QueryByName retrieves the scenario UUID — InsertSeedDataWithDB (called in
//     startScenarioTest) already ran SeedScenariosFromRoot over the canonical
//     deployments/scenarios/ tree, so the row is guaranteed to exist.
//  3. Load executes the transactional swap: deletes scoped rows for the
//     current active scenario, inserts fixture rows for the target, updates
//     the scenarios_active singleton.
//
// Returns the scenario UUID so callers can use it for downstream queries.
// All failure paths call t.Fatalf with the scenario name in the message.
func loadScenarioFixtures(t *testing.T, h *apitest.Test, scenarioName string) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	// Step 1 — verify on-disk existence and enforce D3.
	if _, err := resolveScenarioPath(scenarioName); err != nil {
		t.Fatalf("resolveScenarioPath(%q): %v", scenarioName, err)
	}

	// Step 2 — look up the UUID seeded by startScenarioTest.
	sc, err := h.DB.BusDomain.Scenario.QueryByName(ctx, scenarioName)
	if err != nil {
		t.Fatalf("loadScenarioFixtures(%q): QueryByName: %v", scenarioName, err)
	}

	// Step 3 — transactional activate (Delete scoped rows → ApplyFixtures → SetActive).
	if err := h.DB.BusDomain.Scenario.Load(ctx, sc.ID); err != nil {
		t.Fatalf("loadScenarioFixtures(%q): Load: %v", scenarioName, err)
	}

	return sc.ID
}

// scenarioRoots returns the ordered list of root directories searched by
// resolveScenarioPath. The first entry is the canonical shipped-scenario tree.
// Customer-specific scenario roots are reserved for a future onboarding phase.
//
// Paths are relative to this package's directory
// (api/cmd/services/ichor/tests/floor/scenarios/).
func scenarioRoots() []string {
	return []string{
		"../../../../../../../deployments/scenarios",
		// Customer scenarios:
		//   "../../../../../../../deployments/customers/*/scenarios",
		// Added when customer onboarding lands.
	}
}

// resolveScenarioPath locates the scenario.yaml for name across all
// scenarioRoots. Returns the path to the matching scenario.yaml file.
//
// Fail-fast rules (D3):
//   - Zero matches  → error (scenario not found)
//   - Multiple matches across roots → error (rename one to avoid shadowing)
//
// resolveScenarioPath is testable without a *testing.T: it returns (string, error).
// The caller (loadScenarioFixtures) translates errors into t.Fatalf calls.
func resolveScenarioPath(name string) (string, error) {
	var matches []string
	for _, root := range scenarioRoots() {
		candidate := filepath.Join(root, name, "scenario.yaml")
		if _, err := os.Stat(candidate); err == nil {
			matches = append(matches, candidate)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("scenario %q not found in any root: %v", name, scenarioRoots())
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("scenario %q exists in multiple roots: %v — rename one to avoid shadowing", name, matches)
	}
}
