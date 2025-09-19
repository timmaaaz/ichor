package transferorderbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	TransferID     *uuid.UUID
	ProductID      *uuid.UUID
	FromLocationID *uuid.UUID
	ToLocationID   *uuid.UUID
	RequestedByID  *uuid.UUID
	ApprovedByID   *uuid.UUID
	Quantity       *int
	Status         *string
	TransferDate   *time.Time
	CreatedDate    *time.Time
	UpdatedDate    *time.Time
}
