package brandbus

import (
	"time"

	"github.com/google/uuid"
)

type Brand struct {
	BrandID        uuid.UUID
	Name           string
	ContactInfosID uuid.UUID
	CreatedDate    time.Time
	UpdatedDate    time.Time
}

type NewBrand struct {
	Name           string
	ContactInfosID uuid.UUID
}

type UpdateBrand struct {
	Name           *string
	ContactInfosID *uuid.UUID
}
