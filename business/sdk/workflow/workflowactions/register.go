// business/sdk/workflow/workflowactions/register.go
package workflowactions

import (
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorylocationbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/products/productbus"
	"github.com/timmaaaz/ichor/business/domain/workflow/alertbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/approval"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/control"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/rabbitmq"
)

// TODO: initialize this in main passing all this fun stuff in

// ActionConfig holds all dependencies needed to register workflow actions
type ActionConfig struct {
	Log         *logger.Logger
	DB          *sqlx.DB
	QueueClient *rabbitmq.WorkflowQueue

	// Business layer dependencies
	Buses BusDependencies
}

// BusDependencies contains all business layer dependencies
type BusDependencies struct {
	// Inventory domain
	InventoryItem        *inventoryitembus.Business
	InventoryLocation    *inventorylocationbus.Business
	InventoryTransaction *inventorytransactionbus.Business
	Product              *productbus.Business
	Workflow             *workflow.Business

	// Workflow domain
	Alert *alertbus.Business
}

// RegisterAll registers all standard workflow actions using the config
func RegisterAll(registry *workflow.ActionRegistry, config ActionConfig) {
	// Control flow actions - only need log
	registry.Register(control.NewEvaluateConditionHandler(config.Log))

	// Data actions - only need log and db
	registry.Register(data.NewUpdateFieldHandler(config.Log, config.DB))

	// Approval actions
	registry.Register(approval.NewSeekApprovalHandler(config.Log, config.DB))

	// Communication actions
	registry.Register(communication.NewSendEmailHandler(config.Log, config.DB))
	registry.Register(communication.NewSendNotificationHandler(config.Log, config.DB))
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

// RegisterCoreActions registers action handlers that don't require RabbitMQ or heavy dependencies.
// This should be called even in test environments to enable cascade visualization.
// These handlers implement EntityModifier for cascade detection.
func RegisterCoreActions(registry *workflow.ActionRegistry, log *logger.Logger, db *sqlx.DB) {
	// Control flow actions - only need log
	registry.Register(control.NewEvaluateConditionHandler(log))

	// Data actions - only need log and db, implements EntityModifier for cascade
	registry.Register(data.NewUpdateFieldHandler(log, db))

	// Approval actions - only need log and db
	registry.Register(approval.NewSeekApprovalHandler(log, db))

	// Communication actions that don't need queue
	registry.Register(communication.NewSendEmailHandler(log, db))
	registry.Register(communication.NewSendNotificationHandler(log, db))
}
