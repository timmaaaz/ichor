package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryadjustmentbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ApproveInventoryAdjustmentConfig holds the config for the approve handler.
type ApproveInventoryAdjustmentConfig struct {
	AdjustmentID   string `json:"adjustment_id"`
	ApprovalReason string `json:"approval_reason,omitempty"`
}

// ApproveInventoryAdjustmentHandler handles approve_inventory_adjustment actions.
type ApproveInventoryAdjustmentHandler struct {
	log                    *logger.Logger
	inventoryAdjustmentBus *inventoryadjustmentbus.Business
}

// NewApproveInventoryAdjustmentHandler creates a new approve inventory adjustment handler.
func NewApproveInventoryAdjustmentHandler(log *logger.Logger, inventoryAdjustmentBus *inventoryadjustmentbus.Business) *ApproveInventoryAdjustmentHandler {
	return &ApproveInventoryAdjustmentHandler{log: log, inventoryAdjustmentBus: inventoryAdjustmentBus}
}

// GetType returns the action type.
func (h *ApproveInventoryAdjustmentHandler) GetType() string { return "approve_inventory_adjustment" }

// IsAsync returns false — approve completes inline.
func (h *ApproveInventoryAdjustmentHandler) IsAsync() bool { return false }

// SupportsManualExecution returns true.
func (h *ApproveInventoryAdjustmentHandler) SupportsManualExecution() bool { return true }

// GetDescription returns a human-readable description.
func (h *ApproveInventoryAdjustmentHandler) GetDescription() string {
	return "Approve a pending inventory adjustment, capturing approver and reason for audit trail"
}

// Validate validates the approve inventory adjustment configuration.
func (h *ApproveInventoryAdjustmentHandler) Validate(config json.RawMessage) error {
	var cfg ApproveInventoryAdjustmentConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.AdjustmentID == "" {
		return fmt.Errorf("adjustment_id is required")
	}
	if _, err := uuid.Parse(cfg.AdjustmentID); err != nil {
		return fmt.Errorf("invalid adjustment_id: %w", err)
	}
	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *ApproveInventoryAdjustmentHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "approved", Description: "Adjustment approved successfully", IsDefault: true},
		{Name: "not_found", Description: "Adjustment not found"},
		{Name: "already_approved", Description: "Adjustment was already approved (idempotent)"},
		{Name: "already_rejected", Description: "Adjustment was already rejected — cannot approve"},
		{Name: "failure", Description: "Unexpected error"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *ApproveInventoryAdjustmentHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{EntityName: "inventory.inventory_adjustments", EventType: "on_update", Fields: []string{"approval_status", "approved_by", "approval_reason"}},
	}
}

// Execute approves a pending inventory adjustment.
func (h *ApproveInventoryAdjustmentHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg ApproveInventoryAdjustmentConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return map[string]any{"output": "failure", "error": err.Error()}, nil
	}

	id, err := uuid.Parse(cfg.AdjustmentID)
	if err != nil {
		return map[string]any{"output": "failure", "error": "invalid adjustment_id"}, nil
	}

	ia, err := h.inventoryAdjustmentBus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, inventoryadjustmentbus.ErrNotFound) {
			return map[string]any{"output": "not_found", "adjustment_id": cfg.AdjustmentID}, nil
		}
		return nil, fmt.Errorf("query adjustment: %w", err)
	}

	switch ia.ApprovalStatus {
	case inventoryadjustmentbus.ApprovalStatusApproved:
		return map[string]any{"output": "already_approved", "adjustment_id": cfg.AdjustmentID}, nil
	case inventoryadjustmentbus.ApprovalStatusRejected:
		return map[string]any{"output": "already_rejected", "adjustment_id": cfg.AdjustmentID}, nil
	}

	approved, err := h.inventoryAdjustmentBus.Approve(ctx, ia, execCtx.UserID, cfg.ApprovalReason)
	if err != nil {
		return nil, fmt.Errorf("approve adjustment: %w", err)
	}

	return map[string]any{
		"output":          "approved",
		"adjustment_id":   approved.InventoryAdjustmentID.String(),
		"approved_by":     execCtx.UserID.String(),
		"approval_reason": cfg.ApprovalReason,
	}, nil
}
