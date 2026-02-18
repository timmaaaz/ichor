package inventory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventoryitembus"
	"github.com/timmaaaz/ichor/business/domain/inventory/inventorytransactionbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// ReceiveInventoryConfig represents the configuration for receiving inventory.
type ReceiveInventoryConfig struct {
	ProductID      string `json:"product_id"`
	Quantity       int    `json:"quantity"`
	LocationID     string `json:"location_id"`
	SourceFromPO   bool   `json:"source_from_po,omitempty"`
	POLineItemID   string `json:"po_line_item_id,omitempty"`
	ReferenceNumber string `json:"reference_number,omitempty"`
	Notes          string `json:"notes,omitempty"`
}

// ReceiveInventoryHandler handles receive_inventory actions.
type ReceiveInventoryHandler struct {
	log                *logger.Logger
	db                 *sqlx.DB
	inventoryItemBus   *inventoryitembus.Business
	transactionBus     *inventorytransactionbus.Business
	supplierProductBus *supplierproductbus.Business
}

// NewReceiveInventoryHandler creates a new receive inventory handler.
func NewReceiveInventoryHandler(
	log *logger.Logger,
	db *sqlx.DB,
	inventoryItemBus *inventoryitembus.Business,
	transactionBus *inventorytransactionbus.Business,
	supplierProductBus *supplierproductbus.Business,
) *ReceiveInventoryHandler {
	return &ReceiveInventoryHandler{
		log:                log,
		db:                 db,
		inventoryItemBus:   inventoryItemBus,
		transactionBus:     transactionBus,
		supplierProductBus: supplierProductBus,
	}
}

// GetType returns the action type.
func (h *ReceiveInventoryHandler) GetType() string {
	return "receive_inventory"
}

// SupportsManualExecution returns true - inventory receipt can be triggered manually.
func (h *ReceiveInventoryHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - receive_inventory completes inline.
func (h *ReceiveInventoryHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *ReceiveInventoryHandler) GetDescription() string {
	return "Receive inventory: increase stock quantity and create inbound transaction record"
}

// Validate validates the receive inventory configuration.
func (h *ReceiveInventoryHandler) Validate(config json.RawMessage) error {
	var cfg ReceiveInventoryConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	// location_id is always required (PO line items don't carry location info).
	if cfg.LocationID == "" {
		return fmt.Errorf("location_id is required")
	}
	if _, err := uuid.Parse(cfg.LocationID); err != nil {
		return fmt.Errorf("invalid location_id: %w", err)
	}

	if !cfg.SourceFromPO {
		if cfg.ProductID == "" {
			return fmt.Errorf("product_id is required when source_from_po is false")
		}
		if _, err := uuid.Parse(cfg.ProductID); err != nil {
			return fmt.Errorf("invalid product_id: %w", err)
		}
		if cfg.Quantity <= 0 {
			return fmt.Errorf("quantity must be greater than 0")
		}
	}

	if cfg.POLineItemID != "" {
		if _, err := uuid.Parse(cfg.POLineItemID); err != nil {
			return fmt.Errorf("invalid po_line_item_id: %w", err)
		}
	}

	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *ReceiveInventoryHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "received", Description: "Inventory received and transaction recorded", IsDefault: true},
		{Name: "item_not_found", Description: "No inventory item found for the product/location combination"},
		{Name: "failure", Description: "Unexpected error during receipt"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *ReceiveInventoryHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{
			EntityName: "inventory.inventory_items",
			EventType:  "on_update",
			Fields:     []string{"quantity"},
		},
		{
			EntityName: "inventory.inventory_transactions",
			EventType:  "on_create",
		},
	}
}

