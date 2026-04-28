package dbtest

import (
	"context"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/labels/labelbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
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

	// --- 79 labels split 19 / 20 / 40 across types -----------------------
	pg := page.MustParse("1", "200")
	labels, err := db.BusDomain.Label.Query(ctx, labelbus.QueryFilter{}, labelbus.DefaultOrderBy, pg)
	if err != nil {
		t.Fatalf("label query: %v", err)
	}
	if got, want := len(labels), 79; got != want {
		t.Fatalf("label count: got %d, want %d", got, want)
	}
	typeCounts := map[string]int{}
	for _, l := range labels {
		typeCounts[l.Type]++
	}
	for typ, want := range map[string]int{
		labelbus.TypeLocation:  19,
		labelbus.TypeContainer: 20,
		labelbus.TypeProduct:   40,
	} {
		if got := typeCounts[typ]; got != want {
			t.Errorf("label type %q count: got %d, want %d", typ, got, want)
		}
	}

	// --- 40 products with 28/8/4 tracking distribution -------------------
	products, err := db.BusDomain.Product.Query(ctx, productbus.QueryFilter{}, productbus.DefaultOrderBy, pg)
	if err != nil {
		t.Fatalf("product query: %v", err)
	}
	if got, want := len(products), 40; got != want {
		t.Fatalf("product count: got %d, want %d", got, want)
	}
	trackCounts := map[string]int{}
	for _, p := range products {
		trackCounts[p.TrackingType]++
	}
	for typ, want := range map[string]int{"none": 28, "lot": 8, "serial": 4} {
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
}
