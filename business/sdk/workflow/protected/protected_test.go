package protected_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/timmaaaz/ichor/business/sdk/workflow/protected"
)

func TestRegistry_EmptyAllowsEverything(t *testing.T) {
	r := protected.New()

	if route, blocked := r.Check("procurement.purchase_orders", "purchase_order_status_id"); blocked {
		t.Fatalf("empty registry blocked a write: route=%q blocked=%v", route, blocked)
	}
}

func TestRegistry_ProtectField(t *testing.T) {
	r := protected.New()
	r.ProtectField("procurement.purchase_orders", "purchase_order_status_id", "approve_purchase_order")

	route, blocked := r.Check("procurement.purchase_orders", "purchase_order_status_id")
	if !blocked {
		t.Fatal("expected protected field to be blocked")
	}
	if route != "approve_purchase_order" {
		t.Fatalf("route = %q, want approve_purchase_order", route)
	}

	// A different field on the same entity is not protected.
	if _, blocked := r.Check("procurement.purchase_orders", "notes"); blocked {
		t.Fatal("unrelated field should not be blocked")
	}

	// A different entity is not protected.
	if _, blocked := r.Check("sales.orders", "purchase_order_status_id"); blocked {
		t.Fatal("same field name on a different entity should not be blocked")
	}
}

func TestRegistry_ProtectEntity_WholeTableWins(t *testing.T) {
	r := protected.New()
	r.ProtectEntity("inventory.inventory_transactions", "")

	// Any field on a whole-table-protected entity is blocked.
	if _, blocked := r.Check("inventory.inventory_transactions", "quantity"); !blocked {
		t.Fatal("whole-table protect should block any field")
	}
	if _, blocked := r.Check("inventory.inventory_transactions", "anything_at_all"); !blocked {
		t.Fatal("whole-table protect should block an arbitrary field")
	}
}

func TestRegistry_EntityLevelOverridesField(t *testing.T) {
	r := protected.New()
	r.ProtectField("inventory.inventory_transactions", "quantity", "receive_inventory")
	r.ProtectEntity("inventory.inventory_transactions", "ledger")

	// Entity-level wins for the route on a non-field-specific column...
	if route, blocked := r.Check("inventory.inventory_transactions", "created_date"); !blocked || route != "ledger" {
		t.Fatalf("entity-level should block created_date with route=ledger, got route=%q blocked=%v", route, blocked)
	}
	// ...and the field still blocks (route may be either; assert blocked).
	if _, blocked := r.Check("inventory.inventory_transactions", "quantity"); !blocked {
		t.Fatal("field on a whole-table-protected entity must still be blocked")
	}
}

func TestCollectStructTags(t *testing.T) {
	// Mirrors a db store model: the `db` tag is the authoritative column name. Note the
	// column name (order_fulfillment_status_id) deliberately differs from any json name.
	type sample struct {
		Status    string `db:"order_fulfillment_status_id" protected:"true"`
		Name      string `db:"name"`
		Qty       int    `db:"quantity" protected:"true"`
		NoTag     string `protected:"true"`        // no db tag → not mappable to a column, skip
		Ignored   string `db:"-" protected:"true"` // db:"-" → skip
		NotMarked string `db:"not_marked"`
	}

	r := protected.New()
	protected.CollectStructTags(r, "sales.orders", "route_x", sample{})

	if route, blocked := r.Check("sales.orders", "order_fulfillment_status_id"); !blocked || route != "route_x" {
		t.Fatalf("order_fulfillment_status_id: route=%q blocked=%v, want route_x/true", route, blocked)
	}
	if _, blocked := r.Check("sales.orders", "quantity"); !blocked {
		t.Fatal("quantity should be collected from the protected tag")
	}
	if _, blocked := r.Check("sales.orders", "name"); blocked {
		t.Fatal("untagged field should not be protected")
	}
	if _, blocked := r.Check("sales.orders", "not_marked"); blocked {
		t.Fatal("field without protected tag should not be protected")
	}
	// Fields with no usable db tag are skipped (no panic, no bogus key).
	if _, blocked := r.Check("sales.orders", "NoTag"); blocked {
		t.Fatal("field without a db tag must be skipped")
	}
}

func TestNewError_WrapsSentinelAndIsClear(t *testing.T) {
	err := protected.NewError("procurement.purchase_orders", "purchase_order_status_id", "approve_purchase_order")

	if !errors.Is(err, protected.ErrProtectedField) {
		t.Fatal("NewError must wrap ErrProtectedField for errors.Is")
	}
	msg := err.Error()
	for _, want := range []string{"purchase_order_status_id", "procurement.purchase_orders", "approve_purchase_order"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error message %q missing %q", msg, want)
		}
	}
}

func TestNewError_NoRouteOmitsAction(t *testing.T) {
	err := protected.NewError("inventory.inventory_transactions", "quantity", "")
	if !errors.Is(err, protected.ErrProtectedField) {
		t.Fatal("must wrap sentinel")
	}
	if strings.Contains(err.Error(), "use the") {
		t.Fatalf("no-route error should not suggest an action: %q", err.Error())
	}
}
