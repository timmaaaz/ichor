package dbtest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/config/settingsbus"
	"github.com/timmaaaz/ichor/business/domain/config/settingsbus/levers"
	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
)

// Test_Seed_Integration runs the full Phase 0g.B4 seed chain against a fresh
// dbtest database and asserts the post-seed shape:
//   - 79 label rows (19 location + 20 container + 40 product)
//   - 40 products with tracking distribution {28 none, 8 lot, 4 serial}
//   - ≥1 user with assigned_zones containing "STG-A"
//   - ≥1 user with assigned_zones containing "STG-C"
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
}
