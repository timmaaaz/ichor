package customersapp

import (
	"encoding/json"
	"time"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
)

type QueryParams struct {
	Page              string
	Rows              string
	OrderBy           string
	ID                string
	Name              string
	ContactID         string
	DeliveryAddressID string
	Notes             string
	CreatedBy         string
	UpdatedBy         string
	StartCreatedDate  string
	EndCreatedDate    string
	StartUpdatedDate  string
	EndUpdatedDate    string
}

type Customers struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	ContactID         string `json:"contact_id"`
	DeliveryAddressID string `json:"delivery_address_id"`
	Notes             string `json:"notes"`
	CreatedBy         string `json:"created_by"`
	UpdatedBy         string `json:"updated_by"`
	CreatedDate       string `json:"created_date"`
	UpdatedDate       string `json:"updated_date"`
}

func (app Customers) Encode() ([]byte, string, error) {

	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppCustomer(bus customersbus.Customers) Customers {
	return Customers{
		ID:                bus.ID.String(),
		Name:              bus.Name,
		ContactID:         bus.ContactID.String(),
		DeliveryAddressID: bus.DeliveryAddressID.String(),
		Notes:             bus.Notes,
		CreatedBy:         bus.CreatedBy.String(),
		UpdatedBy:         bus.UpdatedBy.String(),
		CreatedDate:       bus.CreatedDate.Format(time.RFC3339),
		UpdatedDate:       bus.UpdatedDate.Format(time.RFC3339),
	}
}

func ToAppCustomers(bus []customersbus.Customers) []Customers {
	app := make([]Customers, len(bus))
	for i, v := range bus {
		app[i] = ToAppCustomer(v)
	}
	return app
}

// TODO: Go over required fields here
type NewCustomers struct {
	Name              string `json:"name" validate:"required"`
	ContactID         string `json:"contact_id" validate:"required,uuid4"`
	DeliveryAddressID string `json:"delivery_address_id" validate:"required,uuid4"`
	Notes             string `json:"notes" validate:"omitempty,max=500"`
	CreatedBy         string `json:"created_by" validate:"required,uuid4"`
}

func (app *NewCustomers) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewCustomers) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewCustomers(app NewCustomers) (customersbus.NewCustomers, error) {
	dest := customersbus.NewCustomers{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	return dest, err
}

type UpdateCustomers struct {
	Name              *string `json:"name" validate:"omitempty,min=3"`
	ContactID         *string `json:"contact_id" validate:"omitempty,uuid4"`
	DeliveryAddressID *string `json:"delivery_address_id" validate:"omitempty,uuid4"`
	Notes             *string `json:"notes" validate:"omitempty,max=500"`
	UpdatedBy         *string `json:"updated_by" validate:"omitempty,uuid4"`
}

// Decode implements the decoder interface.
func (app *UpdateCustomers) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate checks the data in the model is considered clean.
func (app UpdateCustomers) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusUpdateCustomers(app UpdateCustomers) (customersbus.UpdateCustomers, error) {
	dest := customersbus.UpdateCustomers{}

	err := convert.PopulateTypesFromStrings(app, &dest)

	return dest, err
}
