package workflowactions

import (
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/approval"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/communication"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/data"
	"github.com/timmaaaz/ichor/business/sdk/workflow/workflowactions/inventory"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// RegisterAll registers all standard workflow actions
func RegisterAll(registry *workflow.ActionRegistry, log *logger.Logger, db *sqlx.DB) {
	// Data actions
	registry.Register(data.NewUpdateFieldHandler(log, db))

	// Approval actions
	registry.Register(approval.NewSeekApprovalHandler(log, db))

	// Communication actions
	registry.Register(communication.NewSendEmailHandler(log, db))
	registry.Register(communication.NewSendNotificationHandler(log, db))
	registry.Register(communication.NewCreateAlertHandler(log, db))

	// Inventory actions
	registry.Register(inventory.NewAllocateInventoryHandler(log, db))
}

// RegisterCategory registers actions from a specific category
func RegisterDataActions(registry *workflow.ActionRegistry, log *logger.Logger, db *sqlx.DB) {
	registry.Register(data.NewUpdateFieldHandler(log, db))
}

func RegisterApprovalActions(registry *workflow.ActionRegistry, log *logger.Logger, db *sqlx.DB) {
	registry.Register(approval.NewSeekApprovalHandler(log, db))
}

func RegisterCommunicationActions(registry *workflow.ActionRegistry, log *logger.Logger, db *sqlx.DB) {
	registry.Register(communication.NewSendEmailHandler(log, db))
	registry.Register(communication.NewSendNotificationHandler(log, db))
	registry.Register(communication.NewCreateAlertHandler(log, db))
}

func RegisterInventoryActions(registry *workflow.ActionRegistry, log *logger.Logger, db *sqlx.DB) {
	registry.Register(inventory.NewAllocateInventoryHandler(log, db))
}
