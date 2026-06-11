package actionhandlers_test

// Manifest consistency — "declared == fired" (DB-backed, full firing coverage).
//
// DESIGN §6: the cascade scheme's soundness reduces to "the runtime fires a delegate for
// exactly the mutations the manifest (GetEntityModifications) declares." The structural
// half (business/sdk/workflow/workflowactions/manifest_consistency_test.go) guards the
// declared side for all 19 handlers; this test closes the loop by executing every
// currently-FIRING handler against a real database with a spy on the shared delegate, and
// asserting each declared (non-silent) modification actually fires a matching delegate
// event.
//
// It also pins the two known-silent gaps that P4 will close:
//   - allocate_inventory declares allocation_results (on_create) but workflowbus.
//     CreateAllocationResult has no delegate.Call — the "M2" gap. Asserted silent.
//   - update_field (raw SQL, no bus) fires nothing — the "M1" gap. The trip-wire subtest
//     executes it and asserts NO delegate fires; when P4 wires synthesis this goes RED,
//     forcing P4 to move it out of knownSilent and add the firing assertion.
//
// Handlers are executed DIRECTLY (handler.Execute), not through Temporal: delegate.Call is
// synchronous inside the bus methods, so by the time Execute returns every fire has been
// captured. No worker, rule, edge, trigger, or polling needed.

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/core/currencybus"
	"github.com/timmaaaz/ichor/business/domain/core/userbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/approval"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/procurement"
)

// ─── delegate recorder ───────────────────────────────────────────────────────

type firedEvent struct {
	domain   string
	action   string
	entityID uuid.UUID
}

type delegateRecorder struct {
	mu     sync.Mutex
	events []firedEvent
}

func (r *delegateRecorder) add(e firedEvent) {
	r.mu.Lock()
	r.events = append(r.events, e)
	r.mu.Unlock()
}

func (r *delegateRecorder) mark() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events)
}

func (r *delegateRecorder) since(mark int) []firedEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]firedEvent, len(r.events)-mark)
	copy(out, r.events[mark:])
	return out
}

func (r *delegateRecorder) registerOn(del *delegate.Delegate, domain string) {
	for _, action := range []string{workflow.ActionCreated, workflow.ActionUpdated, workflow.ActionDeleted} {
		a := action
		del.Register(domain, a, func(_ context.Context, data delegate.Data) error {
			var p workflow.DelegateEventParams
			_ = json.Unmarshal(data.RawParams, &p)
			r.add(firedEvent{domain: domain, action: a, entityID: p.EntityID})
			return nil
		})
	}
}

func contains(events []firedEvent, domain, action string) bool {
	for _, e := range events {
		if e.domain == domain && e.action == action {
			return true
		}
	}
	return false
}

// ─── declared-entity → delegate-domain mapping (mirrors all.go RegisterDomain) ──

var entityDomain = map[string]string{
	"workflow.approval_requests":            approvalrequestbus.DomainName,
	"inventory.inventory_adjustments":       inventoryadjustmentbus.DomainName,
	"inventory.transfer_orders":             transferorderbus.DomainName,
	"procurement.purchase_orders":           purchaseorderbus.DomainName,
	"procurement.purchase_order_line_items": purchaseorderlineitembus.DomainName,
	"inventory.inventory_items":             inventoryitembus.DomainName,
	"inventory.put_away_tasks":              putawaytaskbus.DomainName,
}

// knownSilentEntities are declared by a handler but have no delegate today (P4 closes these).
var knownSilentEntities = map[string]bool{
	"allocation_results": true, // M2: workflowbus.CreateAllocationResult has no delegate.Call
}

func actionForEvent(eventType string) string {
	switch eventType {
	case "on_create":
		return workflow.ActionCreated
	case "on_update":
		return workflow.ActionUpdated
	case "on_delete":
		return workflow.ActionDeleted
	default:
		return ""
	}
}