// Execute receives inventory by increasing stock quantity and creating a transaction record.
func (h *ReceiveInventoryHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg ReceiveInventoryConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Extract fields from PO line item event if source_from_po is enabled.
	if cfg.SourceFromPO {
		if err := h.extractFromPOLineItem(ctx, &cfg, execCtx); err != nil {
			return map[string]any{
				"output": "failure",
				"error":  err.Error(),
			}, nil
		}
	}

	// Parse UUIDs — route to failure output port for data issues (not hard errors).
	productID, err := uuid.Parse(cfg.ProductID)
	if err != nil {
		return map[string]any{"output": "failure", "error": fmt.Sprintf("invalid product_id: %s", err)}, nil
	}

	locationID, err := uuid.Parse(cfg.LocationID)
	if err != nil {
		return map[string]any{"output": "failure", "error": fmt.Sprintf("invalid location_id: %s", err)}, nil
	}

	if cfg.Quantity <= 0 {
		return map[string]any{"output": "failure", "error": "quantity must be greater than 0"}, nil
	}

	// Begin transaction for atomic query + update + transaction record creation.
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

	txTransactionBus, err := h.transactionBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("create transactional transaction bus: %w", err)
	}

	// Query inventory item inside the transaction to avoid TOCTOU race.
	filter := inventoryitembus.QueryFilter{
		ProductID:  &productID,
		LocationID: &locationID,
	}
	items, err := txItemBus.Query(ctx, filter, order.NewBy("id", order.ASC), page.MustParse("1", "10"))
	if err != nil {
		return nil, fmt.Errorf("query inventory items: %w", err)
	}

	if len(items) == 0 {
		h.log.Warn(ctx, "receive_inventory: no inventory item found",
			"product_id", productID,
			"location_id", locationID)
		return map[string]any{
			"output":      "item_not_found",
			"product_id":  productID.String(),
			"location_id": locationID.String(),
		}, nil
	}

	// Use the first matching inventory item.
	item := items[0]

	// Update inventory item quantity.
	newQuantity := item.Quantity + cfg.Quantity
	updatedItem, err := txItemBus.Update(ctx, item, inventoryitembus.UpdateInventoryItem{
		Quantity: &newQuantity,
	})
	if err != nil {
		return nil, fmt.Errorf("update inventory item quantity: %w", err)
	}

	// Build reference number from PO line item ID or config.
	referenceNumber := cfg.ReferenceNumber
	if referenceNumber == "" && cfg.POLineItemID != "" {
		referenceNumber = "PO-LINE:" + cfg.POLineItemID
	}

	// Create inbound transaction record.
	newTx := inventorytransactionbus.NewInventoryTransaction{
		ProductID:       productID,
		LocationID:      locationID,
		UserID:          execCtx.UserID,
		Quantity:        cfg.Quantity,
		TransactionType: "inbound",
		ReferenceNumber: referenceNumber,
		TransactionDate: time.Now(),
	}

	txRecord, err := txTransactionBus.Create(ctx, newTx)
	if err != nil {
		return nil, fmt.Errorf("create inventory transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	h.log.Info(ctx, "receive_inventory completed",
		"inventory_item_id", item.ID,
		"product_id", productID,
		"location_id", locationID,
		"quantity_received", cfg.Quantity,
		"new_quantity", updatedItem.Quantity,
		"transaction_id", txRecord.InventoryTransactionID)

	return map[string]any{
		"output":            "received",
		"inventory_item_id": item.ID.String(),
		"transaction_id":    txRecord.InventoryTransactionID.String(),
		"quantity_received":  cfg.Quantity,
		"previous_quantity":  item.Quantity,
		"new_quantity":       updatedItem.Quantity,
		"product_id":         productID.String(),
		"location_id":        locationID.String(),
	}, nil
}

// extractFromPOLineItem extracts product_id and quantity from a PO line item event.
// PO line items have supplier_product_id (not product_id directly), so we resolve
// the product_id via the supplier_product record.
func (h *ReceiveInventoryHandler) extractFromPOLineItem(ctx context.Context, cfg *ReceiveInventoryConfig, execCtx workflow.ActionExecutionContext) error {
	// Extract supplier_product_id from RawData.
	supplierProductIDStr, _ := execCtx.RawData["supplier_product_id"].(string)
	if supplierProductIDStr == "" {
		return fmt.Errorf("supplier_product_id not found in PO line item RawData")
	}

	supplierProductID, err := uuid.Parse(supplierProductIDStr)
	if err != nil {
		return fmt.Errorf("invalid supplier_product_id in RawData: %w", err)
	}

	// Look up supplier product to get the actual product_id.
	if h.supplierProductBus == nil {
		return fmt.Errorf("supplier product bus not available for source_from_po resolution")
	}

	supplierProduct, err := h.supplierProductBus.QueryByID(ctx, supplierProductID)
	if err != nil {
		return fmt.Errorf("query supplier product %s: %w", supplierProductID, err)
	}

	cfg.ProductID = supplierProduct.ProductID.String()

	// Extract quantity_received from RawData (JSON numbers arrive as float64).
	switch v := execCtx.RawData["quantity_received"].(type) {
	case float64:
		cfg.Quantity = int(v)
	case int:
		cfg.Quantity = v
	default:
		// Fall back to quantity_ordered if quantity_received is not set.
		switch v := execCtx.RawData["quantity_ordered"].(type) {
		case float64:
			cfg.Quantity = int(v)
		case int:
			cfg.Quantity = v
		default:
			return fmt.Errorf("quantity_received or quantity_ordered not found in PO line item RawData")
		}
	}

	// Extract PO line item ID for reference.
	if idStr, ok := execCtx.RawData["id"].(string); ok {
		cfg.POLineItemID = idStr
	}

	// location_id must still come from config (PO line items don't have a location).
	if cfg.LocationID == "" {
		return fmt.Errorf("location_id is required in config even when source_from_po is true")
	}

	return nil
}
