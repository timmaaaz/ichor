// business/sdk/workflow/workflowactions/register.go
package workflowactions

import (
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/picktaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/putawaytaskbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderfulfillmentstatusbus"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/approvalrequestbus"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/protected"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/approval"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/integration"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/procurement"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// TODO: initialize this in main passing all this fun stuff in

// ActionConfig holds all dependencies needed to register workflow actions
type ActionConfig struct {
	Log         *logger.Logger
	DB          *sqlx.DB
	QueueClient *rabbitmq.WorkflowQueue

	// EmailClient delivers outbound emails (e.g. via Resend). Nil = graceful degradation.
	EmailClient communication.EmailClient
	// EmailFrom is the sender address used when EmailClient is set.
	EmailFrom string

	// Delegate is fired by the generic data handlers after a successful raw-SQL write so
	// the write cascades to downstream automation (P4 M1). Nil = synthesis disabled.
	Delegate *delegate.Delegate
	// EntityRegistry maps a schema-qualified target table to the delegate (domain, bare
	// entity) the generic handlers fire under. Nil = synthesis disabled. Both this and
	// Delegate must be set (and non-empty) for cascades to fire. Build via
	// workflowdomains.ReverseMap().
	EntityRegistry map[string]data.EntityRef

	// Business layer dependencies
	Buses BusDependencies
}

// BusDependencies contains all business layer dependencies
type BusDependencies struct {
	// Inventory domain
	InventoryItem        *inventoryitembus.Business
	InventoryLocation    *inventorylocationbus.Business
	InventoryTransaction *inventorytransactionbus.Business
	InventoryAdjustment  *inventoryadjustmentbus.Business
	TransferOrder        *transferorderbus.Business
	PutAwayTask          *putawaytaskbus.Business
	PickTask             *picktaskbus.Business
	Product              *productbus.Business
	Workflow             *workflow.Business

	// Sales domain
	Orders                 *ordersbus.Business
	OrderLineItems         *orderlineitemsbus.Business
	OrderFulfillmentStatus *orderfulfillmentstatusbus.Business

	// Procurement domain
	PurchaseOrder         *purchaseorderbus.Business
	PurchaseOrderLineItem *purchaseorderlineitembus.Business
	SupplierProduct       *supplierproductbus.Business

	// Workflow domain
	Alert           *alertbus.Business
	ApprovalRequest *approvalrequestbus.Business
}

// RegisterAll registers all standard workflow actions using the config
func RegisterAll(registry *workflow.ActionRegistry, config ActionConfig) {
	// Protected-field registry: the generic data handlers consult it to reject writes to
	// invariant-bearing fields (DESIGN §10). Populated once all handlers are registered.
	reg := protected.New()

	// Control flow actions - only need log
	registry.Register(control.NewEvaluateConditionHandler(config.Log))

	// Control flow - delay
	registry.Register(control.NewDelayHandler(config.Log))

	// Data actions - only need log and db. The three generic raw-SQL handlers also take
	// the protected registry so they reject writes to guarded fields, and the delegate +
	// entity-registry so a successful write cascades to downstream automation (P4 M1).
	// Nil delegate/registry (e.g. tests) keeps synthesis off.
	dataOpts := []data.Option{
		data.WithProtectedRegistry(reg),
		data.WithDelegate(config.Delegate),
		data.WithEntityRegistry(config.EntityRegistry),
	}
	registry.Register(data.NewUpdateFieldHandler(config.Log, config.DB, dataOpts...))
	registry.Register(data.NewLookupEntityHandler(config.Log, config.DB))
	registry.Register(data.NewCreateEntityHandler(config.Log, config.DB, dataOpts...))
	registry.Register(data.NewTransitionStatusHandler(config.Log, config.DB, dataOpts...))
	registry.Register(data.NewAuditLogHandler(config.Log, config.DB))

	// Approval actions
	registry.Register(approval.NewSeekApprovalHandler(config.Log, config.DB, config.Buses.ApprovalRequest, config.Buses.Alert, config.QueueClient))
	registry.Register(approval.NewResolveApprovalHandler(config.Log, config.Buses.ApprovalRequest))

	// Communication actions
	registry.Register(communication.NewSendEmailHandler(config.Log, config.DB, config.EmailClient, config.EmailFrom))
	registry.Register(communication.NewSendNotificationHandler(config.Log, config.QueueClient))
	registry.Register(communication.NewCreateAlertHandler(config.Log, config.Buses.Alert, config.QueueClient))

	// Inventory actions - need additional dependencies
	registry.Register(inventory.NewAllocateInventoryHandler(
		config.Log,
		config.DB,
		config.Buses.InventoryItem,
		config.Buses.InventoryLocation,
		config.Buses.InventoryTransaction,
		config.Buses.Product,
		config.Buses.Workflow,
	))

	// Granular inventory actions
	RegisterGranularInventoryActions(registry, config)

	// Procurement actions
	RegisterProcurementActions(registry, config)

	// Integration actions
	registry.Register(integration.NewCallWebhookHandler(config.Log))

	// All handlers registered — derive the protected-field set (on_update manifest claims
	// + domain-declared db-model tags) so the generic data handlers enforce it.
	PopulateProtected(reg, registry)
}

// RegisterInventoryActions registers only inventory-related actions
func RegisterInventoryActions(registry *workflow.ActionRegistry, config ActionConfig) {
	registry.Register(inventory.NewAllocateInventoryHandler(
		config.Log,
		config.DB,
		config.Buses.InventoryItem,
		config.Buses.InventoryLocation,
		config.Buses.InventoryTransaction,
		config.Buses.Product,
		config.Buses.Workflow,
	))
}

// RegisterGranularInventoryActions registers the composable inventory actions.
// These are focused, single-purpose actions that can be chained in workflows.
func RegisterGranularInventoryActions(registry *workflow.ActionRegistry, config ActionConfig) {
	registry.Register(inventory.NewCheckInventoryHandler(config.Log, config.Buses.InventoryItem))
	registry.Register(inventory.NewCheckReorderPointHandler(config.Log, config.Buses.InventoryItem))
	registry.Register(inventory.NewReleaseReservationHandler(config.Log, config.DB, config.Buses.InventoryItem))
	registry.Register(inventory.NewCommitAllocationHandler(config.Log, config.DB, config.Buses.InventoryItem))
	registry.Register(inventory.NewReserveInventoryHandler(config.Log, config.DB, config.Buses.InventoryItem, config.Buses.Workflow))
	registry.Register(inventory.NewReceiveInventoryHandler(config.Log, config.DB, config.Buses.InventoryItem, config.Buses.InventoryTransaction, config.Buses.SupplierProduct))

	// release_to_picking flips a customer order PENDING/PROCESSING->PICKING and fans its
	// line items into inventory.pick_tasks. Registered unconditionally (like reserve/receive):
	// the handler nil-guards its buses at Execute time, and GetEntityModifications needs no
	// dependencies. all.go supplies the real Orders/OrderLineItems/PickTask/InventoryItem/
	// OrderFulfillmentStatus buses so the "Release to Picking" button can execute.
	registry.Register(inventory.NewReleaseToPickingHandler(
		config.Log,
		config.DB,
		config.Buses.Orders,
		config.Buses.OrderLineItems,
		config.Buses.PickTask,
		config.Buses.InventoryItem,
		config.Buses.OrderFulfillmentStatus,
	))

	if config.Buses.InventoryAdjustment != nil {
		registry.Register(inventory.NewApproveInventoryAdjustmentHandler(config.Log, config.Buses.InventoryAdjustment))
		registry.Register(inventory.NewRejectInventoryAdjustmentHandler(config.Log, config.Buses.InventoryAdjustment))
	}

	if config.Buses.TransferOrder != nil {
		registry.Register(inventory.NewApproveTransferOrderHandler(config.Log, config.Buses.TransferOrder))
		registry.Register(inventory.NewRejectTransferOrderHandler(config.Log, config.Buses.TransferOrder))
		registry.Register(inventory.NewClaimTransferOrderHandler(config.Log, config.Buses.TransferOrder))
		registry.Register(inventory.NewExecuteTransferOrderHandler(config.Log, config.Buses.TransferOrder))
	}

	if config.Buses.PutAwayTask != nil {
		registry.Register(inventory.NewCreatePutAwayTaskHandler(
			config.Log,
			config.Buses.PutAwayTask,
			config.Buses.SupplierProduct,
			config.Buses.PurchaseOrder,
		))
	}
}

// RegisterProcurementActions registers procurement-domain action handlers.
// These require procurement bus dependencies and are not included in RegisterCoreActions.
func RegisterProcurementActions(registry *workflow.ActionRegistry, config ActionConfig) {
	if config.Buses.PurchaseOrder != nil {
		registry.Register(procurement.NewCreatePurchaseOrderHandler(
			config.Log,
			config.DB,
			config.Buses.PurchaseOrder,
			config.Buses.PurchaseOrderLineItem,
			config.Buses.SupplierProduct,
		))
		registry.Register(procurement.NewApprovePurchaseOrderHandler(config.Log, config.Buses.PurchaseOrder))
		registry.Register(procurement.NewRejectPurchaseOrderHandler(config.Log, config.Buses.PurchaseOrder))
	}
}

// RegisterCoreActions registers action handlers that don't require RabbitMQ or heavy dependencies.
// This should be called even in test environments to enable cascade visualization.
// Entity-modifying handlers implement EntityModifier for cascade detection; communication
// and control flow handlers that don't modify entities are included for completeness.
//
// NOTE: unlike RegisterAll/all.go, this lightweight path does NOT wire the protected-list:
// the generic data handlers are constructed without WithProtectedRegistry and PopulateProtected
// is never called, so protected-field enforcement is a silent no-op here. That is fine for the
// cascade-visualization / loop-detection tests that use this path, but an enforcement test must
// build the registry via RegisterAll (or add the WithProtectedRegistry + PopulateProtected step)
// rather than relying on RegisterCoreActions.
func RegisterCoreActions(registry *workflow.ActionRegistry, log *logger.Logger, db *sqlx.DB) {
	// Control flow actions - only need log
	registry.Register(control.NewEvaluateConditionHandler(log))
	registry.Register(control.NewDelayHandler(log))

	// Data actions - only need log and db, implements EntityModifier for cascade
	registry.Register(data.NewUpdateFieldHandler(log, db))
	registry.Register(data.NewLookupEntityHandler(log, db))
	registry.Register(data.NewCreateEntityHandler(log, db))
	registry.Register(data.NewTransitionStatusHandler(log, db))
	registry.Register(data.NewAuditLogHandler(log, db))

	// Approval actions - nil buses for core path (graceful degradation)
	registry.Register(approval.NewSeekApprovalHandler(log, db, nil, nil, nil))
	registry.Register(approval.NewResolveApprovalHandler(log, nil))

	// Communication actions that don't need queue or email client (nil = graceful degradation)
	registry.Register(communication.NewSendEmailHandler(log, db, nil, ""))
	registry.Register(communication.NewSendNotificationHandler(log, nil))
	registry.Register(communication.NewCreateAlertHandler(log, nil, nil))

	// Inventory actions - nil buses for core path (cascade detection via EntityModifier)
	registry.Register(inventory.NewCreatePutAwayTaskHandler(log, nil, nil, nil))

	// Integration actions - no bus/DB/queue dependencies
	registry.Register(integration.NewCallWebhookHandler(log))
}
