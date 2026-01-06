package validassetbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/assets/validassetbus/types"
)

// JSON tags are required for workflow event serialization. The workflow system
// (via EventPublisher) marshals business models to JSON for RawData in TriggerEvents.
// Without these tags, Go defaults to PascalCase keys, but workflow action handlers
// expect snake_case keys to match API conventions.

// ValidAsset represents information about an individual asset.
type ValidAsset struct {
	ID                  uuid.UUID      `json:"id"`
	TypeID              uuid.UUID      `json:"type_id"`
	Name                string         `json:"name"`
	EstPrice            types.Money    `json:"est_price"`
	Price               types.Money    `json:"price"`
	MaintenanceInterval types.Interval `json:"maintenance_interval"`
	LifeExpectancy      types.Interval `json:"life_expectancy"`
	SerialNumber        string         `json:"serial_number"`
	ModelNumber         string         `json:"model_number"`
	IsEnabled           bool           `json:"is_enabled"`
	CreatedDate         time.Time      `json:"created_date"`
	UpdatedDate         time.Time      `json:"updated_date"`
	CreatedBy           uuid.UUID      `json:"created_by"`
	UpdatedBy           uuid.UUID      `json:"updated_by"`
}

// NewValidAsset contains information needed to create a new asset.
type NewValidAsset struct {
	TypeID              uuid.UUID      `json:"type_id"`
	Name                string         `json:"name"`
	EstPrice            types.Money    `json:"est_price"`
	Price               types.Money    `json:"price"`
	MaintenanceInterval types.Interval `json:"maintenance_interval"`
	LifeExpectancy      types.Interval `json:"life_expectancy"`
	SerialNumber        string         `json:"serial_number"`
	ModelNumber         string         `json:"model_number"`
	IsEnabled           bool           `json:"is_enabled"`
	CreatedBy           uuid.UUID      `json:"created_by"`
	CreatedDate         *time.Time     `json:"created_date,omitempty"` // Optional: if nil, uses time.Now(), otherwise explicit date for seeding
}

// UpdateValidAsset contains information needed to update an asset. Fields that are not
// included are intended to have separate endpoints or permissions to update.
type UpdateValidAsset struct {
	TypeID              *uuid.UUID      `json:"type_id,omitempty"`
	Name                *string         `json:"name,omitempty"`
	EstPrice            *types.Money    `json:"est_price,omitempty"`
	Price               *types.Money    `json:"price,omitempty"`
	MaintenanceInterval *types.Interval `json:"maintenance_interval,omitempty"`
	LifeExpectancy      *types.Interval `json:"life_expectancy,omitempty"`
	SerialNumber        *string         `json:"serial_number,omitempty"`
	ModelNumber         *string         `json:"model_number,omitempty"`
	IsEnabled           *bool           `json:"is_enabled,omitempty"`
	UpdatedBy           *uuid.UUID      `json:"updated_by,omitempty"`
}
