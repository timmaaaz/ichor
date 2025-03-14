package physicalattributebus

import (
	"time"

	"github.com/google/uuid"
)

type PhysicalAttribute struct {
	AttributeID         uuid.UUID
	ProductID           uuid.UUID
	Length              float32
	Width               float32
	Height              float32
	Weight              float32
	WeightUnit          string
	Color               string
	Size                string
	Material            string
	StorageRequirements string
	HazmatClass         string
	ShelfLifeDays       int
	CreatedDate         time.Time
	UpdatedDate         time.Time
}

type NewPhysicalAttribute struct {
	ProductID           uuid.UUID
	Length              float32
	Width               float32
	Height              float32
	Weight              float32
	WeightUnit          string
	Color               string
	Size                string
	Material            string
	StorageRequirements string
	HazmatClass         string
	ShelfLifeDays       int
}

type UpdatePhysicalAttribute struct {
	ProductID           *uuid.UUID
	Length              *float32
	Width               *float32
	Height              *float32
	Weight              *float32
	WeightUnit          *string
	Color               *string
	Size                *string
	Material            *string
	StorageRequirements *string
	HazmatClass         *string
	ShelfLifeDays       *int
}
