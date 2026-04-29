package dbtest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus/yamlload"
)

// seedScenarios is the LAST seeder in the InsertSeedData chain. It reads
// every scenario under deployments/scenarios/ via yamlload and upserts
// scenarios + scenario_fixtures rows.
//
// Per phase 0d Decision 2, the seeder does NOT set scenarios_active —
// fresh-seed state is "no active scenario" (baseline mode); operators opt
// in via POST /v1/scenarios/{id}/load.
//
// Fail-hard semantics: any yamlload validation error aborts seeding. A
// corrupt scenario file is a developer error that must block the dev
// database from being populated silently-wrong.
func seedScenarios(ctx context.Context, busDomain BusDomain) error {
	root, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}
	scenariosDir := filepath.Join(root, "deployments", "scenarios")
	return SeedScenariosFromRoot(ctx, busDomain, scenariosDir)
}

// SeedScenariosFromRoot seeds scenarios from an explicit root directory.
// It is the workhorse called by seedScenarios (which computes the root via
// findRepoRoot) and by integration tests that need to seed from a t.TempDir().
//
// NotFoundErr from yamlload.Load is silently ignored — callers that pass an
// empty temp dir or a path with no scenario subdirectories are not an error.
func SeedScenariosFromRoot(ctx context.Context, busDomain BusDomain, scenariosDir string) error {
	scenarios, err := yamlload.Load(scenariosDir)
	if err != nil {
		if yamlload.IsNotFoundErr(err) {
			// No scenarios on disk yet (e.g. running 0d.6 before 0d.9 lands
			// the rush-receiving fixture). Not an error — later phases add
			// fixtures without reworking this seeder.
			return nil
		}
		return fmt.Errorf("yamlload.Load: %w", err)
	}

	lookups := newRefLookups(busDomain.Product, busDomain.InventoryLocation, busDomain.Label)

	for _, s := range scenarios {
		bus := scenariobus.Scenario{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
		}
		if err := busDomain.Scenario.SeedCreate(ctx, bus); err != nil {
			return fmt.Errorf("seed scenario %s: %w", s.Name, err)
		}

		// Phase 0g.B5 — mirror lever_overrides into config.scenario_setting_overrides.
		if err := busDomain.Scenario.SeedApplyLeverOverrides(ctx, s.ID, s.LeverOverrides); err != nil {
			return fmt.Errorf("seed scenario %s: lever overrides: %w", s.Name, err)
		}

		// Sort state keys so fixture insertion order is deterministic.
		// Go map iteration is randomized; sorted iteration plus slice-ordered
		// rows keeps UUIDs and row identities stable across reseeds.
		keys := make([]string, 0, len(s.State))
		for k := range s.State {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, tableSuffix := range keys {
			targetTable := resolveTargetTable(tableSuffix)
			if targetTable == "" {
				return fmt.Errorf("scenario %s: unknown state key %q (no schema.table mapping)", s.Name, tableSuffix)
			}
			for i, row := range s.State[tableSuffix] {
				resolved, err := resolveRefs(ctx, row, s.ID, lookups)
				if err != nil {
					return fmt.Errorf("scenario %s: resolve refs %s[%d]: %w", s.Name, targetTable, i, err)
				}
				payload, err := yamlload.PayloadJSON(resolved)
				if err != nil {
					return fmt.Errorf("scenario %s: payload marshal: %w", s.Name, err)
				}
				fix := scenariobus.ScenarioFixture{
					ID:          detUUID(fmt.Sprintf("fixture:%s:%s:%d", s.Name, targetTable, i)),
					ScenarioID:  s.ID,
					TargetTable: targetTable,
					PayloadJSON: payload,
				}
				if err := busDomain.Scenario.SeedCreateFixture(ctx, fix); err != nil {
					return fmt.Errorf("seed fixture %s/%s[%d]: %w", s.Name, targetTable, i, err)
				}
			}
		}
	}
	return nil
}

// resolveTargetTable maps a state.yaml top-level key to "<schema>.<table>".
// Source of truth is spec §3.5's 18-table scoped list (Decision 1 includes
// lot_locations). Extend as new floor-scoped tables get scenario_id added.
func resolveTargetTable(suffix string) string {
	m := map[string]string{
		"orders":                     "sales.orders",
		"order_line_items":           "sales.order_line_items",
		"order_fulfillment_statuses": "sales.order_fulfillment_statuses",
		"purchase_orders":            "procurement.purchase_orders",
		"purchase_order_line_items":  "procurement.purchase_order_line_items",
		"transfer_orders":            "inventory.transfer_orders",
		"inventory_transactions":     "inventory.inventory_transactions",
		"inventory_adjustments":      "inventory.inventory_adjustments",
		"inventory_items":            "inventory.inventory_items",
		"lot_trackings":              "inventory.lot_trackings",
		"lot_locations":              "inventory.lot_locations",
		"serial_numbers":             "inventory.serial_numbers",
		"pick_tasks":                 "inventory.pick_tasks",
		"put_away_tasks":             "inventory.put_away_tasks",
		"quality_inspections":        "inventory.quality_inspections",
		"cycle_count_sessions":       "inventory.cycle_count_sessions",
		"cycle_count_items":          "inventory.cycle_count_items",
		"approval_requests":          "workflow.approval_requests",
	}
	return m[suffix]
}

// findRepoRoot walks upward from the current working directory looking for
// go.mod. Seeders usually run from the repo root (make seed-frontend) but
// this is defensive in case the seeder gets invoked from a test or tool in
// a subdirectory.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s upward", dir)
		}
		dir = parent
	}
}
