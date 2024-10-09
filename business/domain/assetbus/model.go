package assetbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assetbus/types"
)

// Asset represents information about an individual asset.
type Asset struct {
	ID                  uuid.UUID
	TypeID              uuid.UUID
	ConditionID         uuid.UUID
	Name                string
	EstPrice            types.Money
	Price               types.Money
	MaintenanceInterval types.Interval
	LifeExpectancy      types.Interval
	SerialNumber        string
	ModelNumber         string
	IsEnabled           bool
	DateCreated         time.Time
	DateUpdated         time.Time
	CreatedBy           uuid.UUID
	UpdatedBy           uuid.UUID
}

// NewAsset contains information needed to create a new asset.
type NewAsset struct {
	TypeID              uuid.UUID
	ConditionID         uuid.UUID
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

// UpdateAsset contains information needed to update an asset. Fields that are not
// included are intended to have separate endpoints or permissions to update.
type UpdateAsset struct {
	TypeID              *uuid.UUID
	ConditionID         *uuid.UUID
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
