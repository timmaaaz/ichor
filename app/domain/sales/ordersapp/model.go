package ordersapp

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/ordersbus"
)

const dateFormat = "2006-01-02"

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	ID                  string
	Number              string
	CustomerID          string
	FulfillmentStatusID string
	CreatedBy           string
	UpdatedBy           string
	StartDueDate        string
	EndDueDate          string
	StartCreatedDate    string
	EndCreatedDate      string
	StartUpdatedDate    string
	EndUpdatedDate      string
}

type Order struct {
	ID                  string `json:"id"`
	Number              string `json:"number"`
	CustomerID          string `json:"customer_id"`
	DueDate             string `json:"due_date"`
	FulfillmentStatusID string `json:"fulfillment_status_id"`
	CreatedBy           string `json:"created_by"`
	UpdatedBy           string `json:"updated_by"`
	CreatedDate         string `json:"created_date"`
	UpdatedDate         string `json:"updated_date"`
}

func (app Order) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppOrder(bus ordersbus.Order) Order {
	return Order{
		ID:                  bus.ID.String(),
		Number:              bus.Number,
		CustomerID:          bus.CustomerID.String(),
		DueDate:             bus.DueDate.Format(dateFormat),
		FulfillmentStatusID: bus.FulfillmentStatusID.String(),
		CreatedBy:           bus.CreatedBy.String(),
		UpdatedBy:           bus.UpdatedBy.String(),
		CreatedDate:         bus.CreatedDate.Format(time.RFC3339),
		UpdatedDate:         bus.UpdatedDate.Format(time.RFC3339),
	}
}

func ToAppOrders(bus []ordersbus.Order) []Order {
	appStatuses := make([]Order, len(bus))
	for i, status := range bus {
		appStatuses[i] = ToAppOrder(status)
	}
	return appStatuses
}

type NewOrder struct {
	Number              string  `json:"number" validate:"required"`
	CustomerID          string  `json:"customer_id" validate:"required,uuid4"`
	DueDate             string  `json:"due_date" validate:"required"`
	FulfillmentStatusID string  `json:"fulfillment_status_id" validate:"required,uuid4"`
	CreatedBy           string  `json:"created_by" validate:"required,uuid4"`
	CreatedDate         *string `json:"created_date"` // Optional: for seeding/import
}

func (app *NewOrder) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewOrder) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewOrder(app NewOrder) (ordersbus.NewOrder, error) {
	customerID, err := uuid.Parse(app.CustomerID)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse customerID: %s", err)
	}

	dueDate, err := time.Parse(dateFormat, app.DueDate)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse dueDate: %s", err)
	}

	fulfillmentStatusID, err := uuid.Parse(app.FulfillmentStatusID)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse fulfillmentStatusID: %s", err)
	}

	createdBy, err := uuid.Parse(app.CreatedBy)
	if err != nil {
		return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse createdBy: %s", err)
	}

	bus := ordersbus.NewOrder{
		Number:              app.Number,
		CustomerID:          customerID,
		DueDate:             dueDate,
		FulfillmentStatusID: fulfillmentStatusID,
		CreatedBy:           createdBy,
		// CreatedDate: nil by default - API always uses server time
	}

	// Handle optional CreatedDate (for imports/admin tools only)
	if app.CreatedDate != nil && *app.CreatedDate != "" {
		createdDate, err := time.Parse(time.RFC3339, *app.CreatedDate)
		if err != nil {
			return ordersbus.NewOrder{}, errs.Newf(errs.InvalidArgument, "parse createdDate: %s", err)
		}
		bus.CreatedDate = &createdDate
	}

	return bus, nil
}

type UpdateOrder struct {
	Number              *string `json:"number" validate:"omitempty"`
	CustomerID          *string `json:"customer_id" validate:"omitempty,uuid4"`
	DueDate             *string `json:"due_date" validate:"omitempty"`
	FulfillmentStatusID *string `json:"fulfillment_status_id" validate:"omitempty,uuid4"`
	UpdatedBy           *string `json:"updated_by" validate:"omitempty,uuid4"`
}

func (app *UpdateOrder) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateOrder) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateOrder(app UpdateOrder) (ordersbus.UpdateOrder, error) {
	var customerID *uuid.UUID
	if app.CustomerID != nil {
		id, err := uuid.Parse(*app.CustomerID)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse customerID: %s", err)
		}
		customerID = &id
	}

	var dueDate *time.Time
	if app.DueDate != nil {
		t, err := time.Parse(dateFormat, *app.DueDate)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse dueDate: %s", err)
		}
		dueDate = &t
	}

	var fulfillmentStatusID *uuid.UUID
	if app.FulfillmentStatusID != nil {
		id, err := uuid.Parse(*app.FulfillmentStatusID)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse fulfillmentStatusID: %s", err)
		}
		fulfillmentStatusID = &id
	}

	var updatedBy *uuid.UUID
	if app.UpdatedBy != nil {
		id, err := uuid.Parse(*app.UpdatedBy)
		if err != nil {
			return ordersbus.UpdateOrder{}, errs.Newf(errs.InvalidArgument, "parse updatedBy: %s", err)
		}
		updatedBy = &id
	}

	bus := ordersbus.UpdateOrder{
		Number:              app.Number,
		CustomerID:          customerID,
		DueDate:             dueDate,
		FulfillmentStatusID: fulfillmentStatusID,
		UpdatedBy:           updatedBy,
	}
	return bus, nil
}
