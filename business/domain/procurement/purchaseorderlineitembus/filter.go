package purchaseorderlineitembus

import (
	"time"

	"github.com/google/uuid"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID                        *uuid.UUID
	PurchaseOrderID           *uuid.UUID
	SupplierProductID         *uuid.UUID
	LineItemStatusID          *uuid.UUID
	CreatedBy                 *uuid.UUID
	UpdatedBy                 *uuid.UUID
	StartExpectedDeliveryDate *time.Time
	EndExpectedDeliveryDate   *time.Time
	StartActualDeliveryDate   *time.Time
	EndActualDeliveryDate     *time.Time
	StartCreatedDate          *time.Time
	EndCreatedDate            *time.Time
	StartUpdatedDate          *time.Time
	EndUpdatedDate            *time.Time
}
