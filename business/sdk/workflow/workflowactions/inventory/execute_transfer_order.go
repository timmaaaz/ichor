package inventory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ExecuteTransferOrderConfig holds the config for the execute transfer order handler.
type ExecuteTransferOrderConfig struct {
	TransferOrderID string `json:"transfer_order_id"`
}

// ExecuteTransferOrderHandler handles execute_transfer_order actions: it completes an
// in_transit transfer order and performs the atomic stock move — status -> completed, a
// TRANSFER_OUT/TRANSFER_IN ledger pair, and source-decrement/destination-increment of
// inventory_items — in a single transaction, mirroring transferorderapp.Execute so the
// button path is equivalent to the REST execute endpoint.
type ExecuteTransferOrderHandler struct {
	log               *logger.Logger
	db                *sqlx.DB
	transferOrderBus  *transferorderbus.Business
	invTransactionBus *inventorytransactionbus.Business
	invItemBus        *inventoryitembus.Business
}

// NewExecuteTransferOrderHandler creates a new execute transfer order handler.
func NewExecuteTransferOrderHandler(
	log *logger.Logger,
	db *sqlx.DB,
	transferOrderBus *transferorderbus.Business,
	invTransactionBus *inventorytransactionbus.Business,
	invItemBus *inventoryitembus.Business,
) *ExecuteTransferOrderHandler {
	return &ExecuteTransferOrderHandler{
		log:               log,
		db:                db,
		transferOrderBus:  transferOrderBus,
		invTransactionBus: invTransactionBus,
		invItemBus:        invItemBus,
	}
}

// GetType returns the action type.
func (h *ExecuteTransferOrderHandler) GetType() string { return "execute_transfer_order" }

// IsAsync returns false — execute completes inline.
func (h *ExecuteTransferOrderHandler) IsAsync() bool { return false }

// SupportsManualExecution returns true.
func (h *ExecuteTransferOrderHandler) SupportsManualExecution() bool { return true }

// GetDescription returns a human-readable description.
func (h *ExecuteTransferOrderHandler) GetDescription() string {
	return "Complete an in-transit transfer order, moving it to completed and recording the completer"
}

// Validate validates the execute transfer order configuration.
func (h *ExecuteTransferOrderHandler) Validate(config json.RawMessage) error {
	var cfg ExecuteTransferOrderConfig
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
func (h *ExecuteTransferOrderHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "executed", Description: "Transfer order completed and stock moved", IsDefault: true},
		{Name: "not_found", Description: "Transfer order not found"},
		{Name: "already_completed", Description: "Transfer order was already completed (idempotent)"},
		{Name: "not_in_transit", Description: "Transfer order is not in_transit — cannot execute"},
		{Name: "insufficient_stock", Description: "Source location has insufficient stock to complete the move"},
		{Name: "failure", Description: "Unexpected error"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *ExecuteTransferOrderHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{
			EntityName: "inventory.transfer_orders",
			EventType:  "on_update",
			Fields:     []string{"status", "completed_by", "completed_at"},
			// status moves to the fixed enum constant. completed_by (runtime user) and
			// completed_at (now) are left indeterminate (no Change entry).
			Changes: []workflow.ProducedChange{
				{FieldName: "status", Operator: workflow.OperatorChangedTo, Value: transferorderbus.StatusCompleted},
			},
		},
	}
}

// Execute completes an in-transit transfer order.
func (h *ExecuteTransferOrderHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg ExecuteTransferOrderConfig
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

	if to.Status == transferorderbus.StatusCompleted {
		return map[string]any{"output": "already_completed", "transfer_order_id": id.String()}, nil
	}
	if to.Status != transferorderbus.StatusInTransit {
		return map[string]any{"output": "not_in_transit", "transfer_order_id": id.String(), "current_status": to.Status}, nil
	}

	// Complete the transfer and move the stock atomically: status -> completed, a
	// TRANSFER_OUT/TRANSFER_IN ledger pair, source decrement, destination increment — all in
	// one transaction, mirroring transferorderapp.Execute so the button equals the REST path.
	tx, err := h.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Carry the tx on ctx so each cascade bus's WriteAtomic JOINs THIS transaction
	// instead of opening its own — the status flip, the TRANSFER_OUT/IN ledger writes,
	// and their outbox.Emit rows all commit or roll back atomically with the stock move.
	ctx = sqldb.WithTx(ctx, tx)

	txTransferBus, err := h.transferOrderBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("tx transfer order bus: %w", err)
	}
	executed, err := txTransferBus.Execute(ctx, to, execCtx.UserID)
	if err != nil {
		if errors.Is(err, transferorderbus.ErrInvalidTransferStatus) {
			return map[string]any{"output": "failure", "error": "transfer order status changed concurrently"}, nil
		}
		return nil, fmt.Errorf("execute transfer order: %w", err)
	}

	txInvTxnBus, err := h.invTransactionBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("tx inventory transaction bus: %w", err)
	}
	refNum := to.TransferID.String()
	now := time.Now()
	if _, err := txInvTxnBus.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       to.ProductID,
		LocationID:      to.FromLocationID,
		UserID:          execCtx.UserID,
		Quantity:        -to.Quantity,
		TransactionType: "TRANSFER_OUT",
		ReferenceNumber: refNum,
		TransactionDate: now,
	}); err != nil {
		return nil, fmt.Errorf("create transfer_out transaction: %w", err)
	}
	if _, err := txInvTxnBus.Create(ctx, inventorytransactionbus.NewInventoryTransaction{
		ProductID:       to.ProductID,
		LocationID:      to.ToLocationID,
		UserID:          execCtx.UserID,
		Quantity:        to.Quantity,
		TransactionType: "TRANSFER_IN",
		ReferenceNumber: refNum,
		TransactionDate: now,
	}); err != nil {
		return nil, fmt.Errorf("create transfer_in transaction: %w", err)
	}

	txItemBus, err := h.invItemBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("tx inventory item bus: %w", err)
	}
	if err := txItemBus.DecrementQuantity(ctx, to.ProductID, to.FromLocationID, to.Quantity); err != nil {
		if errors.Is(err, inventoryitembus.ErrInsufficientStock) {
			return map[string]any{"output": "insufficient_stock", "transfer_order_id": id.String()}, nil
		}
		return nil, fmt.Errorf("decrement source inventory: %w", err)
	}
	if err := txItemBus.UpsertQuantity(ctx, to.ProductID, to.ToLocationID, to.Quantity); err != nil {
		return nil, fmt.Errorf("upsert destination inventory: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return map[string]any{
		"output":            "executed",
		"transfer_order_id": executed.TransferID.String(),
		"completed_by":      execCtx.UserID.String(),
	}, nil
}
