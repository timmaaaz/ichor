package execution_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
	"github.com/timmaaaz/ichor/app/domain/workflow/executionapp"
	"github.com/timmaaaz/ichor/business/domain/core/contactinfosbus"
	"github.com/timmaaaz/ichor/business/domain/geography/citybus"
	"github.com/timmaaaz/ichor/business/domain/geography/regionbus"
	"github.com/timmaaaz/ichor/business/domain/geography/streetbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/domain/inventory/zonebus"
	"github.com/timmaaaz/ichor/business/domain/products/brandbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/products/productcategorybus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/stores/workflowdb"
	workflowtemporal "github.com/timmaaaz/ichor/business/sdk/workflow/temporal"
	"github.com/timmaaaz/ichor/business/sdk/workflow/temporal/stores/edgedb"
	foundationtemporal "github.com/timmaaaz/ichor/foundation/temporal"
	temporalclient "go.temporal.io/sdk/client"
)

// =============================================================================
// Task 9 — End-to-end over-order recovery loop.
//
// Proves the headline of the over-order remediation feature: an over-order
// shortfall routes to the over_order alert + approval branch (Deliverable A/C),
// and RE-RUNNING the execution after stock is restored actually re-attempts the
// reservation against LIVE stock and succeeds — NOT a cached failure. The fresh
// execution id minted by RerunExecution (Task 4) clears the dedup walls
// (allocation_results idempotency key, Temporal workflow-id REJECT_DUPLICATE,
// execution-record upsert), so reserve_inventory runs clean the second time.
//
// This exercises Tasks 1–8 together; it adds NO production code. The only test
// infra it needed beyond Task 7 is StartTestWithTemporalGranular — Task 7's
// StartTestWithTemporal worker registers only 4 core handlers, so a dispatched
// Rule-5 graph would orphan at its first check_inventory/reserve_inventory
// activity. The granular helper registers the full pipeline handler set on the
// SAME production task queue (temporal.TaskQueue) the all.go-built trigger
// dispatches to, sharing the identical composition-root path.
//
// ── check passes but reserve fails (the over-order condition) ────────────────
// check_inventory is quantity-aware: when sourcing from a line item it requires
// available >= max(threshold, requested_qty) (check_inventory.go), and it sums
// available across ALL items for the product. reserve_inventory's
// QueryAvailableForAllocation is capped at LIMIT 10 (inventoryitemdb.go). Both
// use the same per-item availability formula (quantity - reserved - allocated),
// so the divergence is purely that cap:
//
//   12 inventory items for the product, each quantity=1 (available total = 12),
//   line requests 11:
//     • check_inventory: required=max(1,11)=11, available=12 >= 11 -> "sufficient"
//     • reserve_inventory: sees only the first 10 items -> 10 reservable < 11,
//       allow_partial=false -> soft-fail "insufficient_stock" (rolls back; no
//       allocation_results row, no committed reservation).
//
//   Restock (simulate receive): bump every item to quantity=2.
//     • check: available=24 >= 11 -> "sufficient"
//     • reserve: first 10 items -> 20 reservable >= 11 -> "success".
//
// On the insufficient_stock branch the workflow parks on the async seek_approval
// hold, so the ORIGINAL execution stays "running" — we assert the over_order
// alert, not original completion. The rerun takes the success branch (no async
// node), so its execution reaches StatusCompleted and reserves live stock.
// =============================================================================

const (
	overOrderLineQty   = 11 // line item requests 11
	overOrderItemCount = 12 // 12 thin items: check sees all (>=11), reserve caps at 10 (<11)
)

