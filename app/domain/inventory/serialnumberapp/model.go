package serialnumberapp

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/serialnumberbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	SerialID     string
	LotID        string
	ProductID    string
	LocationID   string
	SerialNumber string
	Status       string
	CreatedDate  string
	UpdatedDate  string
}

type SerialNumber struct {
	SerialID     string `json:"serial_id"`
	LotID        string `json:"lot_id"`
	ProductID    string `json:"product_id"`
	LocationID   string `json:"location_id"`
	SerialNumber string `json:"serial_number"`
	Status       string `json:"status"`
	CreatedDate  string `json:"created_date"`
	UpdatedDate  string `json:"updated_date"`
}

func (app SerialNumber) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppSerialNumber(bus serialnumberbus.SerialNumber) SerialNumber {
	return SerialNumber{
		SerialID:     bus.SerialID.String(),
		LotID:        bus.LotID.String(),
		ProductID:    bus.ProductID.String(),
		LocationID:   bus.LocationID.String(),
		SerialNumber: bus.SerialNumber,
		Status:       bus.Status,
		CreatedDate:  bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:  bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppSerialNumbers(bus []serialnumberbus.SerialNumber) []SerialNumber {
	app := make([]SerialNumber, len(bus))
	for i, v := range bus {
		app[i] = ToAppSerialNumber(v)
	}
	return app
}

type NewSerialNumber struct {
	LotID        string `json:"lot_id" validate:"required,min=36,max=36"`
	ProductID    string `json:"product_id" validate:"required,min=36,max=36"`
	LocationID   string `json:"location_id" validate:"required,min=36,max=36"`
	SerialNumber string `json:"serial_number" validate:"required"`
	Status       string `json:"status" validate:"required"`
}

func (app *NewSerialNumber) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewSerialNumber) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func toBusNewSerialNumber(app NewSerialNumber) (serialnumberbus.NewSerialNumber, error) {
	dest := serialnumberbus.NewSerialNumber{}
	err := convert.PopulateTypesFromStrings(app, &dest)

	return dest, err
}

type UpdateSerialNumber struct {
	LotID        *string `json:"lot_id" validate:"omitempty,min=36,max=36"`
	ProductID    *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	LocationID   *string `json:"location_id" validate:"omitempty,min=36,max=36"`
	SerialNumber *string `json:"serial_number" validate:"omitempty"`
	Status       *string `json:"status" validate:"omitempty"`
}

func (app *UpdateSerialNumber) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateSerialNumber) Validate() error {
	if err := errs.Check(app); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	return nil
}

func toBusUpdateSerialNumber(app UpdateSerialNumber) (serialnumberbus.UpdateSerialNumber, error) {
	dest := serialnumberbus.UpdateSerialNumber{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return serialnumberbus.UpdateSerialNumber{}, fmt.Errorf("error populating serial number: %s", err)
	}
	return dest, nil
}
