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
