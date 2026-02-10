package inventory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ReserveInventoryConfig represents the configuration for reserving inventory.
type ReserveInventoryConfig struct {
	ProductID              string `json:"product_id"`
	SourceFromLineItem     bool   `json:"source_from_line_item"`
	Quantity               int    `json:"quantity"`
	WarehouseID            string `json:"warehouse_id,omitempty"`
	LocationID             string `json:"location_id,omitempty"`
	AllocationStrategy     string `json:"allocation_strategy"`
	ReservationDurationHrs int    `json:"reservation_duration_hours,omitempty"`
	AllowPartial           bool   `json:"allow_partial"`
	ReferenceID            string `json:"reference_id,omitempty"`
	ReferenceType          string `json:"reference_type,omitempty"`
}

// ReserveInventoryResult holds the result of an inventory reservation.
type ReserveInventoryResult struct {
	ReservationID   uuid.UUID      `json:"reservation_id"`
	Status          string         `json:"status"` // "success", "partial", "failed"
	ReservedItems   []ReservedItem `json:"reserved_items"`
	FailedItems     []FailedItem   `json:"failed_items"`
	TotalRequested  int            `json:"total_requested"`
	TotalReserved   int            `json:"total_reserved"`
	IdempotencyKey  string         `json:"idempotency_key"`
	ExecutionTimeMs int64          `json:"execution_time_ms"`
	ExpiresAt       *time.Time     `json:"expires_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	CompletedAt     time.Time      `json:"completed_at"`
}

// ReservedItem represents a successfully reserved inventory item.
type ReservedItem struct {
	ProductID         uuid.UUID  `json:"product_id"`
	LocationID        uuid.UUID  `json:"location_id"`
	InventoryItemID   uuid.UUID  `json:"inventory_item_id"`
	RequestedQuantity int        `json:"requested_quantity"`
	ReservedQuantity  int        `json:"reserved_quantity"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
}

// ReserveInventoryHandler handles reserve_inventory actions.
type ReserveInventoryHandler struct {
	log              *logger.Logger
	db               *sqlx.DB
	inventoryItemBus *inventoryitembus.Business
	workflowBus      *workflow.Business
}

// NewReserveInventoryHandler creates a new reserve inventory handler.
func NewReserveInventoryHandler(
	log *logger.Logger,
	db *sqlx.DB,
	inventoryItemBus *inventoryitembus.Business,
	workflowBus *workflow.Business,
) *ReserveInventoryHandler {
	return &ReserveInventoryHandler{
		log:              log,
		db:               db,
		inventoryItemBus: inventoryItemBus,
		workflowBus:      workflowBus,
	}
}

// GetType returns the action type.
func (h *ReserveInventoryHandler) GetType() string {
	return "reserve_inventory"
}

