package inventoryadjustmentbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type InventoryAdjustment struct {
	InventoryAdjustmentID uuid.UUID  `json:"inventory_adjustment_id"`
	ProductID             uuid.UUID  `json:"product_id"`
	LocationID            uuid.UUID  `json:"location_id"`
	AdjustedBy            uuid.UUID  `json:"adjusted_by"`
	ApprovedBy            *uuid.UUID `json:"approved_by"`
	ApprovalStatus        string     `json:"approval_status"`
	QuantityChange        int        `json:"quantity_change"`
	ReasonCode            string     `json:"reason_code"`
	Notes                 string     `json:"notes"`
	AdjustmentDate        time.Time  `json:"adjustment_date"`
	CreatedDate           time.Time  `json:"created_date"`
	UpdatedDate           time.Time  `json:"updated_date"`
}

type NewInventoryAdjustment struct {
	ProductID      uuid.UUID  `json:"product_id"`
	LocationID     uuid.UUID  `json:"location_id"`
	AdjustedBy     uuid.UUID  `json:"adjusted_by"`
	ApprovedBy     *uuid.UUID `json:"approved_by"`
	ApprovalStatus string     `json:"approval_status"`
	QuantityChange int        `json:"quantity_change"`
	ReasonCode     string     `json:"reason_code"`
	Notes          string     `json:"notes"`
	AdjustmentDate time.Time  `json:"adjustment_date"`
}

type UpdateInventoryAdjustment struct {
	ProductID      *uuid.UUID `json:"product_id,omitempty"`
	LocationID     *uuid.UUID `json:"location_id,omitempty"`
	AdjustedBy     *uuid.UUID `json:"adjusted_by,omitempty"`
	ApprovedBy     *uuid.UUID `json:"approved_by,omitempty"`
	ApprovalStatus *string    `json:"approval_status,omitempty"`
	QuantityChange *int       `json:"quantity_change,omitempty"`
	ReasonCode     *string    `json:"reason_code,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	AdjustmentDate *time.Time `json:"adjustment_date,omitempty"`
}
