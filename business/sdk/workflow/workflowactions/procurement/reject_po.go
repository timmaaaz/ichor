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

// RejectPurchaseOrderConfig holds the config for the reject PO handler.
type RejectPurchaseOrderConfig struct {
	PurchaseOrderID string `json:"purchase_order_id"`
	RejectionReason string `json:"rejection_reason"`
}

// RejectPurchaseOrderHandler handles reject_purchase_order actions.
type RejectPurchaseOrderHandler struct {
	log             *logger.Logger
	purchaseOrderBus *purchaseorderbus.Business
}

// NewRejectPurchaseOrderHandler creates a new reject purchase order handler.
func NewRejectPurchaseOrderHandler(log *logger.Logger, purchaseOrderBus *purchaseorderbus.Business) *RejectPurchaseOrderHandler {
	return &RejectPurchaseOrderHandler{log: log, purchaseOrderBus: purchaseOrderBus}
}

// GetType returns the action type.
func (h *RejectPurchaseOrderHandler) GetType() string { return "reject_purchase_order" }

// IsAsync returns false — reject completes inline.
func (h *RejectPurchaseOrderHandler) IsAsync() bool { return false }

// SupportsManualExecution returns true.
func (h *RejectPurchaseOrderHandler) SupportsManualExecution() bool { return true }

// GetDescription returns a human-readable description.
func (h *RejectPurchaseOrderHandler) GetDescription() string {
	return "Reject a purchase order, capturing rejector and reason for audit trail"
}

// Validate validates the reject purchase order configuration.
func (h *RejectPurchaseOrderHandler) Validate(config json.RawMessage) error {
	var cfg RejectPurchaseOrderConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.PurchaseOrderID == "" {
		return fmt.Errorf("purchase_order_id is required")
	}
	if _, err := uuid.Parse(cfg.PurchaseOrderID); err != nil {
		return fmt.Errorf("invalid purchase_order_id: %w", err)
	}
	if cfg.RejectionReason == "" {
		return fmt.Errorf("rejection_reason is required")
	}
	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *RejectPurchaseOrderHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "rejected", Description: "Purchase order rejected successfully", IsDefault: true},
		{Name: "not_found", Description: "Purchase order not found"},
		{Name: "already_approved", Description: "Purchase order was already approved — cannot reject"},
		{Name: "already_rejected", Description: "Purchase order was already rejected (idempotent)"},
		{Name: "failure", Description: "Unexpected error"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *RejectPurchaseOrderHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{EntityName: "procurement.purchase_orders", EventType: "on_update", Fields: []string{"rejected_by", "rejected_date", "rejection_reason"}},
	}
}

// Execute rejects a purchase order.
func (h *RejectPurchaseOrderHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg RejectPurchaseOrderConfig
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

	rejected, err := h.purchaseOrderBus.Reject(ctx, po, execCtx.UserID, cfg.RejectionReason)
	if err != nil {
		if errors.Is(err, purchaseorderbus.ErrAlreadyApproved) {
			return map[string]any{"output": "already_approved", "purchase_order_id": cfg.PurchaseOrderID}, nil
		}
		if errors.Is(err, purchaseorderbus.ErrAlreadyRejected) {
			return map[string]any{"output": "already_rejected", "purchase_order_id": cfg.PurchaseOrderID}, nil
		}
		return nil, fmt.Errorf("reject purchase order: %w", err)
	}

	return map[string]any{
		"output":           "rejected",
		"purchase_order_id": rejected.ID.String(),
		"rejected_by":      execCtx.UserID.String(),
		"rejection_reason": cfg.RejectionReason,
	}, nil
}
