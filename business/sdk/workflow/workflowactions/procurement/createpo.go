package procurement

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// CreatePurchaseOrderConfig represents the configuration for creating a purchase order.
type CreatePurchaseOrderConfig struct {
	SupplierID              string                  `json:"supplier_id,omitempty"`
	PurchaseOrderStatusID   string                  `json:"purchase_order_status_id"`
	DeliveryWarehouseID     string                  `json:"delivery_warehouse_id"`
	DeliveryLocationID      string                  `json:"delivery_location_id"`
	DeliveryStreetID        string                  `json:"delivery_street_id,omitempty"`
	CurrencyID              string                  `json:"currency_id"`
	OrderNumber             string                  `json:"order_number,omitempty"`
	ExpectedDeliveryDays    int                     `json:"expected_delivery_days,omitempty"`
	Notes                   string                  `json:"notes,omitempty"`
	SourceFromEvent         bool                    `json:"source_from_event,omitempty"`
	DefaultLineItemStatusID string                  `json:"default_line_item_status_id,omitempty"`
	LineItems               []CreatePOLineItemConfig `json:"line_items"`
}

// CreatePOLineItemConfig represents a single line item to create.
type CreatePOLineItemConfig struct {
	ProductID         string  `json:"product_id"`
	SupplierProductID string  `json:"supplier_product_id,omitempty"`
	QuantityOrdered   int     `json:"quantity_ordered"`
	UnitCost          float64 `json:"unit_cost,omitempty"`
	Discount          float64 `json:"discount,omitempty"`
	LineItemStatusID  string  `json:"line_item_status_id"`
	Notes             string  `json:"notes,omitempty"`
}

// CreatePurchaseOrderHandler handles create_purchase_order actions.
type CreatePurchaseOrderHandler struct {
	log                *logger.Logger
	db                 *sqlx.DB
	purchaseOrderBus   *purchaseorderbus.Business
	lineItemBus        *purchaseorderlineitembus.Business
	supplierProductBus *supplierproductbus.Business
}

// NewCreatePurchaseOrderHandler creates a new create purchase order handler.
func NewCreatePurchaseOrderHandler(
	log *logger.Logger,
	db *sqlx.DB,
	purchaseOrderBus *purchaseorderbus.Business,
	lineItemBus *purchaseorderlineitembus.Business,
	supplierProductBus *supplierproductbus.Business,
) *CreatePurchaseOrderHandler {
	return &CreatePurchaseOrderHandler{
		log:                log,
		db:                 db,
		purchaseOrderBus:   purchaseOrderBus,
		lineItemBus:        lineItemBus,
		supplierProductBus: supplierProductBus,
	}
}

// GetType returns the action type.
func (h *CreatePurchaseOrderHandler) GetType() string {
	return "create_purchase_order"
}

// SupportsManualExecution returns true - PO creation can be triggered manually.
func (h *CreatePurchaseOrderHandler) SupportsManualExecution() bool {
	return true
}

