package purchaseorderstatusapp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderstatusbus"
)

type QueryParams struct {
	Page        string
	Rows        string
	OrderBy     string
	ID          string
	Name        string
	Description string
}

type PurchaseOrderStatus struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
}

func (app PurchaseOrderStatus) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppPurchaseOrderStatus(bus purchaseorderstatusbus.PurchaseOrderStatus) PurchaseOrderStatus {
	return PurchaseOrderStatus{
		ID:          bus.ID.String(),
		Name:        bus.Name,
		Description: bus.Description,
		SortOrder:   bus.SortOrder,
	}
}

func ToAppPurchaseOrderStatuses(bus []purchaseorderstatusbus.PurchaseOrderStatus) []PurchaseOrderStatus {
	app := make([]PurchaseOrderStatus, len(bus))
	for i, v := range bus {
		app[i] = ToAppPurchaseOrderStatus(v)
	}
	return app
}

type NewPurchaseOrderStatus struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"omitempty"`
	SortOrder   int    `json:"sortOrder" validate:"required"`
}

func (app *NewPurchaseOrderStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewPurchaseOrderStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewPurchaseOrderStatus(app NewPurchaseOrderStatus) (purchaseorderstatusbus.NewPurchaseOrderStatus, error) {
	return purchaseorderstatusbus.NewPurchaseOrderStatus{
		Name:        app.Name,
		Description: app.Description,
		SortOrder:   app.SortOrder,
	}, nil
}

type UpdatePurchaseOrderStatus struct {
	Name        *string `json:"name" validate:"omitempty"`
	Description *string `json:"description" validate:"omitempty"`
	SortOrder   *int    `json:"sortOrder" validate:"omitempty"`
}

func (app *UpdatePurchaseOrderStatus) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdatePurchaseOrderStatus) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdatePurchaseOrderStatus(app UpdatePurchaseOrderStatus) (purchaseorderstatusbus.UpdatePurchaseOrderStatus, error) {
	dest := purchaseorderstatusbus.UpdatePurchaseOrderStatus{}

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

// PurchaseOrderStatuses is a collection wrapper that implements the Encoder interface.
type PurchaseOrderStatuses []PurchaseOrderStatus

func (app PurchaseOrderStatuses) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// QueryByIDsRequest represents a request to query multiple purchase order statuses by their IDs.
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
