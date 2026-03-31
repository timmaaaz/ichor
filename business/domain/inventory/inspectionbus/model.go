package inspectionbus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type Inspection struct {
	InspectionID       uuid.UUID `json:"inspection_id"`
	ProductID          uuid.UUID `json:"product_id"`
	InspectorID        uuid.UUID `json:"inspector_id"`
	LotID              uuid.UUID `json:"lot_id"`
	Status             string    `json:"status"`
	Notes              string    `json:"notes"`
	InspectionDate     time.Time `json:"inspection_date"`
	NextInspectionDate time.Time `json:"next_inspection_date"`
	UpdatedDate        time.Time `json:"updated_date"`
	CreatedDate        time.Time `json:"created_date"`
}

type NewInspection struct {
	ProductID          uuid.UUID `json:"product_id"`
	InspectorID        uuid.UUID `json:"inspector_id"`
	LotID              uuid.UUID `json:"lot_id"`
	Status             string    `json:"status"`
	Notes              string    `json:"notes"`
	InspectionDate     time.Time `json:"inspection_date"`
	NextInspectionDate time.Time `json:"next_inspection_date"`
}

type UpdateInspection struct {
	ProductID          *uuid.UUID `json:"product_id,omitempty"`
	InspectorID        *uuid.UUID `json:"inspector_id,omitempty"`
	LotID              *uuid.UUID `json:"lot_id,omitempty"`
	Status             *string    `json:"status,omitempty"`
	Notes              *string    `json:"notes,omitempty"`
	InspectionDate     *time.Time `json:"inspection_date,omitempty"`
	NextInspectionDate *time.Time `json:"next_inspection_date,omitempty"`
}
