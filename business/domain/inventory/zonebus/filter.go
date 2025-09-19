package zonebus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ZoneID      *uuid.UUID
	WarehouseID *uuid.UUID
	Name        *string
	Description *string
	CreatedDate *time.Time
	UpdatedDate *time.Time
}
