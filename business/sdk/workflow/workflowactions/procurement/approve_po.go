package procurement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ApprovePurchaseOrderConfig holds the config for the approve PO handler.
type ApprovePurchaseOrderConfig struct {
	PurchaseOrderID string `json:"purchase_order_id"`
	ApprovalReason  string `json:"approval_reason,omitempty"`
}

// ApprovePurchaseOrderHandler handles approve_purchase_order actions.
type ApprovePurchaseOrderHandler struct {
	log             *logger.Logger
	purchaseOrderBus *purchaseorderbus.Business
}

// NewApprovePurchaseOrderHandler creates a new approve purchase order handler.
func NewApprovePurchaseOrderHandler(log *logger.Logger, purchaseOrderBus *purchaseorderbus.Business) *ApprovePurchaseOrderHandler {
	return &ApprovePurchaseOrderHandler{log: log, purchaseOrderBus: purchaseOrderBus}
}

// GetType returns the action type.
func (h *ApprovePurchaseOrderHandler) GetType() string { return "approve_purchase_order" }

// IsAsync returns false — approve completes inline.
func (h *ApprovePurchaseOrderHandler) IsAsync() bool { return false }

// SupportsManualExecution returns true.
func (h *ApprovePurchaseOrderHandler) SupportsManualExecution() bool { return true }

// GetDescription returns a human-readable description.
func (h *ApprovePurchaseOrderHandler) GetDescription() string {
	return "Approve a purchase order, capturing approver and date for audit trail"
}

// Validate validates the approve purchase order configuration.
func (h *ApprovePurchaseOrderHandler) Validate(config json.RawMessage) error {
	var cfg ApprovePurchaseOrderConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.PurchaseOrderID == "" {
		return fmt.Errorf("purchase_order_id is required")
	}
	if _, err := uuid.Parse(cfg.PurchaseOrderID); err != nil {
		return fmt.Errorf("invalid purchase_order_id: %w", err)
	}
	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *ApprovePurchaseOrderHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "approved", Description: "Purchase order approved successfully", IsDefault: true},
		{Name: "not_found", Description: "Purchase order not found"},
		{Name: "already_approved", Description: "Purchase order was already approved (idempotent)"},
		{Name: "already_rejected", Description: "Purchase order was already rejected — cannot approve"},
		{Name: "failure", Description: "Unexpected error"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *ApprovePurchaseOrderHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{EntityName: "procurement.purchase_orders", EventType: "on_update", Fields: []string{"approved_by", "approved_date", "approval_reason"}},
	}
}

// Execute approves a purchase order.
func (h *ApprovePurchaseOrderHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg ApprovePurchaseOrderConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return map[string]any{"output": "failure", "error": err.Error()}, nil
	}

	id, err := uuid.Parse(cfg.PurchaseOrderID)
	if err != nil {
		return map[string]any{"output": "failure", "error": "invalid purchase_order_id"}, nil
	}

	po, err := h.purchaseOrderBus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, purchaseorderbus.ErrNotFound) {
			return map[string]any{"output": "not_found", "purchase_order_id": cfg.PurchaseOrderID}, nil
		}
		return nil, fmt.Errorf("query purchase order: %w", err)
	}

	// Guard: already approved
	if po.ApprovedBy != (uuid.UUID{}) {
		return map[string]any{"output": "already_approved", "purchase_order_id": cfg.PurchaseOrderID}, nil
	}
	// Guard: already rejected
	if po.RejectedBy != (uuid.UUID{}) {
		return map[string]any{"output": "already_rejected", "purchase_order_id": cfg.PurchaseOrderID}, nil
	}

	approved, err := h.purchaseOrderBus.Approve(ctx, po, execCtx.UserID, cfg.ApprovalReason)
	if err != nil {
		if errors.Is(err, purchaseorderbus.ErrAlreadyApproved) {
			return map[string]any{"output": "already_approved", "purchase_order_id": cfg.PurchaseOrderID}, nil
		}
		return nil, fmt.Errorf("approve purchase order: %w", err)
	}

	return map[string]any{
		"output":          "approved",
		"purchase_order_id": approved.ID.String(),
		"approved_by":     execCtx.UserID.String(),
		"approval_reason": cfg.ApprovalReason,
	}, nil
}
