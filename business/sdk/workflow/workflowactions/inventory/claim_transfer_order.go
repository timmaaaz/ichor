package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ClaimTransferOrderConfig holds the config for the claim transfer order handler.
type ClaimTransferOrderConfig struct {
	TransferOrderID string `json:"transfer_order_id"`
}

// ClaimTransferOrderHandler handles claim_transfer_order actions: it moves an
// approved transfer order to in_transit, recording who claimed it.
type ClaimTransferOrderHandler struct {
	log              *logger.Logger
	transferOrderBus *transferorderbus.Business
}

// NewClaimTransferOrderHandler creates a new claim transfer order handler.
func NewClaimTransferOrderHandler(log *logger.Logger, transferOrderBus *transferorderbus.Business) *ClaimTransferOrderHandler {
	return &ClaimTransferOrderHandler{log: log, transferOrderBus: transferOrderBus}
}

// GetType returns the action type.
func (h *ClaimTransferOrderHandler) GetType() string { return "claim_transfer_order" }

// IsAsync returns false — claim completes inline.
func (h *ClaimTransferOrderHandler) IsAsync() bool { return false }

// SupportsManualExecution returns true.
func (h *ClaimTransferOrderHandler) SupportsManualExecution() bool { return true }

// GetDescription returns a human-readable description.
func (h *ClaimTransferOrderHandler) GetDescription() string {
	return "Claim an approved transfer order, moving it to in_transit and recording the claimer"
}

// Validate validates the claim transfer order configuration.
func (h *ClaimTransferOrderHandler) Validate(config json.RawMessage) error {
	var cfg ClaimTransferOrderConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	if cfg.TransferOrderID == "" {
		return fmt.Errorf("transfer_order_id is required")
	}
	// A templated transfer_order_id (e.g. "{{entity_id}}") is resolved at runtime from the
	// execution-context entity; only a static value must parse as a UUID here.
	if !strings.Contains(cfg.TransferOrderID, "{{") {
		if _, err := uuid.Parse(cfg.TransferOrderID); err != nil {
			return fmt.Errorf("invalid transfer_order_id: %w", err)
		}
	}
	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *ClaimTransferOrderHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "claimed", Description: "Transfer order claimed (now in_transit)", IsDefault: true},
		{Name: "not_found", Description: "Transfer order not found"},
		{Name: "already_in_transit", Description: "Transfer order was already in_transit (idempotent)"},
		{Name: "already_completed", Description: "Transfer order was already completed — cannot claim"},
		{Name: "not_approved", Description: "Transfer order is not approved — cannot claim"},
		{Name: "failure", Description: "Unexpected error"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *ClaimTransferOrderHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{
			EntityName: "inventory.transfer_orders",
			EventType:  "on_update",
			Fields:     []string{"status", "claimed_by", "claimed_at"},
			// status moves to the fixed enum constant. claimed_by (runtime user) and
			// claimed_at (now) are left indeterminate (no Change entry).
			Changes: []workflow.ProducedChange{
				{FieldName: "status", Operator: workflow.OperatorChangedTo, Value: transferorderbus.StatusInTransit},
			},
		},
	}
}

// Execute claims an approved transfer order.
func (h *ClaimTransferOrderHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg ClaimTransferOrderConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return map[string]any{"output": "failure", "error": err.Error()}, nil
	}

	// Resolve the transfer order id. Default to the execution-context entity (button path,
	// where the page entity IS the transfer order). A static config UUID overrides; a
	// templated value (unresolved "{{...}}") falls back to the entity id.
	id := execCtx.EntityID
	if cfg.TransferOrderID != "" && !strings.Contains(cfg.TransferOrderID, "{{") {
		parsed, err := uuid.Parse(cfg.TransferOrderID)
		if err != nil {
			return map[string]any{"output": "failure", "error": "invalid transfer_order_id"}, nil
		}
		id = parsed
	}
	if id == uuid.Nil {
		return map[string]any{"output": "failure", "error": "no transfer order id"}, nil
	}

	to, err := h.transferOrderBus.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrNotFound) {
			return map[string]any{"output": "not_found", "transfer_order_id": id.String()}, nil
		}
		return nil, fmt.Errorf("query transfer order: %w", err)
	}

	switch to.Status {
	case transferorderbus.StatusInTransit:
		return map[string]any{"output": "already_in_transit", "transfer_order_id": id.String()}, nil
	case transferorderbus.StatusCompleted:
		return map[string]any{"output": "already_completed", "transfer_order_id": id.String()}, nil
	}
	if to.Status != transferorderbus.StatusApproved {
		return map[string]any{"output": "not_approved", "transfer_order_id": id.String(), "current_status": to.Status}, nil
	}

	claimed, err := h.transferOrderBus.Claim(ctx, to, execCtx.UserID)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
			return map[string]any{"output": "failure", "error": "transfer order status changed concurrently"}, nil
		}
		return nil, fmt.Errorf("claim transfer order: %w", err)
	}

	return map[string]any{
		"output":            "claimed",
		"transfer_order_id": claimed.TransferID.String(),
		"claimed_by":        execCtx.UserID.String(),
	}, nil
}
