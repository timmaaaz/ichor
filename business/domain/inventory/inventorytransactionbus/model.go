package inventorytransactionbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type InventoryTransaction struct {
	InventoryTransactionID uuid.UUID  `json:"inventory_transaction_id"`
	ProductID              uuid.UUID  `json:"product_id"`
	LocationID             uuid.UUID  `json:"location_id"`
	UserID                 uuid.UUID  `json:"user_id"`
	LotID                  *uuid.UUID `json:"lot_id,omitempty"`
	Quantity               int        `json:"quantity"`
	TransactionType        string     `json:"transaction_type"`
	ReferenceNumber        string     `json:"reference_number"`
	TransactionDate        time.Time  `json:"transaction_date"`
	CreatedDate            time.Time  `json:"created_date"`
	UpdatedDate            time.Time  `json:"updated_date"`
}

type NewInventoryTransaction struct {
	ProductID       uuid.UUID  `json:"product_id"`
	LocationID      uuid.UUID  `json:"location_id"`
	UserID          uuid.UUID  `json:"user_id"`
	LotID           *uuid.UUID `json:"lot_id,omitempty"`
	Quantity        int        `json:"quantity"`
	TransactionType string     `json:"transaction_type"`
	ReferenceNumber string     `json:"reference_number"`
	TransactionDate time.Time  `json:"transaction_date"`
}

type UpdateInventoryTransaction struct {
	ProductID       *uuid.UUID `json:"product_id,omitempty"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	Quantity        *int       `json:"quantity,omitempty"`
	TransactionType *string    `json:"transaction_type,omitempty"`
	ReferenceNumber *string    `json:"reference_number,omitempty"`
	TransactionDate *time.Time `json:"transaction_date,omitempty"`
}
