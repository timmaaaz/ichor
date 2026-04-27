package ordersbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus/types"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Order struct {
	ID                  uuid.UUID        `json:"id"`
	Number              string           `json:"number"`
	CustomerID          uuid.UUID        `json:"customer_id"`
	DueDate             time.Time        `json:"due_date"`
	FulfillmentStatusID uuid.UUID        `json:"fulfillment_status_id"`
	OrderDate           time.Time        `json:"order_date"`
	BillingAddressID    *uuid.UUID       `json:"billing_address_id,omitempty"`
	ShippingAddressID   *uuid.UUID       `json:"shipping_address_id,omitempty"`
	AssignedTo          *uuid.UUID       `json:"assigned_to,omitempty"`
	Subtotal            types.Money      `json:"subtotal"`
	TaxRate             types.Percentage `json:"tax_rate"`
	TaxAmount           types.Money      `json:"tax_amount"`
	ShippingCost        types.Money      `json:"shipping_cost"`
	TotalAmount         types.Money      `json:"total_amount"`
	CurrencyID          uuid.UUID        `json:"currency_id"`
	PaymentTermID       *uuid.UUID       `json:"payment_term_id,omitempty"`
	Notes               string           `json:"notes"`
	Priority            string           `json:"priority"` // low|medium|high|critical
	CreatedBy           uuid.UUID        `json:"created_by"`
	UpdatedBy           uuid.UUID        `json:"updated_by"`
	CreatedDate         time.Time        `json:"created_date"`
	UpdatedDate         time.Time        `json:"updated_date"`
	ScenarioID          *uuid.UUID       `json:"scenario_id,omitempty"`
}

type NewOrder struct {
	Number              string           `json:"number"`
	CustomerID          uuid.UUID        `json:"customer_id"`
	DueDate             time.Time        `json:"due_date"`
	FulfillmentStatusID uuid.UUID        `json:"fulfillment_status_id"`
	OrderDate           time.Time        `json:"order_date"`
	BillingAddressID    *uuid.UUID       `json:"billing_address_id,omitempty"`
	ShippingAddressID   *uuid.UUID       `json:"shipping_address_id,omitempty"`
	AssignedTo          *uuid.UUID       `json:"assigned_to,omitempty"`
	Subtotal            types.Money      `json:"subtotal"`
	TaxRate             types.Percentage `json:"tax_rate"`
	TaxAmount           types.Money      `json:"tax_amount"`
	ShippingCost        types.Money      `json:"shipping_cost"`
	TotalAmount         types.Money      `json:"total_amount"`
	CurrencyID          uuid.UUID        `json:"currency_id"`
	PaymentTermID       *uuid.UUID       `json:"payment_term_id,omitempty"`
	Notes               string           `json:"notes"`
	Priority            string           `json:"priority,omitempty"` // defaults to "medium" in Create
	CreatedBy           uuid.UUID        `json:"created_by"`
	CreatedDate         *time.Time       `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

type UpdateOrder struct {
	Number              *string           `json:"number,omitempty"`
	CustomerID          *uuid.UUID        `json:"customer_id,omitempty"`
	DueDate             *time.Time        `json:"due_date,omitempty"`
	FulfillmentStatusID *uuid.UUID        `json:"fulfillment_status_id,omitempty"`
	OrderDate           *time.Time        `json:"order_date,omitempty"`
	BillingAddressID    *uuid.UUID        `json:"billing_address_id,omitempty"`
	ShippingAddressID   *uuid.UUID        `json:"shipping_address_id,omitempty"`
	AssignedTo          *uuid.UUID        `json:"assigned_to,omitempty"`
	Subtotal            *types.Money      `json:"subtotal,omitempty"`
	TaxRate             *types.Percentage `json:"tax_rate,omitempty"`
	TaxAmount           *types.Money      `json:"tax_amount,omitempty"`
	ShippingCost        *types.Money      `json:"shipping_cost,omitempty"`
	TotalAmount         *types.Money      `json:"total_amount,omitempty"`
	CurrencyID          *uuid.UUID        `json:"currency_id,omitempty"`
	PaymentTermID       *uuid.UUID        `json:"payment_term_id,omitempty"`
	Notes               *string           `json:"notes,omitempty"`
	Priority            *string           `json:"priority,omitempty"`
	UpdatedBy           *uuid.UUID        `json:"updated_by,omitempty"`
}

// OrderContainerBinding represents the linkage between a sales order and a
// physical container (tote/cart) at a point in time. Active bindings have
// UnboundAt == nil; historical bindings retain UnboundAt.
type OrderContainerBinding struct {
	ID               uuid.UUID  `json:"id"`
	OrderID          uuid.UUID  `json:"order_id"`
	ContainerLabelID uuid.UUID  `json:"container_label_id"`
	BoundAt          time.Time  `json:"bound_at"`
	UnboundAt        *time.Time `json:"unbound_at,omitempty"`
	ScenarioID       *uuid.UUID `json:"scenario_id,omitempty"`
}

// NewOrderContainerBinding is the input shape for BindContainer.
type NewOrderContainerBinding struct {
	OrderID          uuid.UUID  `json:"order_id"`
	ContainerLabelID uuid.UUID  `json:"container_label_id"`
	ScenarioID       *uuid.UUID `json:"scenario_id,omitempty"`
}
