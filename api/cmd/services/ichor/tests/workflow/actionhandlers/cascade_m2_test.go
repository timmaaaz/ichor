package actionhandlers_test

// PL.M2 — the M2 allocation_results cascade, made functional and proven LIVE.
//
// Before the PL.M2 event-contract fix, the M2 delegate passed the tagless AllocationResult
// struct, so the event surfaced only {ID, IdempotencyKey, AllocationData(base64), ...} and a
// trigger on `status` could never match — the seeded Allocation-Success/Failed rules were
// dead end-to-end. event.go now flattens the AllocationData blob into the Entity map and the
// result structs carry reference_id, so `status` and `reference_id` reach a trigger condition.
//
// Coverage (Option A — clean-column positive + assert-blocked compose):
//   - EventContract: the allocation_results.created event RawData now carries status +
//     reference_id (the Part-1 fix, asserted directly via a delegate recorder).
//   - LiveCascade: allocate(status=success, reference_id=order) → allocation_results.created →
//     a rule gated status==success → update_field a CLEAN order_line_items column
//     (short_pick_reason) WHERE order_id={{reference_id}} → the line item is updated end-to-end
//     through the real worker. Also the read-after-commit observation (DESIGN §9).
//   - ProtectedTargetBlocked: the REAL seeded rule writes line_item_fulfillment_statuses_id,
//     which P3 protects — assert the protected-list rejects it (so the advertised seeded rule
//     stays gated until F3 builds the typed fulfillment action; M2 + guard compose correctly).

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/domain/sales/lineitemfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus/stores/orderlineitemsdb"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/protected"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"

	"github.com/timmaaaz/ichor/api/cmd/services/ichor/build/all/workflowdomains"
)

// ── EventContract + ProtectedTargetBlocked (no worker) ─────────────────────────────────

func TestCascade_M2_EventContract(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeM2Contract")
	ctx := context.Background()
	base := seedConsistencyBase(t, ctx, db)
	uid := base.userID

	// PL.M2 Part-1: the allocation_results.created event now surfaces status + reference_id.
	t.Run("allocation_event_surfaces_status_and_reference", func(t *testing.T) {
		if _, err := inventoryitembus.TestSeedInventoryItems(ctx, 1, []uuid.UUID{base.loc0}, []uuid.UUID{base.productIDs[0]}, db.BusDomain.InventoryItem); err != nil {
			t.Fatalf("seeding inventory item: %v", err)
		}

		refID := uuid.New().String()
		var entity map[string]any
		var fired bool
		db.BusDomain.Delegate.Register(workflow.AllocationResultDomainName, workflow.ActionCreated,
			func(_ context.Context, d delegate.Data) error {
				var p workflow.DelegateEventParams
				if err := json.Unmarshal(d.RawParams, &p); err != nil {
					return nil
				}
				if m, ok := p.Entity.(map[string]any); ok {
					entity = m
					fired = true
				}
				return nil
			})

		h := inventory.NewAllocateInventoryHandler(db.Log, db.DB, db.BusDomain.InventoryItem,
			db.BusDomain.InventoryLocation, db.BusDomain.InventoryTransaction, db.BusDomain.Product, db.BusDomain.Workflow)
		cfg := mustJSON(t, map[string]any{
			"inventory_items":     []map[string]any{{"product_id": base.productIDs[0].String(), "quantity": 10}},
			"allocation_mode":     "allocate",
			"allocation_strategy": "fifo",
			"allow_partial":       false,
			"priority":            "high",
			"reference_id":        refID,
		})
		ruleID := uuid.New()
		if _, err := h.Execute(ctx, cfg, workflow.ActionExecutionContext{UserID: uid, ExecutionID: uuid.New(), RuleID: &ruleID}); err != nil {
			t.Fatalf("allocate Execute: %v", err)
		}

		if !fired {
			t.Fatal("allocation_results.created delegate never fired")
		}
		if got := entity["status"]; got != "success" {
			t.Errorf("event RawData status = %v, want \"success\" — PL.M2 blob-flatten must surface status", got)
		}
		if got := entity["reference_id"]; got != refID {
			t.Errorf("event RawData reference_id = %v, want %s — result struct must carry reference_id", got, refID)
		}
	})

	// Option A compose: the real seeded rule's protected target is rejected by P3.
	t.Run("seeded_rule_protected_target_blocked", func(t *testing.T) {
		reg := protected.New()
		orderlineitemsdb.RegisterProtected(reg) // the real store's protected:"true" columns
		h := data.NewUpdateFieldHandler(db.Log, db.DB, data.WithProtectedRegistry(reg))

		// Exactly what the seeded "Allocation Success" rule writes.
		cfg := mustJSON(t, map[string]any{
			"target_entity": "sales.order_line_items",
			"target_field":  "line_item_fulfillment_statuses_id",
			"new_value":     "ALLOCATED",
			"field_type":    "foreign_key",
			"conditions":    []map[string]any{cond("order_id", "equals", uuid.New().String())},
		})
		_, err := h.Execute(ctx, cfg, workflow.ActionExecutionContext{UserID: uid})
		if !errors.Is(err, protected.ErrProtectedField) {
			t.Fatalf("seeded rule's protected target must be rejected (ErrProtectedField), got %v", err)
		}
		t.Log("seeded Allocation-Success rule's line_item_fulfillment_statuses_id write is correctly P3-blocked; the real fulfillment write awaits F3 (typed action)")
	})
}

