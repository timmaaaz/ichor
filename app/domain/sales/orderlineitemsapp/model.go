package orderlineitemsapp

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/sales/orderlineitemsbus"
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
	OrderID                       string  `json:"order_id" validate:"required,uuid4"`
	ProductID                     string  `json:"product_id" validate:"required,uuid4"`
	Quantity                      string  `json:"quantity" validate:"required,numeric"`
	Discount                      string  `json:"discount" validate:"omitempty,numeric"`
	LineItemFulfillmentStatusesID string  `json:"line_item_fulfillment_statuses_id" validate:"required,uuid4"`
	CreatedBy                     string  `json:"created_by" validate:"required,uuid4"`
	CreatedDate                   *string `json:"created_date"` // Optional: for seeding/import
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
	orderID, err := uuid.Parse(app.OrderID)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse orderID: %s", err)
	}

	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
	}

	quantity, err := strconv.Atoi(app.Quantity)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
	}

	discount, err := strconv.ParseFloat(app.Discount, 64)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse discount: %s", err)
	}

	lineItemFulfillmentStatusesID, err := uuid.Parse(app.LineItemFulfillmentStatusesID)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse lineItemFulfillmentStatusesID: %s", err)
	}

	createdBy, err := uuid.Parse(app.CreatedBy)
	if err != nil {
		return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse createdBy: %s", err)
	}

	bus := orderlineitemsbus.NewOrderLineItem{
		OrderID:                       orderID,
		ProductID:                     productID,
		Quantity:                      quantity,
		Discount:                      discount,
		LineItemFulfillmentStatusesID: lineItemFulfillmentStatusesID,
		CreatedBy:                     createdBy,
		// CreatedDate: nil by default - API always uses server time
	}

	// Handle optional CreatedDate (for imports/admin tools only)
	if app.CreatedDate != nil && *app.CreatedDate != "" {
		createdDate, err := time.Parse(time.RFC3339, *app.CreatedDate)
		if err != nil {
			return orderlineitemsbus.NewOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse createdDate: %s", err)
		}
		bus.CreatedDate = &createdDate
	}

	return bus, nil
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
	var orderID *uuid.UUID
	if app.OrderID != nil {
		id, err := uuid.Parse(*app.OrderID)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse orderID: %s", err)
		}
		orderID = &id
	}

	var productID *uuid.UUID
	if app.ProductID != nil {
		id, err := uuid.Parse(*app.ProductID)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
		}
		productID = &id
	}

	var quantity *int
	if app.Quantity != nil {
		q, err := strconv.Atoi(*app.Quantity)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse quantity: %s", err)
		}
		quantity = &q
	}

	var discount *float64
	if app.Discount != nil {
		d, err := strconv.ParseFloat(*app.Discount, 64)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse discount: %s", err)
		}
		discount = &d
	}

	var lineItemFulfillmentStatusesID *uuid.UUID
	if app.LineItemFulfillmentStatusesID != nil {
		id, err := uuid.Parse(*app.LineItemFulfillmentStatusesID)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse lineItemFulfillmentStatusesID: %s", err)
		}
		lineItemFulfillmentStatusesID = &id
	}

	var updatedBy *uuid.UUID
	if app.UpdatedBy != nil {
		id, err := uuid.Parse(*app.UpdatedBy)
		if err != nil {
			return orderlineitemsbus.UpdateOrderLineItem{}, errs.Newf(errs.InvalidArgument, "parse updatedBy: %s", err)
		}
		updatedBy = &id
	}

	bus := orderlineitemsbus.UpdateOrderLineItem{
		OrderID:                       orderID,
		ProductID:                     productID,
		Quantity:                      quantity,
		Discount:                      discount,
		LineItemFulfillmentStatusesID: lineItemFulfillmentStatusesID,
		UpdatedBy:                     updatedBy,
	}
	return bus, nil
}
