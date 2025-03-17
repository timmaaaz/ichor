package zonebus

import (
	"time"

	"github.com/google/uuid"
)

type Zone struct {
	ID          uuid.UUID
	WarehouseID uuid.UUID
	Name        string
	Description string
	IsActive    bool
	DateCreated time.Time
	DateUpdated time.Time
	CreatedBy   uuid.UUID
	UpdatedBy   uuid.UUID
}

type NewZone struct {
	WarehouseID uuid.UUID
	Name        string
	Description string
	CreatedBy   uuid.UUID
}

type UpdateZone struct {
	Name        *string
	Description *string
	IsActive    *bool
	UpdatedBy   *uuid.UUID
}