// assertDeclaredFired checks that every non-silent declared modification produced a matching
// delegate event in the window, and that silent ones are accounted for (not asserted to fire).
func assertDeclaredFired(t *testing.T, actionType string, declared []workflow.EntityModification, window []firedEvent) {
	t.Helper()
	if len(declared) == 0 {
		t.Fatalf("%s declared no modifications", actionType)
	}
	for _, mod := range declared {
		if knownSilentEntities[mod.EntityName] {
			t.Logf("%s: %s/%s is known-silent (no delegate today) — P4 will enable + assert it", actionType, mod.EntityName, mod.EventType)
			continue
		}
		domain, ok := entityDomain[mod.EntityName]
		if !ok {
			t.Fatalf("%s: no delegate-domain mapping for declared entity %q (update entityDomain / all.go wiring)", actionType, mod.EntityName)
		}
		action := actionForEvent(mod.EventType)
		if !contains(window, domain, action) {
			t.Errorf("%s: declared %s/%s did NOT fire a delegate (window=%v)", actionType, mod.EntityName, mod.EventType, window)
		}
	}
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	return b
}

// ─── shared fixtures ───────────────────────────────────────────────────────────

type baseFixtures struct {
	userID            uuid.UUID
	productIDs        []uuid.UUID
	loc0, loc1        uuid.UUID
	warehouseID       uuid.UUID
	streetID          uuid.UUID
	supplierID        uuid.UUID
	supplierProductID uuid.UUID
	poStatusID        uuid.UUID
	lineItemStatusID  uuid.UUID
	currencyID        uuid.UUID
}

func seedConsistencyBase(t *testing.T, ctx context.Context, db *dbtest.Database) baseFixtures {
	t.Helper()

	admins, err := userbus.TestSeedUsersWithNoFKs(ctx, 1, userbus.Roles.Admin, db.BusDomain.User)
	if err != nil {
		t.Fatalf("seeding users: %v", err)
	}
	userID := admins[0].ID

	regions, err := db.BusDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "3"))
	if err != nil || len(regions) == 0 {
		t.Fatalf("querying regions: %v", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 2, regionIDs, db.BusDomain.City)
	if err != nil {
		t.Fatalf("seeding cities: %v", err)
	}
	cityIDs := make([]uuid.UUID, len(cities))
	for i, c := range cities {
		cityIDs[i] = c.ID
	}

	streets, err := streetbus.TestSeedStreets(ctx, 3, cityIDs, db.BusDomain.Street)
	if err != nil {
		t.Fatalf("seeding streets: %v", err)
	}
	streetIDs := make([]uuid.UUID, len(streets))
	for i, s := range streets {
		streetIDs[i] = s.ID
	}

	tzs, err := db.BusDomain.Timezone.QueryAll(ctx)
	if err != nil || len(tzs) == 0 {
		t.Fatalf("querying timezones: %v", err)
	}
	tzIDs := make([]uuid.UUID, len(tzs))
	for i, tz := range tzs {
		tzIDs[i] = tz.ID
	}

	contactInfos, err := contactinfosbus.TestSeedContactInfos(ctx, 2, streetIDs, tzIDs, db.BusDomain.ContactInfos)
	if err != nil {
		t.Fatalf("seeding contact infos: %v", err)
	}
	contactInfoIDs := make([]uuid.UUID, len(contactInfos))
	for i, ci := range contactInfos {
		contactInfoIDs[i] = ci.ID
	}

	suppliers, err := supplierbus.TestSeedSuppliers(ctx, 1, contactInfoIDs, db.BusDomain.Supplier)
	if err != nil {
		t.Fatalf("seeding suppliers: %v", err)
	}
	supplierID := suppliers[0].SupplierID

	brands, err := brandbus.TestSeedBrands(ctx, 2, contactInfoIDs, db.BusDomain.Brand)
	if err != nil {
		t.Fatalf("seeding brands: %v", err)
	}
	brandIDs := make(uuid.UUIDs, len(brands))
	for i, b := range brands {
		brandIDs[i] = b.BrandID
	}

	productCategories, err := productcategorybus.TestSeedProductCategories(ctx, 2, db.BusDomain.ProductCategory)
	if err != nil {
		t.Fatalf("seeding product categories: %v", err)
	}
	categoryIDs := make(uuid.UUIDs, len(productCategories))
	for i, pc := range productCategories {
		categoryIDs[i] = pc.ProductCategoryID
	}

	products, err := productbus.TestSeedProducts(ctx, 5, brandIDs, categoryIDs, db.BusDomain.Product)
	if err != nil || len(products) < 2 {
		t.Fatalf("seeding products: %v", err)
	}
	productIDs := []uuid.UUID{products[0].ProductID, products[1].ProductID}

	supplierProducts, err := supplierproductbus.TestSeedSupplierProducts(ctx, 1, []uuid.UUID{productIDs[0]}, []uuid.UUID{supplierID}, db.BusDomain.SupplierProduct)
	if err != nil {
		t.Fatalf("seeding supplier products: %v", err)
	}
	supplierProductID := supplierProducts[0].SupplierProductID

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 1, userID, streetIDs, db.BusDomain.Warehouse)
	if err != nil {
		t.Fatalf("seeding warehouses: %v", err)
	}
	warehouseIDs := []uuid.UUID{warehouses[0].ID}

	zones, err := zonebus.TestSeedZone(ctx, 1, warehouseIDs, db.BusDomain.Zones)
	if err != nil {
		t.Fatalf("seeding zones: %v", err)
	}

	locations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, 2, warehouseIDs, zones, db.BusDomain.InventoryLocation)
	if err != nil || len(locations) < 2 {
		t.Fatalf("seeding inventory locations: %v", err)
	}

	poStatuses, err := purchaseorderstatusbus.TestSeedPurchaseOrderStatuses(ctx, 1, db.BusDomain.PurchaseOrderStatus)
	if err != nil {
		t.Fatalf("seeding PO statuses: %v", err)
	}

	lineItemStatuses, err := purchaseorderlineitemstatusbus.TestSeedPurchaseOrderLineItemStatuses(ctx, 1, db.BusDomain.PurchaseOrderLineItemStatus)
	if err != nil {
		t.Fatalf("seeding PO line item statuses: %v", err)
	}

	currencies, err := currencybus.TestSeedCurrencies(ctx, 1, db.BusDomain.Currency)
	if err != nil {
		t.Fatalf("seeding currencies: %v", err)
	}

	return baseFixtures{
		userID:            userID,
		productIDs:        productIDs,
		loc0:              locations[0].LocationID,
		loc1:              locations[1].LocationID,
		warehouseID:       warehouseIDs[0],
		streetID:          streetIDs[0],
		supplierID:        supplierID,
		supplierProductID: supplierProductID,
		poStatusID:        poStatuses[0].ID,
		lineItemStatusID:  lineItemStatuses[0].ID,
		currencyID:        currencies[0].ID,
	}
}

