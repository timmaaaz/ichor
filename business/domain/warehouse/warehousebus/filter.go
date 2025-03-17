package warehousebus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID               *uuid.UUID
	StreetID         *uuid.UUID
	Name             *string
	IsActive         *bool
	StartCreatedDate *time.Time
	EndCreatedDate   *time.Time
	StartUpdatedDate *time.Time
	EndUpdatedDate   *time.Time
	CreatedBy        *uuid.UUID
	UpdatedBy        *uuid.UUID
}
