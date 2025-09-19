package transferorderbus

import (
	"time"

	"github.com/google/uuid"
)

type TransferOrder struct {
	TransferID     uuid.UUID // TODO: these should be id
	ProductID      uuid.UUID
	FromLocationID uuid.UUID
	ToLocationID   uuid.UUID
	RequestedByID  uuid.UUID
	ApprovedByID   uuid.UUID
	Quantity       int
	Status         string
	TransferDate   time.Time
	CreatedDate    time.Time
	UpdatedDate    time.Time
}

type NewTransferOrder struct {
	ProductID      uuid.UUID
	FromLocationID uuid.UUID
	ToLocationID   uuid.UUID
	RequestedByID  uuid.UUID
	ApprovedByID   uuid.UUID
	Quantity       int
	Status         string
	TransferDate   time.Time
}

type UpdateTransferOrder struct {
	ProductID      *uuid.UUID
	FromLocationID *uuid.UUID
	ToLocationID   *uuid.UUID
	RequestedByID  *uuid.UUID
	ApprovedByID   *uuid.UUID
	Quantity       *int
	Status         *string
	TransferDate   *time.Time
}
