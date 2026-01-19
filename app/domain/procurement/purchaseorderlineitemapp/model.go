package purchaseorderlineitemapp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/procurement/purchaseorderlineitembus"
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
	PurchaseOrderID      string `json:"purchase_order_id"`
	SupplierProductID    string `json:"supplier_product_id"`
	QuantityOrdered      string `json:"quantity_ordered"`
	QuantityReceived     string `json:"quantity_received"`
	QuantityCancelled    string `json:"quantity_cancelled"`
	UnitCost             string `json:"unit_cost"`
	Discount             string `json:"discount"`
	LineTotal            string `json:"line_total"`
	LineItemStatusID     string `json:"line_item_status_id"`
	ExpectedDeliveryDate string `json:"expected_delivery_date"`
	ActualDeliveryDate   string `json:"actual_delivery_date"`
	Notes                string `json:"notes"`
	CreatedBy            string `json:"created_by"`
	UpdatedBy            string `json:"updated_by"`
	CreatedDate          string `json:"created_date"`
	UpdatedDate          string `json:"updated_date"`
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
	PurchaseOrderID      string `json:"purchase_order_id" validate:"required"`
	SupplierProductID    string `json:"supplier_product_id" validate:"required"`
	QuantityOrdered      string `json:"quantity_ordered" validate:"required"`
	UnitCost             string `json:"unit_cost" validate:"required"`
	Discount             string `json:"discount" validate:"required"`
	LineTotal            string `json:"line_total" validate:"required"`
	LineItemStatusID     string `json:"line_item_status_id" validate:"required"`
	ExpectedDeliveryDate string `json:"expected_delivery_date" validate:"required"`
	Notes                string `json:"notes"`
	CreatedBy            string `json:"created_by" validate:"required"`
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
	purchaseOrderID, err := uuid.Parse(app.PurchaseOrderID)
	if err != nil {
		return purchaseorderlineitembus.NewPurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse purchaseOrderId: %s", err)
	}

	supplierProductID, err := uuid.Parse(app.SupplierProductID)
	if err != nil {
		return purchaseorderlineitembus.NewPurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse supplierProductId: %s", err)
	}

	quantityOrdered, err := strconv.Atoi(app.QuantityOrdered)
	if err != nil {
		return purchaseorderlineitembus.NewPurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse quantityOrdered: %s", err)
	}

	unitCost, err := strconv.ParseFloat(app.UnitCost, 64)
	if err != nil {
		return purchaseorderlineitembus.NewPurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse unitCost: %s", err)
	}

	discount, err := strconv.ParseFloat(app.Discount, 64)
	if err != nil {
		return purchaseorderlineitembus.NewPurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse discount: %s", err)
	}

	lineTotal, err := strconv.ParseFloat(app.LineTotal, 64)
	if err != nil {
		return purchaseorderlineitembus.NewPurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse lineTotal: %s", err)
	}

	lineItemStatusID, err := uuid.Parse(app.LineItemStatusID)
	if err != nil {
		return purchaseorderlineitembus.NewPurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse lineItemStatusId: %s", err)
	}

	expectedDeliveryDate, err := time.Parse(timeutil.FORMAT, app.ExpectedDeliveryDate)
	if err != nil {
		return purchaseorderlineitembus.NewPurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse expectedDeliveryDate: %s", err)
	}

	createdBy, err := uuid.Parse(app.CreatedBy)
	if err != nil {
		return purchaseorderlineitembus.NewPurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse createdBy: %s", err)
	}

	bus := purchaseorderlineitembus.NewPurchaseOrderLineItem{
		PurchaseOrderID:      purchaseOrderID,
		SupplierProductID:    supplierProductID,
		QuantityOrdered:      quantityOrdered,
		UnitCost:             unitCost,
		Discount:             discount,
		LineTotal:            lineTotal,
		LineItemStatusID:     lineItemStatusID,
		ExpectedDeliveryDate: expectedDeliveryDate,
		Notes:                app.Notes,
		CreatedBy:            createdBy,
	}

	return bus, nil
}

// UpdatePurchaseOrderLineItem contains information needed to update a purchase order line item.
type UpdatePurchaseOrderLineItem struct {
	SupplierProductID    *string `json:"supplier_product_id" validate:"omitempty"`
	QuantityOrdered      *string `json:"quantity_ordered" validate:"omitempty"`
	QuantityReceived     *string `json:"quantity_received" validate:"omitempty"`
	QuantityCancelled    *string `json:"quantity_cancelled" validate:"omitempty"`
	UnitCost             *string `json:"unit_cost" validate:"omitempty"`
	Discount             *string `json:"discount" validate:"omitempty"`
	LineTotal            *string `json:"line_total" validate:"omitempty"`
	LineItemStatusID     *string `json:"line_item_status_id" validate:"omitempty"`
	ExpectedDeliveryDate *string `json:"expected_delivery_date" validate:"omitempty"`
	ActualDeliveryDate   *string `json:"actual_delivery_date" validate:"omitempty"`
	Notes                *string `json:"notes" validate:"omitempty"`
	UpdatedBy            *string `json:"updated_by" validate:"omitempty"`
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
	bus := purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}

	if app.SupplierProductID != nil {
		id, err := uuid.Parse(*app.SupplierProductID)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse supplierProductId: %s", err)
		}
		bus.SupplierProductID = &id
	}

	if app.QuantityOrdered != nil {
		qty, err := strconv.Atoi(*app.QuantityOrdered)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse quantityOrdered: %s", err)
		}
		bus.QuantityOrdered = &qty
	}

	if app.QuantityReceived != nil {
		qty, err := strconv.Atoi(*app.QuantityReceived)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse quantityReceived: %s", err)
		}
		bus.QuantityReceived = &qty
	}

	if app.QuantityCancelled != nil {
		qty, err := strconv.Atoi(*app.QuantityCancelled)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse quantityCancelled: %s", err)
		}
		bus.QuantityCancelled = &qty
	}

	if app.UnitCost != nil {
		cost, err := strconv.ParseFloat(*app.UnitCost, 64)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse unitCost: %s", err)
		}
		bus.UnitCost = &cost
	}

	if app.Discount != nil {
		discount, err := strconv.ParseFloat(*app.Discount, 64)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse discount: %s", err)
		}
		bus.Discount = &discount
	}

	if app.LineTotal != nil {
		total, err := strconv.ParseFloat(*app.LineTotal, 64)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse lineTotal: %s", err)
		}
		bus.LineTotal = &total
	}

	if app.LineItemStatusID != nil {
		id, err := uuid.Parse(*app.LineItemStatusID)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse lineItemStatusId: %s", err)
		}
		bus.LineItemStatusID = &id
	}

	if app.ExpectedDeliveryDate != nil {
		date, err := time.Parse(timeutil.FORMAT, *app.ExpectedDeliveryDate)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse expectedDeliveryDate: %s", err)
		}
		bus.ExpectedDeliveryDate = &date
	}

	if app.ActualDeliveryDate != nil {
		date, err := time.Parse(timeutil.FORMAT, *app.ActualDeliveryDate)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse actualDeliveryDate: %s", err)
		}
		bus.ActualDeliveryDate = &date
	}

	if app.UpdatedBy != nil {
		id, err := uuid.Parse(*app.UpdatedBy)
		if err != nil {
			return purchaseorderlineitembus.UpdatePurchaseOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse updatedBy: %s", err)
		}
		bus.UpdatedBy = &id
	}

	bus.Notes = app.Notes

	return bus, nil
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
	ReceivedBy string `json:"received_by" validate:"required"`
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
