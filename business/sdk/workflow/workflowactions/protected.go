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

	// (4) whole-table: tables a generic workflow action must NEVER mutate. These are
	// protected against the WORKFLOW write path only — the normal domain CRUD path
	// (bus.Create/Update, admin UI, REST endpoints, seeds) never consults this registry,
	// so humans still curate them freely. routeAction is "" (no typed action substitutes).
	//
	// ENGINE — a rule that rewrites rules/edges/executions makes the cascade-loop guard
	// undecidable and can forge the P1 lineage. The engine's own state is off-limits to
	// the automations it runs.
	reg.ProtectEntity("workflow.automation_rules", "")
	reg.ProtectEntity("workflow.rule_actions", "")
	reg.ProtectEntity("workflow.action_templates", "")
	reg.ProtectEntity("workflow.rule_dependencies", "")
	reg.ProtectEntity("workflow.trigger_types", "")
	reg.ProtectEntity("workflow.entity_types", "")
	reg.ProtectEntity("workflow.entities", "")
	reg.ProtectEntity("workflow.automation_executions", "")
	reg.ProtectEntity("workflow.notification_deliveries", "")

	// TABLE BUILDER — config.table_configs drives the dynamic table-builder UI; a workflow
	// rewriting it would reshape the app's own configuration surface.
	reg.ProtectEntity("config.table_configs", "")

	// RBAC — writing roles / role assignments / table access from a workflow is privilege
	// escalation. Access control is curated by humans, never by an automation.
	reg.ProtectEntity("core.roles", "")
	reg.ProtectEntity("core.user_roles", "")
	reg.ProtectEntity("core.table_access", "")

	// STATUS / REFERENCE DEFINITIONS — these rows' values are wired into both frontend and
	// backend code (status-name constants, badges); mutating a definition breaks those tie-ins.
	// Workflows still READ them as FK lookups to resolve an id by name (ProtectEntity blocks
	// writes TO the table, not its use as a ForeignKeyConfig.ReferenceTable); they must not
	// rewrite the definitions themselves.
	reg.ProtectEntity("sales.order_fulfillment_statuses", "")
	reg.ProtectEntity("sales.line_item_fulfillment_statuses", "")
	reg.ProtectEntity("procurement.purchase_order_statuses", "")
	reg.ProtectEntity("procurement.purchase_order_line_item_statuses", "")
	reg.ProtectEntity("hr.user_approval_status", "")

	// WAREHOUSE STRUCTURE — inventory_items reference locations by id, so a workflow that
	// restructured warehouses/zones/locations would orphan physical inventory. Automations
	// allocate/putaway INTO existing locations (writing inventory_items, which stays writable);
	// they must not create or restructure the facility itself.
	reg.ProtectEntity("inventory.warehouses", "")
	reg.ProtectEntity("inventory.zones", "")
	reg.ProtectEntity("inventory.inventory_locations", "")
}
