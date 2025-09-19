package orderlineitemsapp

import (
	"encoding/json"
	"strconv"

	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
	"github.com/timmaaaz/ichor/business/sdk/convert"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	ID                            string
	OrderID                       string
	ProductID                     string
	Quantity                      string
	Discount                      string
	LineItemFulfillmentStatusesID string
	CreatedBy                     string
	StartCreatedDate              string
	EndCreatedDate                string
	UpdatedBy                     string
	StartUpdatedDate              string
	EndUpdatedDate                string
}

type OrderLineItem struct {
	ID                            string
	OrderID                       string
	ProductID                     string
	Quantity                      string
	Discount                      string
	LineItemFulfillmentStatusesID string
	CreatedBy                     string
	CreatedDate                   string
	UpdatedBy                     string
	UpdatedDate                   string
}

func (app OrderLineItem) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppOrderLineItem(bus orderlineitemsbus.OrderLineItem) OrderLineItem {
	return OrderLineItem{
		ID:                            bus.ID.String(),
		OrderID:                       bus.OrderID.String(),
		ProductID:                     bus.ProductID.String(),
		Quantity:                      strconv.Itoa(bus.Quantity),
		Discount:                      strconv.FormatFloat(bus.Discount, 'f', 2, 64),
		LineItemFulfillmentStatusesID: bus.LineItemFulfillmentStatusesID.String(),
		CreatedBy:                     bus.CreatedBy.String(),
		CreatedDate:                   bus.CreatedDate.String(),
		UpdatedBy:                     bus.UpdatedBy.String(),
		UpdatedDate:                   bus.UpdatedDate.String(),
	}
}

func ToAppOrderLineItems(bus []orderlineitemsbus.OrderLineItem) []OrderLineItem {
	appStatuses := make([]OrderLineItem, len(bus))
	for i, status := range bus {
		appStatuses[i] = ToAppOrderLineItem(status)
	}
	return appStatuses
}

type NewOrderLineItem struct {
	OrderID                       string `json:"order_id" validate:"required,uuid4"`
	ProductID                     string `json:"product_id" validate:"required,uuid4"`
	Quantity                      string `json:"quantity" validate:"required,numeric"`
	Discount                      string `json:"discount" validate:"omitempty,numeric"`
	LineItemFulfillmentStatusesID string `json:"line_item_fulfillment_statuses_id" validate:"required,uuid4"`
	CreatedBy                     string `json:"created_by" validate:"required,uuid4"`
}

func (app *NewOrderLineItem) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewOrderLineItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewOrderLineItem(app NewOrderLineItem) (orderlineitemsbus.NewOrderLineItem, error) {
	var dest orderlineitemsbus.NewOrderLineItem
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "toBusNewOrderLineItem: %s", err)
	}

	return dest, nil
}

type UpdateOrderLineItem struct {
	OrderID                       *string `json:"order_id" validate:"omitempty,uuid4"`
	ProductID                     *string `json:"product_id" validate:"omitempty,uuid4"`
	Quantity                      *string `json:"quantity" validate:"omitempty,numeric"`
	Discount                      *string `json:"discount" validate:"omitempty,numeric"`
	LineItemFulfillmentStatusesID *string `json:"line_item_fulfillment_statuses_id" validate:"omitempty,uuid4"`
	UpdatedBy                     *string `json:"updated_by" validate:"omitempty,uuid4"`
}

func (app *UpdateOrderLineItem) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateOrderLineItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateOrderLineItem(app UpdateOrderLineItem) (orderlineitemsbus.UpdateOrderLineItem, error) {
	var dest orderlineitemsbus.UpdateOrderLineItem
	err := convert.PopulateTypesFromStrings(app, &dest)
	if err != nil {
		return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "toBusUpdateOrderLineItem: %s", err)
	}

	return dest, nil
}
