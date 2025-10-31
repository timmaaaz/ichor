package purchaseorderlineitemapp

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams holds the query parameters for filtering purchase order line items.
type QueryParams struct {
	Page                      string
	Rows                      string
	OrderBy                   string
	ID                        string
	PurchaseOrderID           string
	SupplierProductID         string
	LineItemStatusID          string
	CreatedBy                 string
	UpdatedBy                 string
	StartExpectedDeliveryDate string
	EndExpectedDeliveryDate   string
	StartActualDeliveryDate   string
	EndActualDeliveryDate     string
	StartCreatedDate          string
	EndCreatedDate            string
	StartUpdatedDate          string
	EndUpdatedDate            string
}

// PurchaseOrderLineItem represents a purchase order line item response.
type PurchaseOrderLineItem struct {
	ID                   string `json:"id"`
	PurchaseOrderID      string `json:"purchaseOrderId"`
	SupplierProductID    string `json:"supplierProductId"`
	QuantityOrdered      string `json:"quantityOrdered"`
	QuantityReceived     string `json:"quantityReceived"`
	QuantityCancelled    string `json:"quantityCancelled"`
	UnitCost             string `json:"unitCost"`
	Discount             string `json:"discount"`
	LineTotal            string `json:"lineTotal"`
	LineItemStatusID     string `json:"lineItemStatusId"`
	ExpectedDeliveryDate string `json:"expectedDeliveryDate"`
	ActualDeliveryDate   string `json:"actualDeliveryDate"`
	Notes                string `json:"notes"`
	CreatedBy            string `json:"createdBy"`
	UpdatedBy            string `json:"updatedBy"`
	CreatedDate          string `json:"createdDate"`
	UpdatedDate          string `json:"updatedDate"`
}

