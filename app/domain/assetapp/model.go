package assetapp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/assetbus"
	"github.com/timmaaaz/ichor/business/domain/assetbus/types"
)

type QueryParams struct {
	Page                string
	Rows                string
	OrderBy             string
	ID                  string
	TypeID              string
	Name                string
	EstPrice            string
	Price               string
	MaintenanceInterval string
	LifeExpectancy      string
	ModelNumber         string
	IsEnabled           string
	StartDateCreated    string
	EndDateCreated      string
	StartDateUpdated    string
	EndDateUpdated      string
	CreatedBy           string
	UpdatedBy           string
}

// =============================================================================

type Asset struct {
	ID                  string `json:"id"`
	TypeID              string `json:"type_id"`
	Name                string `json:"name"`
	EstPrice            string `json:"est_price"`
	Price               string `json:"price"`
	MaintenanceInterval string `json:"maintenance_interval"`
	LifeExpectancy      string `json:"life_expectancy"`
	ModelNumber         string `json:"model_number"`
	IsEnabled           bool   `json:"is_enabled"`
	DateCreated         string `json:"date_created"`
	DateUpdated         string `json:"date_updated"`
	CreatedBy           string `json:"created_by"`
	UpdatedBy           string `json:"updated_by"`
}

// Encode implements the encoder interface.
func (app Asset) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppAsset(bus assetbus.Asset) Asset {
	return Asset{
		ID:                  bus.ID.String(),
		TypeID:              bus.TypeID.String(),
		Name:                bus.Name,
		EstPrice:            bus.EstPrice.Value(),
		Price:               bus.Price.Value(),
		MaintenanceInterval: bus.MaintenanceInterval.Value(),
		LifeExpectancy:      bus.LifeExpectancy.Value(),
		ModelNumber:         bus.ModelNumber,
		IsEnabled:           bus.IsEnabled,
		DateCreated:         bus.DateCreated.Format(time.RFC3339),
		DateUpdated:         bus.DateUpdated.Format(time.RFC3339),
		CreatedBy:           bus.CreatedBy.String(),
		UpdatedBy:           bus.UpdatedBy.String(),
	}
}

func ToAppAssets(bus []assetbus.Asset) []Asset {
	app := make([]Asset, len(bus))
	for i, v := range bus {
		app[i] = ToAppAsset(v)
	}
	return app
}

// =============================================================================

type NewAsset struct {
	TypeID              string `json:"type_id" validate:"required,uuid"`
	Name                string `json:"name" validate:"required"`
	EstPrice            string `json:"est_price"`
	Price               string `json:"price"`
	MaintenanceInterval string `json:"maintenance_interval"`
	LifeExpectancy      string `json:"life_expectancy"`
	ModelNumber         string `json:"model_number"`
	IsEnabled           bool   `json:"is_enabled"`
	CreatedBy           string `json:"created_by" validate:"required"`
}

// Decode implements the decoder interface.
func (app *NewAsset) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewAsset) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewAsset(app NewAsset) (assetbus.NewAsset, error) {
	var typeID uuid.UUID
	var estPrice types.Money
	var price types.Money
	var maintenanceInterval types.Interval
	var lifeExpectancy types.Interval

	var err error

	if app.TypeID != "" {
		typeID, err = uuid.Parse(app.TypeID)
		if err != nil {
			return assetbus.NewAsset{}, fmt.Errorf("tobusnewasset: %w", err)
		}
	}

	if app.EstPrice != "" {
		estPrice, err = types.ParseMoney(app.EstPrice)
		if err != nil {
			return assetbus.NewAsset{}, fmt.Errorf("tobusnewasset: %w", err)
		}
	}

	if app.Price != "" {
		price, err = types.ParseMoney(app.Price)
		if err != nil {
			return assetbus.NewAsset{}, fmt.Errorf("tobusnewasset: %w", err)
		}
	}

	if app.MaintenanceInterval != "" {
		maintenanceInterval, err = types.ParseInterval(app.MaintenanceInterval)
		if err != nil {
			return assetbus.NewAsset{}, fmt.Errorf("tobusnewasset: %w", err)
		}
	}

	if app.LifeExpectancy != "" {
		lifeExpectancy, err = types.ParseInterval(app.LifeExpectancy)
		if err != nil {
			return assetbus.NewAsset{}, fmt.Errorf("tobusnewasset: %w", err)
		}
	}

	createdBy, err := uuid.Parse(app.CreatedBy)
	if err != nil {
		return assetbus.NewAsset{}, fmt.Errorf("tobusnewasset: %w", err)
	}

	return assetbus.NewAsset{
		TypeID:              typeID,
		Name:                app.Name,
		EstPrice:            estPrice,
		Price:               price,
		MaintenanceInterval: maintenanceInterval,
		LifeExpectancy:      lifeExpectancy,
		ModelNumber:         app.ModelNumber,
		IsEnabled:           app.IsEnabled,
		CreatedBy:           createdBy,
	}, nil
}