func Test_OverOrder_RerunSucceedsAfterRestock(t *testing.T) {
	// Intentionally NOT parallel: this worker polls the production task queue
	// (temporal.TaskQueue), shared with Test_Execution_Rerun in this package.
	// Running serially avoids cross-test activity contention on that queue.

	test := apitest.StartTestWithTemporalGranular(t, "Test_OverOrder_RerunSucceedsAfterRestock")
	db := test.DB
	ctx := context.Background()

	// Seed the full Go seed chain (migrate.Seed / seed.sql alone does NOT create
	// Rule 5 — only seedWorkflow, reached via InsertSeedDataWithDB, does).
	if err := dbtest.InsertSeedDataWithDB(db.Log, db.DB); err != nil {
		t.Fatalf("InsertSeedDataWithDB (seeds Rule 5): %v", err)
	}

	// Admin token + permissions for the rerun POST. The /rerun route runs through
	// AuthorizeTable (OPA RuleAdminOnly + a table-access permission on
	// workflow.automation_executions), so a bare admin OPA role is not enough — the
	// user also needs a role with table access. insertSeedData (seed_test.go) wires
	// exactly that (admin user -> role -> tableaccess on the executions table); reuse
	// it so the auth flow matches the existing rerun_test.go path.
	sd, err := insertSeedData(db, test.Auth)
	if err != nil {
		t.Fatalf("insertSeedData (admin + table permissions): %v", err)
	}
	adminUserID := sd.Users[0].User.ID
	adminToken := sd.Users[0].Token

	// ── Locate Rule 5 (Granular Inventory Pipeline) ──────────────────────────
	const ruleName = "Line Item Created - Granular Inventory Pipeline"
	rules, err := db.BusDomain.Workflow.QueryActiveRules(ctx)
	if err != nil {
		t.Fatalf("QueryActiveRules: %v", err)
	}
	var rule workflow.AutomationRule
	for _, r := range rules {
		if r.Name == ruleName {
			rule = r
			break
		}
	}
	if rule.ID == uuid.Nil {
		t.Fatalf("rule %q not found among %d active rules", ruleName, len(rules))
	}

	// ── Seed a clean product with controlled, low on-hand stock ──────────────
	productID, locationID, items := seedOverOrderProduct(t, ctx, db, adminUserID)
	orderID := uuid.New() // synthetic order reference; reserve only uses it as reference_id

	// ── Build our own production-queue trigger to fire the ORIGINAL event ─────
	// all.go's trigger processor (inside the mux) is Initialize()d at mux-build
	// time, BEFORE InsertSeedDataWithDB created Rule 5, so it cannot match Rule 5
	// for a relay-driven create. We dispatch the original event ourselves through
	// a trigger Initialize()d AFTER seeding, on the SAME production task queue
	// (WithTaskQueue(temporal.TaskQueue)), so the granular worker runs the graph.
	// The rerun, by contrast, loads the graph fresh from the DB by rule id and
	// needs no rule-cache entry, so it works straight through the mux.
	trigger := newProductionQueueTrigger(t, db)

	originalEvent := workflow.TriggerEvent{
		EventType:  "on_create",
		EntityName: "order_line_items",
		EntityID:   uuid.New(), // the line item id
		Timestamp:  time.Now(),
		UserID:     adminUserID,
		RawData: map[string]any{
			"product_id": productID.String(),
			"quantity":   float64(overOrderLineQty),
			"order_id":   orderID.String(),
		},
	}
	if err := trigger.OnEntityEvent(ctx, originalEvent); err != nil {
		t.Fatalf("firing original over-order event: %v", err)
	}

	// ── Assert the over_order alert exists, SCOPED to Rule 5 ─────────────────
	// (workflow-alerts.md: never assert against global alert totals.)
	overOrder := "over_order"
	var alert alertbus.Alert
	if !pollFor(20, 500*time.Millisecond, func() bool {
		alerts, qErr := db.BusDomain.Alert.Query(ctx, alertbus.QueryFilter{
			AlertType:    &overOrder,
			SourceRuleID: &rule.ID,
		}, alertbus.DefaultOrderBy, page.MustParse("1", "10"))
		if qErr != nil {
			t.Fatalf("querying over_order alerts: %v", qErr)
		}
		if len(alerts) > 0 {
			alert = alerts[0]
			return true
		}
		return false
	}) {
		t.Fatal("timeout: no over_order alert (scoped to Rule 5) after the original run — " +
			"reserve_inventory did not soft-fail to the over-order branch")
	}
	t.Logf("over_order alert created: id=%s title=%q action_url=%q source_rule=%s",
		alert.ID, alert.Title, alert.ActionURL, alert.SourceRuleID)

	// Find the ORIGINATING execution: the over_order alert deep-links to it via
	// its ActionURL "/workflow/executions/{{execution_id}}" (Task 2 enrichment).
	// This is the operator's path too (click the alert -> open the execution ->
	// rerun), so resolving the id from the alert mirrors the real flow exactly.
	originalExecID := executionIDFromActionURL(t, alert.ActionURL)
	t.Logf("originating execution (from alert action_url): %s", originalExecID)

	// ── Restock: raise on-hand so the requested qty now fits (simulate receive) ─
	for _, it := range items {
		newQty := 2
		if _, uErr := db.BusDomain.InventoryItem.Update(ctx, it, inventoryitembus.UpdateInventoryItem{
			Quantity: &newQty,
		}); uErr != nil {
			t.Fatalf("restocking inventory item %s: %v", it.ID, uErr)
		}
	}
	t.Logf("restocked %d items to quantity=2 each (location=%s product=%s)", len(items), locationID, productID)

	// Snapshot the over_order alert count (scoped to Rule 5) BEFORE the rerun, so
	// we can assert the success-branch rerun adds none (robust against any unrelated
	// re-fire of the rule).
	overOrderCountBefore := countOverOrderAlerts(t, ctx, db, rule.ID, overOrder)

	// ── POST /rerun (admin) -> 200 + fresh new_execution_id ──────────────────
	var rerunResp executionapp.RerunResponse
	doRerun(t, test, adminToken, originalExecID, &rerunResp)
	if rerunResp.OriginalExecutionID != originalExecID {
		t.Fatalf("rerun original_execution_id = %s, want %s", rerunResp.OriginalExecutionID, originalExecID)
	}
	if rerunResp.NewExecutionID == uuid.Nil {
		t.Fatal("rerun new_execution_id must not be nil")
	}
	if rerunResp.NewExecutionID == originalExecID {
		t.Fatal("rerun new_execution_id must differ from the original (fresh id clears the dedup walls)")
	}
	newExecID := rerunResp.NewExecutionID
	t.Logf("rerun dispatched: original=%s new=%s", originalExecID, newExecID)

	// ── Poll the NEW execution: assert reserve now SUCCEEDS against live stock ─
	// Proof it is a fresh attempt (not the cached insufficient_stock): the new
	// execution reaches StatusCompleted (the success branch has no async node),
	// AND the product's reserved_quantity is now > 0 (the original run rolled its
	// reservation back, so this can only come from the re-run's live reserve).
	if !pollFor(60, 500*time.Millisecond, func() bool {
		exec, qErr := db.BusDomain.Workflow.QueryExecutionByID(ctx, newExecID)
		if qErr != nil {
			return false
		}
		return exec.Status == workflow.StatusCompleted
	}) {
		exec, _ := db.BusDomain.Workflow.QueryExecutionByID(ctx, newExecID)
		t.Fatalf("timeout: re-run execution %s did not reach completed (status=%q err=%q) — "+
			"reserve_inventory may not have succeeded against live stock", newExecID, exec.Status, exec.ErrorMessage)
	}

	totalReserved := totalReservedForProduct(t, ctx, db, productID)
	if totalReserved < overOrderLineQty {
		t.Fatalf("re-run reserved %d of product %s, want >= %d — the fresh attempt did not reserve live stock",
			totalReserved, productID, overOrderLineQty)
	}

	// And the success branch must NOT have raised any new over_order alert: the
	// re-run reserved cleanly, so the over_order count (scoped to Rule 5) is
	// unchanged from before the rerun.
	overOrderCountAfter := countOverOrderAlerts(t, ctx, db, rule.ID, overOrder)
	if overOrderCountAfter != overOrderCountBefore {
		t.Fatalf("over_order alert count changed across the rerun (before=%d after=%d) — "+
			"the re-run should have taken the success branch, not re-alerted",
			overOrderCountBefore, overOrderCountAfter)
	}

	t.Logf("SUCCESS: over-order recovery loop verified — original run alerted over_order, "+
		"re-run (execution %s) reserved %d units of live stock and completed.", newExecID, totalReserved)
}

