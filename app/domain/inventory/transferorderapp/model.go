package transferorderapp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/transferorderbus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	TransferID     string
	ProductID      string
	FromLocationID string
	ToLocationID   string
	RequestedByID  string
	ApprovedByID   string
	Quantity       string
	Status         string
	TransferDate   string
	CreatedDate    string
	UpdatedDate    string
}

type TransferOrder struct {
	TransferID     string `json:"transfer_id"`
	ProductID      string `json:"product_id"`
	FromLocationID string `json:"from_location_id"`
	ToLocationID   string `json:"to_location_id"`
	RequestedByID  string `json:"requested_by"`
	ApprovedByID   string `json:"approved_by"`
	Quantity       string `json:"quantity"`
	Status         string `json:"status"`
	TransferDate   string `json:"transfer_date"`
	CreatedDate    string `json:"created_date"`
	UpdatedDate    string `json:"updated_date"`
}

func (app TransferOrder) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppTransferOrder(bus transferorderbus.TransferOrder) TransferOrder {
	return TransferOrder{
		TransferID:     bus.TransferID.String(),
		ProductID:      bus.ProductID.String(),
		FromLocationID: bus.FromLocationID.String(),
		ToLocationID:   bus.ToLocationID.String(),
		RequestedByID:  bus.RequestedByID.String(),
		ApprovedByID:   bus.ApprovedByID.String(),
		Quantity:       fmt.Sprintf("%d", bus.Quantity),
		Status:         bus.Status,
		TransferDate:   bus.TransferDate.Format(timeutil.FORMAT),
		CreatedDate:    bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:    bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppTransferOrders(bus []transferorderbus.TransferOrder) []TransferOrder {
	app := make([]TransferOrder, len(bus))
	for i, v := range bus {
		app[i] = ToAppTransferOrder(v)
	}
	return app
}

type NewTransferOrder struct {
	ProductID      string `json:"product_id" validate:"required,min=36,max=36"`
	FromLocationID string `json:"from_location_id" validate:"required,min=36,max=36"`
	ToLocationID   string `json:"to_location_id" validate:"required,min=36,max=36"`
	RequestedByID  string `json:"requested_by" validate:"required,min=36,max=36"`
	ApprovedByID   string `json:"approved_by" validate:"required,min=36,max=36"`
	Quantity       string `json:"quantity" validate:"required"`
	Status         string `json:"status" validate:"required"`
	TransferDate   string `json:"transfer_date" validate:"required"`
}

func (app *NewTransferOrder) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewTransferOrder) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewTransferOrder(app NewTransferOrder) (transferorderbus.NewTransferOrder, error) {
	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return transferorderbus.NewTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
	}

	fromLocationID, err := uuid.Parse(app.FromLocationID)
	if err != nil {
		return transferorderbus.NewTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse fromLocationID: %s", err)
	}

	toLocationID, err := uuid.Parse(app.ToLocationID)
	if err != nil {
		return transferorderbus.NewTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse toLocationID: %s", err)
	}

	requestedByID, err := uuid.Parse(app.RequestedByID)
	if err != nil {
		return transferorderbus.NewTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse requestedByID: %s", err)
	}

	approvedByID, err := uuid.Parse(app.ApprovedByID)
	if err != nil {
		return transferorderbus.NewTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse approvedByID: %s", err)
	}

	quantity, err := strconv.Atoi(app.Quantity)
	if err != nil {
		return transferorderbus.NewTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
	}

	transferDate, err := time.Parse(timeutil.FORMAT, app.TransferDate)
	if err != nil {
		return transferorderbus.NewTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse transferDate: %s", err)
	}

	bus := transferorderbus.NewTransferOrder{
		ProductID:      productID,
		FromLocationID: fromLocationID,
		ToLocationID:   toLocationID,
		RequestedByID:  requestedByID,
		ApprovedByID:   approvedByID,
		Quantity:       quantity,
		Status:         app.Status,
		TransferDate:   transferDate,
	}
	return bus, nil
}

type UpdateTransferOrder struct {
	ProductID      *string `json:"product_id" validate:"omitempty,min=36,max=36"`
	FromLocationID *string `json:"from_location_id" validate:"omitempty,min=36,max=36"`
	ToLocationID   *string `json:"to_location_id" validate:"omitempty,min=36,max=36"`
	RequestedByID  *string `json:"requested_by" validate:"omitempty,min=36,max=36"`
	ApprovedByID   *string `json:"approved_by" validate:"omitempty,min=36,max=36"`
	Quantity       *string `json:"quantity" validate:"omitempty"`
	Status         *string `json:"status" validate:"omitempty"`
	TransferDate   *string `json:"transfer_date" validate:"omitempty"`
}

func (app *UpdateTransferOrder) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateTransferOrder) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateTransferOrder(app UpdateTransferOrder) (transferorderbus.UpdateTransferOrder, error) {
	bus := transferorderbus.UpdateTransferOrder{
		Status: app.Status,
	}

	if app.ProductID != nil {
		productID, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return transferorderbus.UpdateTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
		}
		bus.ProductID = &productID
	}

	if app.FromLocationID != nil {
		fromLocationID, err := uuid.Parse(*app.FromLocationID)
		if err != nil {
			return transferorderbus.UpdateTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse fromLocationID: %s", err)
		}
		bus.FromLocationID = &fromLocationID
	}

	if app.ToLocationID != nil {
		toLocationID, err := uuid.Parse(*app.ToLocationID)
		if err != nil {
			return transferorderbus.UpdateTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse toLocationID: %s", err)
		}
		bus.ToLocationID = &toLocationID
	}

	if app.RequestedByID != nil {
		requestedByID, err := uuid.Parse(*app.RequestedByID)
		if err != nil {
			return transferorderbus.UpdateTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse requestedByID: %s", err)
		}
		bus.RequestedByID = &requestedByID
	}

	if app.ApprovedByID != nil {
		approvedByID, err := uuid.Parse(*app.ApprovedByID)
		if err != nil {
			return transferorderbus.UpdateTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse approvedByID: %s", err)
		}
		bus.ApprovedByID = &approvedByID
	}

	if app.Quantity != nil {
		quantity, err := strconv.Atoi(*app.Quantity)
		if err != nil {
			return transferorderbus.UpdateTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
		}
		bus.Quantity = &quantity
	}

	if app.TransferDate != nil {
		transferDate, err := time.Parse(timeutil.FORMAT, *app.TransferDate)
		if err != nil {
			return transferorderbus.UpdateTransferOrder{}, errs.Newf(errs.InvalidArgument, "parse transferDate: %s", err)
		}
		bus.TransferDate = &transferDate
	}

	return bus, nil
}
