package assetbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assetbus/types"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value.
type QueryFilter struct {
	ID                  *uuid.UUID
	TypeID              *uuid.UUID
	Name                *string
	EstPrice            *types.Money
	Price               *types.Money
	MaintenanceInterval *types.Interval
	LifeExpectancy      *types.Interval
	ModelNumber         *string
	IsEnabled           *bool
	SerialNumber        *string

	// Date filters
	StartDateCreated *time.Time
	EndDateCreated   *time.Time
	StartDateUpdated *time.Time
	EndDateUpdated   *time.Time

	// User filters
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
}
