package workflowactions

import (
	userdb "github.com/timmaaaz/ichor/business/domain/core/userbus/stores/userdb"
	transferorderdb "github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus/stores/transferorderdb"
	purchaseorderdb "github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus/stores/purchaseorderdb"
	purchaseorderlineitemdb "github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus/stores/purchaseorderlineitemdb"
	orderlineitemsdb "github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus/stores/orderlineitemsdb"
	ordersdb "github.com/timmaaaz/ichor/business/domain/sales/ordersbus/stores/ordersdb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/protected"
)

// genericDataActions are the polymorphic raw-SQL handlers that the protected-list guards.
// They are the SUBJECT of the block, never a source of protected fields, so the auto-source
// skips them (with a nil config they would yield an empty entity name anyway).
var genericDataActions = map[string]bool{
	"update_field":      true,
	"create_entity":     true,
	"transition_status": true,
}

// PopulateProtected fills the protected-field registry from its two domain-declared sources
// (DESIGN §10), and must be called AFTER every action handler has been registered:
//
//  1. Auto-source — fields a typed action claims via an on_update GetEntityModifications
//     manifest are protected and routed to that action. on_create claims (the create
//     constructors, e.g. create_purchase_order / create_put_away_task) are deliberately
//     EXCLUDED: protected-ness guards illegal state transitions on existing rows, while
//     create-time validation is the bus-routing follow-up (FOLLOW_UP F1), not this list.
//  2. Domain-declared db-model tags — each store package's RegisterProtected reflects its
//     own `protected:"true"` columns (the authoritative names the raw-SQL handlers target),
//     covering fields no on_update action claims yet (drift / Protect-only / NEEDS-NEW-ACTION).
//
// Plus the append-only inventory ledger, which is never generic-writable at all.
func PopulateProtected(reg *protected.Registry, registry *workflow.ActionRegistry) {
	// (1) auto-source: on_update manifest claims.
	for _, actionType := range registry.GetAll() {
		if genericDataActions[actionType] {
			continue
		}
		h, ok := registry.Get(actionType)
		if !ok {
			continue
		}
		em, ok := h.(workflow.EntityModifier)
		if !ok {
			continue
		}
		for _, mod := range em.GetEntityModifications(nil) {
			if mod.EventType != "on_update" || mod.EntityName == "" {
				continue
			}
			for _, field := range mod.Fields {
				reg.ProtectField(mod.EntityName, field, actionType)
			}
		}
	}

	// (2) domain-declared protected columns (co-located db-model tags).
	purchaseorderdb.RegisterProtected(reg)
	transferorderdb.RegisterProtected(reg)
	ordersdb.RegisterProtected(reg)
	orderlineitemsdb.RegisterProtected(reg)
	purchaseorderlineitemdb.RegisterProtected(reg)
	userdb.RegisterProtected(reg)

	// (3) whole-table: the inventory transaction ledger is append-only.
	reg.ProtectEntity("inventory.inventory_transactions", "")
}
