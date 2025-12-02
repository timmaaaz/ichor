package warehouseapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
)

type QueryParams struct {
	Page             string
	Rows             string
	OrderBy          string
	ID               string
	Code             string
	StreetID         string
	Name             string
	IsActive         string
	StartCreatedDate string
	EndCreatedDate   string
	StartUpdatedDate string
	EndUpdatedDate   string
	CreatedBy        string
	UpdatedBy        string
}

type Warehouse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	StreetID    string `json:"street_id"`
	Name        string `json:"name"`
	IsActive    bool   `json:"is_active"`
	CreatedDate string `json:"created_date"`
	UpdatedDate string `json:"updated_date"`
	CreatedBy   string `json:"created_by"`
	UpdatedBy   string `json:"updated_by"`
}

func (app Warehouse) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppWarehouse(bus warehousebus.Warehouse) Warehouse {
	return Warehouse{
		ID:          bus.ID.String(),
		Code:        bus.Code,
		StreetID:    bus.StreetID.String(),
		Name:        bus.Name,
		IsActive:    bus.IsActive,
		CreatedDate: bus.CreatedDate.Format(time.RFC3339),
		UpdatedDate: bus.UpdatedDate.Format(time.RFC3339),
		CreatedBy:   bus.CreatedBy.String(),
		UpdatedBy:   bus.UpdatedBy.String(),
	}
}

func ToAppWarehouses(bus []warehousebus.Warehouse) []Warehouse {
	app := make([]Warehouse, len(bus))
	for i, v := range bus {
		app[i] = ToAppWarehouse(v)
	}
	return app
}

// =========================================================================

type NewWarehouse struct {
	Code      string `json:"code" validate:"omitempty"`
	StreetID  string `json:"street_id" validate:"required,uuid"`
	Name      string `json:"name" validate:"required"`
	CreatedBy string `json:"created_by" validate:"required,uuid"`
}

func (app *NewWarehouse) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewWarehouse) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewWarehouse(app NewWarehouse) (warehousebus.NewWarehouse, error) {
	streetID, err := uuid.Parse(app.StreetID)
	if err != nil {
		return warehousebus.NewWarehouse{}, errs.Newf(errs.InvalidArgument, "parse streetID: %s", err)
	}

	createdBy, err := uuid.Parse(app.CreatedBy)
	if err != nil {
		return warehousebus.NewWarehouse{}, errs.Newf(errs.InvalidArgument, "parse createdBy: %s", err)
	}

	bus := warehousebus.NewWarehouse{
		Code:      app.Code,
		StreetID:  streetID,
		Name:      app.Name,
		CreatedBy: createdBy,
	}
	return bus, nil
}

// =========================================================================

type UpdateWarehouse struct {
	Code      *string `json:"code"`
	StreetID  *string `json:"street_id" validate:"omitempty,uuid"`
	Name      *string `json:"name"`
	IsActive  *bool   `json:"is_active"`
	UpdatedBy string  `json:"updated_by" validate:"required,uuid"`
}

func (app *UpdateWarehouse) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateWarehouse) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateWarehouse(app UpdateWarehouse) (warehousebus.UpdateWarehouse, error) {
	bus := warehousebus.UpdateWarehouse{
		Code:     app.Code,
		Name:     app.Name,
		IsActive: app.IsActive,
	}

	if app.StreetID != nil {
		streetID, err := uuid.Parse(*app.StreetID)
		if err != nil {
			return warehousebus.UpdateWarehouse{}, errs.Newf(errs.InvalidArgument, "parse streetID: %s", err)
		}
		bus.StreetID = &streetID
	}

	updatedBy, err := uuid.Parse(app.UpdatedBy)
	if err != nil {
		return warehousebus.UpdateWarehouse{}, errs.Newf(errs.InvalidArgument, "parse updatedBy: %s", err)
	}
	bus.UpdatedBy = &updatedBy

	return bus, nil
}