// Encode implements the Encoder interface.
func (app PurchaseOrderLineItem) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppPurchaseOrderLineItem converts a business purchase order line item to an app purchase order line item.
func ToAppPurchaseOrderLineItem(bus purchaseorderlineitembus.PurchaseOrderLineItem) PurchaseOrderLineItem {
	actualDeliveryDate := ""
	if !bus.ActualDeliveryDate.IsZero() {
		actualDeliveryDate = bus.ActualDeliveryDate.Format(timeutil.FORMAT)
	}

	return PurchaseOrderLineItem{
		ID:                   bus.ID.String(),
		PurchaseOrderID:      bus.PurchaseOrderID.String(),
		SupplierProductID:    bus.SupplierProductID.String(),
		QuantityOrdered:      fmt.Sprintf("%d", bus.QuantityOrdered),
		QuantityReceived:     fmt.Sprintf("%d", bus.QuantityReceived),
		QuantityCancelled:    fmt.Sprintf("%d", bus.QuantityCancelled),
		UnitCost:             fmt.Sprintf("%.2f", bus.UnitCost),
		Discount:             fmt.Sprintf("%.2f", bus.Discount),
		LineTotal:            fmt.Sprintf("%.2f", bus.LineTotal),
		LineItemStatusID:     bus.LineItemStatusID.String(),
		ExpectedDeliveryDate: bus.ExpectedDeliveryDate.Format(timeutil.FORMAT),
		ActualDeliveryDate:   actualDeliveryDate,
		Notes:                bus.Notes,
		CreatedBy:            bus.CreatedBy.String(),
		UpdatedBy:            bus.UpdatedBy.String(),
		CreatedDate:          bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:          bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

// ToAppPurchaseOrderLineItems converts a slice of business purchase order line items to app purchase order line items.
func ToAppPurchaseOrderLineItems(bus []purchaseorderlineitembus.PurchaseOrderLineItem) []PurchaseOrderLineItem {
	app := make([]PurchaseOrderLineItem, len(bus))
	for i, v := range bus {
		app[i] = ToAppPurchaseOrderLineItem(v)
	}
	return app
}

// NewPurchaseOrderLineItem contains information needed to create a new purchase order line item.
type NewPurchaseOrderLineItem struct {
	PurchaseOrderID      string `json:"purchaseOrderId" validate:"required"`
	SupplierProductID    string `json:"supplierProductId" validate:"required"`
	QuantityOrdered      string `json:"quantityOrdered" validate:"required"`
	UnitCost             string `json:"unitCost" validate:"required"`
	Discount             string `json:"discount" validate:"required"`
	LineTotal            string `json:"lineTotal" validate:"required"`
	LineItemStatusID     string `json:"lineItemStatusId" validate:"required"`
	ExpectedDeliveryDate string `json:"expectedDeliveryDate" validate:"required"`
	Notes                string `json:"notes"`
	CreatedBy            string `json:"createdBy" validate:"required"`
}

// Decode implements the Decoder interface.
func (app *NewPurchaseOrderLineItem) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate validates the NewPurchaseOrderLineItem fields.
func (app NewPurchaseOrderLineItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// toBusNewPurchaseOrderLineItem converts an app NewPurchaseOrderLineItem to a business NewPurchaseOrderLineItem.
func toBusNewPurchaseOrderLineItem(app NewPurchaseOrderLineItem) (purchaseorderlineitembus.NewPurchaseOrderLineItem, error) {
	dest := purchaseorderlineitembus.NewPurchaseOrderLineItem{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return purchaseorderlineitembus.NewPurchaseOrderLineItem{}, fmt.Errorf("toBusNewPurchaseOrderLineItem: %w", err)
	}

	return dest, nil
}

// UpdatePurchaseOrderLineItem contains information needed to update a purchase order line item.
type UpdatePurchaseOrderLineItem struct {
	SupplierProductID    *string `json:"supplierProductId" validate:"omitempty"`
	QuantityOrdered      *string `json:"quantityOrdered" validate:"omitempty"`
	QuantityReceived     *string `json:"quantityReceived" validate:"omitempty"`
	QuantityCancelled    *string `json:"quantityCancelled" validate:"omitempty"`
	UnitCost             *string `json:"unitCost" validate:"omitempty"`
	Discount             *string `json:"discount" validate:"omitempty"`
	LineTotal            *string `json:"lineTotal" validate:"omitempty"`
	LineItemStatusID     *string `json:"lineItemStatusId" validate:"omitempty"`
	ExpectedDeliveryDate *string `json:"expectedDeliveryDate" validate:"omitempty"`
	ActualDeliveryDate   *string `json:"actualDeliveryDate" validate:"omitempty"`
	Notes                *string `json:"notes" validate:"omitempty"`
	UpdatedBy            *string `json:"updatedBy" validate:"omitempty"`
}

// Decode implements the Decoder interface.
func (app *UpdatePurchaseOrderLineItem) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate validates the UpdatePurchaseOrderLineItem fields.
func (app UpdatePurchaseOrderLineItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// toBusUpdatePurchaseOrderLineItem converts an app UpdatePurchaseOrderLineItem to a business UpdatePurchaseOrderLineItem.
func toBusUpdatePurchaseOrderLineItem(app UpdatePurchaseOrderLineItem) (purchaseorderlineitembus.UpdatePurchaseOrderLineItem, error) {
	dest := purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, fmt.Errorf("toBusUpdatePurchaseOrderLineItem: %w", err)
	}

	return dest, nil
}

// PurchaseOrderLineItems is a collection wrapper that implements the Encoder interface.
type PurchaseOrderLineItems []PurchaseOrderLineItem

// Encode implements the Encoder interface.
func (app PurchaseOrderLineItems) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// QueryByIDsRequest represents a request to query multiple purchase order line items by their IDs.
type QueryByIDsRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// Decode implements the Decoder interface.
func (app *QueryByIDsRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate validates the QueryByIDsRequest fields.
func (app QueryByIDsRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

// ReceiveQuantityRequest represents a request to receive quantity for a line item.
type ReceiveQuantityRequest struct {
	Quantity   string `json:"quantity" validate:"required"`
	ReceivedBy string `json:"receivedBy" validate:"required"`
}

// Decode implements the Decoder interface.
func (app *ReceiveQuantityRequest) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// Validate validates the ReceiveQuantityRequest fields.
func (app ReceiveQuantityRequest) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
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
