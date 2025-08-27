package validassetbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus/types"
)

// ValidAsset represents information about an individual asset.
type ValidAsset struct {
	ID                  uuid.UUID
	TypeID              uuid.UUID
	Name                string
	EstPrice            types.Money
	Price               types.Money
	MaintenanceInterval types.Interval
	LifeExpectancy      types.Interval
	SerialNumber        string
	ModelNumber         string
	IsEnabled           bool
	CreatedDate         time.Time
	UpdatedDate         time.Time
	CreatedBy           uuid.UUID
	UpdatedBy           uuid.UUID
}

// NewValidAsset contains information needed to create a new asset.
type NewValidAsset struct {
	TypeID              uuid.UUID
	Name                string
	EstPrice            types.Money
	Price               types.Money
	MaintenanceInterval types.Interval
	LifeExpectancy      types.Interval
	SerialNumber        string
	ModelNumber         string
	IsEnabled           bool
	CreatedBy           uuid.UUID
}

// UpdateValidAsset contains information needed to update an asset. Fields that are not
// included are intended to have separate endpoints or permissions to update.
type UpdateValidAsset struct {
	TypeID              *uuid.UUID
	Name                *string
	EstPrice            *types.Money
	Price               *types.Money
	MaintenanceInterval *types.Interval
	LifeExpectancy      *types.Interval
	SerialNumber        *string
	ModelNumber         *string
	IsEnabled           *bool
	UpdatedBy           *uuid.UUID
}
