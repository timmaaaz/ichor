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
	// Data actions - only need log and db
	registry.Register(data.NewUpdateFieldHandler(config.Log, config.DB))

	// Approval actions
	registry.Register(approval.NewSeekApprovalHandler(config.Log, config.DB))

	// Communication actions
	registry.Register(communication.NewSendEmailHandler(config.Log, config.DB))
	registry.Register(communication.NewSendNotificationHandler(config.Log, config.DB))
	registry.Register(communication.NewCreateAlertHandler(config.Log, config.Buses.Alert))

	// Inventory actions - need additional dependencies
	registry.Register(inventory.NewAllocateInventoryHandler(
		config.Log,
		config.DB,
		config.QueueClient,
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
		config.QueueClient,
		config.Buses.InventoryItem,
		config.Buses.InventoryLocation,
		config.Buses.InventoryTransaction,
		config.Buses.Product,
		config.Buses.Workflow,
	))
}
