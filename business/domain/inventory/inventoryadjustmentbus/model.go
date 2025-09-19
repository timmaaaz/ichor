package inventoryadjustmentbus

import (
	"time"

	"github.com/google/uuid"
)

type InventoryAdjustment struct {
	InventoryAdjustmentID uuid.UUID
	ProductID             uuid.UUID
	LocationID            uuid.UUID
	AdjustedBy            uuid.UUID
	ApprovedBy            uuid.UUID
	QuantityChange        int
	ReasonCode            string
	Notes                 string
	AdjustmentDate        time.Time
	CreatedDate           time.Time
	UpdatedDate           time.Time
}

type NewInventoryAdjustment struct {
	ProductID      uuid.UUID
	LocationID     uuid.UUID
	AdjustedBy     uuid.UUID
	ApprovedBy     uuid.UUID
	QuantityChange int
	ReasonCode     string
	Notes          string
	AdjustmentDate time.Time
}

type UpdateInventoryAdjustment struct {
	ProductID      *uuid.UUID
	LocationID     *uuid.UUID
	AdjustedBy     *uuid.UUID
	ApprovedBy     *uuid.UUID
	QuantityChange *int
	ReasonCode     *string
	Notes          *string
	AdjustmentDate *time.Time
}