// IsAsync returns false - create_purchase_order completes inline.
func (h *CreatePurchaseOrderHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description for discovery APIs.
func (h *CreatePurchaseOrderHandler) GetDescription() string {
	return "Create a purchase order with line items, optionally auto-resolving supplier from product"
}

// Validate validates the create purchase order configuration.
func (h *CreatePurchaseOrderHandler) Validate(config json.RawMessage) error {
	var cfg CreatePurchaseOrderConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	if cfg.PurchaseOrderStatusID == "" {
		return fmt.Errorf("purchase_order_status_id is required")
	}
	if _, err := uuid.Parse(cfg.PurchaseOrderStatusID); err != nil {
		return fmt.Errorf("invalid purchase_order_status_id: %w", err)
	}

	if cfg.DeliveryWarehouseID == "" {
		return fmt.Errorf("delivery_warehouse_id is required")
	}
	if _, err := uuid.Parse(cfg.DeliveryWarehouseID); err != nil {
		return fmt.Errorf("invalid delivery_warehouse_id: %w", err)
	}

	if cfg.DeliveryLocationID == "" {
		return fmt.Errorf("delivery_location_id is required")
	}
	if _, err := uuid.Parse(cfg.DeliveryLocationID); err != nil {
		return fmt.Errorf("invalid delivery_location_id: %w", err)
	}

	if cfg.DeliveryStreetID != "" {
		if _, err := uuid.Parse(cfg.DeliveryStreetID); err != nil {
			return fmt.Errorf("invalid delivery_street_id: %w", err)
		}
	}

	if cfg.CurrencyID == "" {
		return fmt.Errorf("currency_id is required")
	}
	if _, err := uuid.Parse(cfg.CurrencyID); err != nil {
		return fmt.Errorf("invalid currency_id: %w", err)
	}

	if cfg.SupplierID != "" {
		if _, err := uuid.Parse(cfg.SupplierID); err != nil {
			return fmt.Errorf("invalid supplier_id: %w", err)
		}
	}

	if !cfg.SourceFromEvent && len(cfg.LineItems) == 0 {
		return fmt.Errorf("at least one line item is required when source_from_event is false")
	}

	// When source_from_event is true, require a default line item status ID.
	if cfg.SourceFromEvent {
		if cfg.DefaultLineItemStatusID == "" {
			return fmt.Errorf("default_line_item_status_id is required when source_from_event is true")
		}
		if _, err := uuid.Parse(cfg.DefaultLineItemStatusID); err != nil {
			return fmt.Errorf("invalid default_line_item_status_id: %w", err)
		}
		return nil
	}

	for i, li := range cfg.LineItems {
		if li.ProductID == "" {
			return fmt.Errorf("line_items[%d].product_id is required", i)
		}
		if _, err := uuid.Parse(li.ProductID); err != nil {
			return fmt.Errorf("line_items[%d].product_id is invalid: %w", i, err)
		}
		if li.QuantityOrdered <= 0 {
			return fmt.Errorf("line_items[%d].quantity_ordered must be greater than 0", i)
		}
		if li.LineItemStatusID == "" {
			return fmt.Errorf("line_items[%d].line_item_status_id is required", i)
		}
		if _, err := uuid.Parse(li.LineItemStatusID); err != nil {
			return fmt.Errorf("line_items[%d].line_item_status_id is invalid: %w", i, err)
		}
		if li.SupplierProductID != "" {
			if _, err := uuid.Parse(li.SupplierProductID); err != nil {
				return fmt.Errorf("line_items[%d].supplier_product_id is invalid: %w", i, err)
			}
		}
	}

	return nil
}

// GetOutputPorts implements workflow.OutputPortProvider.
func (h *CreatePurchaseOrderHandler) GetOutputPorts() []workflow.OutputPort {
	return []workflow.OutputPort{
		{Name: "created", Description: "Purchase order created successfully", IsDefault: true},
		{Name: "no_supplier_found", Description: "No supplier product found for a given product"},
		{Name: "failure", Description: "Unexpected error during PO creation"},
	}
}

// GetEntityModifications implements workflow.EntityModifier.
func (h *CreatePurchaseOrderHandler) GetEntityModifications(config json.RawMessage) []workflow.EntityModification {
	return []workflow.EntityModification{
		{
			EntityName: "procurement.purchase_orders",
			EventType:  "on_create",
		},
		{
			EntityName: "procurement.purchase_order_line_items",
			EventType:  "on_create",
		},
	}
}

// Execute creates a purchase order with line items.
func (h *CreatePurchaseOrderHandler) Execute(ctx context.Context, config json.RawMessage, execCtx workflow.ActionExecutionContext) (any, error) {
	var cfg CreatePurchaseOrderConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Extract line item data from trigger event if source_from_event is enabled.
	if cfg.SourceFromEvent {
		if err := h.extractFromEvent(&cfg, execCtx); err != nil {
			return map[string]any{
				"output": "failure",
				"error":  err.Error(),
			}, nil
		}
	}

	// Parse header UUIDs.
	poStatusID, err := uuid.Parse(cfg.PurchaseOrderStatusID)
	if err != nil {
		return map[string]any{"output": "failure", "error": fmt.Sprintf("invalid purchase_order_status_id: %s", err)}, nil
	}

	warehouseID, err := uuid.Parse(cfg.DeliveryWarehouseID)
	if err != nil {
		return map[string]any{"output": "failure", "error": fmt.Sprintf("invalid delivery_warehouse_id: %s", err)}, nil
	}

	locationID, err := uuid.Parse(cfg.DeliveryLocationID)
	if err != nil {
		return map[string]any{"output": "failure", "error": fmt.Sprintf("invalid delivery_location_id: %s", err)}, nil
	}

	currencyID, err := uuid.Parse(cfg.CurrencyID)
	if err != nil {
		return map[string]any{"output": "failure", "error": fmt.Sprintf("invalid currency_id: %s", err)}, nil
	}

	// DeliveryStreetID defaults to uuid.Nil if not provided.
	var streetID uuid.UUID
	if cfg.DeliveryStreetID != "" {
		streetID, err = uuid.Parse(cfg.DeliveryStreetID)
		if err != nil {
			return map[string]any{"output": "failure", "error": fmt.Sprintf("invalid delivery_street_id: %s", err)}, nil
		}
	}

	// Resolve supplier_id: use config value or auto-lookup from first line item's product.
	var supplierID uuid.UUID
	if cfg.SupplierID != "" {
		supplierID, err = uuid.Parse(cfg.SupplierID)
		if err != nil {
			return map[string]any{"output": "failure", "error": fmt.Sprintf("invalid supplier_id: %s", err)}, nil
		}
	}

	// Pre-resolve supplier products for each line item.
	type resolvedLineItem struct {
		supplierProductID uuid.UUID
		supplierID        uuid.UUID
		unitCost          float64
		discount          float64
		quantityOrdered   int
		lineItemStatusID  uuid.UUID
		notes             string
	}

	resolved := make([]resolvedLineItem, 0, len(cfg.LineItems))
	for i, li := range cfg.LineItems {
		productID, err := uuid.Parse(li.ProductID)
		if err != nil {
			return map[string]any{"output": "failure", "error": fmt.Sprintf("line_items[%d]: invalid product_id: %s", i, err)}, nil
		}

		liStatusID, err := uuid.Parse(li.LineItemStatusID)
		if err != nil {
			return map[string]any{"output": "failure", "error": fmt.Sprintf("line_items[%d]: invalid line_item_status_id: %s", i, err)}, nil
		}

		var spID uuid.UUID
		var spSupplierID uuid.UUID
		unitCost := li.UnitCost

		if li.SupplierProductID != "" {
			// Explicit supplier_product_id provided.
			spID, err = uuid.Parse(li.SupplierProductID)
			if err != nil {
				return map[string]any{"output": "failure", "error": fmt.Sprintf("line_items[%d]: invalid supplier_product_id: %s", i, err)}, nil
			}
		} else {
			// Auto-lookup: find primary supplier product for this product.
			sp, found, lookupErr := h.lookupSupplierProduct(ctx, productID)
			if lookupErr != nil {
				return nil, fmt.Errorf("line_items[%d]: supplier product lookup: %w", i, lookupErr)
			}
			if !found {
				h.log.Warn(ctx, "create_purchase_order: no supplier product found",
					"product_id", productID,
					"line_item_index", i)
				return map[string]any{
					"output":     "no_supplier_found",
					"product_id": productID.String(),
				}, nil
			}
			spID = sp.SupplierProductID
			spSupplierID = sp.SupplierID

			// Use supplier product unit cost if not overridden in config.
			if unitCost == 0 {
				costStr := sp.UnitCost.Value()
				if costStr != "" {
					unitCost, _ = strconv.ParseFloat(costStr, 64)
				}
			}
		}

		// If we resolved a supplier from the product lookup, use it for the PO header.
		if supplierID == uuid.Nil && spSupplierID != uuid.Nil {
			supplierID = spSupplierID
		}

		resolved = append(resolved, resolvedLineItem{
			supplierProductID: spID,
			supplierID:        spSupplierID,
			unitCost:          unitCost,
			discount:          li.Discount,
			quantityOrdered:   li.QuantityOrdered,
			lineItemStatusID:  liStatusID,
			notes:             li.Notes,
		})
	}

	if supplierID == uuid.Nil {
		return map[string]any{
			"output": "no_supplier_found",
			"error":  "could not determine supplier_id from config or line items",
		}, nil
	}

	// Validate all auto-resolved line items have the same supplier.
	for i, rli := range resolved {
		if rli.supplierID != uuid.Nil && rli.supplierID != supplierID {
			return map[string]any{
				"output": "failure",
				"error":  fmt.Sprintf("line_items[%d] resolved to supplier %s, but PO header uses %s; all line items must share the same supplier", i, rli.supplierID, supplierID),
			}, nil
		}
	}

	// Compute financial totals with discount applied.
	var subtotal float64
	lineTotals := make([]float64, len(resolved))
	for i, rli := range resolved {
		lt := float64(rli.quantityOrdered) * rli.unitCost * (1.0 - rli.discount)
		lineTotals[i] = lt
		subtotal += lt
	}

	now := time.Now()

	expectedDeliveryDays := cfg.ExpectedDeliveryDays
	if expectedDeliveryDays <= 0 {
		expectedDeliveryDays = 7
	}
	expectedDelivery := now.AddDate(0, 0, expectedDeliveryDays)

	orderNumber := cfg.OrderNumber
	if orderNumber == "" {
		orderNumber = fmt.Sprintf("PO-%s", uuid.New().String()[:8])
	}

	// Begin transaction for atomic PO + line item creation.
	tx, err := h.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	txPOBus, err := h.purchaseOrderBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("create transactional po bus: %w", err)
	}

	txLineItemBus, err := h.lineItemBus.NewWithTx(tx)
	if err != nil {
		return nil, fmt.Errorf("create transactional line item bus: %w", err)
	}

	// Create the purchase order header.
	po, err := txPOBus.Create(ctx, purchaseorderbus.NewPurchaseOrder{
		OrderNumber:           orderNumber,
		SupplierID:            supplierID,
		PurchaseOrderStatusID: poStatusID,
		DeliveryWarehouseID:   warehouseID,
		DeliveryLocationID:    locationID,
		DeliveryStreetID:      streetID,
		OrderDate:             now,
		ExpectedDeliveryDate:  expectedDelivery,
		Subtotal:              subtotal,
		TotalAmount:           subtotal,
		CurrencyID:            currencyID,
		RequestedBy:           execCtx.UserID,
		Notes:                 cfg.Notes,
		CreatedBy:             execCtx.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("create purchase order: %w", err)
	}

	// Create line items.
	lineItemIDs := make([]string, 0, len(resolved))
	for i, rli := range resolved {
		li, err := txLineItemBus.Create(ctx, purchaseorderlineitembus.NewPurchaseOrderLineItem{
			PurchaseOrderID:      po.ID,
			SupplierProductID:    rli.supplierProductID,
			QuantityOrdered:      rli.quantityOrdered,
			UnitCost:             rli.unitCost,
			LineTotal:            lineTotals[i],
			LineItemStatusID:     rli.lineItemStatusID,
			ExpectedDeliveryDate: expectedDelivery,
			Notes:                rli.notes,
			CreatedBy:            execCtx.UserID,
		})
		if err != nil {
			return nil, fmt.Errorf("create line item %d: %w", i, err)
		}
		lineItemIDs = append(lineItemIDs, li.ID.String())
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	h.log.Info(ctx, "create_purchase_order completed",
		"purchase_order_id", po.ID,
		"order_number", po.OrderNumber,
		"supplier_id", supplierID,
		"line_item_count", len(lineItemIDs))

	return map[string]any{
		"output":            "created",
		"purchase_order_id": po.ID.String(),
		"order_number":      po.OrderNumber,
		"supplier_id":       supplierID.String(),
		"line_item_ids":     lineItemIDs,
		"subtotal":          subtotal,
		"total_amount":      subtotal,
	}, nil
}

// lookupSupplierProduct finds the primary supplier product for a given product ID.
// Returns the supplier product, whether one was found, and any error.
func (h *CreatePurchaseOrderHandler) lookupSupplierProduct(ctx context.Context, productID uuid.UUID) (supplierproductbus.SupplierProduct, bool, error) {
	if h.supplierProductBus == nil {
		return supplierproductbus.SupplierProduct{}, false, fmt.Errorf("supplier product bus not available")
	}

	isPrimary := true
	filter := supplierproductbus.QueryFilter{
		ProductID:         &productID,
		IsPrimarySupplier: &isPrimary,
	}

	results, err := h.supplierProductBus.Query(ctx, filter, order.NewBy(supplierproductbus.OrderByUnitCost, order.ASC), page.MustParse("1", "1"))
	if err != nil {
		return supplierproductbus.SupplierProduct{}, false, fmt.Errorf("query supplier products: %w", err)
	}

	if len(results) == 0 {
		// Fall back to any supplier for this product.
		filter = supplierproductbus.QueryFilter{
			ProductID: &productID,
		}
		results, err = h.supplierProductBus.Query(ctx, filter, order.NewBy(supplierproductbus.OrderByUnitCost, order.ASC), page.MustParse("1", "1"))
		if err != nil {
			return supplierproductbus.SupplierProduct{}, false, fmt.Errorf("query supplier products (fallback): %w", err)
		}
		if len(results) == 0 {
			return supplierproductbus.SupplierProduct{}, false, nil
		}
	}

	return results[0], true, nil
}

// extractFromEvent extracts line item data from the workflow trigger event.
// This supports the reorder chain use case where the trigger event carries
// product and quantity information.
func (h *CreatePurchaseOrderHandler) extractFromEvent(cfg *CreatePurchaseOrderConfig, execCtx workflow.ActionExecutionContext) error {
	productIDStr, _ := execCtx.RawData["product_id"].(string)
	if productIDStr == "" {
		return fmt.Errorf("product_id not found in event RawData")
	}

	var quantity int
	switch v := execCtx.RawData["quantity"].(type) {
	case float64:
		quantity = int(v)
	case int:
		quantity = v
	default:
		// Try reorder_quantity or quantity_ordered.
		switch v := execCtx.RawData["reorder_quantity"].(type) {
		case float64:
			quantity = int(v)
		case int:
			quantity = v
		default:
			return fmt.Errorf("quantity not found in event RawData (tried quantity, reorder_quantity)")
		}
	}

	if quantity <= 0 {
		return fmt.Errorf("extracted quantity must be greater than 0, got %d", quantity)
	}

	cfg.LineItems = []CreatePOLineItemConfig{
		{
			ProductID:        productIDStr,
			QuantityOrdered:  quantity,
			LineItemStatusID: cfg.DefaultLineItemStatusID,
		},
	}

	return nil
}
