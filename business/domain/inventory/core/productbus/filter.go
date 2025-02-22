package productbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID                   *uuid.UUID
	SKU                  *string
	BrandID              *uuid.UUID
	ProductCategoryID    *uuid.UUID
	Name                 *string
	Description          *string
	ModelNumber          *string
	UpcCode              *string
	Status               *string
	IsActive             *bool
	IsPerishable         *bool
	HandlingInstructions *string
	UnitsPerCase         *int
	CreatedDate          *time.Time
	UpdatedDate          *time.Time
}
