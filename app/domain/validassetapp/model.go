package validassetapp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/validassetbus"
	"github.com/timmaaaz/ichor/business/domain/validassetbus/types"
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

type ValidAsset struct {
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
func (app ValidAsset) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppValidAsset(bus validassetbus.ValidAsset) ValidAsset {
	return ValidAsset{
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

func ToAppValidAssets(bus []validassetbus.ValidAsset) []ValidAsset {
	app := make([]ValidAsset, len(bus))
	for i, v := range bus {
		app[i] = ToAppValidAsset(v)
	}
	return app
}

// =============================================================================

type NewValidAsset struct {
	TypeID              string `json:"type_id" validate:"required"`
	Name                string `json:"name" validate:"required"`
	EstPrice            string `json:"est_price"`
	Price               string `json:"price"`
	MaintenanceInterval string `json:"maintenance_interval"`
	LifeExpectancy      string `json:"life_expectancy"`
	ModelNumber         string `json:"model_number"`
	IsEnabled           bool   `json:"is_enabled"`
	CreatedBy           string `json:"created_by"`
}

// Decode implements the decoder interface.
func (app *NewValidAsset) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app NewValidAsset) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewValidAsset(app NewValidAsset) (validassetbus.NewValidAsset, error) {
	var typeID uuid.UUID
	var estPrice types.Money
	var price types.Money
	var maintenanceInterval types.Interval
	var lifeExpectancy types.Interval

	var err error

	if app.TypeID != "" {
		typeID, err = uuid.Parse(app.TypeID)
		if err != nil {
			return validassetbus.NewValidAsset{}, fmt.Errorf("tobusnewvalidasset`: %w", err)
		}
	}

	if app.EstPrice != "" {
		estPrice, err = types.ParseMoney(app.EstPrice)
		if err != nil {
			return validassetbus.NewValidAsset{}, fmt.Errorf("tobusnewvalidasset: %w", err)
		}
	}

	if app.Price != "" {
		price, err = types.ParseMoney(app.Price)
		if err != nil {
			return validassetbus.NewValidAsset{}, fmt.Errorf("tobusnewvalidasset: %w", err)
		}
	}

	if app.MaintenanceInterval != "" {
		maintenanceInterval, err = types.ParseInterval(app.MaintenanceInterval)
		if err != nil {
			return validassetbus.NewValidAsset{}, fmt.Errorf("tobusnewvalidasset: %w", err)
		}
	}

	if app.LifeExpectancy != "" {
		lifeExpectancy, err = types.ParseInterval(app.LifeExpectancy)
		if err != nil {
			return validassetbus.NewValidAsset{}, fmt.Errorf("tobusnewvalidasset: %w", err)
		}
	}

	createdBy, err := uuid.Parse(app.CreatedBy)
	if err != nil {
		return validassetbus.NewValidAsset{}, fmt.Errorf("tobusnewvalidasset: %w", err)
	}

	return validassetbus.NewValidAsset{
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

// UpdateValidAsset contains information needed to update an asset.
type UpdateValidAsset struct {
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
func (app *UpdateValidAsset) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateValidAsset) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateValidAsset(app UpdateValidAsset) (validassetbus.UpdateValidAsset, error) {
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
			return validassetbus.UpdateValidAsset{}, fmt.Errorf("tobusupdatevalidasset: %w", err)
		}
		typeID = &id
	}

	if app.Name != nil {
		name = app.Name
	}

	if app.EstPrice != nil {
		estPrice, err = types.ParseMoneyPtr(*app.EstPrice)
		if err != nil {
			return validassetbus.UpdateValidAsset{}, fmt.Errorf("tobusupdatevalidasset: %w", err)
		}
	}

	if app.Price != nil {
		price, err = types.ParseMoneyPtr(*app.Price)
		if err != nil {
			return validassetbus.UpdateValidAsset{}, fmt.Errorf("tobusupdatevalidasset: %w", err)
		}
	}

	if app.MaintenanceInterval != nil {
		maintenanceInterval, err = types.ParseIntervalPtr(*app.MaintenanceInterval)
		if err != nil {
			return validassetbus.UpdateValidAsset{}, fmt.Errorf("tobusupdatevalidasset: %w", err)
		}
	}

	if app.LifeExpectancy != nil {
		lifeExpectancy, err = types.ParseIntervalPtr(*app.LifeExpectancy)
		if err != nil {
			return validassetbus.UpdateValidAsset{}, fmt.Errorf("tobusupdatevalidasset: %w", err)
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
			return validassetbus.UpdateValidAsset{}, fmt.Errorf("tobusupdatevalidasset: %w", err)
		}
		updatedBy = &id
	}

	return validassetbus.UpdateValidAsset{
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
