package workflowactions_test

// PG — Guard Verification backstop (deliberately deferred from P3).
//
// The protected-list rests on a SILENT-DROP hazard: a protected (entity, field) only blocks if
// `field` is a REAL column on a REAL table. The db-tag source reflects `protected:"true"` off the
// db STORE models reading their `db` tag — but bus json tags ≠ DB columns, and a tag that drifts
// off a real column makes the protection a silent no-op (it stops blocking and nobody notices).
// Nothing else catches that: the auto-source is backstopped by the manifest-consistency test
// (declared == fired against real bus writes), but the db-tag source has no such check.
//
// This test closes it: build the production protected registry, enumerate every registered
// (entity, field), and assert each resolves to a real column (IsValidTableName + information_schema
// column existence). Whole-table protections assert the table exists. A drifted db tag → RED.

import (
	"context"
	"strings"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/protected"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/procurement"
)

// Test_ProtectedFields_ResolveToRealColumns enumerates the production protected registry and
// asserts every guarded (entity, field) maps to a real DB column on a whitelisted table.
func Test_ProtectedFields_ResolveToRealColumns(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ProtectedFields_Columns")
	ctx := context.Background()
	log := newLog()

	// Build the production registry the way RegisterAll/all.go do: register the EntityModifier
	// handlers (auto-source) + PopulateProtected (which also pulls the db-tag stores + the
	// whole-table ledger). nil buses are fine — GetEntityModifications ignores them.
	reg := workflow.NewActionRegistry()
	reg.Register(inventory.NewReserveInventoryHandler(log, nil, nil, nil))  // on_update inventory_items.reserved_quantity
	reg.Register(procurement.NewApprovePurchaseOrderHandler(log, nil))      // on_update purchase_orders.approved_by/...
	reg.Register(inventory.NewCreatePutAwayTaskHandler(log, nil, nil, nil)) // on_create — excluded from auto-source
	reg.Register(data.NewUpdateFieldHandler(log, nil))                      // generic — skipped

	preg := protected.New()
	workflowactions.PopulateProtected(preg, reg)

	entries := preg.Entries()
	if len(entries) == 0 {
		t.Fatal("registry is empty — PopulateProtected produced no protected (entity,field) pairs; the backstop would vacuously pass")
	}

	const colExistsQ = `SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2 AND column_name = $3)`
	const tableExistsQ = `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = $1 AND table_name = $2)`

	for _, e := range entries {
		// 1. The protected entity must be a table the generic handlers can actually target;
		//    a protected entity outside the whitelist is dead protection.
		if !data.IsValidTableName(e.Entity) {
			t.Errorf("protected entity %q is not in the generic-handler table whitelist (IsValidTableName=false)", e.Entity)
			continue
		}

		// 2. The entity must be schema-qualified (the form the registry + handlers use).
		schema, table, ok := strings.Cut(e.Entity, ".")
		if !ok {
			t.Errorf("protected entity %q is not schema-qualified (expected schema.table)", e.Entity)
			continue
		}

		if e.Field == "" {
			// Whole-table protection — assert the table exists.
			var exists bool
			if err := db.DB.QueryRowContext(ctx, tableExistsQ, schema, table).Scan(&exists); err != nil {
				t.Fatalf("querying table existence for %q: %v", e.Entity, err)
			}
			if !exists {
				t.Errorf("whole-table protected entity %q resolves to NO real table", e.Entity)
			}
			continue
		}

		// 3. Field-level protection — assert the column really exists (the silent-drop guard).
		var exists bool
		if err := db.DB.QueryRowContext(ctx, colExistsQ, schema, table, e.Field).Scan(&exists); err != nil {
			t.Fatalf("querying column existence for %s.%s: %v", e.Entity, e.Field, err)
		}
		if !exists {
			t.Errorf("protected field %q on %q resolves to NO real column — protection silently no-ops (db tag drifted?)", e.Field, e.Entity)
		}
	}
}