// =============================================================================
// Helpers
// =============================================================================

// newProductionQueueTrigger builds a WorkflowTrigger that dispatches to the
// production task queue (temporal.TaskQueue) — the SAME queue the granular helper's
// worker polls — with a trigger processor Initialize()d now (after Rule 5 was
// seeded), so it matches Rule 5 for the synthetic order_line_items event.
func newProductionQueueTrigger(t *testing.T, db *dbtest.Database) *workflowtemporal.WorkflowTrigger {
	t.Helper()
	ctx := context.Background()

	container := foundationtemporal.GetTestContainer(t)
	tc, err := temporalclient.Dial(temporalclient.Options{HostPort: container.HostPort})
	if err != nil {
		t.Fatalf("dialing temporal for trigger: %v", err)
	}
	t.Cleanup(func() { tc.Close() })

	workflowStore := workflowdb.NewStore(db.Log, db.DB)
	workflowBus := workflow.NewBusiness(db.Log, db.BusDomain.Delegate, workflowStore)
	edgeStore := edgedb.NewStore(db.Log, db.DB)

	tp := workflow.NewTriggerProcessor(db.Log, db.DB, workflowBus)
	if err := tp.Initialize(ctx); err != nil {
		t.Fatalf("initializing trigger processor: %v", err)
	}

	return workflowtemporal.NewWorkflowTrigger(db.Log, tc, tp, edgeStore, workflowStore).
		WithTaskQueue(workflowtemporal.TaskQueue)
}

