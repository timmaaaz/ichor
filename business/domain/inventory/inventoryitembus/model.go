package inventoryitembus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type InventoryItem struct {
	ID                    uuid.UUID `json:"id"`
	ProductID             uuid.UUID `json:"product_id"`
	LocationID            uuid.UUID `json:"location_id"`
	Quantity              int       `json:"quantity"`
	ReservedQuantity      int       `json:"reserved_quantity"`
	AllocatedQuantity     int       `json:"allocated_quantity"`
	MinimumStock          int       `json:"minimum_stock"`
	MaximumStock          int       `json:"maximum_stock"`
	ReorderPoint          int       `json:"reorder_point"`
	EconomicOrderQuantity int       `json:"economic_order_quantity"`
	SafetyStock           int       `json:"safety_stock"`
	AvgDailyUsage         int       `json:"avg_daily_usage"`
	CreatedDate           time.Time `json:"created_date"`
	UpdatedDate           time.Time `json:"updated_date"`
}

type NewInventoryItem struct {
	ProductID             uuid.UUID `json:"product_id"`
	LocationID            uuid.UUID `json:"location_id"`
	Quantity              int       `json:"quantity"`
	ReservedQuantity      int       `json:"reserved_quantity"`
	AllocatedQuantity     int       `json:"allocated_quantity"`
	MinimumStock          int       `json:"minimum_stock"`
	MaximumStock          int       `json:"maximum_stock"`
	ReorderPoint          int       `json:"reorder_point"`
	EconomicOrderQuantity int       `json:"economic_order_quantity"`
	SafetyStock           int       `json:"safety_stock"`
	AvgDailyUsage         int       `json:"avg_daily_usage"`
}

// ItemWithProduct represents an inventory item at a location with product details.
type ItemWithProduct struct {
	ProductID    uuid.UUID `json:"product_id"`
	ProductName  string    `json:"product_name"`
	ProductSKU   string    `json:"product_sku"`
	TrackingType string    `json:"tracking_type"`
	Quantity     int       `json:"quantity"`
}

// InventoryItemWithLocation embeds InventoryItem with location context fields.
type InventoryItemWithLocation struct {
	InventoryItem
	LocationCode  string `json:"location_code"`
	Aisle         string `json:"aisle"`
	Rack          string `json:"rack"`
	Shelf         string `json:"shelf"`
	Bin           string `json:"bin"`
	ZoneName      string `json:"zone_name"`
	WarehouseName string `json:"warehouse_name"`
}

type UpdateInventoryItem struct {
	ProductID             *uuid.UUID `json:"product_id,omitempty"`
	LocationID            *uuid.UUID `json:"location_id,omitempty"`
	Quantity              *int       `json:"quantity,omitempty"`
	ReservedQuantity      *int       `json:"reserved_quantity,omitempty"`
	AllocatedQuantity     *int       `json:"allocated_quantity,omitempty"`
	MinimumStock          *int       `json:"minimum_stock,omitempty"`
	MaximumStock          *int       `json:"maximum_stock,omitempty"`
	ReorderPoint          *int       `json:"reorder_point,omitempty"`
	EconomicOrderQuantity *int       `json:"economic_order_quantity,omitempty"`
	SafetyStock           *int       `json:"safety_stock,omitempty"`
	AvgDailyUsage         *int       `json:"avg_daily_usage,omitempty"`
}
