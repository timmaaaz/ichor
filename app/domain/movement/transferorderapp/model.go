package transferorderapp

import (
	"encoding/json"
	"fmt"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/movement/transferorderbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
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
	dest := transferorderbus.NewTransferOrder{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return transferorderbus.NewTransferOrder{}, fmt.Errorf("toBusNewTransferOrder: %w", err)
	}
	return dest, err
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
	dest := transferorderbus.UpdateTransferOrder{}

	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return transferorderbus.UpdateTransferOrder{}, fmt.Errorf("toBusUpdateTransferOrder: %w", err)
	}
	return dest, err
}
