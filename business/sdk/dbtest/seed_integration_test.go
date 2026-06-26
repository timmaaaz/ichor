package dbtest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/assets/approvalstatusbus"
	"github.com/timmaaaz/ichor/business/domain/assets/fulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus/levers"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/dbtest/seedmodels"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// Test_Seed_Integration runs the full Phase 0g.B4 seed chain against a fresh
// dbtest database and asserts the post-seed shape:
//   - 79 label rows (19 location + 20 container + 40 product)
//   - 40 products with tracking distribution {28 none, 8 lot, 4 serial}
//   - ≥1 user with assigned_zones containing "STG-A"
//   - ≥1 user with assigned_zones containing "STG-C"
//   - the Release-to-Picking button gated on the PENDING/PROCESSING status UUIDs
func Test_Seed_Integration(t *testing.T) {
	t.Parallel()

	db := NewDatabase(t, "Test_Seed_Integration")

	if err := InsertSeedDataWithDB(db.Log, db.DB); err != nil {
		t.Fatalf("InsertSeedDataWithDB: %v", err)
	}

	ctx := context.Background()

	// --- labels split across types --------------------------------------
	expectedLabelsByType := map[string]int{
		labelbus.TypeLocation:  19,
		labelbus.TypeContainer: 20,
		labelbus.TypeProduct:   40,
	}
	expectedLabelTotal := 0
	for _, n := range expectedLabelsByType {
		expectedLabelTotal += n
	}

	pg := page.MustParse("1", "200")
	labels, err := db.BusDomain.Label.Query(ctx, labelbus.QueryFilter{}, labelbus.DefaultOrderBy, pg)
	if err != nil {
		t.Fatalf("label query: %v", err)
	}
	if got := len(labels); got != expectedLabelTotal {
		t.Fatalf("label count: got %d, want %d", got, expectedLabelTotal)
	}
	typeCounts := map[string]int{}
	for _, l := range labels {
		typeCounts[l.Type]++
	}
	for typ, want := range expectedLabelsByType {
		if got := typeCounts[typ]; got != want {
			t.Errorf("label type %q count: got %d, want %d", typ, got, want)
		}
	}

	// --- products with tracking distribution ----------------------------
	expectedProductsByTracking := map[string]int{"none": 28, "lot": 8, "serial": 4}
	expectedProductTotal := 0
	for _, n := range expectedProductsByTracking {
		expectedProductTotal += n
	}

	products, err := db.BusDomain.Product.Query(ctx, productbus.QueryFilter{}, productbus.DefaultOrderBy, pg)
	if err != nil {
		t.Fatalf("product query: %v", err)
	}
	if got := len(products); got != expectedProductTotal {
		t.Fatalf("product count: got %d, want %d", got, expectedProductTotal)
	}
	trackCounts := map[string]int{}
	for _, p := range products {
		trackCounts[p.TrackingType]++
	}
	for typ, want := range expectedProductsByTracking {
		if got := trackCounts[typ]; got != want {
			t.Errorf("tracking %q count: got %d, want %d", typ, got, want)
		}
	}

	// --- ≥1 user in STG-A and ≥1 user in STG-C ---------------------------
	zonedA, err := db.BusDomain.User.QueryByAssignedZone(ctx, "STG-A")
	if err != nil {
		t.Fatalf("query STG-A: %v", err)
	}
	if len(zonedA) < 1 {
		t.Errorf("expected ≥1 user assigned to STG-A, got %d", len(zonedA))
	}
	zonedC, err := db.BusDomain.User.QueryByAssignedZone(ctx, "STG-C")
	if err != nil {
		t.Fatalf("query STG-C: %v", err)
	}
	if len(zonedC) < 1 {
		t.Errorf("expected ≥1 user assigned to STG-C, got %d", len(zonedC))
	}

	// --- Phase 0g.B5 — assert 11 lever defaults present --------------------
	leverRows, err := db.BusDomain.Settings.Query(ctx, settingsbus.QueryFilter{}, order.NewBy(settingsbus.OrderByKey, order.ASC), page.MustParse("1", "100"))
	if err != nil {
		t.Fatalf("query settings: %v", err)
	}

	wantKeys := map[string]string{}
	for _, k := range levers.KnownKeys {
		wantKeys[k] = levers.Defaults[k]
	}

	gotKeys := map[string]string{}
	for _, s := range leverRows {
		// migrate.sql v2.01 pre-seeds non-lever numeric rows (e.g.
		// inventory.variance_threshold_units = 5) that would panic the
		// string unmarshal below. Skip anything not in the lever set.
		if _, isLever := wantKeys[s.Key]; !isLever {
			continue
		}
		var v string
		if err := json.Unmarshal(s.Value, &v); err != nil {
			t.Fatalf("unmarshal value for %q: %v", s.Key, err)
		}
		gotKeys[s.Key] = v
	}

	for k, want := range wantKeys {
		got, ok := gotKeys[k]
		if !ok {
			t.Errorf("missing lever key %q in config.settings after reseed", k)
			continue
		}
		if got != want {
			t.Errorf("lever %q value: got %q, want %q", k, got, want)
		}
	}

	// --- asset statuses are owned by seed.sql, not re-created by seed-frontend -
	// Regression guard: seed_assets.go used to Create approval/fulfillment
	// statuses that seed.sql already inserts. Those tables have no UNIQUE(name),
	// so both layers ran and left 10 rows / 5 duplicate names each (silent — no
	// error). seed_assets.go now Queries them instead, so the count must equal
	// the canonical name list exactly.
	statusPg := page.MustParse("1", "50")
	approvals, err := db.BusDomain.ApprovalStatus.Query(ctx, approvalstatusbus.QueryFilter{}, approvalstatusbus.DefaultOrderBy, statusPg)
	if err != nil {
		t.Fatalf("query approval statuses: %v", err)
	}
	if got, want := len(approvals), len(seedmodels.ApprovalStatusNames); got != want {
		t.Errorf("approval_status count: got %d, want %d (seed.sql owns these; seed-frontend must not duplicate)", got, want)
	}
	fulfillments, err := db.BusDomain.FulfillmentStatus.Query(ctx, fulfillmentstatusbus.QueryFilter{}, fulfillmentstatusbus.DefaultOrderBy, statusPg)
	if err != nil {
		t.Fatalf("query fulfillment statuses: %v", err)
	}
	if got, want := len(fulfillments), len(seedmodels.FulfillmentStatusNames); got != want {
		t.Errorf("fulfillment_status count: got %d, want %d (seed.sql owns these; seed-frontend must not duplicate)", got, want)
	}

	// --- Release-to-Picking button gates on PENDING/PROCESSING status UUIDs -----
	// The button's visibility is driven by valid_from_statuses in its action_config
	// (frontend canShowExecute string-matches it against the order's
	// order_fulfillment_status_id UUID). release_to_picking only accepts PENDING or
	// PROCESSING orders, so seedReleaseToPickingButton must resolve those two statuses
	// by name (IDs are random per environment) and embed their UUIDs.
	wantStatusIDs := map[string]bool{}
	for _, name := range []string{"PENDING", "PROCESSING"} {
		n := name
		st, err := db.BusDomain.OrderFulfillmentStatus.Query(ctx, orderfulfillmentstatusbus.QueryFilter{Name: &n}, orderfulfillmentstatusbus.DefaultOrderBy, page.MustParse("1", "1"))
		if err != nil {
			t.Fatalf("query %q fulfillment status: %v", name, err)
		}
		if len(st) != 1 {
			t.Fatalf("expected exactly 1 %q fulfillment status, got %d", name, len(st))
		}
		wantStatusIDs[st[0].ID.String()] = true
	}

	actions, err := db.BusDomain.PageAction.Query(ctx, pageactionbus.QueryFilter{}, pageactionbus.DefaultOrderBy, page.MustParse("1", "1000"))
	if err != nil {
		t.Fatalf("query page actions: %v", err)
	}
	var releaseBtn *pageactionbus.ButtonAction
	for i := range actions {
		if b := actions[i].Button; b != nil && b.ActionType == "release_to_picking" {
			releaseBtn = b
			break
		}
	}
	if releaseBtn == nil {
		t.Fatalf("release_to_picking button not found among %d page actions", len(actions))
	}
	var releaseCfg struct {
		OrderID           string   `json:"order_id"`
		ValidFromStatuses []string `json:"valid_from_statuses"`
	}
	if err := json.Unmarshal(releaseBtn.ActionConfig, &releaseCfg); err != nil {
		t.Fatalf("unmarshal release button action_config: %v", err)
	}
	// order_id is what the release_to_picking handler actually consumes; the gate must not
	// have displaced it.
	if releaseCfg.OrderID != "{{entity_id}}" {
		t.Errorf("release button order_id: got %q, want %q", releaseCfg.OrderID, "{{entity_id}}")
	}
	if got, want := len(releaseCfg.ValidFromStatuses), len(wantStatusIDs); got != want {
		t.Fatalf("release button valid_from_statuses: got %d entries %v, want %d (PENDING/PROCESSING IDs)", got, releaseCfg.ValidFromStatuses, want)
	}
	for _, id := range releaseCfg.ValidFromStatuses {
		if !wantStatusIDs[id] {
			t.Errorf("release button valid_from_statuses contains unexpected id %q (want PENDING/PROCESSING status IDs %v)", id, wantStatusIDs)
		}
	}
}
