package physicalattributebus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID                  *uuid.UUID
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
	CreatedDate         *time.Time
	UpdatedDate         *time.Time
}