// ─── the test ──────────────────────────────────────────────────────────────────

func Test_ManifestConsistency_DeclaredEqualsFired(t *testing.T) {
	t.Parallel()

	db := dbtest.NewDatabase(t, "Test_ManifestConsistency_Fired")
	ctx := context.Background()

	rec := &delegateRecorder{}
	for _, d := range []string{
		approvalrequestbus.DomainName, inventoryadjustmentbus.DomainName, transferorderbus.DomainName,
		purchaseorderbus.DomainName, purchaseorderlineitembus.DomainName, inventoryitembus.DomainName,
		putawaytaskbus.DomainName,
	} {
		rec.registerOn(db.BusDomain.Delegate, d)
	}

	base := seedConsistencyBase(t, ctx, db)
	uid := base.userID

	// run executes a handler and asserts its declared mods fired.
	run := func(t *testing.T, actionType string, h interface {
		Execute(context.Context, json.RawMessage, workflow.ActionExecutionContext) (any, error)
		GetEntityModifications(json.RawMessage) []workflow.EntityModification
	}, cfg json.RawMessage, execCtx workflow.ActionExecutionContext) {
		t.Helper()
		mark := rec.mark()
		if _, err := h.Execute(ctx, cfg, execCtx); err != nil {
			t.Fatalf("%s Execute: %v", actionType, err)
		}
		assertDeclaredFired(t, actionType, h.GetEntityModifications(cfg), rec.since(mark))
	}

	// 1. resolve_approval_request → approvalrequest.updated
	t.Run("resolve_approval_request", func(t *testing.T) {
		wf, err := workflow.TestSeedFullWorkflow(ctx, uid, db.BusDomain.Workflow)
		if err != nil {
			t.Fatalf("seeding full workflow: %v", err)
		}
		if len(wf.AutomationExecutions) == 0 || len(wf.AutomationRules) == 0 {
			t.Fatalf("TestSeedFullWorkflow produced no executions/rules")
		}
		req, err := db.BusDomain.ApprovalRequest.Create(ctx, approvalrequestbus.NewApprovalRequest{
			ExecutionID:     wf.AutomationExecutions[0].ID,
			RuleID:          wf.AutomationRules[0].ID,
			ActionName:      "resolve_consistency",
			Approvers:       []uuid.UUID{uid},
			ApprovalType:    approvalrequestbus.ApprovalTypeAny,
			TimeoutHours:    48,
			TaskToken:       "tok-consistency-" + uuid.NewString()[:8],
			ApprovalMessage: "consistency test",
		})
		if err != nil {
			t.Fatalf("seeding pending approval request: %v", err)
		}
		h := approval.NewResolveApprovalHandler(db.Log, db.BusDomain.ApprovalRequest)
		cfg := mustJSON(t, map[string]any{"approval_request_id": req.ID.String(), "resolution": approvalrequestbus.StatusApproved, "reason": "ok"})
		run(t, "resolve_approval_request", h, cfg, workflow.ActionExecutionContext{UserID: uid})
	})

	// 2. approve_inventory_adjustment → inventoryadjustment.updated
	t.Run("approve_inventory_adjustment", func(t *testing.T) {
		adj, err := db.BusDomain.InventoryAdjustment.Create(ctx, inventoryadjustmentbus.NewInventoryAdjustment{
			ProductID: base.productIDs[0], LocationID: base.loc0, AdjustedBy: uid,
			QuantityChange: 10, ReasonCode: "cycle_count", Notes: "consistency", AdjustmentDate: time.Now(),
		})
		if err != nil {
			t.Fatalf("seeding pending adjustment: %v", err)
		}
		h := inventory.NewApproveInventoryAdjustmentHandler(db.Log, db.BusDomain.InventoryAdjustment)
		cfg := mustJSON(t, map[string]any{"adjustment_id": adj.InventoryAdjustmentID.String(), "approval_reason": "ok"})
		run(t, "approve_inventory_adjustment", h, cfg, workflow.ActionExecutionContext{UserID: uid})
	})

	// 3. reject_inventory_adjustment → inventoryadjustment.updated
	t.Run("reject_inventory_adjustment", func(t *testing.T) {
		adj, err := db.BusDomain.InventoryAdjustment.Create(ctx, inventoryadjustmentbus.NewInventoryAdjustment{
			ProductID: base.productIDs[0], LocationID: base.loc0, AdjustedBy: uid,
			QuantityChange: 5, ReasonCode: "cycle_count", Notes: "consistency", AdjustmentDate: time.Now(),
		})
		if err != nil {
			t.Fatalf("seeding pending adjustment: %v", err)
		}
		h := inventory.NewRejectInventoryAdjustmentHandler(db.Log, db.BusDomain.InventoryAdjustment)
		cfg := mustJSON(t, map[string]any{"adjustment_id": adj.InventoryAdjustmentID.String(), "rejection_reason": "bad count"})
		run(t, "reject_inventory_adjustment", h, cfg, workflow.ActionExecutionContext{UserID: uid})
	})

	// 4. approve_transfer_order → transferorder.updated
	t.Run("approve_transfer_order", func(t *testing.T) {
		to, err := db.BusDomain.TransferOrder.Create(ctx, transferorderbus.NewTransferOrder{
			ProductID: base.productIDs[0], FromLocationID: base.loc0, ToLocationID: base.loc1,
			RequestedByID: uid, Quantity: 5, Status: transferorderbus.StatusPending, TransferDate: time.Now(),
		})
		if err != nil {
			t.Fatalf("seeding pending transfer order: %v", err)
		}
		h := inventory.NewApproveTransferOrderHandler(db.Log, db.BusDomain.TransferOrder)
		cfg := mustJSON(t, map[string]any{"transfer_order_id": to.TransferID.String(), "approval_reason": "ok"})
		run(t, "approve_transfer_order", h, cfg, workflow.ActionExecutionContext{UserID: uid})
	})

	// 5. reject_transfer_order → transferorder.updated
	t.Run("reject_transfer_order", func(t *testing.T) {
		to, err := db.BusDomain.TransferOrder.Create(ctx, transferorderbus.NewTransferOrder{
			ProductID: base.productIDs[0], FromLocationID: base.loc0, ToLocationID: base.loc1,
			RequestedByID: uid, Quantity: 7, Status: transferorderbus.StatusPending, TransferDate: time.Now(),
		})
		if err != nil {
			t.Fatalf("seeding pending transfer order: %v", err)
		}
		h := inventory.NewRejectTransferOrderHandler(db.Log, db.BusDomain.TransferOrder)
		cfg := mustJSON(t, map[string]any{"transfer_order_id": to.TransferID.String(), "rejection_reason": "bad"})
		run(t, "reject_transfer_order", h, cfg, workflow.ActionExecutionContext{UserID: uid})
	})

	// 6. approve_purchase_order → purchaseorder.updated
	t.Run("approve_purchase_order", func(t *testing.T) {
		po, err := db.BusDomain.PurchaseOrder.Create(ctx, newPendingPO(base, "APPROVE"))
		if err != nil {
			t.Fatalf("seeding pending PO: %v", err)
		}
		h := procurement.NewApprovePurchaseOrderHandler(db.Log, db.BusDomain.PurchaseOrder)
		cfg := mustJSON(t, map[string]any{"purchase_order_id": po.ID.String(), "approval_reason": "ok"})
		run(t, "approve_purchase_order", h, cfg, workflow.ActionExecutionContext{UserID: uid})
	})

	// 7. reject_purchase_order → purchaseorder.updated
	t.Run("reject_purchase_order", func(t *testing.T) {
		po, err := db.BusDomain.PurchaseOrder.Create(ctx, newPendingPO(base, "REJECT"))
		if err != nil {
			t.Fatalf("seeding pending PO: %v", err)
		}
		h := procurement.NewRejectPurchaseOrderHandler(db.Log, db.BusDomain.PurchaseOrder)
		cfg := mustJSON(t, map[string]any{"purchase_order_id": po.ID.String(), "rejection_reason": "wrong supplier"})
		run(t, "reject_purchase_order", h, cfg, workflow.ActionExecutionContext{UserID: uid})
	})

	// 8. allocate_inventory → inventory_items fires; allocation_results known-silent (M2).
	t.Run("allocate_inventory", func(t *testing.T) {
		if _, err := inventoryitembus.TestSeedInventoryItems(ctx, 1, []uuid.UUID{base.loc0}, []uuid.UUID{base.productIDs[0]}, db.BusDomain.InventoryItem); err != nil {
			t.Fatalf("seeding inventory item: %v", err)
		}
		h := inventory.NewAllocateInventoryHandler(db.Log, db.DB, db.BusDomain.InventoryItem, db.BusDomain.InventoryLocation, db.BusDomain.InventoryTransaction, db.BusDomain.Product, db.BusDomain.Workflow)
		cfg := mustJSON(t, map[string]any{
			"inventory_items":     []map[string]any{{"product_id": base.productIDs[0].String(), "quantity": 10}},
			"allocation_mode":     "allocate",
			"allocation_strategy": "fifo",
			"allow_partial":       false,
			"priority":            "high",
		})
		ruleID := uuid.New()
		run(t, "allocate_inventory", h, cfg, workflow.ActionExecutionContext{UserID: uid, ExecutionID: uuid.New(), RuleID: &ruleID})
	})

	// 9. reserve_inventory → inventory_items fires.
	t.Run("reserve_inventory", func(t *testing.T) {
		if _, err := inventoryitembus.TestSeedInventoryItems(ctx, 1, []uuid.UUID{base.loc1}, []uuid.UUID{base.productIDs[1]}, db.BusDomain.InventoryItem); err != nil {
			t.Fatalf("seeding inventory item: %v", err)
		}
		h := inventory.NewReserveInventoryHandler(db.Log, db.DB, db.BusDomain.InventoryItem, db.BusDomain.Workflow)
		cfg := mustJSON(t, map[string]any{
			"product_id":                 base.productIDs[1].String(),
			"quantity":                   10,
			"allocation_strategy":        "fifo",
			"reservation_duration_hours": 24,
		})
		ruleID := uuid.New()
		run(t, "reserve_inventory", h, cfg, workflow.ActionExecutionContext{UserID: uid, ExecutionID: uuid.New(), RuleID: &ruleID})
	})

	// 10. create_purchase_order → purchaseorder.created + purchaseorderlineitem.created.
	t.Run("create_purchase_order", func(t *testing.T) {
		h := procurement.NewCreatePurchaseOrderHandler(db.Log, db.DB, db.BusDomain.PurchaseOrder, db.BusDomain.PurchaseOrderLineItem, db.BusDomain.SupplierProduct)
		cfg := mustJSON(t, map[string]any{
			"supplier_id":              base.supplierID.String(),
			"purchase_order_status_id": base.poStatusID.String(),
			"delivery_warehouse_id":    base.warehouseID.String(),
			"delivery_location_id":     base.loc0.String(),
			"currency_id":              base.currencyID.String(),
			"expected_delivery_days":   7,
			"notes":                    "consistency PO",
			"line_items": []map[string]any{{
				"product_id":          base.productIDs[0].String(),
				"supplier_product_id": base.supplierProductID.String(),
				"quantity_ordered":    10,
				"unit_cost":           25.00,
				"line_item_status_id": base.lineItemStatusID.String(),
				"notes":               "test line item",
			}},
		})
		run(t, "create_purchase_order", h, cfg, workflow.ActionExecutionContext{UserID: uid})
	})

	// 11. create_put_away_task → putawaytask.created (needs quantity_received delta > 0).
	t.Run("create_put_away_task", func(t *testing.T) {
		h := inventory.NewCreatePutAwayTaskHandler(db.Log, db.BusDomain.PutAwayTask, db.BusDomain.SupplierProduct, db.BusDomain.PurchaseOrder)
		cfg := mustJSON(t, map[string]any{
			"source_from_po":    false,
			"product_id":        base.productIDs[0].String(),
			"location_strategy": "static",
			"location_id":       base.loc0.String(),
			"reference_number":  "STATIC-CONSISTENCY",
		})
		execCtx := workflow.ActionExecutionContext{
			UserID: uid,
			FieldChanges: map[string]workflow.FieldChange{
				"quantity_received": {OldValue: float64(0), NewValue: float64(10)},
			},
		}
		run(t, "create_put_away_task", h, cfg, execCtx)
	})

	// 12. update_field known-silent trip-wire: raw SQL, fires NOTHING today (M1 gap).
	t.Run("update_field_known_silent", func(t *testing.T) {
		items, err := inventoryitembus.TestSeedInventoryItems(ctx, 1, []uuid.UUID{base.loc1}, []uuid.UUID{base.productIDs[0]}, db.BusDomain.InventoryItem)
		if err != nil {
			t.Fatalf("seeding inventory item: %v", err)
		}
		h := data.NewUpdateFieldHandler(db.Log, db.DB)
		cfg := mustJSON(t, map[string]any{
			"target_entity": "inventory.inventory_items",
			"target_field":  "quantity",
			"new_value":     999,
			"conditions":    []map[string]any{{"field_name": "id", "operator": "equals", "value": items[0].ID.String()}},
		})
		mark := rec.mark()
		res, err := h.Execute(ctx, cfg, workflow.ActionExecutionContext{UserID: uid})
		if err != nil {
			t.Fatalf("update_field Execute: %v", err)
		}
		if m, ok := res.(map[string]any); ok && m["status"] != "success" {
			t.Fatalf("update_field did not succeed (status=%v) — trip-wire would be a false negative; fix the config", m["status"])
		}
		window := rec.since(mark)
		if contains(window, inventoryitembus.DomainName, workflow.ActionUpdated) {
			t.Errorf("update_field fired %s.updated — M1 is no longer silent. P4 has enabled cascade synthesis: remove update_field from knownSilent and add its declared==fired assertion.", inventoryitembus.DomainName)
		}
	})
}

// newPendingPO builds an unapproved/unrejected PO (the approve/reject precondition).
func newPendingPO(base baseFixtures, tag string) purchaseorderbus.NewPurchaseOrder {
	now := time.Now().UTC()
	return purchaseorderbus.NewPurchaseOrder{
		OrderNumber:           "PO-CONSISTENCY-" + tag + "-" + uuid.NewString()[:8],
		SupplierID:            base.supplierID,
		PurchaseOrderStatusID: base.poStatusID,
		DeliveryWarehouseID:   base.warehouseID,
		DeliveryLocationID:    uuid.Nil,
		DeliveryStreetID:      base.streetID,
		OrderDate:             now,
		ExpectedDeliveryDate:  now.Add(14 * 24 * time.Hour),
		Subtotal:              1000.00,
		TaxAmount:             80.00,
		ShippingCost:          50.00,
		TotalAmount:           1130.00,
		CurrencyID:            base.currencyID,
		RequestedBy:           base.userID,
		Notes:                 "consistency PO",
		CreatedBy:             base.userID,
	}
}