// =============================================================================

// UpdateAsset contains information needed to update an asset.
type UpdateAsset struct {
	TypeID              *string `json:"type_id"`
	Name                *string `json:"name"`
	EstPrice            *string `json:"est_price"`
	Price               *string `json:"price"`
	MaintenanceInterval *string `json:"maintenance_interval"`
	LifeExpectancy      *string `json:"life_expectancy"`
	ModelNumber         *string `json:"model_number"`
	IsEnabled           *bool   `json:"is_enabled"`
	UpdatedBy           *string `json:"updated_by"`
}

// Decode implements the decoder interface.
func (app *UpdateAsset) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateAsset) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateAsset(app UpdateAsset) (assetbus.UpdateAsset, error) {
	var typeID *uuid.UUID
	var name *string
	var estPrice *types.Money
	var price *types.Money
	var maintenanceInterval *types.Interval
	var lifeExpectancy *types.Interval
	var modelNumber *string
	var isEnabled *bool
	var updatedBy *uuid.UUID

	var err error

	if app.TypeID != nil {
		id, err := uuid.Parse(*app.TypeID)
		if err != nil {
			return assetbus.UpdateAsset{}, fmt.Errorf("tobusupdateasset: %w", err)
		}
		typeID = &id
	}

	if app.Name != nil {
		name = app.Name
	}

	if app.EstPrice != nil {
		estPrice, err = types.ParseMoneyPtr(*app.EstPrice)
		if err != nil {
			return assetbus.UpdateAsset{}, fmt.Errorf("tobusupdateasset: %w", err)
		}
	}

	if app.Price != nil {
		price, err = types.ParseMoneyPtr(*app.Price)
		if err != nil {
			return assetbus.UpdateAsset{}, fmt.Errorf("tobusupdateasset: %w", err)
		}
	}

	if app.MaintenanceInterval != nil {
		maintenanceInterval, err = types.ParseIntervalPtr(*app.MaintenanceInterval)
		if err != nil {
			return assetbus.UpdateAsset{}, fmt.Errorf("tobusupdateasset: %w", err)
		}
	}

	if app.LifeExpectancy != nil {
		lifeExpectancy, err = types.ParseIntervalPtr(*app.LifeExpectancy)
		if err != nil {
			return assetbus.UpdateAsset{}, fmt.Errorf("tobusupdateasset: %w", err)
		}
	}

	if app.ModelNumber != nil {
		modelNumber = app.ModelNumber
	}

	if app.IsEnabled != nil {
		isEnabled = app.IsEnabled
	}

	if app.UpdatedBy != nil {
		id, err := uuid.Parse(*app.UpdatedBy)
		if err != nil {
			return assetbus.UpdateAsset{}, fmt.Errorf("tobusupdateasset: %w", err)
		}
		updatedBy = &id
	}

	return assetbus.UpdateAsset{
		TypeID:              typeID,
		Name:                name,
		EstPrice:            estPrice,
		Price:               price,
		MaintenanceInterval: maintenanceInterval,
		LifeExpectancy:      lifeExpectancy,
		ModelNumber:         modelNumber,
		IsEnabled:           isEnabled,
		UpdatedBy:           updatedBy,
	}, nil
}
