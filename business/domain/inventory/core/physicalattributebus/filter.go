package physicalattributebus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID                  *uuid.UUID
	ProductID           *uuid.UUID
	Length              *float64
	Width               *float64
	Height              *float64
	Weight              *float64
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
