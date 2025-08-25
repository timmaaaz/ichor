package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// AllocateInventoryHandler handles allocate_inventory actions
type AllocateInventoryHandler struct {
	log *logger.Logger
	db  *sqlx.DB
}

// NewAllocateInventoryHandler creates a new allocate inventory handler
func NewAllocateInventoryHandler(log *logger.Logger, db *sqlx.DB) *AllocateInventoryHandler {
	return &AllocateInventoryHandler{
		log: log,
		db:  db,
	}
}

func (h *AllocateInventoryHandler) GetType() string {
	return "allocate_inventory"
}

func (h *AllocateInventoryHandler) Validate(config json.RawMessage) error {
	var cfg struct {
		InventoryItems []struct {
			ItemID   string `json:"item_id"`
			Quantity int    `json:"quantity"`
		} `json:"inventory_items"`
		AllocationStrategy string `json:"allocation_strategy"`
	}

	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if len(cfg.InventoryItems) == 0 {
		return fmt.Errorf("inventory_items list is required and must not be empty")
	}

	validStrategies := map[string]bool{
		"fifo": true, "lifo": true, "nearest_expiry": true, "lowest_cost": true,
	}
	if !validStrategies[cfg.AllocationStrategy] {
		return fmt.Errorf("invalid allocation_strategy")
	}

	return nil
}

func (h *AllocateInventoryHandler) Execute(ctx context.Context, config json.RawMessage, context workflow.ActionExecutionContext) (interface{}, error) {
	h.log.Info(ctx, "Executing allocate_inventory action",
		"entityID", context.EntityID,
		"ruleName", context.RuleName)

	result := map[string]interface{}{
		"allocation_id": fmt.Sprintf("alloc_%d", time.Now().Unix()),
		"status":        "allocated",
		"allocated_at":  time.Now().Format(time.RFC3339),
	}

	return result, nil
}
