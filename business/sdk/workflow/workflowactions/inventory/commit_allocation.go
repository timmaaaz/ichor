package inventory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// CommitAllocationConfig represents the configuration for committing a reservation to allocation.
type CommitAllocationConfig struct {
	ProductID  string `json:"product_id"`
	LocationID string `json:"location_id"`
	Quantity   int    `json:"quantity"`
}

// CommitAllocationResult holds the result of a commit allocation operation.
type CommitAllocationResult struct {
	ProductID         string `json:"product_id"`
	LocationID        string `json:"location_id"`
	PreviousReserved  int    `json:"previous_reserved"`
	NewReserved       int    `json:"new_reserved"`
	PreviousAllocated int    `json:"previous_allocated"`
	NewAllocated      int    `json:"new_allocated"`
	QuantityCommitted int    `json:"quantity_committed"`
}

// CommitAllocationHandler handles commit_allocation actions.
type CommitAllocationHandler struct {
	log              *logger.Logger
	db               *sqlx.DB
	inventoryItemBus *inventoryitembus.Business
}

// NewCommitAllocationHandler creates a new commit allocation handler.
func NewCommitAllocationHandler(
	log *logger.Logger,
	db *sqlx.DB,
	inventoryItemBus *inventoryitembus.Business,
) *CommitAllocationHandler {
	return &CommitAllocationHandler{
		log:              log,
		db:               db,
		inventoryItemBus: inventoryItemBus,
	}
}

// GetType returns the action type.
func (h *CommitAllocationHandler) GetType() string {
	return "commit_allocation"
}

// SupportsManualExecution returns true - allocations can be committed manually.
func (h *CommitAllocationHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - commit_allocation is a synchronous transactional operation.
func (h *CommitAllocationHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *CommitAllocationHandler) GetDescription() string {
	return "Commit reserved inventory to allocated status"
}

// Validate validates the commit allocation configuration.
func (h *CommitAllocationHandler) Validate(config json.RawMessage) error {
	var cfg CommitAllocationConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.ProductID == "" {
		return errors.New("product_id is required")
	}
	if _, err := uuid.Parse(cfg.ProductID); err != nil {
		return fmt.Errorf("invalid product_id: %w", err)
	}

	if cfg.LocationID == "" {
		return errors.New("location_id is required")
	}
	if _, err := uuid.Parse(cfg.LocationID); err != nil {
		return fmt.Errorf("invalid location_id: %w", err)
	}

	if cfg.Quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}

	return nil
}

// Execute commits reserved inventory to allocated within a transaction.
// Atomically: reserved -= N, allocated += N.
func (h *CommitAllocationHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg CommitAllocationConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return CommitAllocationResult{}, fmt.Errorf("failed to parse config: %w", err)
	}

	productID, err := uuid.Parse(cfg.ProductID)
	if err != nil {
		return CommitAllocationResult{}, fmt.Errorf("invalid product_id: %w", err)
	}

	locationID, err := uuid.Parse(cfg.LocationID)
	if err != nil {
		return CommitAllocationResult{}, fmt.Errorf("invalid location_id: %w", err)
	}

	// Begin transaction.
	tx, err := h.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return CommitAllocationResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	txItemBus, err := h.inventoryItemBus.NewWithTx(tx)
	if err != nil {
		return CommitAllocationResult{}, fmt.Errorf("create transactional bus: %w", err)
	}

	// Query the inventory item by product and location.
	filter := inventoryitembus.QueryFilter{
		ProductID:  &productID,
		LocationID: &locationID,
	}

	items, err := txItemBus.Query(ctx, filter, order.NewBy("id", order.ASC), page.MustParse("1", "1"))
	if err != nil {
		return CommitAllocationResult{}, fmt.Errorf("query inventory item: %w", err)
	}

	if len(items) == 0 {
		return CommitAllocationResult{}, fmt.Errorf("no inventory item found for product %s at location %s", productID, locationID)
	}

	item := items[0]

	// Validate sufficient reserved quantity.
	if item.ReservedQuantity < cfg.Quantity {
		return CommitAllocationResult{}, fmt.Errorf("insufficient reserved quantity: have %d, need %d", item.ReservedQuantity, cfg.Quantity)
	}

	previousReserved := item.ReservedQuantity
	previousAllocated := item.AllocatedQuantity
	newReserved := previousReserved - cfg.Quantity
	newAllocated := previousAllocated + cfg.Quantity

	// Atomically move from reserved to allocated.
	_, err = txItemBus.Update(ctx, item, inventoryitembus.UpdateInventoryItem{
		ReservedQuantity:  &newReserved,
		AllocatedQuantity: &newAllocated,
	})
	if err != nil {
		return CommitAllocationResult{}, fmt.Errorf("update inventory item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return CommitAllocationResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	h.log.Info(ctx, "commit_allocation completed",
		"product_id", productID,
		"location_id", locationID,
		"previous_reserved", previousReserved,
		"new_reserved", newReserved,
		"previous_allocated", previousAllocated,
		"new_allocated", newAllocated,
		"quantity_committed", cfg.Quantity)

	return CommitAllocationResult{
		ProductID:         cfg.ProductID,
		LocationID:        cfg.LocationID,
		PreviousReserved:  previousReserved,
		NewReserved:       newReserved,
		PreviousAllocated: previousAllocated,
		NewAllocated:      newAllocated,
		QuantityCommitted: cfg.Quantity,
	}, nil
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *CommitAllocationHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{
			EntityName: "inventory.inventory_items",
			EventType:  "on_update",
			Fields:     []string{"reserved_quantity", "allocated_quantity"},
		},
	}
}
