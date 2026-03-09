package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// RejectTransferOrderConfig holds the config for the reject transfer order handler.
type RejectTransferOrderConfig struct {
	TransferOrderID string `json:"transfer_order_id"`
	RejectionReason string `json:"rejection_reason"`
}

// RejectTransferOrderHandler handles reject_transfer_order actions.
type RejectTransferOrderHandler struct {
	log              *logger.Logger
	transferOrderBus *transferorderbus.Business
}

// NewRejectTransferOrderHandler creates a new reject transfer order handler.
func NewRejectTransferOrderHandler(log *logger.Logger, transferOrderBus *transferorderbus.Business) *RejectTransferOrderHandler {
	return &RejectTransferOrderHandler{log: log, transferOrderBus: transferOrderBus}
}

// GetType returns the action type.
func (h *RejectTransferOrderHandler) GetType() string { return "reject_transfer_order" }

// IsAsync returns false — reject completes inline.
func (h *RejectTransferOrderHandler) IsAsync() bool { return false }

// SupportsManualExecution returns true.
func (h *RejectTransferOrderHandler) SupportsManualExecution() bool { return true }

// GetDescription returns a human-readable description.
func (h *RejectTransferOrderHandler) GetDescription() string {
	return "Reject a pending transfer order, capturing rejector and reason for audit trail"
}

// Validate validates the reject transfer order configuration.
func (h *RejectTransferOrderHandler) Validate(config json.RawMessage) error {
	var cfg RejectTransferOrderConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.TransferOrderID == "" {
		return fmt.Errorf("transfer_order_id is required")
	}
	if _, err := uuid.Parse(cfg.TransferOrderID); err != nil {
		return fmt.Errorf("invalid transfer_order_id: %w", err)
	}
	if cfg.RejectionReason == "" {
		return fmt.Errorf("rejection_reason is required")
	}
	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *RejectTransferOrderHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "rejected", Description: "Transfer order rejected successfully", IsDefault: true},
		{Name: "not_found", Description: "Transfer order not found"},
		{Name: "already_approved", Description: "Transfer order was already approved — cannot reject"},
		{Name: "already_rejected", Description: "Transfer order was already rejected (idempotent)"},
		{Name: "failure", Description: "Unexpected error"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *RejectTransferOrderHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{EntityName: "inventory.transfer_orders", EventType: "on_update", Fields: []string{"status", "rejected_by_id", "rejection_reason"}},
	}
}

// Execute rejects a pending transfer order.
func (h *RejectTransferOrderHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg RejectTransferOrderConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return map[string]any{"output": "failure", "error": err.Error()}, nil
	}

	id, err := uuid.Parse(cfg.TransferOrderID)
	if err != nil {
		return map[string]any{"output": "failure", "error": "invalid transfer_order_id"}, nil
	}

	to, err := h.transferOrderBus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return map[string]any{"output": "not_found", "transfer_order_id": cfg.TransferOrderID}, nil
		}
		return nil, fmt.Errorf("query transfer order: %w", err)
	}

	switch to.Status {
	case "approved":
		return map[string]any{"output": "already_approved", "transfer_order_id": cfg.TransferOrderID}, nil
	case "rejected":
		return map[string]any{"output": "already_rejected", "transfer_order_id": cfg.TransferOrderID}, nil
	}

	rejected, err := h.transferOrderBus.Reject(ctx, to, execCtx.UserID, cfg.RejectionReason)
	if err != nil {
		return nil, fmt.Errorf("reject transfer order: %w", err)
	}

	return map[string]any{
		"output":            "rejected",
		"transfer_order_id": rejected.TransferID.String(),
		"rejected_by":       execCtx.UserID.String(),
		"rejection_reason":  cfg.RejectionReason,
	}, nil
}
