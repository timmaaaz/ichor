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

// ReleaseReservationConfig represents the configuration for releasing a reservation.
type ReleaseReservationConfig struct {
	ProductID  string `json:"product_id"`
	LocationID string `json:"location_id"`
	Quantity   int    `json:"quantity"`
}

// ReleaseReservationResult holds the result of a reservation release.
type ReleaseReservationResult struct {
	ProductID            string `json:"product_id"`
	LocationID           string `json:"location_id"`
	PreviousReserved     int    `json:"previous_reserved"`
	NewReserved          int    `json:"new_reserved"`
	QuantityReleased     int    `json:"quantity_released"`
}

// ReleaseReservationHandler handles release_reservation actions.
type ReleaseReservationHandler struct {
	log              *logger.Logger
	db               *sqlx.DB
	inventoryItemBus *inventoryitembus.Business
}

// NewReleaseReservationHandler creates a new release reservation handler.
func NewReleaseReservationHandler(
	log *logger.Logger,
	db *sqlx.DB,
	inventoryItemBus *inventoryitembus.Business,
) *ReleaseReservationHandler {
	return &ReleaseReservationHandler{
		log:              log,
		db:               db,
		inventoryItemBus: inventoryItemBus,
	}
}

// GetType returns the action type.
func (h *ReleaseReservationHandler) GetType() string {
	return "release_reservation"
}

// SupportsManualExecution returns true - reservations can be released manually.
func (h *ReleaseReservationHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - release_reservation is a synchronous transactional operation.
func (h *ReleaseReservationHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *ReleaseReservationHandler) GetDescription() string {
	return "Release reserved inventory quantity back to available stock"
}

// Validate validates the release reservation configuration.
func (h *ReleaseReservationHandler) Validate(config json.RawMessage) error {
	var cfg ReleaseReservationConfig
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

// Execute releases reserved inventory within a transaction.
func (h *ReleaseReservationHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg ReleaseReservationConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return ReleaseReservationResult{}, fmt.Errorf("failed to parse config: %w", err)
	}

	productID, err := uuid.Parse(cfg.ProductID)
	if err != nil {
		return ReleaseReservationResult{}, fmt.Errorf("invalid product_id: %w", err)
	}

	locationID, err := uuid.Parse(cfg.LocationID)
	if err != nil {
		return ReleaseReservationResult{}, fmt.Errorf("invalid location_id: %w", err)
	}

	// Begin transaction.
	tx, err := h.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return ReleaseReservationResult{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	txItemBus, err := h.inventoryItemBus.NewWithTx(tx)
	if err != nil {
		return ReleaseReservationResult{}, fmt.Errorf("create transactional bus: %w", err)
	}

	// Query the inventory item by product and location.
	filter := inventoryitembus.QueryFilter{
		ProductID:  &productID,
		LocationID: &locationID,
	}

	items, err := txItemBus.Query(ctx, filter, order.NewBy("id", order.ASC), page.MustParse("1", "1"))
	if err != nil {
		return ReleaseReservationResult{}, fmt.Errorf("query inventory item: %w", err)
	}

	if len(items) == 0 {
		return ReleaseReservationResult{}, fmt.Errorf("no inventory item found for product %s at location %s", productID, locationID)
	}

	item := items[0]

	// Validate sufficient reserved quantity.
	if item.ReservedQuantity < cfg.Quantity {
		return ReleaseReservationResult{}, fmt.Errorf("insufficient reserved quantity: have %d, need %d", item.ReservedQuantity, cfg.Quantity)
	}

	previousReserved := item.ReservedQuantity
	newReserved := previousReserved - cfg.Quantity

	// Update the inventory item.
	_, err = txItemBus.Update(ctx, item, inventoryitembus.UpdateInventoryItem{
		ReservedQuantity: &newReserved,
	})
	if err != nil {
		return ReleaseReservationResult{}, fmt.Errorf("update inventory item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return ReleaseReservationResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	h.log.Info(ctx, "release_reservation completed",
		"product_id", productID,
		"location_id", locationID,
		"previous_reserved", previousReserved,
		"new_reserved", newReserved,
		"quantity_released", cfg.Quantity)

	return ReleaseReservationResult{
		ProductID:        cfg.ProductID,
		LocationID:       cfg.LocationID,
		PreviousReserved: previousReserved,
		NewReserved:      newReserved,
		QuantityReleased: cfg.Quantity,
	}, nil
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *ReleaseReservationHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{
			EntityName: "inventory.inventory_items",
			EventType:  "on_update",
			Fields:     []string{"reserved_quantity"},
		},
	}
}
