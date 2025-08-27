package warehousebus

import (
	"time"

	"github.com/google/uuid"
)

type Warehouse struct {
	ID          uuid.UUID
	StreetID    uuid.UUID
	Name        string
	IsActive    bool
	CreatedDate time.Time
	UpdatedDate time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}

type NewWarehouse struct {
	StreetID  uuid.UUID
	Name      string
	CreatedBy uuid.UUID
}

type UpdateWarehouse struct {
	StreetID  *uuid.UUID
	Name      *string
	IsActive  *bool
	UpdatedBy *uuid.UUID
}
