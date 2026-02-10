package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// CheckReorderPointConfig represents the configuration for checking reorder points.
type CheckReorderPointConfig struct {
	ProductID          string `json:"product_id"`
	SourceFromLineItem bool   `json:"source_from_line_item"`
	CustomThreshold    *int   `json:"custom_threshold,omitempty"`
	WarehouseID        string `json:"warehouse_id,omitempty"`
	LocationID         string `json:"location_id,omitempty"`
}

// CheckReorderPointResult holds the result of a reorder point check.
type CheckReorderPointResult struct {
	Available    int  `json:"available"`
	ReorderPoint int  `json:"reorder_point"`
	NeedsReorder bool `json:"needs_reorder"`
	TotalOnHand  int  `json:"total_on_hand"`
	Reserved     int  `json:"reserved"`
	Allocated    int  `json:"allocated"`
}

// CheckReorderPointHandler handles check_reorder_point actions.
type CheckReorderPointHandler struct {
	log              *logger.Logger
	inventoryItemBus *inventoryitembus.Business
}

// NewCheckReorderPointHandler creates a new check reorder point handler.
func NewCheckReorderPointHandler(
	log *logger.Logger,
	inventoryItemBus *inventoryitembus.Business,
) *CheckReorderPointHandler {
	return &CheckReorderPointHandler{
		log:              log,
		inventoryItemBus: inventoryItemBus,
	}
}

// GetType returns the action type.
func (h *CheckReorderPointHandler) GetType() string {
	return "check_reorder_point"
}

// SupportsManualExecution returns false - check_reorder_point is read-only and used in automation.
func (h *CheckReorderPointHandler) SupportsManualExecution() bool {
	return false
}

// IsAsync returns false - check_reorder_point is a synchronous read-only operation.
func (h *CheckReorderPointHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *CheckReorderPointHandler) GetDescription() string {
	return "Check if inventory is below its reorder point and branch accordingly"
}

// Validate validates the check reorder point configuration.
func (h *CheckReorderPointHandler) Validate(config json.RawMessage) error {
	var cfg CheckReorderPointConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.ProductID == "" && !cfg.SourceFromLineItem {
		return errors.New("product_id is required when source_from_line_item is false")
	}

	if cfg.ProductID != "" {
		if _, err := uuid.Parse(cfg.ProductID); err != nil {
			return fmt.Errorf("invalid product_id: %w", err)
		}
	}

	if cfg.CustomThreshold != nil && *cfg.CustomThreshold < 0 {
		return errors.New("custom_threshold must be non-negative")
	}

	if cfg.WarehouseID != "" {
		if _, err := uuid.Parse(cfg.WarehouseID); err != nil {
			return fmt.Errorf("invalid warehouse_id: %w", err)
		}
	}

	if cfg.LocationID != "" {
		if _, err := uuid.Parse(cfg.LocationID); err != nil {
			return fmt.Errorf("invalid location_id: %w", err)
		}
	}

	return nil
}

// Execute checks if inventory is below the reorder point and returns a ConditionResult.
// true_branch = below reorder point (needs reorder), false_branch = sufficient.
func (h *CheckReorderPointHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg CheckReorderPointConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return workflow.ConditionResult{}, fmt.Errorf("failed to parse config: %w", err)
	}

	// Extract product_id from RawData if sourcing from line item.
	productIDStr := cfg.ProductID
	if cfg.SourceFromLineItem {
		raw, _ := execContext.RawData["product_id"].(string)
		if raw == "" {
			return workflow.ConditionResult{}, errors.New("product_id not found in line item RawData")
		}
		productIDStr = raw
	}

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return workflow.ConditionResult{}, fmt.Errorf("invalid product_id: %w", err)
	}

	// Build filter for querying inventory items.
	filter := inventoryitembus.QueryFilter{
		ProductID: &productID,
	}

	if cfg.LocationID != "" {
		locID, err := uuid.Parse(cfg.LocationID)
		if err != nil {
			return workflow.ConditionResult{}, fmt.Errorf("invalid location_id: %w", err)
		}
		filter.LocationID = &locID
	}

	// Query inventory items (read-only, no locks needed).
	items, err := h.inventoryItemBus.Query(ctx, filter, order.NewBy("id", order.ASC), page.MustParse("1", "100"))
	if err != nil {
		return workflow.ConditionResult{}, fmt.Errorf("query inventory items: %w", err)
	}

	// Sum available quantities and determine the reorder point.
	var totalOnHand, totalReserved, totalAllocated int
	var maxReorderPoint int
	for _, item := range items {
		totalOnHand += item.Quantity
		totalReserved += item.ReservedQuantity
		totalAllocated += item.AllocatedQuantity
		if item.ReorderPoint > maxReorderPoint {
			maxReorderPoint = item.ReorderPoint
		}
	}
	available := totalOnHand - totalReserved - totalAllocated

	// Use custom threshold if provided, otherwise use the item's reorder point.
	reorderPoint := maxReorderPoint
	if cfg.CustomThreshold != nil {
		reorderPoint = *cfg.CustomThreshold
	}

	needsReorder := available < reorderPoint

	branchTaken := workflow.EdgeTypeFalseBranch
	if needsReorder {
		branchTaken = workflow.EdgeTypeTrueBranch
	}

	h.log.Info(ctx, "check_reorder_point completed",
		"product_id", productID,
		"available", available,
		"reorder_point", reorderPoint,
		"needs_reorder", needsReorder,
		"branch", branchTaken)

	return workflow.ConditionResult{
		Evaluated:   true,
		Result:      needsReorder,
		BranchTaken: branchTaken,
	}, nil
}

// GetEntityModifications implements workflow.EntityModifier.
// check_reorder_point is read-only and does not modify any entities.
func (h *CheckReorderPointHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return nil
}
