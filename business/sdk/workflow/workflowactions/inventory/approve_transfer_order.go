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

// ApproveTransferOrderConfig holds the config for the approve transfer order handler.
type ApproveTransferOrderConfig struct {
	TransferOrderID string `json:"transfer_order_id"`
	ApprovalReason  string `json:"approval_reason,omitempty"`
}

// ApproveTransferOrderHandler handles approve_transfer_order actions.
type ApproveTransferOrderHandler struct {
	log              *logger.Logger
	transferOrderBus *transferorderbus.Business
}

// NewApproveTransferOrderHandler creates a new approve transfer order handler.
func NewApproveTransferOrderHandler(log *logger.Logger, transferOrderBus *transferorderbus.Business) *ApproveTransferOrderHandler {
	return &ApproveTransferOrderHandler{log: log, transferOrderBus: transferOrderBus}
}

// GetType returns the action type.
func (h *ApproveTransferOrderHandler) GetType() string { return "approve_transfer_order" }

// IsAsync returns false — approve completes inline.
func (h *ApproveTransferOrderHandler) IsAsync() bool { return false }

// SupportsManualExecution returns true.
func (h *ApproveTransferOrderHandler) SupportsManualExecution() bool { return true }

// GetDescription returns a human-readable description.
func (h *ApproveTransferOrderHandler) GetDescription() string {
	return "Approve a pending transfer order, capturing approver and reason for audit trail"
}

// Validate validates the approve transfer order configuration.
func (h *ApproveTransferOrderHandler) Validate(config json.RawMessage) error {
	var cfg ApproveTransferOrderConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.TransferOrderID == "" {
		return fmt.Errorf("transfer_order_id is required")
	}
	if _, err := uuid.Parse(cfg.TransferOrderID); err != nil {
		return fmt.Errorf("invalid transfer_order_id: %w", err)
	}
	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *ApproveTransferOrderHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "approved", Description: "Transfer order approved successfully", IsDefault: true},
		{Name: "not_found", Description: "Transfer order not found"},
		{Name: "already_approved", Description: "Transfer order was already approved (idempotent)"},
		{Name: "already_rejected", Description: "Transfer order was already rejected — cannot approve"},
		{Name: "failure", Description: "Unexpected error"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *ApproveTransferOrderHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{EntityName: "inventory.transfer_orders", EventType: "on_update", Fields: []string{"status", "approved_by_id", "approval_reason"}},
	}
}

// Execute approves a pending transfer order.
func (h *ApproveTransferOrderHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg ApproveTransferOrderConfig
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
	case transferorderbus.StatusApproved:
		return map[string]any{"output": "already_approved", "transfer_order_id": cfg.TransferOrderID}, nil
	case transferorderbus.StatusRejected:
		return map[string]any{"output": "already_rejected", "transfer_order_id": cfg.TransferOrderID}, nil
	}

	approved, err := h.transferOrderBus.Approve(ctx, to, execCtx.UserID, cfg.ApprovalReason)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
			return map[string]any{"output": "failure", "error": "transfer order status changed concurrently"}, nil
		}
		return nil, fmt.Errorf("approve transfer order: %w", err)
	}

	return map[string]any{
		"output":            "approved",
		"transfer_order_id": approved.TransferID.String(),
		"approved_by":       execCtx.UserID.String(),
		"approval_reason":   cfg.ApprovalReason,
	}, nil
}
