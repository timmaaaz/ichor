package physicalattributebus

import (
	"time"

	"github.com/google/uuid"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

type PhysicalAttribute struct {
	AttributeID         uuid.UUID `json:"attribute_id"`
	ProductID           uuid.UUID `json:"product_id"`
	Length              Dimension `json:"length"`
	Width               Dimension `json:"width"`
	Height              Dimension `json:"height"`
	Weight              Dimension `json:"weight"`
	WeightUnit          string    `json:"weight_unit"`
	Color               string    `json:"color"`
	Size                string    `json:"size"`
	Material            string    `json:"material"`
	StorageRequirements string    `json:"storage_requirements"`
	HazmatClass         string    `json:"hazmat_class"`
	ShelfLifeDays       int       `json:"shelf_life_days"`
	CreatedDate         time.Time `json:"created_date"`
	UpdatedDate         time.Time `json:"updated_date"`
}

type NewPhysicalAttribute struct {
	ProductID           uuid.UUID `json:"product_id"`
	Length              Dimension `json:"length"`
	Width               Dimension `json:"width"`
	Height              Dimension `json:"height"`
	Weight              Dimension `json:"weight"`
	WeightUnit          string    `json:"weight_unit"`
	Color               string    `json:"color"`
	Size                string    `json:"size"`
	Material            string    `json:"material"`
	StorageRequirements string    `json:"storage_requirements"`
	HazmatClass         string    `json:"hazmat_class"`
	ShelfLifeDays       int       `json:"shelf_life_days"`
}

type UpdatePhysicalAttribute struct {
	ProductID           *uuid.UUID `json:"product_id,omitempty"`
	Length              *Dimension `json:"length,omitempty"`
	Width               *Dimension `json:"width,omitempty"`
	Height              *Dimension `json:"height,omitempty"`
	Weight              *Dimension `json:"weight,omitempty"`
	WeightUnit          *string    `json:"weight_unit,omitempty"`
	Color               *string    `json:"color,omitempty"`
	Size                *string    `json:"size,omitempty"`
	Material            *string    `json:"material,omitempty"`
	StorageRequirements *string    `json:"storage_requirements,omitempty"`
	HazmatClass         *string    `json:"hazmat_class,omitempty"`
	ShelfLifeDays       *int       `json:"shelf_life_days,omitempty"`
}
