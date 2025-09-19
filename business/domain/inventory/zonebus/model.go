package zonebus

import (
	"time"

	"github.com/google/uuid"
)

type Zone struct {
	ZoneID      uuid.UUID
	WarehouseID uuid.UUID
	Name        string
	Description string
	CreatedDate time.Time
	UpdatedDate time.Time
}

type NewZone struct {
	WarehouseID uuid.UUID
	Name        string
	Description string
}

type UpdateZone struct {
	WarehouseID *uuid.UUID
	Name        *string
	Description *string
}
