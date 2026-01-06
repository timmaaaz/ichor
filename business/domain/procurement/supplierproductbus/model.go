package supplierproductbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierproductbus/types"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type SupplierProduct struct {
	SupplierProductID  uuid.UUID   `json:"supplier_product_id"`
	SupplierID         uuid.UUID   `json:"supplier_id"`
	ProductID          uuid.UUID   `json:"product_id"`
	SupplierPartNumber string      `json:"supplier_part_number"`
	MinOrderQuantity   int         `json:"min_order_quantity"`
	MaxOrderQuantity   int         `json:"max_order_quantity"`
	LeadTimeDays       int         `json:"lead_time_days"`
	UnitCost           types.Money `json:"unit_cost"`
	IsPrimarySupplier  bool        `json:"is_primary_supplier"`
	CreatedDate        time.Time   `json:"created_date"`
	UpdatedDate        time.Time   `json:"updated_date"`
}

type NewSupplierProduct struct {
	SupplierID         uuid.UUID   `json:"supplier_id"`
	ProductID          uuid.UUID   `json:"product_id"`
	SupplierPartNumber string      `json:"supplier_part_number"`
	MinOrderQuantity   int         `json:"min_order_quantity"`
	MaxOrderQuantity   int         `json:"max_order_quantity"`
	LeadTimeDays       int         `json:"lead_time_days"`
	UnitCost           types.Money `json:"unit_cost"`
	IsPrimarySupplier  bool        `json:"is_primary_supplier"`
}

type UpdateSupplierProduct struct {
	SupplierID         *uuid.UUID   `json:"supplier_id,omitempty"`
	ProductID          *uuid.UUID   `json:"product_id,omitempty"`
	SupplierPartNumber *string      `json:"supplier_part_number,omitempty"`
	MinOrderQuantity   *int         `json:"min_order_quantity,omitempty"`
	MaxOrderQuantity   *int         `json:"max_order_quantity,omitempty"`
	LeadTimeDays       *int         `json:"lead_time_days,omitempty"`
	UnitCost           *types.Money `json:"unit_cost,omitempty"`
	IsPrimarySupplier  *bool        `json:"is_primary_supplier,omitempty"`
}
