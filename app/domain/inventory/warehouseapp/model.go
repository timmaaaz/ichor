package warehouseapp

import (
	"encoding/json"
	"time"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/warehousebus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
)

type QueryParams struct {
	Page             string
	Rows             string
	OrderBy          string
	ID               string
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
	dest := warehousebus.NewWarehouse{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}

// =========================================================================

type UpdateWarehouse struct {
	StreetID  string `json:"street_id" validate:"omitempty,uuid"`
	Name      string `json:"name" validate:"omitempty"`
	IsActive  bool   `json:"is_active" validate:"omitempty"`
	UpdatedBy string `json:"updated_by" validate:"required,uuid"`
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
	dest := warehousebus.UpdateWarehouse{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}
