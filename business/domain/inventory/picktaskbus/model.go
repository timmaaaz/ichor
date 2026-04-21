package picktaskbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// marshals business models to JSON for RawData in TriggerEvents.

// PickTask represents a single directed work instruction for a floor worker:
// pick QuantityToPick units of ProductID from LocationID to fulfill a sales order line item.
type PickTask struct {
	ID                   uuid.UUID  `json:"id"`
	TaskNumber           *string    `json:"task_number,omitempty"`
	SalesOrderID         uuid.UUID  `json:"sales_order_id"`
	SalesOrderLineItemID uuid.UUID  `json:"sales_order_line_item_id"`
	ProductID            uuid.UUID  `json:"product_id"`
	LotID                *uuid.UUID `json:"lot_id,omitempty"`
	SerialID             *uuid.UUID `json:"serial_id,omitempty"`
	LocationID           uuid.UUID  `json:"location_id"`
	QuantityToPick       int        `json:"quantity_to_pick"`
	QuantityPicked       int        `json:"quantity_picked"`
	Status               Status     `json:"status"`
	AssignedTo           uuid.UUID  `json:"assigned_to"`
	AssignedAt           time.Time  `json:"assigned_at"`
	CompletedBy          uuid.UUID  `json:"completed_by"`
	CompletedAt          time.Time  `json:"completed_at"`
	ShortPickReason      string     `json:"short_pick_reason,omitempty"`
	CreatedBy            uuid.UUID  `json:"created_by"`
	CreatedDate          time.Time  `json:"created_date"`
	UpdatedDate          time.Time  `json:"updated_date"`
	ScenarioID           *uuid.UUID `json:"scenario_id,omitempty"`
}

// NewPickTask contains the information needed to create a new pick task.
// Status is always set to Statuses.Pending by the business layer.
type NewPickTask struct {
	TaskNumber           *string    `json:"task_number,omitempty"`
	SalesOrderID         uuid.UUID  `json:"sales_order_id"`
	SalesOrderLineItemID uuid.UUID  `json:"sales_order_line_item_id"`
	ProductID            uuid.UUID  `json:"product_id"`
	LotID                *uuid.UUID `json:"lot_id,omitempty"`
	SerialID             *uuid.UUID `json:"serial_id,omitempty"`
	LocationID           uuid.UUID  `json:"location_id"`
	QuantityToPick       int        `json:"quantity_to_pick"`
	CreatedBy            uuid.UUID  `json:"created_by"`
}

// UpdatePickTask contains the information that can be changed on a pick task.
// All fields are optional pointers; nil means "do not update this field."
type UpdatePickTask struct {
	TaskNumber      *string    `json:"task_number,omitempty"`
	LotID           *uuid.UUID `json:"lot_id,omitempty"`
	SerialID        *uuid.UUID `json:"serial_id,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	QuantityToPick  *int       `json:"quantity_to_pick,omitempty"`
	QuantityPicked  *int       `json:"quantity_picked,omitempty"`
	Status          *Status    `json:"status,omitempty"`
	AssignedTo      *uuid.UUID `json:"assigned_to,omitempty"`
	AssignedAt      *time.Time `json:"assigned_at,omitempty"`
	CompletedBy     *uuid.UUID `json:"completed_by,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	ShortPickReason *string    `json:"short_pick_reason,omitempty"`
}
