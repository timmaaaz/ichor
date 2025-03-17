package zonebus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID               *uuid.UUID
	WarehouseID      *uuid.UUID
	Name             *string
	Description      *string
	IsActive         *bool
	StartCreatedDate *time.Time
	EndCreatedDate   *time.Time
	StartUpdatedDate *time.Time
	EndUpdatedDate   *time.Time
	CreatedBy        *uuid.UUID
	UpdatedBy        *uuid.UUID
}
