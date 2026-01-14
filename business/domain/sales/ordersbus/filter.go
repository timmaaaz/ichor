package ordersbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID                  *uuid.UUID
	Number              *string
	CustomerID          *uuid.UUID
	FulfillmentStatusID *uuid.UUID
	BillingAddressID    *uuid.UUID
	ShippingAddressID   *uuid.UUID
	Currency            *string
	PaymentTermID       *uuid.UUID
	CreatedBy           *uuid.UUID
	UpdatedBy           *uuid.UUID
	StartDueDate        *time.Time
	EndDueDate          *time.Time
	StartOrderDate      *time.Time
	EndOrderDate        *time.Time
	StartCreatedDate    *time.Time
	EndCreatedDate      *time.Time
	StartUpdatedDate    *time.Time
	EndUpdatedDate      *time.Time
	// Monetary range filters
	MinSubtotal     *string
	MaxSubtotal     *string
	MinTaxAmount    *string
	MaxTaxAmount    *string
	MinShippingCost *string
	MaxShippingCost *string
	MinTotalAmount  *string
	MaxTotalAmount  *string
}
