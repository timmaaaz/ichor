package inventorytransactionbus

import (
	"time"

	"github.com/google/uuid"
)

type InventoryTransaction struct {
	InventoryTransactionID uuid.UUID
	ProductID              uuid.UUID
	LocationID             uuid.UUID
	UserID                 uuid.UUID
	Quantity               int
	TransactionType        string
	ReferenceNumber        string
	TransactionDate        time.Time
	CreatedDate            time.Time
	UpdatedDate            time.Time
}

type NewInventoryTransaction struct {
	ProductID       uuid.UUID
	LocationID      uuid.UUID
	UserID          uuid.UUID
	Quantity        int
	TransactionType string
	ReferenceNumber string
	TransactionDate time.Time
}

type UpdateInventoryTransaction struct {
	ProductID       *uuid.UUID
	LocationID      *uuid.UUID
	UserID          *uuid.UUID
	Quantity        *int
	TransactionType *string
	ReferenceNumber *string
	TransactionDate *time.Time
}
