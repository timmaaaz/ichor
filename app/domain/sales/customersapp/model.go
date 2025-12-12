package customersapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/customersbus"
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
	Name              string  `json:"name" validate:"required"`
	ContactID         string  `json:"contact_id" validate:"required,uuid4"`
	DeliveryAddressID string  `json:"delivery_address_id" validate:"required,uuid4"`
	Notes             string  `json:"notes" validate:"omitempty,max=500"`
	CreatedBy         string  `json:"created_by" validate:"required,uuid4"`
	CreatedDate       *string `json:"created_date"` // Optional: for seeding/import
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
	contactID, err := uuid.Parse(app.ContactID)
	if err != nil {
		return customersbus.NewCustomers{}, errs.Newf(errs.InvalidArgument, "parse contactID: %s", err)
	}

	deliveryAddressID, err := uuid.Parse(app.DeliveryAddressID)
	if err != nil {
		return customersbus.NewCustomers{}, errs.Newf(errs.InvalidArgument, "parse deliveryAddressID: %s", err)
	}

	createdBy, err := uuid.Parse(app.CreatedBy)
	if err != nil {
		return customersbus.NewCustomers{}, errs.Newf(errs.InvalidArgument, "parse createdBy: %s", err)
	}

	bus := customersbus.NewCustomers{
		Name:              app.Name,
		ContactID:         contactID,
		DeliveryAddressID: deliveryAddressID,
		Notes:             app.Notes,
		CreatedBy:         createdBy,
		// CreatedDate: nil by default - API always uses server time
	}

	// Handle optional CreatedDate (for imports/admin tools only)
	if app.CreatedDate != nil && *app.CreatedDate != "" {
		createdDate, err := time.Parse(time.RFC3339, *app.CreatedDate)
		if err != nil {
			return customersbus.NewCustomers{}, errs.Newf(errs.InvalidArgument, "parse createdDate: %s", err)
		}
		bus.CreatedDate = &createdDate
	}

	return bus, nil
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
	var contactID *uuid.UUID
	if app.ContactID != nil {
		id, err := uuid.Parse(*app.ContactID)
		if err != nil {
			return customersbus.UpdateCustomers{}, errs.Newf(errs.InvalidArgument, "parse contactID: %s", err)
		}
		contactID = &id
	}

	var deliveryAddressID *uuid.UUID
	if app.DeliveryAddressID != nil {
		id, err := uuid.Parse(*app.DeliveryAddressID)
		if err != nil {
			return customersbus.UpdateCustomers{}, errs.Newf(errs.InvalidArgument, "parse deliveryAddressID: %s", err)
		}
		deliveryAddressID = &id
	}

	var updatedBy *uuid.UUID
	if app.UpdatedBy != nil {
		id, err := uuid.Parse(*app.UpdatedBy)
		if err != nil {
			return customersbus.UpdateCustomers{}, errs.Newf(errs.InvalidArgument, "parse updatedBy: %s", err)
		}
		updatedBy = &id
	}

	bus := customersbus.UpdateCustomers{
		Name:              app.Name,
		ContactID:         contactID,
		DeliveryAddressID: deliveryAddressID,
		Notes:             app.Notes,
		UpdatedBy:         updatedBy,
	}
	return bus, nil
}