// seedOverOrderProduct creates a fresh product (full brand/category chain) at a
// fresh warehouse location with overOrderItemCount thin inventory items, each
// quantity=1. Returns the product id, the location id, and the seeded items.
func seedOverOrderProduct(t *testing.T, ctx context.Context, db *dbtest.Database, adminID uuid.UUID) (uuid.UUID, uuid.UUID, []inventoryitembus.InventoryItem) {
	t.Helper()

	regions, err := db.BusDomain.Region.Query(ctx, regionbus.QueryFilter{}, regionbus.DefaultOrderBy, page.MustParse("1", "3"))
	if err != nil || len(regions) == 0 {
		t.Fatalf("querying regions: %v", err)
	}
	regionIDs := make([]uuid.UUID, len(regions))
	for i, r := range regions {
		regionIDs[i] = r.ID
	}

	cities, err := citybus.TestSeedCities(ctx, 1, regionIDs, db.BusDomain.City)
	if err != nil {
		t.Fatalf("seeding cities: %v", err)
	}
	streets, err := streetbus.TestSeedStreets(ctx, 2, []uuid.UUID{cities[0].ID}, db.BusDomain.Street)
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
	tzIDs := []uuid.UUID{tzs[0].ID}

	contacts, err := contactinfosbus.TestSeedContactInfos(ctx, 1, streetIDs, tzIDs, db.BusDomain.ContactInfos)
	if err != nil {
		t.Fatalf("seeding contact infos: %v", err)
	}
	brands, err := brandbus.TestSeedBrands(ctx, 1, []uuid.UUID{contacts[0].ID}, db.BusDomain.Brand)
	if err != nil {
		t.Fatalf("seeding brands: %v", err)
	}
	cats, err := productcategorybus.TestSeedProductCategories(ctx, 1, db.BusDomain.ProductCategory)
	if err != nil {
		t.Fatalf("seeding product categories: %v", err)
	}

	// Use productbus.Create directly (random v4 id + a unique SKU/UPC) rather than
	// TestSeedProducts: the latter mints deterministic seedid.Stable ids + fixed
	// SKU-0001, which collide with the products InsertSeedDataWithDB already seeded.
	uniq := uuid.New().String()[:8]
	product, err := db.BusDomain.Product.Create(ctx, productbus.NewProduct{
		SKU:               "OO-" + uniq,
		BrandID:           brands[0].BrandID,
		ProductCategoryID: cats[0].ProductCategoryID,
		Name:              "Over-Order Test Product " + uniq,
		Description:       "Product seeded by the over-order rerun e2e test",
		ModelNumber:       "OO-MDL-" + uniq,
		UpcCode:           "OO" + uniq,
		Status:            "active",
		IsActive:          true,
		TrackingType:      "none",
	})
	if err != nil {
		t.Fatalf("creating product: %v", err)
	}
	productID := product.ProductID

	warehouses, err := warehousebus.TestSeedWarehouses(ctx, 1, adminID, streetIDs, db.BusDomain.Warehouse)
	if err != nil {
		t.Fatalf("seeding warehouses: %v", err)
	}
	// 6 zones cover all 6 distinct zone codes (RCV/QA/STG/PCK/PKG/SHP), which is
	// what TestSeedInventoryLocations needs to emit all 19 spec-catalogue locations
	// (>= overOrderItemCount).
	zones, err := zonebus.TestSeedZone(ctx, 6, []uuid.UUID{warehouses[0].ID}, db.BusDomain.Zones)
	if err != nil {
		t.Fatalf("seeding zones: %v", err)
	}
	// inventory_items has a UNIQUE(product_id, location_id), so the thin items must
	// each live at a DISTINCT location. Seed overOrderItemCount locations.
	locations, err := inventorylocationbus.TestSeedInventoryLocations(ctx, overOrderItemCount, []uuid.UUID{warehouses[0].ID}, zones, db.BusDomain.InventoryLocation)
	if err != nil {
		t.Fatalf("seeding inventory locations: %v", err)
	}
	if len(locations) < overOrderItemCount {
		t.Fatalf("seeded %d locations, need %d", len(locations), overOrderItemCount)
	}

	// Thin items: one per location, each quantity=1. check_inventory (no location
	// filter) sums all overOrderItemCount (>= requested -> sufficient); reserve's
	// QueryAvailableForAllocation is capped at LIMIT 10 (< requested -> insufficient).
	items := make([]inventoryitembus.InventoryItem, 0, overOrderItemCount)
	for i := 0; i < overOrderItemCount; i++ {
		it, err := db.BusDomain.InventoryItem.Create(ctx, inventoryitembus.NewInventoryItem{
			LocationID:            locations[i].LocationID,
			ProductID:             productID,
			Quantity:              1,
			ReservedQuantity:      0,
			AllocatedQuantity:     0,
			MinimumStock:          0,
			MaximumStock:          100,
			ReorderPoint:          0,
			EconomicOrderQuantity: 50,
			SafetyStock:           0,
			AvgDailyUsage:         1,
		})
		if err != nil {
			t.Fatalf("creating thin inventory item %d: %v", i, err)
		}
		items = append(items, it)
	}

	t.Logf("seeded product=%s with %d items across %d locations (qty 1 each, total available=%d); line requests %d",
		productID, len(items), len(locations), len(items), overOrderLineQty)
	return productID, locations[0].LocationID, items
}

