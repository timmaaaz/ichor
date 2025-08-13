package customersbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID                *uuid.UUID
	Name              *string
	ContactID         *uuid.UUID
	DeliveryAddressID *uuid.UUID
	Notes             *string
	CreatedBy         *uuid.UUID
	UpdatedBy         *uuid.UUID

	// DateFilters
	StartCreatedDate *time.Time
	EndCreatedDate   *time.Time
	StartUpdatedDate *time.Time
	EndUpdatedDate   *time.Time
}