// ── LiveCascade (worker): allocate → allocation_results → update order_line_items ──────────

func TestCascade_M2_LiveCascade(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_CascadeM2Live")
	ctx := context.Background()
	base := seedConsistencyBase(t, ctx, db)
	uid := base.userID

	orderID := seedOrderWithLineItem(t, ctx, db, base)

	// The cascaded action is update_field (wired to fire synthesized events).
	registry := workflow.NewActionRegistry()
	registry.Register(data.NewUpdateFieldHandler(db.Log, db.DB,
		data.WithDelegate(db.BusDomain.Delegate),
		data.WithEntityRegistry(workflowdomains.ReverseMap())))
	rig := startCascadeRig(t, ctx, db, registry)

	entityType, err := rig.workflowBus.QueryEntityTypeByName(ctx, "table")
	if err != nil {
		t.Fatalf("querying entity type: %v", err)
	}
	onCreate, err := rig.workflowBus.QueryTriggerTypeByName(ctx, "on_create")
	if err != nil {
		t.Fatalf("querying on_create trigger type: %v", err)
	}
	allocEntity, err := rig.workflowBus.QueryEntityByName(ctx, "allocation_results")
	if err != nil {
		t.Fatalf("querying allocation_results entity: %v", err)
	}

	// A test-copy of the seeded Allocation-Success rule, but writing a CLEAN (non-protected)
	// order_line_items column so the cascade completes a real write WHERE order_id={{reference_id}}.
	const marker = "ALLOCATED-VIA-CASCADE"
	seedActiveRule(t, ctx, rig.workflowBus, uid, "alloc-success-copy",
		allocEntity.ID, entityType.ID, onCreate.ID,
		fieldConditions(cond("status", "equals", "success")),
		"update_field", map[string]any{
			"target_entity": "sales.order_line_items",
			"target_field":  "short_pick_reason",
			"new_value":     marker,
			"conditions":    []map[string]any{cond("order_id", "equals", "{{reference_id}}")},
		})
	if err := rig.triggerProcessor.RefreshRules(ctx); err != nil {
		t.Fatalf("refreshing rules: %v", err)
	}

	// Fire the allocation. It fires the M2 delegate PRE-COMMIT (allocate.go:479 < tx.Commit:488) —
	// DESIGN §9 best-effort. The cascade nonetheless observes committed state below because the
	// worker dispatch lag exceeds the commit; the read-your-writes guarantee is F2's job (outbox).
	if _, err := inventoryitembus.TestSeedInventoryItems(ctx, 1, []uuid.UUID{base.loc0}, []uuid.UUID{base.productIDs[0]}, db.BusDomain.InventoryItem); err != nil {
		t.Fatalf("seeding inventory item: %v", err)
	}
	h := inventory.NewAllocateInventoryHandler(db.Log, db.DB, db.BusDomain.InventoryItem,
		db.BusDomain.InventoryLocation, db.BusDomain.InventoryTransaction, db.BusDomain.Product, db.BusDomain.Workflow)
	cfg := mustJSON(t, map[string]any{
		"inventory_items":     []map[string]any{{"product_id": base.productIDs[0].String(), "quantity": 10}},
		"allocation_mode":     "allocate",
		"allocation_strategy": "fifo",
		"allow_partial":       false,
		"priority":            "high",
		"reference_id":        orderID.String(),
	})
	ruleID := uuid.New()
	if _, err := h.Execute(ctx, cfg, workflow.ActionExecutionContext{UserID: uid, ExecutionID: uuid.New(), RuleID: &ruleID}); err != nil {
		t.Fatalf("allocate Execute: %v", err)
	}

	if !eventually(20*time.Second, 400*time.Millisecond, func() bool {
		return lineItemShortPick(t, ctx, db, orderID) == marker
	}) {
		t.Fatalf("live M2 cascade never updated the order line item short_pick_reason to %q "+
			"(allocate→allocation_results→update_field WHERE order_id={{reference_id}})", marker)
	}
	t.Logf("live M2 cascade proven: allocate(status=success, ref=%s) → order_line_items.short_pick_reason=%q "+
		"(cascade observed committed state despite the pre-commit fire — §9 best-effort holds here; F2 owns the guarantee)", orderID, marker)
}

