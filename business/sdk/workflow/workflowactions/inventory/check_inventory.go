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

// CheckInventoryConfig represents the configuration for checking inventory availability.
type CheckInventoryConfig struct {
	ProductID          string `json:"product_id"`
	SourceFromLineItem bool   `json:"source_from_line_item"`
	Threshold          int    `json:"threshold"`
	WarehouseID        string `json:"warehouse_id,omitempty"`
	LocationID         string `json:"location_id,omitempty"`
}

// CheckInventoryResult holds the result of an inventory availability check.
type CheckInventoryResult struct {
	Available   int  `json:"available"`
	Threshold   int  `json:"threshold"`
	Sufficient  bool `json:"sufficient"`
	TotalOnHand int  `json:"total_on_hand"`
	Reserved    int  `json:"reserved"`
	Allocated   int  `json:"allocated"`
}

// CheckInventoryHandler handles check_inventory actions.
type CheckInventoryHandler struct {
	log              *logger.Logger
	inventoryItemBus *inventoryitembus.Business
}

// NewCheckInventoryHandler creates a new check inventory handler.
func NewCheckInventoryHandler(
	log *logger.Logger,
	inventoryItemBus *inventoryitembus.Business,
) *CheckInventoryHandler {
	return &CheckInventoryHandler{
		log:              log,
		inventoryItemBus: inventoryItemBus,
	}
}

// GetType returns the action type.
func (h *CheckInventoryHandler) GetType() string {
	return "check_inventory"
}

// SupportsManualExecution returns false - check_inventory is read-only and used in automation.
func (h *CheckInventoryHandler) SupportsManualExecution() bool {
	return false
}

// IsAsync returns false - check_inventory is a synchronous read-only operation.
func (h *CheckInventoryHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *CheckInventoryHandler) GetDescription() string {
	return "Check inventory availability against a threshold and branch accordingly"
}

// Validate validates the check inventory configuration.
func (h *CheckInventoryHandler) Validate(config json.RawMessage) error {
	var cfg CheckInventoryConfig
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

	if cfg.Threshold < 0 {
		return errors.New("threshold must be non-negative")
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

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *CheckInventoryHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "sufficient", Description: "Inventory meets or exceeds threshold", IsDefault: true},
		{Name: "insufficient", Description: "Inventory is below threshold"},
	}
}

// Execute checks inventory availability and returns the result with an output port.
// output=sufficient when inventory meets threshold, output=insufficient otherwise.
func (h *CheckInventoryHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg CheckInventoryConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Extract product_id from RawData if sourcing from line item.
	productIDStr := cfg.ProductID
	if cfg.SourceFromLineItem {
		raw, _ := execContext.RawData["product_id"].(string)
		if raw == "" {
			return nil, errors.New("product_id not found in line item RawData")
		}
		productIDStr = raw
	}

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid product_id: %w", err)
	}

	// Build filter for querying inventory items.
	filter := inventoryitembus.QueryFilter{
		ProductID: &productID,
	}

	if cfg.LocationID != "" {
		locID, err := uuid.Parse(cfg.LocationID)
		if err != nil {
			return nil, fmt.Errorf("invalid location_id: %w", err)
		}
		filter.LocationID = &locID
	}

	// Query inventory items (read-only, no locks needed).
	items, err := h.inventoryItemBus.Query(ctx, filter, order.NewBy("id", order.ASC), page.MustParse("1", "100"))
	if err != nil {
		return nil, fmt.Errorf("query inventory items: %w", err)
	}

	// Sum available quantities across all matching items.
	var totalOnHand, totalReserved, totalAllocated int
	for _, item := range items {
		totalOnHand += item.Quantity
		totalReserved += item.ReservedQuantity
		totalAllocated += item.AllocatedQuantity
	}
	available := totalOnHand - totalReserved - totalAllocated

	sufficient := available >= cfg.Threshold

	output := "insufficient"
	if sufficient {
		output = "sufficient"
	}

	h.log.Info(ctx, "check_inventory completed",
		"product_id", productID,
		"available", available,
		"threshold", cfg.Threshold,
		"sufficient", sufficient,
		"output", output)

	return map[string]any{
		"available":    available,
		"threshold":    cfg.Threshold,
		"sufficient":   sufficient,
		"total_on_hand": totalOnHand,
		"reserved":     totalReserved,
		"allocated":    totalAllocated,
		"output":       output,
	}, nil
}

// GetEntityModifications implements workflow.EntityModifier.
// check_inventory is read-only and does not modify any entities.
func (h *CheckInventoryHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return nil
}
