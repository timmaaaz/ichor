package inventoryadjustmentbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	InventoryAdjustmentID *uuid.UUID
	ProductID             *uuid.UUID
	LocationID            *uuid.UUID
	AdjustedBy            *uuid.UUID
	ApprovedBy            *uuid.UUID
	ApprovalStatus        *string
	QuantityChange        *int
	ReasonCode            *string
	Notes                 *string
	AdjustmentDate        *time.Time
	CreatedDate           *time.Time
	UpdatedDate           *time.Time
}
