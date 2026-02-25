package purchaseorderbus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID                      *uuid.UUID
	OrderNumber             *string
	SupplierID              *uuid.UUID
	PurchaseOrderStatusID   *uuid.UUID
	DeliveryWarehouseID     *uuid.UUID
	RequestedBy             *uuid.UUID
	ApprovedBy              *uuid.UUID
	StartOrderDate          *time.Time
	EndOrderDate            *time.Time
	StartExpectedDelivery   *time.Time
	EndExpectedDelivery     *time.Time
	StartActualDeliveryDate *time.Time
	EndActualDeliveryDate   *time.Time
	IsUndelivered           *bool
}