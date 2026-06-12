package workflowactions_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/protected"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/procurement"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func newLog() *logger.Logger {
	var buf bytes.Buffer
	return logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string {
		return otel.GetTraceID(context.Background())
	})
}

func TestPopulateProtected(t *testing.T) {
	t.Parallel()
	log := newLog()

	reg := workflow.NewActionRegistry()
	reg.Register(inventory.NewReserveInventoryHandler(log, nil, nil, nil))  // on_update inventory_items.reserved_quantity
	reg.Register(procurement.NewApprovePurchaseOrderHandler(log, nil))      // on_update purchase_orders.approved_by/...
	reg.Register(inventory.NewCreatePutAwayTaskHandler(log, nil, nil, nil)) // on_create put_away_tasks (must be EXCLUDED)
	reg.Register(data.NewUpdateFieldHandler(log, nil))                      // generic (must be skipped)

	preg := protected.New()
	workflowactions.PopulateProtected(preg, reg)

	mustBlock := func(entity, field, wantRoute string) {
		t.Helper()
		route, blocked := preg.Check(entity, field)
		if !blocked {
			t.Fatalf("expected (%s,%s) blocked", entity, field)
		}
		if wantRoute != "" && route != wantRoute {
			t.Fatalf("(%s,%s) route=%q want %q", entity, field, route, wantRoute)
		}
	}
	mustAllow := func(entity, field string) {
		t.Helper()
		if _, blocked := preg.Check(entity, field); blocked {
			t.Fatalf("expected (%s,%s) NOT blocked", entity, field)
		}
	}

	// --- auto-source: on_update manifest claims, routed to the claiming action ---
	mustBlock("inventory.inventory_items", "reserved_quantity", "reserve_inventory")
	mustBlock("procurement.purchase_orders", "approved_by", "approve_purchase_order")

	// --- on_create claims are NOT auto-protected (create-time validation is Option B) ---
	mustAllow("inventory.put_away_tasks", "anything")

	// --- generic handler claims are skipped; no bogus empty-entity key leaks ---
	mustAllow("", "")

	// --- domain-declared db-tag source (per-store RegisterProtected) ---
	mustBlock("procurement.purchase_orders", "purchase_order_status_id", "approve_purchase_order / reject_purchase_order")
	mustBlock("sales.orders", "order_fulfillment_status_id", "")
	mustBlock("sales.order_line_items", "picked_quantity", "")
	mustBlock("inventory.transfer_orders", "claimed_by", "")
	mustBlock("inventory.transfer_orders", "completed_by", "")
	mustBlock("procurement.purchase_order_line_items", "quantity_received", "")
	mustBlock("core.users", "user_approval_status_id", "")

	// --- whole-table ledger protection ---
	mustBlock("inventory.inventory_transactions", "quantity", "")
	mustBlock("inventory.inventory_transactions", "created_date", "")
}

// Test_PopulateProtected_WholeTableInvariants is a regression guard for the control-plane,
// status-definition, and warehouse-structure whole-table protections. A generic workflow
// action must never mutate these — see PopulateProtected (4). The guard catches a future
// whitelist addition (data/tables.go) that forgets to also protect the table here: any of
// these becoming generic-writable is a privilege-escalation / cascade-undecidability / broken
// frontend-tie-in / orphaned-inventory regression. Because ProtectEntity blocks every column,
// an arbitrary "any_field" probe is sufficient.
func Test_PopulateProtected_WholeTableInvariants(t *testing.T) {
	t.Parallel()

	// The protections are unconditional in PopulateProtected, but build the FULL production
	// registry so this asserts what ships, not a hand-picked subset.
	reg := buildFullRegistry(t)
	preg := protected.New()
	workflowactions.PopulateProtected(preg, reg)

	categories := map[string][]string{
		// ENGINE — a rule rewriting the engine's own state makes the cascade-loop guard
		// undecidable and can forge the P1 lineage.
		"engine": {
			"workflow.automation_rules",
			"workflow.rule_actions",
			"workflow.action_templates",
			"workflow.rule_dependencies",
			"workflow.trigger_types",
			"workflow.entity_types",
			"workflow.entities",
			"workflow.automation_executions",
			"workflow.notification_deliveries",
		},
		// TABLE BUILDER — the dynamic table-builder config surface.
		"table_builder": {
			"config.table_configs",
		},
		// RBAC — writing these from a workflow is privilege escalation.
		"rbac": {
			"core.roles",
			"core.user_roles",
			"core.table_access",
		},
		// STATUS / REFERENCE DEFINITIONS — values wired into frontend + backend code.
		"status_defs": {
			"sales.order_fulfillment_statuses",
			"sales.line_item_fulfillment_statuses",
			"procurement.purchase_order_statuses",
			"procurement.purchase_order_line_item_statuses",
			"hr.user_approval_status",
		},
		// WAREHOUSE STRUCTURE — restructuring these orphans physical inventory_items.
		"warehouse_structure": {
			"inventory.warehouses",
			"inventory.zones",
			"inventory.inventory_locations",
		},
	}

	total := 0
	for category, tables := range categories {
		for _, table := range tables {
			total++
			route, blocked := preg.Check(table, "any_field")
			if !blocked {
				t.Errorf("[%s] %q is NOT whole-table protected — a generic workflow action can mutate it", category, table)
			}
			if route != "" {
				t.Errorf("[%s] %q whole-table protection has route %q, want \"\" (no typed action substitutes)", category, table, route)
			}
		}
	}
	if total != 21 {
		t.Fatalf("expected 21 protected tables across all categories, got %d", total)
	}

	// Counterpart guard — the OTHER direction the list can drift: a future over-broad protection
	// silently catching a table we DELIBERATELY left workflow-writable (WRITE_PATH §1b). Nothing
	// else guards this; the 21-protected loop above only checks presence. The generic handlers are
	// a pure pass-through to Check, so a registry assertion is the whole story (no handler needed).
	// Whole-table-protecting core.users (onboarding workflows write it) or products.brands trips here.
	writable := []struct{ table, field string }{
		{"core.users", "first_name"},    // sensitive cols are FIELD-protected via userdb; the table is not
		{"products.brands", "name"},     // taxonomy: investigated 2026-06-12, no code tie-ins → left open
	}
	for _, w := range writable {
		if _, blocked := preg.Check(w.table, w.field); blocked {
			t.Errorf("deliberately-writable %s.%s is protected — over-broad protection? (WRITE_PATH §1b)", w.table, w.field)
		}
	}
}
