package purchaseorderlineitemstatusapp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitemstatusbus"
)

type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ID          string
	Name        string
	Description string
}

type PurchaseOrderLineItemStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

func (app PurchaseOrderLineItemStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppPurchaseOrderLineItemStatus(bus purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus) PurchaseOrderLineItemStatus {
	return PurchaseOrderLineItemStatus{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
		SortOrder:   bus.SortOrder,
	}
}

func ToAppPurchaseOrderLineItemStatuses(bus []purchaseorderlineitemstatusbus.PurchaseOrderLineItemStatus) []PurchaseOrderLineItemStatus {
	app := make([]PurchaseOrderLineItemStatus, len(bus))
	for i, v := range bus {
		app[i] = ToAppPurchaseOrderLineItemStatus(v)
	}
	return app
}

type NewPurchaseOrderLineItemStatus struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty"`
	SortOrder   int    `json:"sortOrder" validate:"required"`
}

func (app *NewPurchaseOrderLineItemStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewPurchaseOrderLineItemStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewPurchaseOrderLineItemStatus(app NewPurchaseOrderLineItemStatus) (purchaseorderlineitemstatusbus.NewPurchaseOrderLineItemStatus, error) {
	return purchaseorderlineitemstatusbus.NewPurchaseOrderLineItemStatus{
		Name:        app.Name,
		Description: app.Description,
		SortOrder:   app.SortOrder,
	}, nil
}

type UpdatePurchaseOrderLineItemStatus struct {
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	SortOrder   *int    `json:"sortOrder" validate:"omitempty"`
}

func (app *UpdatePurchaseOrderLineItemStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdatePurchaseOrderLineItemStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdatePurchaseOrderLineItemStatus(app UpdatePurchaseOrderLineItemStatus) (purchaseorderlineitemstatusbus.UpdatePurchaseOrderLineItemStatus, error) {
	dest := purchaseorderlineitemstatusbus.UpdatePurchaseOrderLineItemStatus{}

	if app.Name != nil {
		dest.Name = app.Name
	}

	if app.Description != nil {
		dest.Description = app.Description
	}

	if app.SortOrder != nil {
		dest.SortOrder = app.SortOrder
	}

	return dest, nil
}

func toBusIDs(ids []string) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("parse id[%d]: %w", i, err)
		}
		uuids[i] = uid
	}
	return uuids, nil
}

// PurchaseOrderLineItemStatuses is a collection wrapper that implements the Encoder interface.
type PurchaseOrderLineItemStatuses []PurchaseOrderLineItemStatus

func (app PurchaseOrderLineItemStatuses) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// QueryByIDsRequest represents a request to query multiple purchase order line item statuses by their IDs.
type QueryByIDsRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

func (app *QueryByIDsRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app QueryByIDsRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}
