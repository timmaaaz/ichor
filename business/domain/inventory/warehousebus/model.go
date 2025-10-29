package warehousebus

import (
	"time"

	"github.com/google/uuid"
)

type Warehouse struct {
	ID          uuid.UUID
	Code        string
	StreetID    uuid.UUID
	Name        string
	IsActive    bool
	CreatedDate time.Time
	UpdatedDate time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}

type NewWarehouse struct {
	Code      string
	StreetID  uuid.UUID
	Name      string
	CreatedBy uuid.UUID
}

type UpdateWarehouse struct {
	Code      *string
	StreetID  *uuid.UUID
	Name      *string
	IsActive  *bool
	UpdatedBy *uuid.UUID
}