// ── seeding helpers ────────────────────────────────────────────────────────────────────

// seedOrderWithLineItem builds the minimum sales chain (contacts → customer → fulfillment
// statuses → order → one line item) and returns the order id. The line item's order_id is the
// returned id, so an update_field WHERE order_id={{reference_id}} (reference_id = order id) hits it.
func seedOrderWithLineItem(t *testing.T, ctx context.Context, db *dbtest.Database, base baseFixtures) uuid.UUID {
	t.Helper()

	tzs, err := db.BusDomain.Timezone.QueryAll(ctx)
	if err != nil || len(tzs) == 0 {
		t.Fatalf("querying timezones: %v", err)
	}
	tzIDs := make([]uuid.UUID, len(tzs))
	for i, tz := range tzs {
		tzIDs[i] = tz.ID
	}

	contacts, err := contactinfosbus.TestSeedContactInfos(ctx, 1, []uuid.UUID{base.streetID}, tzIDs, db.BusDomain.ContactInfos)
	if err != nil {
		t.Fatalf("seeding contact infos: %v", err)
	}
	contactIDs := []uuid.UUID{contacts[0].ID}

	customers, err := customersbus.TestSeedCustomers(ctx, 1, []uuid.UUID{base.streetID}, contactIDs, []uuid.UUID{base.userID}, db.BusDomain.Customers)
	if err != nil {
		t.Fatalf("seeding customers: %v", err)
	}
	customerIDs := []uuid.UUID{customers[0].ID}

	ofs, err := orderfulfillmentstatusbus.TestSeedOrderFulfillmentStatuses(ctx, db.BusDomain.OrderFulfillmentStatus)
	if err != nil || len(ofs) == 0 {
		t.Fatalf("seeding order fulfillment statuses: %v", err)
	}
	ofIDs := []uuid.UUID{ofs[0].ID}

	lifs, err := lineitemfulfillmentstatusbus.TestSeedLineItemFulfillmentStatuses(ctx, db.BusDomain.LineItemFulfillmentStatus)
	if err != nil || len(lifs) == 0 {
		t.Fatalf("seeding line item fulfillment statuses: %v", err)
	}
	lifIDs := []uuid.UUID{lifs[0].ID}

	orders, err := ordersbus.TestSeedOrders(ctx, 1, []uuid.UUID{base.userID}, customerIDs, ofIDs, []uuid.UUID{base.currencyID}, db.BusDomain.Order)
	if err != nil || len(orders) == 0 {
		t.Fatalf("seeding orders: %v", err)
	}
	orderID := orders[0].ID

	if _, err := orderlineitemsbus.TestSeedOrderLineItems(ctx, 1, []uuid.UUID{orderID},
		[]uuid.UUID{base.productIDs[0]}, lifIDs, []uuid.UUID{base.userID}, db.BusDomain.OrderLineItem); err != nil {
		t.Fatalf("seeding order line items: %v", err)
	}
	return orderID
}

// lineItemShortPick reads the (nullable) short_pick_reason of the line item for an order.
func lineItemShortPick(t *testing.T, ctx context.Context, db *dbtest.Database, orderID uuid.UUID) string {
	t.Helper()
	var s sql.NullString
	err := db.DB.QueryRowContext(ctx,
		`SELECT short_pick_reason FROM sales.order_line_items WHERE order_id = $1`, orderID).Scan(&s)
	if err != nil {
		t.Fatalf("reading order line item short_pick_reason: %v", err)
	}
	return s.String
}
