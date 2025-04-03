package inventorytransactionbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	InventoryTransactionID *uuid.UUID
	ProductID              *uuid.UUID
	LocationID             *uuid.UUID
	UserID                 *uuid.UUID
	Quantity               *int
	TransactionType        *string
	ReferenceNumber        *string
	TransactionDate        *time.Time
	CreatedDate            *time.Time
	UpdatedDate            *time.Time
}
