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

// RejectInventoryAdjustmentConfig holds the config for the reject handler.
type RejectInventoryAdjustmentConfig struct {
	AdjustmentID    string `json:"adjustment_id"`
	RejectionReason string `json:"rejection_reason"`
}

// RejectInventoryAdjustmentHandler handles reject_inventory_adjustment actions.
type RejectInventoryAdjustmentHandler struct {
	log                    *logger.Logger
	inventoryAdjustmentBus *inventoryadjustmentbus.Business
}

// NewRejectInventoryAdjustmentHandler creates a new reject inventory adjustment handler.
func NewRejectInventoryAdjustmentHandler(log *logger.Logger, inventoryAdjustmentBus *inventoryadjustmentbus.Business) *RejectInventoryAdjustmentHandler {
	return &RejectInventoryAdjustmentHandler{log: log, inventoryAdjustmentBus: inventoryAdjustmentBus}
}

// GetType returns the action type.
func (h *RejectInventoryAdjustmentHandler) GetType() string { return "reject_inventory_adjustment" }

// IsAsync returns false — reject completes inline.
func (h *RejectInventoryAdjustmentHandler) IsAsync() bool { return false }

// SupportsManualExecution returns true.
func (h *RejectInventoryAdjustmentHandler) SupportsManualExecution() bool { return true }

// GetDescription returns a human-readable description.
func (h *RejectInventoryAdjustmentHandler) GetDescription() string {
	return "Reject a pending inventory adjustment, capturing rejector and reason for audit trail"
}

// Validate validates the reject inventory adjustment configuration.
func (h *RejectInventoryAdjustmentHandler) Validate(config json.RawMessage) error {
	var cfg RejectInventoryAdjustmentConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.AdjustmentID == "" {
		return fmt.Errorf("adjustment_id is required")
	}
	if _, err := uuid.Parse(cfg.AdjustmentID); err != nil {
		return fmt.Errorf("invalid adjustment_id: %w", err)
	}
	if cfg.RejectionReason == "" {
		return fmt.Errorf("rejection_reason is required")
	}
	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *RejectInventoryAdjustmentHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "rejected", Description: "Adjustment rejected successfully", IsDefault: true},
		{Name: "not_found", Description: "Adjustment not found"},
		{Name: "already_approved", Description: "Adjustment was already approved — cannot reject"},
		{Name: "already_rejected", Description: "Adjustment was already rejected (idempotent)"},
		{Name: "failure", Description: "Unexpected error"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *RejectInventoryAdjustmentHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{EntityName: "inventory.inventory_adjustments", EventType: "on_update", Fields: []string{"approval_status", "rejected_by", "rejection_reason"}},
	}
}

// Execute rejects a pending inventory adjustment.
func (h *RejectInventoryAdjustmentHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg RejectInventoryAdjustmentConfig
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
	case "approved":
		return map[string]any{"output": "already_approved", "adjustment_id": cfg.AdjustmentID}, nil
	case "rejected":
		return map[string]any{"output": "already_rejected", "adjustment_id": cfg.AdjustmentID}, nil
	}

	rejected, err := h.inventoryAdjustmentBus.Reject(ctx, ia, execCtx.UserID, cfg.RejectionReason)
	if err != nil {
		return nil, fmt.Errorf("reject adjustment: %w", err)
	}

	return map[string]any{
		"output":           "rejected",
		"adjustment_id":    rejected.InventoryAdjustmentID.String(),
		"rejected_by":      execCtx.UserID.String(),
		"rejection_reason": cfg.RejectionReason,
	}, nil
}