// executionIDFromActionURL parses the originating execution id out of the
// over_order alert's deep-link ActionURL ("/workflow/executions/{id}"), the same
// link an operator clicks to reach the rerun control. This is the durable signal
// for the originating run (Task 2 enrichment guarantees the {{execution_id}}
// substitution), and it avoids QueryExecutionHistory — which fails to scan the
// trigger-created record's NULL actions_executed column.
func executionIDFromActionURL(t *testing.T, actionURL string) uuid.UUID {
	t.Helper()
	const prefix = "/workflow/executions/"
	idx := strings.Index(actionURL, prefix)
	if idx < 0 {
		t.Fatalf("alert action_url %q does not contain %q", actionURL, prefix)
	}
	idStr := actionURL[idx+len(prefix):]
	id, err := uuid.Parse(idStr)
	if err != nil {
		t.Fatalf("parsing execution id from action_url %q: %v", actionURL, err)
	}
	return id
}

// countOverOrderAlerts returns the number of over_order alerts scoped to the rule.
func countOverOrderAlerts(t *testing.T, ctx context.Context, db *dbtest.Database, ruleID uuid.UUID, overOrder string) int {
	t.Helper()
	alerts, err := db.BusDomain.Alert.Query(ctx, alertbus.QueryFilter{
		AlertType:    &overOrder,
		SourceRuleID: &ruleID,
	}, alertbus.DefaultOrderBy, page.MustParse("1", "50"))
	if err != nil {
		t.Fatalf("counting over_order alerts: %v", err)
	}
	return len(alerts)
}

// totalReservedForProduct sums reserved_quantity across all inventory items for
// the product.
func totalReservedForProduct(t *testing.T, ctx context.Context, db *dbtest.Database, productID uuid.UUID) int {
	t.Helper()
	items, err := db.BusDomain.InventoryItem.Query(ctx, inventoryitembus.QueryFilter{
		ProductID: &productID,
	}, inventoryitembus.DefaultOrderBy, page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("querying inventory items for reserved total: %v", err)
	}
	total := 0
	for _, it := range items {
		total += it.ReservedQuantity
	}
	return total
}

// doRerun POSTs /v1/workflow/executions/{id}/rerun with the admin token and
// decodes the 200 RerunResponse.
func doRerun(t *testing.T, test *apitest.Test, token string, execID uuid.UUID, out *executionapp.RerunResponse) {
	t.Helper()

	url := fmt.Sprintf("/v1/workflow/executions/%s/rerun", execID)
	r := httptest.NewRequest(http.MethodPost, url, nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	test.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("POST %s = %d, want 200; body=%s", url, w.Code, w.Body.String())
	}
	if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
		t.Fatalf("decoding rerun response %q: %v", w.Body.String(), err)
	}
}

// pollFor calls fn up to attempts times, sleeping interval between, returning
// true on the first success.
func pollFor(attempts int, interval time.Duration, fn func() bool) bool {
	for i := 0; i < attempts; i++ {
		if fn() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}
