package brandbus

import (
	"time"

	"github.com/google/uuid"
)

type Brand struct {
	BrandID        uuid.UUID
	Name           string
	ManufacturerID uuid.UUID
	ContactInfo    uuid.UUID
	CreatedDate    time.Time
	UpdatedDate    time.Time
}

type NewBrand struct {
	Name           string
	ManufacturerID uuid.UUID
	ContactInfo    uuid.UUID
	CreatedDate    time.Time
	UpdatedDate    time.Time
}

type UpdateBrand struct {
	Name           *string
	ManufacturerID *uuid.UUID
	ContactInfo    *uuid.UUID
	CreatedDate    *time.Time
	UpdatedDate    *time.Time
}