// SupportsManualExecution returns true - inventory reservation can be triggered manually.
func (h *ReserveInventoryHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - reserve_inventory is synchronous (Temporal handles durability).
func (h *ReserveInventoryHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *ReserveInventoryHandler) GetDescription() string {
	return "Reserve inventory items for future allocation with idempotency guarantees"
}

// Validate validates the reserve inventory configuration.
func (h *ReserveInventoryHandler) Validate(config json.RawMessage) error {
	var cfg ReserveInventoryConfig
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

	if cfg.Quantity <= 0 && !cfg.SourceFromLineItem {
		return errors.New("quantity must be greater than 0")
	}

	validStrategies := map[string]bool{
		"fifo": true, "lifo": true, "nearest_expiry": true, "lowest_cost": true,
	}
	if cfg.AllocationStrategy != "" && !validStrategies[cfg.AllocationStrategy] {
		return fmt.Errorf("invalid allocation_strategy: %s", cfg.AllocationStrategy)
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

// Execute reserves inventory with idempotency.
func (h *ReserveInventoryHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg ReserveInventoryConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return ReserveInventoryResult{}, fmt.Errorf("failed to parse config: %w", err)
	}

	// Default strategy.
	if cfg.AllocationStrategy == "" {
		cfg.AllocationStrategy = "fifo"
	}

	// Default reservation duration.
	if cfg.ReservationDurationHrs <= 0 {
		cfg.ReservationDurationHrs = 24
	}

	// Extract product/quantity from RawData if sourcing from line item.
	if cfg.SourceFromLineItem {
		productIDStr, _ := execContext.RawData["product_id"].(string)
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			return ReserveInventoryResult{}, fmt.Errorf("invalid product_id in line item: %w", err)
		}

		quantity, ok := execContext.RawData["quantity"].(float64)
		if !ok || quantity <= 0 {
			if qInt, ok := execContext.RawData["quantity"].(int); ok {
				quantity = float64(qInt)
			}
		}
		if quantity <= 0 {
			return ReserveInventoryResult{}, errors.New("quantity must be greater than 0")
		}

		cfg.ProductID = productID.String()
		cfg.Quantity = int(quantity)

		orderIDStr, _ := execContext.RawData["order_id"].(string)
		cfg.ReferenceID = orderIDStr
		cfg.ReferenceType = "order"
	}

	// Generate idempotency key.
	ruleIDStr := "manual"
	if execContext.RuleID != nil {
		ruleIDStr = execContext.RuleID.String()
	}
	idempotencyKey := fmt.Sprintf("%s_%s_%s", execContext.ExecutionID, ruleIDStr, h.GetType())

	// Check idempotency.
	existing, idempotencyResult, err := h.workflowBus.QueryAllocationResultByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return ReserveInventoryResult{}, fmt.Errorf("idempotency check: %w", err)
	}

	switch idempotencyResult {
	case workflow.IdempotencyExists:
		var cachedResult ReserveInventoryResult
		if err := json.Unmarshal(existing.AllocationData, &cachedResult); err != nil {
			return ReserveInventoryResult{}, fmt.Errorf("failed to unmarshal cached result: %w", err)
		}
		return cachedResult, nil
	case workflow.IdempotencyNotFound:
		// Proceed with reservation.
	}

	// Process reservation.
	result, err := h.processReservation(ctx, cfg, execContext, idempotencyKey)
	if err != nil {
		return ReserveInventoryResult{}, err
	}

	return *result, nil
}

