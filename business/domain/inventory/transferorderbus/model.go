package transferorderbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type TransferOrder struct {
	TransferID      uuid.UUID  `json:"transfer_id"` // TODO: these should be id
	ProductID       uuid.UUID  `json:"product_id"`
	FromLocationID  uuid.UUID  `json:"from_location_id"`
	ToLocationID    uuid.UUID  `json:"to_location_id"`
	RequestedByID   uuid.UUID  `json:"requested_by_id"`
	ApprovedByID    *uuid.UUID `json:"approved_by_id"`
	RejectedByID    *uuid.UUID `json:"rejected_by_id"`
	ApprovalReason  string     `json:"approval_reason"`
	RejectionReason string     `json:"rejection_reason"`
	ClaimedByID     *uuid.UUID `json:"claimed_by_id"`
	ClaimedAt       *time.Time `json:"claimed_at"`
	CompletedByID   *uuid.UUID `json:"completed_by_id"`
	CompletedAt     *time.Time `json:"completed_at"`
	Quantity        int        `json:"quantity"`
	Status          string     `json:"status"`
	TransferDate    time.Time  `json:"transfer_date"`
	CreatedDate     time.Time  `json:"created_date"`
	UpdatedDate     time.Time  `json:"updated_date"`
}

type NewTransferOrder struct {
	ProductID      uuid.UUID  `json:"product_id"`
	FromLocationID uuid.UUID  `json:"from_location_id"`
	ToLocationID   uuid.UUID  `json:"to_location_id"`
	RequestedByID  uuid.UUID  `json:"requested_by_id"`
	ApprovedByID   *uuid.UUID `json:"approved_by_id"`
	Quantity       int        `json:"quantity"`
	Status         string     `json:"status"`
	TransferDate   time.Time  `json:"transfer_date"`
}

type UpdateTransferOrder struct {
	ProductID       *uuid.UUID `json:"product_id,omitempty"`
	FromLocationID  *uuid.UUID `json:"from_location_id,omitempty"`
	ToLocationID    *uuid.UUID `json:"to_location_id,omitempty"`
	RequestedByID   *uuid.UUID `json:"requested_by_id,omitempty"`
	ApprovedByID    *uuid.UUID `json:"approved_by_id,omitempty"`
	RejectedByID    *uuid.UUID `json:"rejected_by_id,omitempty"`
	ApprovalReason  *string    `json:"approval_reason,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	ClaimedByID     *uuid.UUID `json:"claimed_by_id,omitempty"`
	ClaimedAt       *time.Time `json:"claimed_at,omitempty"`
	CompletedByID   *uuid.UUID `json:"completed_by_id,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	Quantity        *int       `json:"quantity,omitempty"`
	Status          *string    `json:"status,omitempty"`
	TransferDate    *time.Time `json:"transfer_date,omitempty"`
}
