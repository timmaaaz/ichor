package scenariodb

import (
	"slices"
	"strings"
	"testing"
)

// Test_orderForInsert_Empty asserts that an empty fixture set produces an
// empty slice and no error.
func Test_orderForInsert_Empty(t *testing.T) {
	t.Parallel()

	got, err := orderForInsert(map[string]struct{}{})
	if err != nil {
		t.Fatalf("orderForInsert(empty): unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("orderForInsert(empty) = %v, want empty slice", got)
	}
}

// Test_orderForInsert_SingleTable asserts that a single-table fixture set
// returns a one-element slice containing exactly that table.
func Test_orderForInsert_SingleTable(t *testing.T) {
	t.Parallel()

	in := map[string]struct{}{
		"inventory.inventory_items": {},
	}
	got, err := orderForInsert(in)
	if err != nil {
		t.Fatalf("orderForInsert(single): unexpected error: %v", err)
	}
	want := []string{"inventory.inventory_items"}
	if !slices.Equal(got, want) {
		t.Errorf("orderForInsert(single) = %v, want %v", got, want)
	}
}

// Test_orderForInsert_TwoDeepFKChain is the canonical regression case for
// follow-up B. The receive-lot-tracking scenario authors fixtures across
// inventory.lot_trackings (parent) and inventory.lot_locations (child).
// Because the child sorts alphabetically before the parent, the pre-fix
// SELECT DISTINCT order could (under any future PG plan change) return the
// child first and break the FK INSERT. orderForInsert must always produce
// parent-then-child by reversing scopedTables.
func Test_orderForInsert_TwoDeepFKChain(t *testing.T) {
	t.Parallel()

	in := map[string]struct{}{
		"inventory.lot_locations": {},
		"inventory.lot_trackings": {},
	}
	got, err := orderForInsert(in)
	if err != nil {
		t.Fatalf("orderForInsert(two-deep): unexpected error: %v", err)
	}
	want := []string{"inventory.lot_trackings", "inventory.lot_locations"}
	if !slices.Equal(got, want) {
		t.Errorf("orderForInsert(two-deep) = %v, want %v (parent must precede child)", got, want)
	}
}

// Test_orderForInsert_ThreeDeepFKChain covers the transfer-lot-tracked case
// where supplier_products → lot_trackings → lot_locations forms a 3-deep FK
// chain. Parents must come first.
func Test_orderForInsert_ThreeDeepFKChain(t *testing.T) {
	t.Parallel()

	in := map[string]struct{}{
		"inventory.lot_locations":       {},
		"inventory.lot_trackings":       {},
		"procurement.supplier_products": {},
	}
	got, err := orderForInsert(in)
	if err != nil {
		t.Fatalf("orderForInsert(three-deep): unexpected error: %v", err)
	}
	want := []string{
		"procurement.supplier_products",
		"inventory.lot_trackings",
		"inventory.lot_locations",
	}
	if !slices.Equal(got, want) {
		t.Errorf("orderForInsert(three-deep) = %v, want %v", got, want)
	}
}

// Test_orderForInsert_UnknownTable asserts a loud error when a fixture
// references a table not registered in scopedTables. This is the safety
// net for the "new scoped table added in a migration but missed in
// scopedTables" case.
func Test_orderForInsert_UnknownTable(t *testing.T) {
	t.Parallel()

	const bad = "inventory.does_not_exist"
	in := map[string]struct{}{
		bad: {},
	}
	got, err := orderForInsert(in)
	if err == nil {
		t.Fatalf("orderForInsert(unknown) = %v, want error", got)
	}
	if !strings.Contains(err.Error(), bad) {
		t.Errorf("orderForInsert(unknown) error %q does not mention offending table %q", err.Error(), bad)
	}
}