// processReservation handles the actual reservation logic within a transaction.
func (h *ReserveInventoryHandler) processReservation(
	ctx context.Context,
	cfg ReserveInventoryConfig,
	execContext workflow.ActionExecutionContext,
	idempotencyKey string,
) (*ReserveInventoryResult, error) {
	startTime := time.Now()

	productID, err := uuid.Parse(cfg.ProductID)
	if err != nil {
		return nil, fmt.Errorf("invalid product_id: %w", err)
	}

	// Parse optional location/warehouse filters.
	var locationID, warehouseID *uuid.UUID
	if cfg.LocationID != "" {
		lid, err := uuid.Parse(cfg.LocationID)
		if err != nil {
			return nil, fmt.Errorf("invalid location_id: %w", err)
		}
		locationID = &lid
	}
	if cfg.WarehouseID != "" {
		wid, err := uuid.Parse(cfg.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("invalid warehouse_id: %w", err)
		}
		warehouseID = &wid
	}

	// Begin transaction.
	tx, err := h.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	txItemBus, err := h.inventoryItemBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("create transactional item bus: %w", err)
	}

	// Query available inventory (FOR UPDATE via the specialized method).
	items, err := txItemBus.QueryAvailableForAllocation(
		ctx,
		productID,
		locationID,
		warehouseID,
		cfg.AllocationStrategy,
		10,
	)
	if err != nil {
		return nil, fmt.Errorf("query available inventory: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(cfg.ReservationDurationHrs) * time.Hour)

	result := &ReserveInventoryResult{
		ReservationID:  uuid.New(),
		Status:         "processing",
		ReservedItems:  []ReservedItem{},
		FailedItems:    []FailedItem{},
		TotalRequested: cfg.Quantity,
		IdempotencyKey: idempotencyKey,
		ExpiresAt:      &expiresAt,
		CreatedAt:      time.Now(),
	}

	remaining := cfg.Quantity

	if len(items) == 0 {
		result.FailedItems = append(result.FailedItems, FailedItem{
			ProductID:         productID,
			RequestedQuantity: cfg.Quantity,
			AvailableQuantity: 0,
			Reason:            "insufficient_inventory",
			ErrorMessage:      "No inventory available",
		})
	} else {
		for _, invItem := range items {
			if remaining <= 0 {
				break
			}

			available := invItem.Quantity - invItem.ReservedQuantity - invItem.AllocatedQuantity
			if available <= 0 {
				continue
			}

			toReserve := min(remaining, available)
			newReserved := invItem.ReservedQuantity + toReserve

			_, err := txItemBus.Update(ctx, invItem, inventoryitembus.UpdateInventoryItem{
				ReservedQuantity: &newReserved,
			})
			if err != nil {
				result.FailedItems = append(result.FailedItems, FailedItem{
					ProductID:         productID,
					RequestedQuantity: cfg.Quantity,
					Reason:            "update_failed",
					ErrorMessage:      err.Error(),
				})
				if !cfg.AllowPartial {
					return nil, fmt.Errorf("reservation failed for product %s: %w", productID, err)
				}
				continue
			}

			result.ReservedItems = append(result.ReservedItems, ReservedItem{
				ProductID:         productID,
				LocationID:        invItem.LocationID,
				InventoryItemID:   invItem.ID,
				RequestedQuantity: toReserve,
				ReservedQuantity:  toReserve,
				ExpiresAt:         &expiresAt,
			})

			remaining -= toReserve
			result.TotalReserved += toReserve
		}
	}

	// Check if we need all and couldn't get it.
	if remaining > 0 && !cfg.AllowPartial && len(result.FailedItems) == 0 {
		return nil, fmt.Errorf("insufficient inventory: requested %d, available %d",
			cfg.Quantity, cfg.Quantity-remaining)
	}

	// Handle remaining as a failed item for partial.
	if remaining > 0 && cfg.AllowPartial {
		totalAvailable := cfg.Quantity - remaining
		result.FailedItems = append(result.FailedItems, FailedItem{
			ProductID:         productID,
			RequestedQuantity: remaining,
			AvailableQuantity: totalAvailable,
			Reason:            "insufficient_inventory",
			ErrorMessage:      fmt.Sprintf("Only %d available, %d requested", totalAvailable, cfg.Quantity),
		})
	}

	// Determine final status.
	if len(result.FailedItems) == 0 && result.TotalReserved == result.TotalRequested {
		result.Status = "success"
	} else if result.TotalReserved > 0 {
		result.Status = "partial"
	} else {
		result.Status = "failed"
	}

	// Store idempotency result.
	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}

	txWorkflowBus, err := h.workflowBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("create transactional workflow bus: %w", err)
	}

	_, err = txWorkflowBus.CreateAllocationResult(ctx, workflow.NewAllocationResult{
		IdempotencyKey: idempotencyKey,
		AllocationData: data,
	})
	if err != nil {
		return nil, fmt.Errorf("store reservation result: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	result.CompletedAt = time.Now()
	result.ExecutionTimeMs = time.Since(startTime).Milliseconds()

	h.log.Info(ctx, "reserve_inventory completed",
		"reservation_id", result.ReservationID,
		"status", result.Status,
		"total_reserved", result.TotalReserved,
		"total_requested", result.TotalRequested,
		"execution_time_ms", result.ExecutionTimeMs)

	return result, nil
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *ReserveInventoryHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{
			EntityName: "inventory.inventory_items",
			EventType:  "on_update",
			Fields:     []string{"reserved_quantity"},
		},
	}
}
